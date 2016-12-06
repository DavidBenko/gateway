package request

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"gateway/model"

	b64 "encoding/base64"
	aperrors "gateway/errors"
	sql "gateway/sql"
)

const (
	CreateType   = "create"
	GenerateType = "generate"
	DeleteType   = "delete"
	RsaKey       = "rsa"
	EcdsaKey     = "ecdsa"
)

type genericKeyRequest struct {
	ReqType string `json:"_reqtype"`
}

type KeyCreateRequest struct {
	endpoint *model.RemoteEndpoint
	db       *sql.DB
	Name     string `json:"name"`
	Contents string `json:"contents"`
	Password string `json:"password"`
	Pkcs12   bool   `json:"pkcs12"`
}

type KeyGenerateRequest struct {
	endpoint       *model.RemoteEndpoint
	db             *sql.DB
	PrivateKeyName string `json:"privateKeyName"`
	PublicKeyName  string `json:"publicKeyName"`
	KeyType        string `json:"keytype"`
	Bits           int    `json:"bits"`
}

type KeyDeleteRequest struct {
	endpoint *model.RemoteEndpoint
	db       *sql.DB
	Name     string `json:"name"`
}

type KeyResponse struct {
	Data  map[string]interface{} `json:"data"`
	Error string                 `json:"error,omitempty"`
}

func (r *KeyResponse) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

func (r *KeyResponse) Log() string {
	return ""
}

func NewKeyRequest(db *sql.DB, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	generic := &genericKeyRequest{}
	if err := json.Unmarshal(*data, generic); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal request json: %v", err)
	}

	switch generic.ReqType {
	case CreateType:
		return newKeyCreateRequest(db, endpoint, data)
	case DeleteType:
		return newKeyDeleteRequest(db, endpoint, data)
	case GenerateType:
		return newKeyGenerateRequest(db, endpoint, data)
	default:
		return nil, fmt.Errorf("%s not a supported action", generic.ReqType)
	}
}

/*
 * ===============
 * Generate Request
 * ===============
 */
func newKeyGenerateRequest(db *sql.DB, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := &KeyGenerateRequest{}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("unable to unmarshal request json: %v", err)
	}
	request.endpoint = endpoint
	request.db = db
	return request, nil
}

func (r *KeyGenerateRequest) JSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *KeyGenerateRequest) Log(devMode bool) string {
	return fmt.Sprintf("generating %s keypair \"%s\"/\"%s\"", r.KeyType, r.PrivateKeyName, r.PublicKeyName)
}

func insertKeyPair(db *sql.DB, endpoint *model.RemoteEndpoint, private *model.Key, public *model.Key) (aperrors.Errors, error) {
	validationErrors := make(aperrors.Errors)
	// Private key validation
	if validationErrors := private.Validate(true); !validationErrors.Empty() {
		return validationErrors, nil
	}
	//Public key validation
	if validationErrors := public.Validate(true); !validationErrors.Empty() {
		return validationErrors, nil

	}
	err := db.DoInTransaction(func(tx *sql.Tx) error {
		err := private.Insert(endpoint.AccountID, endpoint.UserID, endpoint.APIID, tx)
		if err != nil {
			validationErrors = private.ValidateFromDatabaseError(err)
			return err

		}
		err = public.Insert(endpoint.AccountID, endpoint.UserID, endpoint.APIID, tx)
		if err != nil {
			validationErrors = public.ValidateFromDatabaseError(err)
			return err
		}
		return nil
	})

	return validationErrors, err
}

func (r *KeyGenerateRequest) Perform() Response {
	response := &KeyResponse{}
	response.Data = make(map[string]interface{})
	response.Data["success"] = false

	switch r.KeyType {
	case RsaKey:
		private, public, err := r.genRsa()
		if err != nil {
			response.Error = err.Error()
		}

		validationErrors, err := insertKeyPair(r.db, r.endpoint, private, public)
		if !validationErrors.Empty() {
			response.Error = validationErrors.String()
			return response
		}
		if err != nil {
			response.Error = err.Error()
			return response
		}

		return response
	case EcdsaKey:
		private, public, err := r.genEcdsa()
		if err != nil {
			response.Error = err.Error()
		}

		validationErrors, err := insertKeyPair(r.db, r.endpoint, private, public)
		if !validationErrors.Empty() {
			response.Error = validationErrors.String()
			return response
		}
		if err != nil {
			response.Error = err.Error()
			return response
		}

		return response
	default:
		response.Error = fmt.Sprintf("%s not supported", r.KeyType)
		return response
	}
}

