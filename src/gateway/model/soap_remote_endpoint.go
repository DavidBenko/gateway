package model

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gateway/config"
	aperrors "gateway/errors"
	"gateway/logreport"
	"gateway/soap"
	apsql "gateway/sql"
)

const (
	soapRemoteEndpoints = "soap_remote_endpoints"
	filePrefix          = "file:///"
)

// SoapRemoteEndpoint contains attributes for a remote endpoint of type Soap
type SoapRemoteEndpoint struct {
	RemoteEndpointID int64 `json:"-" db:"remote_endpoint_id"`

	ID                    int64  `json:"-" db:"id"`
	Wsdl                  string `json:"-"`
	wsdlContent           []byte
	WsdlContentThumbprint apsql.NullString `json:"-" db:"wsdl_content_thumbprint"`

	RemoteEndpoint *RemoteEndpoint `json:"-"`
}

// Copy creates a shallow copy of a SoapRemoteEndpoint and returns a reference to the copy
func (e *SoapRemoteEndpoint) Copy() *SoapRemoteEndpoint {
	return &SoapRemoteEndpoint{
		RemoteEndpointID:      e.RemoteEndpointID,
		ID:                    e.ID,
		Wsdl:                  e.Wsdl,
		wsdlContent:           e.wsdlContent,
		WsdlContentThumbprint: e.WsdlContentThumbprint,
		RemoteEndpoint:        e.RemoteEndpoint,
	}
}

type soapNotificationListener struct {
	*apsql.DB
}

// Notify tells the listener a particular notification was fired
func (l *soapNotificationListener) Notify(n *apsql.Notification) {
	if n.Table != soapRemoteEndpoints {
		return
	}

	switch n.Event {
	case apsql.Update, apsql.Insert:
		err := CacheWsdlFile(l.DB, n.APIID)
		if err != nil {
			logreport.Printf("%s Error caching WSDL for api %d: %v", config.System, n.APIID, err)
		}
	case apsql.Delete:
		var remoteEndpointID int64
		var ok bool
		if remoteEndpointID, ok = n.Messages[0].(int64); !ok {
			tmp, ok := n.Messages[0].(float64)
			remoteEndpointID = int64(tmp)

			if !ok {
				logreport.Printf("%s Error deleting WSDL for api %d: %v", config.System, n.APIID, "deletion message did not come in expected format")
				return
			}
		}

		err := DeleteWsdlFile(remoteEndpointID)

		if err != nil && !os.IsNotExist(err) {
			logreport.Printf("%s Error deleting WSDL for api %d: %v", config.System, n.APIID, err)
		}
	}
}

// Reconnect tells the listener that we may have been disconnected, but
// have reconnected. They should update all state that could have changed.
func (l *soapNotificationListener) Reconnect() {
	// Nothing to do here
}

func writeToWsdlFile(bytes []byte, filename string) error {
	return ioutil.WriteFile(filename, bytes, os.ModeDir|0600)
}

func DeleteWsdlFile(soapRemoteEndpointID int64) error {
	logreport.Printf("Received a request to delete WSDL file for soapRemoteEndpointID %d", soapRemoteEndpointID)
	wsdlDir, err := soap.EnsureWsdlPath()
	if err != nil {
		return err
	}

	wsdlFileName := path.Join(wsdlDir, fmt.Sprintf("%d.wsdl", soapRemoteEndpointID))
	return os.Remove(wsdlFileName)
}

// CacheAllJarFiles iterates through all SoapRemoteEndpoints, and copies the
// generatedJar to the file system in the appropriate file location if the file
// is missing.
func CacheAllWsdlFiles(db *apsql.DB) error {
	endpoints, err := allSoapRemoteEndpoints(db)
	if err != nil {
		return err
	}
	for _, endpoint := range endpoints {
		exists, err := endpoint.WsdlExists()
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if err := CacheWsdlFile(db, endpoint.ID); err != nil {
			return err
		}
	}
	return nil
}

