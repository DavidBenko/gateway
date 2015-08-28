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
	"gateway/model/remote_endpoint"
	"gateway/soap"
	apsql "gateway/sql"

	"github.com/vincent-petithory/dataurl"
)

const (
	// SoapRemoteEndpointStatusUninitialized is one of the possible statuses for the Status field on
	// the SoapRemoteEndpoint struct.  Uninitialized indicates that no processing has yet been attempted
	// on the SoapRemoteEndpoint
	SoapRemoteEndpointStatusUninitialized = "Uninitialized"
	// SoapRemoteEndpointStatusProcessing is one of the possible statuses for the Status field on
	// the SoapRemoteEndpoint struct.  Processing indicates that the WSDL file is actively being processed,
	// and is pending final outcome.
	SoapRemoteEndpointStatusProcessing = "Processing"
	// SoapRemoteEndpointStatusFailed is one of the possible statuses for the Status field on the
	// SoapRemoteEndpoint struct.  Failed indicates that there was a failure encountered processing the
	// WSDL.  The user ought to correct the problem with their WSDL and attempt to process it again.
	SoapRemoteEndpointStatusFailed = "Failed"
	// SoapRemoteEndpointStatusSuccess is one of the possible statuses for the Status field on the
	// SoapRemoteEndpoint struct.  Success indicates taht processing on the WSDL file has been completed
	// successfully, and the SOAP service is ready to be invoked.
	SoapRemoteEndpointStatusSuccess = "Success"
)

// SoapRemoteEndpoint contains attributes for a remote endpoint of type Soap
type SoapRemoteEndpoint struct {
	RemoteEndpointID int64 `json:"remote_endpoint_id,omitempty" db:"remote_endpoint_id"`

	ID                     int64 `json:"id,omitempty"`
	wsdl                   string
	generatedJar           []byte
	generatedJarThumbprint string
	Status                 string `json:"status"`
	Message                string `json:"message"`
}

type notificationListener struct {
	*apsql.DB
}