func (r *KeyGenerateRequest) genRsa() (*model.Key, *model.Key, error) {
	rsa, err := rsa.GenerateKey(rand.Reader, r.Bits)
	if err != nil {
		return nil, nil, err
	}

	err = rsa.Validate()
	if err != nil {
		return nil, nil, err
	}

	block := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsa),
	}

	private := &model.Key{
		AccountID: r.endpoint.AccountID,
		Key:       pem.EncodeToMemory(&block),
		Name:      r.PrivateKeyName,
	}

	encodedPublic, err := x509.MarshalPKIXPublicKey(rsa.Public())
	if err != nil {
		return nil, nil, err
	}
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: encodedPublic,
	}

	public := &model.Key{
		AccountID: r.endpoint.AccountID,
		Key:       pem.EncodeToMemory(&publicBlock),
		Name:      r.PublicKeyName,
	}

	return private, public, nil
}

func (r *KeyGenerateRequest) genEcdsa() (*model.Key, *model.Key, error) {
	ecdsa, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	encodedPrivate, err := x509.MarshalECPrivateKey(ecdsa)
	if err != nil {
		return nil, nil, err
	}

	block := pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: encodedPrivate,
	}

	private := &model.Key{
		AccountID: r.endpoint.AccountID,
		Key:       pem.EncodeToMemory(&block),
		Name:      r.PrivateKeyName,
	}

	encodedPublic, err := x509.MarshalPKIXPublicKey(ecdsa.Public())
	if err != nil {
		return nil, nil, err
	}
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: encodedPublic,
	}

	public := &model.Key{
		AccountID: r.endpoint.AccountID,
		Key:       pem.EncodeToMemory(&publicBlock),
		Name:      r.PublicKeyName,
	}

	return private, public, nil
}

/*
 * ===============
 * Delete Request
 * ===============
 */
func newKeyDeleteRequest(db *sql.DB, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := &KeyDeleteRequest{}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("unable to unmarshal request json: %v", err)
	}
	request.endpoint = endpoint
	request.db = db
	return request, nil
}

func (r *KeyDeleteRequest) JSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *KeyDeleteRequest) Log(devMode bool) string {
	return fmt.Sprintf("deleting key \"%s\"", r.Name)
}

func (r *KeyDeleteRequest) Perform() Response {
	response := &KeyResponse{}
	response.Data = make(map[string]interface{})
	response.Data["success"] = false
	key := &model.Key{AccountID: r.endpoint.AccountID, Name: r.Name}
	if err := r.db.DoInTransaction(func(tx *sql.Tx) error {
		return key.DeleteByName(r.endpoint.AccountID, r.endpoint.UserID, r.endpoint.APIID, r.db, tx)
	}); err != nil {
		validationErrors := key.ValidateFromDatabaseError(err)
		response.Error = validationErrors.String()
		return response
	}
	response.Data["success"] = true
	return response
}

/*
 * ===============
 * Create Request
 * ===============
 */
func newKeyCreateRequest(db *sql.DB, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := &KeyCreateRequest{}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("unable to unmarshal request json: %v", err)
	}
	request.endpoint = endpoint
	request.db = db
	return request, nil
}

func (r *KeyCreateRequest) JSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *KeyCreateRequest) Log(devMode bool) string {
	return fmt.Sprintf("creating key \"%s\"", r.Name)
}

func (r *KeyCreateRequest) Perform() Response {
	response := &KeyResponse{}
	response.Data = make(map[string]interface{})
	response.Data["success"] = false
	if r.Pkcs12 == true {
		contents, err := b64.StdEncoding.DecodeString(r.Contents)
		if err != nil {
			response.Error = err.Error()
			return response
		}
		block, err := model.ParsePkcs12(contents, r.Password)
		if err != nil {
			response.Error = err.Error()
			return response
		}
		encoded := pem.EncodeToMemory(block)
		key := &model.Key{AccountID: r.endpoint.AccountID, Name: r.Name, Key: []byte(encoded), Password: r.Password}
		if err := insert(r.db, r.endpoint, key); err != nil {
			validationErrors := key.ValidateFromDatabaseError(err)
			response.Error = validationErrors.String()
			return response
		}
	} else {
		key := &model.Key{AccountID: r.endpoint.AccountID, Name: r.Name, Key: []byte(r.Contents), Password: r.Password}
		if validationErrors := key.Validate(true); !validationErrors.Empty() {
			response.Error = validationErrors.String()
			return response
		}

		if err := insert(r.db, r.endpoint, key); err != nil {
			validationErrors := key.ValidateFromDatabaseError(err)
			response.Error = validationErrors.String()
			return response
		}

		response.Data["success"] = true
	}
	return response
}

func insert(db *sql.DB, endpoint *model.RemoteEndpoint, key *model.Key) error {
	return db.DoInTransaction(func(tx *sql.Tx) error {
		return key.Insert(endpoint.AccountID, endpoint.UserID, endpoint.APIID, tx)
	})
}