// JarExists checks for the existence of the JAR file corresponding to the
// given soapRemoteEndpointID on the file system.
func (s *SoapRemoteEndpoint) WsdlExists() (bool, error) {
	wsdlURL, err := soap.WsdlURLForSoapRemoteEndpointID(s.ID)
	if err != nil {
		return false, err
	}
	if strings.HasPrefix(wsdlURL, filePrefix) {
		wsdlURL = filepath.FromSlash(wsdlURL[len(filePrefix):])
	}
	if _, err := os.Stat(wsdlURL); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// Caches the WSDL file for the given soapRemoteEndpointID on the file system.
func CacheWsdlFile(db *apsql.DB, soapRemoteEndpointID int64) error {
	logreport.Printf("Caching WSDL file for soap_remote_endpoint with ID %v", soapRemoteEndpointID)
	wsdlDir, err := soap.EnsureWsdlPath()
	if err != nil {
		return err
	}

	// check if jar exists
	wsdlFileName := path.Join(wsdlDir, fmt.Sprintf("%d.wsdl", soapRemoteEndpointID))
	fileBytes, err := ioutil.ReadFile(wsdlFileName)

	var hexsum string
	switch {
	case err != nil && os.IsNotExist(err):
		// copy to file system
		hexsum = ""
	case err != nil:
		return fmt.Errorf("Unable to open WSDL file: %v", err)
	default:
		// jar exists! get its MD5 hash and compare against record from DB.
		checksum := md5.Sum(fileBytes)
		hexsum = hex.EncodeToString(checksum[:])
	}

	fileBytes, err = GetGeneratedJarBytes(db, soapRemoteEndpointID, hexsum)
	if err != nil {
		return fmt.Errorf("Unable to get bytes from database: %v", err)
	}

	if fileBytes != nil {
		// copy to file system
		err := writeToWsdlFile(fileBytes, wsdlFileName)
		if err != nil {
			return fmt.Errorf("Unable to write WSDL %s to file system: %v", wsdlFileName, err)
		}
	}

	return nil
}

// StartSoapRemoteEndpointUpdateListener registers a listener for updates from
func StartSoapRemoteEndpointUpdateListener(db *apsql.DB) {
	listener := &soapNotificationListener{db}
	db.RegisterListener(listener)
}

// NewSoapRemoteEndpoint creates a new SoapRemoteEndpoint struct
func NewSoapRemoteEndpoint(remoteEndpoint *RemoteEndpoint) *SoapRemoteEndpoint {
	return &SoapRemoteEndpoint{RemoteEndpointID: remoteEndpoint.ID, RemoteEndpoint: remoteEndpoint}
}

// GetGeneratedJarBytes returns the bytes of the generated_jar file for the soap_remote_endpoint with the specified ID,
// if and only if the checksum does not match the generated_jar_thumbprint stored in the DB.  If the thumbprints match,
// then they are the same file, and so no bytes will be returned.
func GetGeneratedJarBytes(db *apsql.DB, soapRemoteEndpointID int64, checksum string) ([]byte, error) {
	query := `SELECT wsdl_content FROM soap_remote_endpoints WHERE id = ? AND wsdl_content_thumbprint != ?`

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
func (e *SoapRemoteEndpoint) Insert(tx *apsql.Tx) error {
	query := `INSERT INTO soap_remote_endpoints(remote_endpoint_id, wsdl)
            VALUES (?, ?)`

	var err error
	id, err := tx.InsertOne(query, e.RemoteEndpointID, e.Wsdl)
	if err != nil {
		return fmt.Errorf("Unable to insert SoapRemoteEndpoint record %v: %v", e.RemoteEndpointID, err)
	}
	e.ID = id

	tag := tx.TopTag()
	tx.AddPostCommitHook(func(t *apsql.Tx) {
		// copy to ensure there's no chance of concurrency issues
		go afterSave(t, e.Copy(), tag)
	})

	return tx.Notify(
		soapRemoteEndpoints,
		e.RemoteEndpoint.AccountID,
		e.RemoteEndpoint.UserID,
		e.RemoteEndpoint.APIID,
		0,
		e.ID,
		apsql.Insert,
	)
}

func allSoapRemoteEndpoints(db *apsql.DB) ([]*SoapRemoteEndpoint, error) {
	query := `SELECT id, remote_endpoint_id, wsdl_content_thumbprint
						FROM soap_remote_endpoints`

	soapRemoteEndpoints := []*SoapRemoteEndpoint{}
	err := db.Select(&soapRemoteEndpoints, query)
	return soapRemoteEndpoints, err
}

// FindSoapRemoteEndpointByRemoteEndpointID finds a SoapRemoteEndpoint given a remoteEndpoint
func FindSoapRemoteEndpointByRemoteEndpointID(db *apsql.DB, remoteEndpointID int64) (*SoapRemoteEndpoint, error) {
	query := `SELECT id, remote_endpoint_id, wsdl_content_thumbprint
            FROM soap_remote_endpoints
            WHERE remote_endpoint_id = ?`

	soapRemoteEndpoint := &SoapRemoteEndpoint{}
	if err := db.Get(soapRemoteEndpoint, query, remoteEndpointID); err != nil {
		return nil, fmt.Errorf("Unable to fetch SoapRemoteEndpoint from the database for remote_endpoint_id %d: %v", remoteEndpointID, err)
	}

	return soapRemoteEndpoint, nil
}

// Update updates an existing SoapRemoteEndpoint record in the database
func (e *SoapRemoteEndpoint) Update(tx *apsql.Tx) error {
	return e.update(tx, true, false)
}

func (e *SoapRemoteEndpoint) update(tx *apsql.Tx, fireAfterSave, updateGeneratedJar bool) error {
	var query string
	if updateGeneratedJar {
		query = `UPDATE soap_remote_endpoints
              SET wsdl = ?, wsdl_content = ?, wsdl_content_thumbprint = ?
              WHERE id = ?`
	} else {
		query = `UPDATE soap_remote_endpoints
              SET wsdl = ?
              WHERE id = ?`
	}

	var err error
	if updateGeneratedJar {
		err = tx.UpdateOne(query, e.Wsdl, e.wsdlContent, e.WsdlContentThumbprint, e.ID)
	} else {
		err = tx.UpdateOne(query, e.Wsdl, e.ID)
	}

	if err != nil {
		return fmt.Errorf("Unable to update SoapRemoteEndpoint: %v", err)
	}

	if fireAfterSave {
		tag := tx.TopTag()
		tx.AddPostCommitHook(func(t *apsql.Tx) {
			// copy to ensure there's no chance of concurrency issues
			go afterSave(t, e.Copy(), tag)
		})
	}

	return tx.Notify(soapRemoteEndpoints, e.RemoteEndpoint.AccountID, e.RemoteEndpoint.UserID, e.RemoteEndpoint.APIID, 0, e.ID, apsql.Update)
}

func afterSave(origTx *apsql.Tx, e *SoapRemoteEndpoint, tag string) {
	db := origTx.DB
	processWsdl(db, e, tag)
}

func processWsdl(db *apsql.DB, e *SoapRemoteEndpoint, tag string) {
	e.RemoteEndpoint.Status = apsql.MakeNullString(RemoteEndpointStatusProcessing)
	e.RemoteEndpoint.StatusMessage = apsql.MakeNullStringNull()

	// In a new transaction, update the status to processing before we do anything
	err := db.DoInTransaction(func(tx *apsql.Tx) error {
		tx.PushTag(tag)
		defer tx.PopTag()

		err := e.RemoteEndpoint.update(tx, false)
		if err != nil {
			return aperrors.NewWrapped("soap_remote_endpoint.go: updating remote endpoint", err)
		}
		err = e.update(tx, false, false)
		if err != nil {
			return aperrors.NewWrapped("soap_remote_endpoint.go: updating soap remote endpoint", err)
		}
		return nil
	})

	if err != nil {
		logreport.Printf("%s Unable to update status to %v.  Processing will continue anyway.", config.System, RemoteEndpointStatusProcessing)
	}

	err = db.DoInTransaction(func(tx *apsql.Tx) error {
		tx.PushTag(tag)
		defer tx.PopTag()

		return ingestWsdl(tx, e)
	})

	if err != nil {
		logreport.Printf("%s WSDL could not be processed due to error encountered:  %v", config.System, err)
		e.RemoteEndpoint.Status = apsql.MakeNullString(RemoteEndpointStatusFailed)
		e.RemoteEndpoint.StatusMessage = apsql.MakeNullString(err.Error())
	} else {
		e.RemoteEndpoint.Status = apsql.MakeNullString(RemoteEndpointStatusSuccess)
		e.RemoteEndpoint.StatusMessage = apsql.MakeNullStringNull()
	}

	err = db.DoInTransaction(func(tx *apsql.Tx) error {
		tx.PushTag(tag)
		defer tx.PopTag()

		err := e.RemoteEndpoint.update(tx, false)
		if err != nil {
			return aperrors.NewWrapped("soap_remote_endpoint.go: Updating final state on remote endpoint", err)
		}
		err = e.update(tx, false, false)
		if err != nil {
			return aperrors.NewWrapped("soap_remote_endpoint.go: Updating final state on soap remote endpoint", err)
		}
		return nil
	})

	if err != nil {
		logreport.Printf("%s Unable to update status of RemoteSoapEndpoint due to error: %v", config.System, err)
	}
}

func ingestWsdl(tx *apsql.Tx, e *SoapRemoteEndpoint) error {
	logreport.Printf("%s Starting wsdl ingestion", "[debug]")

	dir, err := soap.EnsureWsdlPath()
	if err != nil {
		return err
	}

	// write wsdl file to directory
	filePerm := os.FileMode(os.ModeDir | 0600)
	filename := path.Join(dir, fmt.Sprintf("%d.wsdl", e.ID))

	err = ioutil.WriteFile(filename, []byte(e.Wsdl), filePerm)
	if err != nil {
		return err
	}

	// update the DB with the wsdl
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		logreport.Printf("%s Couldn't read bytes! %v", "[debug]", err)
		return err
	}
	e.wsdlContent = bytes
	checksum := md5.Sum(bytes)
	e.WsdlContentThumbprint = apsql.MakeNullString(hex.EncodeToString(checksum[:]))

	err = e.update(tx, false, true)
	if err != nil {
		return err
	}

	err = tx.Notify(soapRemoteEndpoints, e.RemoteEndpoint.AccountID, e.RemoteEndpoint.UserID, e.RemoteEndpoint.APIID, 0, e.ID, apsql.Update)
	if err != nil {
		logreport.Printf("%s Unable to notify of update: %v", "[debug]", err)
		return err
	}

	// clean up wsdl
	err = os.Remove(filename)
	if err != nil {
		return err
	}

	logreport.Printf("%s Finished wsdl ingestion", "[debug]")

	return nil
}
