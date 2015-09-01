package model

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"gateway/config"
	aperrors "gateway/errors"
	"gateway/model/remote_endpoint"
	"gateway/soap"
	apsql "gateway/sql"

	"github.com/vincent-petithory/dataurl"
)

const (
	// RemoteEndpointStatusPending is one of the possible statuses for the Status field on
	// the SoapRemoteEndpoint struct.  Pending indicates that no processing has yet been attempted
	// on the SoapRemoteEndpoint
	RemoteEndpointStatusPending = "Pending"
	// RemoteEndpointStatusProcessing is one of the possible statuses for the Status field on
	// the SoapRemoteEndpoint struct.  Processing indicates that the WSDL file is actively being processed,
	// and is pending final outcome.
	RemoteEndpointStatusProcessing = "Processing"
	// RemoteEndpointStatusFailed is one of the possible statuses for the Status field on the
	// SoapRemoteEndpoint struct.  Failed indicates that there was a failure encountered processing the
	// WSDL.  The user ought to correct the problem with their WSDL and attempt to process it again.
	RemoteEndpointStatusFailed = "Failed"
	// RemoteEndpointStatusSuccess is one of the possible statuses for the Status field on the
	// SoapRemoteEndpoint struct.  Success indicates taht processing on the WSDL file has been completed
	// successfully, and the SOAP service is ready to be invoked.
	RemoteEndpointStatusSuccess = "Success"
)

// SoapRemoteEndpoint contains attributes for a remote endpoint of type Soap
type SoapRemoteEndpoint struct {
	RemoteEndpointID int64 `json:"-" db:"remote_endpoint_id"`

	ID                     int64  `json:"-" db:"id"`
	Wsdl                   string `json:"-"`
	generatedJar           []byte
	GeneratedJarThumbprint string `json:"-" db:"generated_jar_thumbprint"`

	RemoteEndpoint *RemoteEndpoint `json:"-"`
}

type notificationListener struct {
	*apsql.DB
}

// Notify tells the listener a particular notification was fired
func (listener *notificationListener) Notify(n *apsql.Notification) {
	switch {
	case n.Table == "soap_remote_endpoints" && (n.Event == apsql.Update || n.Event == apsql.Insert):
		err := cacheJarFile(listener.DB, n.APIID)
		if err != nil {
			log.Printf("%s Error caching jarfile for api %d: %v", config.System, n.APIID, err)
		}
	case n.Table == "soap_remote_endpoints" && n.Event == apsql.Delete:
		// TODO
		log.Printf("Received delete notification for %v", n)
	}
}

// Reconnect tells the listener that we may have been disconnected, but
// have reconnected. They should update all state that could have changed.
func (listener *notificationListener) Reconnect() {
	// TODO anything to do here?
}

func writeToJarFile(bytes []byte, filename string) error {
	return ioutil.WriteFile(filename, bytes, os.ModeDir|0600)
}

func cacheJarFile(db *apsql.DB, soapRemoteEndpointID int64) error {
	jarDir, err := ensureJarPath()
	if err != nil {
		return err
	}

	// check if jar exists
	jarFileName := path.Join(jarDir, fmt.Sprintf("%d.jar", soapRemoteEndpointID))
	fileBytes, err := ioutil.ReadFile(jarFileName)

	var hexsum string
	if err != nil && os.IsNotExist(err) {
		// copy to file system
		hexsum = ""
	} else if err != nil {
		return fmt.Errorf("Unable to open jar file: %v", err)
	} else {
		// jar exists! get its MD5 hash and compare against record from DB.
		checksum := md5.Sum(fileBytes)
		hexsum = hex.EncodeToString(checksum[:])
	}

	fileBytes, err = GetSoapRemoteEndpointGeneratedJarBytesIfStale(db, soapRemoteEndpointID, hexsum)
	if err != nil {
		return fmt.Errorf("Unable to get bytes from database: %v", err)
	}

	if fileBytes != nil {
		// copy to file system
		err := writeToJarFile(fileBytes, jarFileName)
		if err != nil {
			return fmt.Errorf("Unable to write jar %s to file system: %v", jarFileName, err)
		}
	}

	return nil
}

func ensureJarPath() (string, error) {
	dirPerm := os.FileMode(os.ModeDir | 0700)

	dir := path.Clean(path.Join(".", "tmp", "jaxws"))
	err := os.MkdirAll(dir, dirPerm)
	return dir, err
}

