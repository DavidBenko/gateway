package remote_endpoint

import (
	aperrors "gateway/errors"

	"github.com/vincent-petithory/dataurl"
)

// Docker represents a configuration for a remote Docker endpoint
type Docker struct {
	Endpoint string `json:"endpoint"`
	Image    string `json:"image"`
	Command  string `json:"command"`
	UseTLS   bool   `json:"use_tls"`
	TLS      struct {
		EncodedCA          string `json:"encoded_ca,omitempty"`
		EncodedCertificate string `json:"encoded_certificate,omitempty"`
		EncodedPrivateKey  string `json:"encoded_private_key,omitempty"`
		CA                 string `json:"ca,omitempty"`
		Certificate        string `json:"certificate,omitempty"`
		PrivateKey         string `json:"private_key,omitempty"`
	} `json:"tls,omitempty"`
}

func (d *Docker) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)

	if d.Endpoint == "" {
		errors.Add("endpoint", "must not be blank")
	}

	if d.Image == "" {
		errors.Add("image", "must not be blank")
	}

	if d.Command == "" {
		errors.Add("command", "must not be blank")
	}

	if !d.UseTLS {
		if d.TLS.EncodedCertificate != "" {
			cert, err := dataurl.DecodeString(d.TLS.EncodedCertificate)
			if err != nil {
				errors.Add("tls.encoded_certificate", "Must be data-uri encoded")
			} else {
				d.TLS.Certificate = string(cert.Data)
				d.TLS.EncodedCertificate = ""
			}
		}

		if d.TLS.EncodedCA != "" {
			ca, err := dataurl.DecodeString(d.TLS.EncodedCA)
			if err != nil {
				errors.Add("tls.encoded_ca", "Must be data-uri encoded")
			} else {
				d.TLS.CA = string(ca.Data)
				d.TLS.EncodedCA = ""
			}
		}

		if d.TLS.EncodedPrivateKey != "" {
			pk, err := dataurl.DecodeString(d.TLS.EncodedPrivateKey)
			if err != nil {
				errors.Add("tls.encoded_private_key", "Must be data-uri encoded")
			} else {
				d.TLS.PrivateKey = string(pk.Data)
				d.TLS.EncodedPrivateKey = ""
			}
		}
	}

	if !errors.Empty() {
		return errors
	}

	return nil
}
