package push

import (
	"gateway/model"
	re "gateway/model/remote_endpoint"

	gcm "github.com/google/go-gcm"
)

type GooglePusher struct {
	apiKey string
}

func NewGooglePusher(platform *re.PushPlatform) *GooglePusher {
	return &GooglePusher{
		apiKey: platform.APIKey,
	}
}

func (p *GooglePusher) Push(channel *model.PushChannel, device *model.PushDevice, data interface{}) error {
	dat := gcm.Data(data.(map[string]interface{}))
	message := gcm.HttpMessage{
		To:   device.Token,
		Data: dat,
	}
	if n, ok := dat["notification"].(map[string]interface{}); ok {
		notification := gcm.Notification{}
		if title, ok := n["title"].(string); ok {
			notification.Title = title
		}
		if body, ok := n["body"].(string); ok {
			notification.Body = body
		}
		if icon, ok := n["icon"].(string); ok {
			notification.Icon = icon
		}
		if sound, ok := n["sound"].(string); ok {
			notification.Sound = sound
		}
		if badge, ok := n["badge"].(string); ok {
			notification.Badge = badge
		}
		if tag, ok := n["tag"].(string); ok {
			notification.Tag = tag
		}
		if color, ok := n["color"].(string); ok {
			notification.Color = color
		}
		if clickAction, ok := n["click_action"].(string); ok {
			notification.ClickAction = clickAction
		}
		if bodyLocKey, ok := n["body_loc_key"].(string); ok {
			notification.BodyLocKey = bodyLocKey
		}
		if bodyLocArgs, ok := n["body_loc_args"].(string); ok {
			notification.BodyLocArgs = bodyLocArgs
		}
		if titleLocArgs, ok := n["title_loc_args"].(string); ok {
			notification.TitleLocArgs = titleLocArgs
		}
		if titleLocKey, ok := n["title_loc_key"].(string); ok {
			notification.TitleLocKey = titleLocKey
		}
		message.Notification = &notification
		delete(dat, "notification")
	}
	_, err := gcm.SendHttp(p.apiKey, message)
	return err
}