// StartSoapRemoteEndpointUpdateListener registers a listener for updates from
func StartSoapRemoteEndpointUpdateListener(db *apsql.DB) {
	listener := &notificationListener{db}
	db.RegisterListener(listener)
}

// NewSoapRemoteEndpoint creates a new SoapRemoteEndpoint struct
func NewSoapRemoteEndpoint(remoteEndpoint *RemoteEndpoint) (SoapRemoteEndpoint, error) {
	soap := SoapRemoteEndpoint{RemoteEndpointID: remoteEndpoint.ID, RemoteEndpoint: remoteEndpoint}

	data := remoteEndpoint.Data
	soapConfig, err := remote_endpoint.SoapConfig(data)
	if err != nil {
		return soap, fmt.Errorf("Encountered an error attempting to extract soap config: %v", err)
	}

	if soapConfig.WSDL != "" {
		decoded, err := dataurl.DecodeString(soapConfig.WSDL)
		if err != nil {
			return soap, fmt.Errorf("Encountered an error attempting to decode WSDL: %v", err)
		}

		soap.Wsdl = string(decoded.Data)
	}

	return soap, nil
}

// GetSoapRemoteEndpointGeneratedJarBytesIfStale returns the bytes of the generated_jar file for the soap_remote_endpoint with the specified ID,
// if and only if the checksum does not matched the generated_jar_thumbprint stored in the DB
func GetSoapRemoteEndpointGeneratedJarBytesIfStale(db *apsql.DB, soapRemoteEndpointID int64, checksum string) ([]byte, error) {
	query := `SELECT generated_jar FROM soap_remote_endpoints WHERE id = ? AND generated_jar_thumbprint != ?`

	var dest []byte
	err := db.Get(&dest, query, soapRemoteEndpointID, checksum)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("Unable to get generated jar bytes from soap_remote_entpoints for ID %v: %v", soapRemoteEndpointID, err)
	}

	return dest, nil
}

// Insert inserts a new SoapRemoteEndpoint record into the database
func (endpoint *SoapRemoteEndpoint) Insert(tx *apsql.Tx) error {
	query := `INSERT INTO soap_remote_endpoints(remote_endpoint_id, wsdl)
            VALUES (?, ?)`

	var err error
	endpoint.ID, err = tx.InsertOne(query, endpoint.RemoteEndpointID, endpoint.Wsdl)
	if err != nil {
		return fmt.Errorf("Unable to insert SoapRemoteEndpoint record: %v", err)
	}

	endpoint.afterSave(tx)

	return tx.Notify("soap_remote_endpoints", endpoint.ID, apsql.Insert)
}

// FindSoapRemoteEndpointByRemoteEndpointID finds a SoapRemoteEndpoint given a remoteEndpoint
func FindSoapRemoteEndpointByRemoteEndpointID(db *apsql.DB, remoteEndpointID int64) (*SoapRemoteEndpoint, error) {
	query := `SELECT id, remote_endpoint_id, generated_jar_thumbprint
            FROM soap_remote_endpoints
            WHERE remote_endpoint_id = ?`

	soapRemoteEndpoint := &SoapRemoteEndpoint{}
	err := db.Get(soapRemoteEndpoint, query, remoteEndpointID)
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch SoapRemoteEndpoint from the database for remote_endpoint_id %d: %v", remoteEndpointID, err)
	}

	return soapRemoteEndpoint, nil
}

// Update updates an existing SoapRemoteEndpoint record in the database
func (endpoint *SoapRemoteEndpoint) Update(tx *apsql.Tx) error {
	return endpoint.update(tx, true, false)
}

func (endpoint *SoapRemoteEndpoint) update(tx *apsql.Tx, fireAfterSave, updateGeneratedJar bool) error {
	var query string
	if updateGeneratedJar {
		query = `UPDATE soap_remote_endpoints
              SET wsdl = ?, generated_jar = ?, generated_jar_thumbprint = ?
              WHERE id = ?`
	} else {
		query = `UPDATE soap_remote_endpoints
              SET wsdl = ?
              WHERE id = ?`
	}

	var err error
	if updateGeneratedJar {
		err = tx.UpdateOne(query, endpoint.Wsdl, endpoint.generatedJar, endpoint.GeneratedJarThumbprint, endpoint.ID)
	} else {
		err = tx.UpdateOne(query, endpoint.Wsdl, endpoint.ID)
	}

	if err != nil {
		return fmt.Errorf("Unable to update SoapRemoteEndpoint: %v", err)
	}

	if fireAfterSave {
		endpoint.afterSave(tx)
	}

	return tx.Notify("soap_remote_endpoints", endpoint.ID, apsql.Update)
}

