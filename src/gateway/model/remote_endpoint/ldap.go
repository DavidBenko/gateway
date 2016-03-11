package remote_endpoint

import (
	aperrors "gateway/errors"

	"github.com/vincent-petithory/dataurl"
)

const defaultPort = 389

type LDAP struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	UseTLS   bool   `json:"use_tls"`
	TLS      struct {
		EncodedCertificate string `json:"encoded_certificate,omitempty"`
		EncodedPrivateKey  string `json:"encoded_private_key,omitempty"`
		PrivateKeyPassword string `json:"private_key_password,omitempty"`
		ServerName         string `json:"server_name,omitempty"`
		PrivateKey         string `json:"private_key,omitempty"`
		Certificate        string `json:"certificate,omitempty"`
	} `json:"tls,omitempty"`
}

func (l *LDAP) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)

	if l.Host == "" {
		errors.Add("host", "must not be blank")
	}

	if l.Port == 0 {
		l.Port = defaultPort
	}

	if l.TLS.EncodedCertificate != "" {
		decoded, err := dataurl.DecodeString(l.TLS.EncodedCertificate)
		if err != nil {
			errors.Add("tls.encoded_certificate", "Must be data-uri encoded")
		} else {
			l.TLS.Certificate = string(decoded.Data)
			l.TLS.EncodedCertificate = ""
		}
	}

	if l.TLS.EncodedPrivateKey != "" {
		decoded, err := dataurl.DecodeString(l.TLS.EncodedPrivateKey)
		if err != nil {
			errors.Add("tls.encoded_private_key", "Must be data-uri encoded")
		} else {
			l.TLS.PrivateKey = string(decoded.Data)
			l.TLS.EncodedPrivateKey = ""
		}
	}

	if !errors.Empty() {
		return errors
	}

	return nil
}
