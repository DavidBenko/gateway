package remote_endpoint

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"

	aperrors "gateway/errors"

	"github.com/sideshow/apns2/certificate"
	"github.com/vincent-petithory/dataurl"
)

const (
	PushTypeOSX  = "osx"
	PushTypeIOS  = "ios"
	PushTypeGCM  = "gcm"
	PushTypeMQTT = "mqtt"

	PushCertificateTypePKCS12 = "application/x-pkcs12"
	PushCertificateTypeX509   = "application/x-x509-ca-cert"
)

type Push struct {
	PublishEndpoint bool           `json:"publish_endpoint"`
	PushPlatforms   []PushPlatform `json:"push_platforms"`
}

type PushPlatform struct {
	Name           string `json:"name"`
	Codename       string `json:"codename"`
	Type           string `json:"type"`
	Certificate    string `json:"certificate"`
	Password       string `json:"password"`
	Topic          string `json:"topic"`
	Development    bool   `json:"development"`
	APIKey         string `json:"api_key"`
	ConnectTimeout int    `json:"connect_timeout"`
	AckTimeout     int    `json:"ack_timeout"`
	TimeoutRetries int    `json:"timeout_retries"`
}

func (p *Push) UpdateWith(parent *Push) {
	p.PublishEndpoint = p.PublishEndpoint || parent.PublishEndpoint
	length := len(p.PushPlatforms)
	for i := range parent.PushPlatforms {
		found := false
		for j := 0; j < length; j++ {
			if parent.PushPlatforms[i].Codename == p.PushPlatforms[j].Codename {
				found = true
				break
			}
		}
		if !found {
			p.PushPlatforms = append(p.PushPlatforms, parent.PushPlatforms[i])
		}
	}
}

// https://developer.apple.com/certificationauthority/AppleWWDRCA.cer
const AppleCertificate = `
-----BEGIN CERTIFICATE-----
MIIEIjCCAwqgAwIBAgIIAd68xDltoBAwDQYJKoZIhvcNAQEFBQAwYjELMAkGA1UE
BhMCVVMxEzARBgNVBAoTCkFwcGxlIEluYy4xJjAkBgNVBAsTHUFwcGxlIENlcnRp
ZmljYXRpb24gQXV0aG9yaXR5MRYwFAYDVQQDEw1BcHBsZSBSb290IENBMB4XDTEz
MDIwNzIxNDg0N1oXDTIzMDIwNzIxNDg0N1owgZYxCzAJBgNVBAYTAlVTMRMwEQYD
VQQKDApBcHBsZSBJbmMuMSwwKgYDVQQLDCNBcHBsZSBXb3JsZHdpZGUgRGV2ZWxv
cGVyIFJlbGF0aW9uczFEMEIGA1UEAww7QXBwbGUgV29ybGR3aWRlIERldmVsb3Bl
ciBSZWxhdGlvbnMgQ2VydGlmaWNhdGlvbiBBdXRob3JpdHkwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQDKOFSmy1aqyCQ5SOmM7uxfuH8mkbw0U3rOfGOA
YXdkXqUHI7Y5/lAtFVZYcC1+xG7BSoU+L/DehBqhV8mvexj/avoVEkkVCBmsqtsq
Mu2WY2hSFT2Miuy/axiV4AOsAX2XBWfODoWVN2rtCbauZ81RZJ/GXNG8V25nNYB2
NqSHgW44j9grFU57Jdhav06DwY3Sk9UacbVgnJ0zTlX5ElgMhrgWDcHld0WNUEi6
Ky3klIXh6MSdxmilsKP8Z35wugJZS3dCkTm59c3hTO/AO0iMpuUhXf1qarunFjVg
0uat80YpyejDi+l5wGphZxWy8P3laLxiX27Pmd3vG2P+kmWrAgMBAAGjgaYwgaMw
HQYDVR0OBBYEFIgnFwmpthhgi+zruvZHWcVSVKO3MA8GA1UdEwEB/wQFMAMBAf8w
HwYDVR0jBBgwFoAUK9BpR5R2Cf70a40uQKb3R01/CF4wLgYDVR0fBCcwJTAjoCGg
H4YdaHR0cDovL2NybC5hcHBsZS5jb20vcm9vdC5jcmwwDgYDVR0PAQH/BAQDAgGG
MBAGCiqGSIb3Y2QGAgEEAgUAMA0GCSqGSIb3DQEBBQUAA4IBAQBPz+9Zviz1smwv
j+4ThzLoBTWobot9yWkMudkXvHcs1Gfi/ZptOllc34MBvbKuKmFysa/Nw0Uwj6OD
Dc4dR7Txk4qjdJukw5hyhzs+r0ULklS5MruQGFNrCk4QttkdUGwhgAqJTleMa1s8
Pab93vcNIx0LSiaHP7qRkkykGRIZbVf1eliHe2iK5IaMSuviSRSqpd1VAKmuu0sw
ruGgsbwpgOYJd+W+NKIByn/c4grmO7i77LpilfMFY0GCzQ87HUyVpNur+cmV6U/k
TecmmYHpvPm0KdIBembhLoz2IYrF+Hjhga6/05Cdqa3zr/04GpZnMBxRpVzscYqC
tGwPDBUf
-----END CERTIFICATE-----`

