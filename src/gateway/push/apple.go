package push

import (
	"crypto/tls"
	"encoding/json"

	"gateway/logreport"
	"gateway/model"
	re "gateway/model/remote_endpoint"

	apns "github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/vincent-petithory/dataurl"
)

type ApplePusher struct {
	connection *apns.Client
	topic      string
}

func NewApplePusher(platform *re.PushPlatform) *ApplePusher {
	var cert tls.Certificate
	dataURL, err := dataurl.DecodeString(platform.Certificate)
	if err != nil {
		logreport.Fatal(err)
	}
	switch dataURL.MediaType.ContentType() {
	case re.PushCertificateTypePKCS12:
		cert, err = certificate.FromP12Bytes(dataURL.Data, platform.Password)
		if err != nil {
			logreport.Fatal(err)
		}
	case re.PushCertificateTypeX509:
		cert, err = certificate.FromPemBytes(dataURL.Data, platform.Password)
		if err != nil {
			logreport.Fatal(err)
		}
	default:
		logreport.Fatal("invalid apple certificate type")
	}
	client := apns.NewClient(cert)
	if platform.Development {
		client = client.Development()
	} else {
		client = client.Production()
	}
	return &ApplePusher{
		connection: client,
		topic:      platform.Topic,
	}
}

func (p *ApplePusher) Push(channel *model.PushChannel, device *model.PushDevice, data interface{}) error {
	notification := &apns.Notification{}
	notification.DeviceToken = device.Token
	notification.Topic = p.topic
	payload, err := json.Marshal(data)
	if err != nil {
		logreport.Fatal(err)
	}
	notification.Payload = payload
	_, err = p.connection.Push(notification)
	return err
}