// Notify tells the listener a particular notification was fired
func (listener *notificationListener) Notify(n *apsql.Notification) {
	switch {
	case n.Table == "soap_remote_endpoints" && n.Event == apsql.Update:
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
	soap := SoapRemoteEndpoint{RemoteEndpointID: remoteEndpoint.ID, Status: SoapRemoteEndpointStatusUninitialized}

	data := remoteEndpoint.Data
	soapConfig, err := remote_endpoint.SoapConfig(data)
	if err != nil {
		return soap, fmt.Errorf("Encountered an error attempting to extract soap config: %v", err)
	}

	decoded, err := dataurl.DecodeString(soapConfig.WSDL)
	if err != nil {
		return soap, fmt.Errorf("Encountered an error attempting to decode WSDL: %v", err)
	}

	soap.wsdl = string(decoded.Data)

	return soap, nil
}

func (endpoint *SoapRemoteEndpoint) validate() error {
	switch endpoint.Status {
	case SoapRemoteEndpointStatusFailed, SoapRemoteEndpointStatusProcessing,
		SoapRemoteEndpointStatusSuccess, SoapRemoteEndpointStatusUninitialized:
	default:
		return fmt.Errorf("Invalid value for soap.Status: %s", endpoint.Status)
	}

	return nil
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
	query := `INSERT INTO soap_remote_endpoints(remote_endpoint_id, wsdl, status, message)
            VALUES (?, ?, ?, ?)`
	err := endpoint.validate()
	if err != nil {
		return fmt.Errorf("Can't insert SoapRemoteEndpoint due to validation error: %v", err)
	}

	endpoint.ID, err = tx.InsertOne(query, endpoint.RemoteEndpointID, endpoint.wsdl, endpoint.Status, endpoint.Message)
	if err != nil {
		return fmt.Errorf("Unable to insert SoapRemoteEndpoint record: %v", err)
	}

	err = endpoint.afterSave(tx)
	if err != nil {
		return fmt.Errorf("Error occured in afterSave hook for SoapRemoteEndpoint while trying to insert: %v", err)
	}

	return tx.Notify("soap_remote_endpoints", endpoint.ID, apsql.Insert)
}

// Update updates an existing SoapRemoteEndpoint record in the database
func (endpoint *SoapRemoteEndpoint) Update(tx *apsql.Tx) error {
	return endpoint.update(tx, true)
}

func (endpoint *SoapRemoteEndpoint) update(tx *apsql.Tx, fireAfterSave bool) error {
	query := `UPDATE soap_remote_endpoints
            SET wsdl = ?, generated_jar = ?, generated_jar_thumbprint = ?, status = ?, message = ?
            WHERE id = ?`
	err := endpoint.validate()
	if err != nil {
		return fmt.Errorf("Can't insert SoapRemoteEndpoint due to validation error: %v", err)
	}

	err = tx.UpdateOne(query, endpoint.wsdl, endpoint.generatedJar, endpoint.generatedJarThumbprint, endpoint.Status, endpoint.Message, endpoint.ID)
	if err != nil {
		return fmt.Errorf("Unable to update SoapRemoteEndpoint: %v", err)
	}

	if fireAfterSave {
		err = endpoint.afterSave(tx)
		if err != nil {
			return fmt.Errorf("Error occured in afterSave hook for SoapRemoteEndpoint while trying to update: %v", err)
		}
	}

	return nil
}

func (endpoint *SoapRemoteEndpoint) afterSave(tx *apsql.Tx) error {
	go func() {
		log.Printf("%s Starting wsdl ingestion", "[debug]")

		// TODO - wrap this inside a panic/recover block
		dir, err := ensureJarPath()

		// write wsdl file to directory
		filePerm := os.FileMode(os.ModeDir | 0600)
		filename := path.Join(dir, fmt.Sprintf("%d.wsdl", endpoint.ID))
		err = ioutil.WriteFile(filename, []byte(endpoint.wsdl), filePerm)
		if err != nil {
			// TODO
		}

		// invoke wsimport
		outputfile := path.Join(dir, fmt.Sprintf("%d.jar", endpoint.ID))
		err = soap.Wsimport(filename, outputfile)
		if err != nil {
			// TODO
			log.Printf("%s Tried to invoke wsimport: %v", "[debug]", err)
		}

		// update the DB with the generated jars bytes!
		bytes, err := ioutil.ReadFile(outputfile)
		if err != nil {
			// TODO
			log.Printf("%s Couldn't read bytes! %v", "[debug]", err)
		}
		endpoint.generatedJar = bytes
		checksum := md5.Sum(bytes)
		endpoint.generatedJarThumbprint = hex.EncodeToString(checksum[:])

		// TODO
		tx, err := tx.DB.Begin()
		if err != nil {
			// TODO
			log.Printf("%s Unable to begin tx! %v", "[debug]", err)
		}
		err = endpoint.update(tx, false)
		if err != nil {
			log.Printf("%s Unable to execute update! %v", "[debug]", err)
			// TODO
			rbErr := tx.Rollback()
			if rbErr != nil {
				// TODO
				log.Printf("%s Unable to rollback tx! %v", "[debug]", rbErr)
			}
			// TODO
		}

		err = tx.Notify("soap_remote_endpoints", endpoint.ID, apsql.Update)
		if err != nil {
			// TODO
			log.Printf("%s Unable to notify of update: %v", "[debug]", err)
			// TODO actually handle notification on listener side ...
		}

		// report success or failure
		// TODO
		endpoint.Status = SoapRemoteEndpointStatusSuccess
		err = endpoint.update(tx, false)
		if err != nil {
			// TODO
			log.Printf("%s Unable to update status for SoapRemoteEndpoint: %v", "[debug]", err)
		}

		// TODO
		err = tx.Commit()
		if err != nil {
			log.Printf("%s Unable to commit tx! %v", "[debug]", err)
		}

		// clean up wsdl
		err = os.Remove(filename)
		if err != nil {
			// TODO
		}

		log.Printf("%s Finished wsdl ingestion", "[debug]")
	}()
	return nil
}