func validateCertificate(cert tls.Certificate, errors aperrors.Errors) {
	block, _ := pem.Decode([]byte(AppleCertificate))
	appleCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		errors.Add("certificate", fmt.Sprintf("error loading apple certificate: %v", err))
	}
	err = cert.Leaf.CheckSignatureFrom(appleCert)
	if err != nil {
		errors.Add("certificate", fmt.Sprintf("error checking apple signature: %v", err))
	}
}

func validateKey(key string, errors aperrors.Errors) {
	client := &http.Client{}
	body := bytes.NewReader([]byte(`{"registration_ids":["ABC"]}`))
	req, err := http.NewRequest("POST", "https://gcm-http.googleapis.com/gcm/send", body)
	if err != nil {
		errors.Add("api_key", fmt.Sprintf("validate key: %v", err))
	}
	req.Header.Add("Authorization", fmt.Sprintf("key=%v", key))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		errors.Add("api_key", fmt.Sprintf("validate key: %v", err))
	}
	if resp.StatusCode != http.StatusOK {
		errors.Add("api_key", "invalid key")
	}
}

func (p *Push) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)
	names, codenames := make(map[string]bool), make(map[string]bool)
	for i := range p.PushPlatforms {
		name := p.PushPlatforms[i].Name
		if name == "" {
			errors.Add("name", "must not be blank")
		}
		if _, has := names[name]; has {
			errors.Add("name", "must be unique")
		}
		names[name] = true
		codename := p.PushPlatforms[i].Codename
		if codename == "" {
			errors.Add("codename", "must not be blank")
		}
		if _, has := codenames[codename]; has {
			errors.Add("codename", "must be unique")
		}
		codenames[codename] = true
		switch p.PushPlatforms[i].Type {
		case PushTypeOSX:
			fallthrough
		case PushTypeIOS:
			dataURL, err := dataurl.DecodeString(p.PushPlatforms[i].Certificate)
			if err != nil {
				errors.Add("certificate", fmt.Sprintf("invalid data url: %v", err))
				break
			}
			switch dataURL.MediaType.ContentType() {
			case PushCertificateTypePKCS12:
				cert, err := certificate.FromP12Bytes(dataURL.Data, p.PushPlatforms[i].Password)
				if err != nil {
					errors.Add("certificate", fmt.Sprintf("invalid certificate: %v", err))
				}
				validateCertificate(cert, errors)
			case PushCertificateTypeX509:
				cert, err := certificate.FromPemBytes(dataURL.Data, p.PushPlatforms[i].Password)
				if err != nil {
					errors.Add("certificate", fmt.Sprintf("invalid certificate: %v", err))
				}
				validateCertificate(cert, errors)
			default:
				errors.Add("certificate", "must be pkcs12 or pem format")
			}
		case PushTypeGCM:
			key := p.PushPlatforms[i].APIKey
			if key == "" {
				errors.Add("api_key", "must not be blank")
			}
			validateKey(key, errors)
		case PushTypeMQTT:
			if p.PushPlatforms[i].ConnectTimeout == 0 {
				errors.Add("connect_timeout", "must not be zero")
			}
			if p.PushPlatforms[i].AckTimeout == 0 {
				errors.Add("ack_timeout", "must not be zero")
			}
			if p.PushPlatforms[i].TimeoutRetries == 0 {
				errors.Add("timeout_retries", "must not be zero")
			}
		default:
			errors.Add("type", "must be a valid type")
		}
	}
	return errors
}
