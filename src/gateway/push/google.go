package push

import (
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

func (p *GooglePusher) Push(token string, data interface{}) error {
	message := gcm.NewMessage(data.(map[string]interface{}), token)
	_, err := p.connection.Send(message, 3)
	return err
}
