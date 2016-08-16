package push

import (
	"gateway/model"
	re "gateway/model/remote_endpoint"

	"github.com/alexjlockwood/gcm"
)

type GooglePusher struct {
	connection *gcm.Sender
}

func NewGooglePusher(platform *re.PushPlatform) *GooglePusher {
	return &GooglePusher{
		connection: &gcm.Sender{
			ApiKey: platform.APIKey,
		},
	}
}

func (p *GooglePusher) Push(channel *model.PushChannel, device *model.PushDevice, data interface{}) error {
	message := gcm.NewMessage(data.(map[string]interface{}), device.Token)
	_, err := p.connection.Send(message, 3)
	return err
}
