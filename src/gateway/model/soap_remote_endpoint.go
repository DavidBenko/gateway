package model

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"gateway/model/remote_endpoint"
	"gateway/soap"

	"github.com/vincent-petithory/dataurl"
)
import apsql "gateway/sql"

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
		log.Println("Starting wsdl ingestion")

		// TODO - wrap this inside a panic/recover block

		dirPerm := os.FileMode(os.ModeDir | 0700)
		filePerm := os.FileMode(os.ModeDir | 0600)

		// ensure directory exists
		dir := path.Clean(path.Join(".", "tmp", "jaxws"))
		err := os.MkdirAll(dir, dirPerm)
		if err != nil {
			// TODO
		}

		// write wsdl file to directory
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
			log.Printf("Tried to invoke wsimport: %v", err)
		}

		// update the DB with the generated jars bytes!
		bytes, err := ioutil.ReadFile(outputfile)
		if err != nil {
			// TODO
			log.Printf("Couldn't read bytes! %v", err)
		}
		endpoint.generatedJar = bytes
		checksum := md5.Sum(bytes)
		endpoint.generatedJarThumbprint = hex.EncodeToString(checksum[:])

		// TODO
		tx, err := tx.DB.Begin()
		if err != nil {
			// TODO
			log.Printf("Unable to begin tx! %v", err)
		}
		err = endpoint.update(tx, false)
		if err != nil {
			log.Printf("Unable to execute update! %v", err)
			// TODO
			rbErr := tx.Rollback()
			if rbErr != nil {
				// TODO
				log.Printf("Unable to rollback tx! %v", rbErr)
			}
			// TODO
		}

		err = tx.Notify("soap_remote_endpoints", endpoint.ID, apsql.Update)
		if err != nil {
			// TODO
			log.Printf("Unable to notify of update: %v", err)
			// TODO actually handle notification on listener side ...
		}

		// report success or failure
		// TODO
		endpoint.Status = SoapRemoteEndpointStatusSuccess
		err = endpoint.update(tx, false)
		if err != nil {
			// TODO
			log.Printf("Unable to update status for SoapRemoteEndpoint: %v", err)
		}

		// TODO
		err = tx.Commit()
		if err != nil {
			log.Printf("Unable to commit tx! %v", err)
		}

		// clean up wsdl
		err = os.Remove(filename)
		if err != nil {
			// TODO
		}

		log.Println("Finished wsdl ingestion")
	}()
	return nil
}