func (endpoint *SoapRemoteEndpoint) afterSave(origTx *apsql.Tx) {
	db := origTx.DB
	go func() {
		endpoint.RemoteEndpoint.Status = apsql.MakeNullString(RemoteEndpointStatusProcessing)
		endpoint.RemoteEndpoint.StatusMessage = apsql.MakeNullStringNull()

		// In a new transaction, update the status to processing before we do anything
		err := db.DoInTransaction(func(tx *apsql.Tx) error {
			err := endpoint.RemoteEndpoint.Update(tx)
			if err != nil {
				return aperrors.NewWrapped("soap_remote_endpoint.go: updating remote endpoint", err)
			}
			err = endpoint.update(tx, false, false)
			if err != nil {
				return aperrors.NewWrapped("soap_remote_endpoint.go: updating soap remote endpoint", err)
			}
			return nil
		})

		if err != nil {
			log.Printf("%s Unable to update status to %v.  Processing will continue anyway.", config.System, RemoteEndpointStatusProcessing)
		}

		err = db.DoInTransaction(func(tx *apsql.Tx) error {
			return endpoint.ingestWsdl(tx)
		})

		if err != nil {
			log.Printf("%s WSDL could not be processed due to error encountered:  %v", config.System, err)
			endpoint.RemoteEndpoint.Status = apsql.MakeNullString(RemoteEndpointStatusFailed)
			endpoint.RemoteEndpoint.StatusMessage = apsql.MakeNullString(err.Error())
		} else {
			endpoint.RemoteEndpoint.Status = apsql.MakeNullString(RemoteEndpointStatusSuccess)
			endpoint.RemoteEndpoint.StatusMessage = apsql.MakeNullStringNull()
		}

		err = db.DoInTransaction(func(tx *apsql.Tx) error {
			err := endpoint.RemoteEndpoint.Update(tx)
			if err != nil {
				return aperrors.NewWrapped("soap_remote_endpoint.go: Updating final state on remote endpoint", err)
			}
			err = endpoint.update(tx, false, false)
			if err != nil {
				return aperrors.NewWrapped("soap_remote_endpoint.go: Updating final state on soap remote endpoint", err)
			}
			return nil
		})

		if err != nil {
			log.Printf("%s Unable to update status of RemoteSoapEndpoint due to error: %v", config.System, err)
		}
	}()
}

func (endpoint *SoapRemoteEndpoint) ingestWsdl(tx *apsql.Tx) error {
	log.Printf("%s Starting wsdl ingestion", "[debug]")

	dir, err := ensureJarPath()
	if err != nil {
		return err
	}

	// write wsdl file to directory
	filePerm := os.FileMode(os.ModeDir | 0600)
	filename := path.Join(dir, fmt.Sprintf("%d.wsdl", endpoint.ID))

	err = ioutil.WriteFile(filename, []byte(endpoint.Wsdl), filePerm)
	if err != nil {
		return err
	}

	// invoke wsimport
	outputfile := path.Join(dir, fmt.Sprintf("%d.jar", endpoint.ID))
	err = soap.Wsimport(filename, outputfile)
	if err != nil {
		log.Printf("%s Tried to invoke wsimport: %v", "[debug]", err)
		return err
	}

	// update the DB with the generated jars bytes!
	bytes, err := ioutil.ReadFile(outputfile)
	if err != nil {
		log.Printf("%s Couldn't read bytes! %v", "[debug]", err)
		return err
	}
	endpoint.generatedJar = bytes
	checksum := md5.Sum(bytes)
	endpoint.GeneratedJarThumbprint = hex.EncodeToString(checksum[:])

	err = endpoint.update(tx, false, true)
	if err != nil {
		return err
	}

	err = tx.Notify("soap_remote_endpoints", endpoint.ID, apsql.Update)
	if err != nil {
		log.Printf("%s Unable to notify of update: %v", "[debug]", err)
		return err
	}

	// clean up wsdl
	err = os.Remove(filename)
	if err != nil {
		return err
	}

	log.Printf("%s Finished wsdl ingestion", "[debug]")

	return nil
}
