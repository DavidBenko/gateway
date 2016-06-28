package push

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"gateway/config"
	"gateway/logreport"
	"gateway/model"
	re "gateway/model/remote_endpoint"
	"gateway/queue"
	"gateway/queue/mangos"
	apsql "gateway/sql"

	"github.com/AnyPresence/surgemq/auth"
	"github.com/AnyPresence/surgemq/service"
	"github.com/surgemq/message"
)

type MQTTPusher struct {
	push chan<- []byte
}

type MQTT struct {
	DB     *apsql.DB
	Server *service.Server
	Broker *mangos.Broker
}

type Context struct {
	RemoteEndpointID     int64
	PushPlatformCodename string
}

func (c *Context) String() string {
	return fmt.Sprintf("%v,%v", c.RemoteEndpointID, c.PushPlatformCodename)
}

func NewMQTTPusher(pool *PushPool, platform *re.PushPlatform) *MQTTPusher {
	push, _ := pool.push.Channels()
	return &MQTTPusher{
		push: push,
	}
}

func (p *MQTTPusher) Push(channel *model.PushChannel, device *model.PushDevice, data interface{}) error {
	message := &PushMessage{
		Channel: channel,
		Device:  device,
		Data:    data,
	}
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	p.push <- payload
	return nil
}

func SetupMQTT(db *apsql.DB, conf config.Push) *MQTT {
	server := &service.Server{
		KeepAlive:        300,
		ConnectTimeout:   2,
		SessionsProvider: "mem",
		Authenticator:    "gateway",
		TopicsProvider:   "mem",
	}

	mqtt := &MQTT{
		DB:     db,
		Server: server,
	}

	auth.Register("gateway", mqtt)

	go func() {
		if err := server.ListenAndServe("tcp://:1883"); err != nil {
			logreport.Fatal(err)
		}
	}()

	if conf.EnableBroker {
		var err error
		mqtt.Broker, err = mangos.NewBroker(mangos.XPubXSub, mangos.TCP, conf.XPub(), conf.XSub())
		if err != nil {
			log.Fatal(err)
		}
	}

	go func() {
		receive, err := queue.Subscribe(
			conf.XPub(),
			mangos.Sub,
			mangos.SubTCP,
		)
		if err != nil {
			logreport.Fatal(err)
		}
		messages, errs := receive.Channels()
		defer func() {
			receive.Close()
		}()
		go func() {
			for err := range errs {
				logreport.Printf("[mqtt] %v", err)
			}
		}()
		for msg := range messages {
			push := &PushMessage{}
			err := json.Unmarshal(msg, push)
			if err != nil {
				log.Fatal(err)
			}
			context := &Context{
				RemoteEndpointID:     push.Channel.RemoteEndpointID,
				PushPlatformCodename: push.Device.Type,
			}
			pubmsg := message.NewPublishMessage()
			pubmsg.SetTopic([]byte(fmt.Sprintf("/%s", push.Channel.Name)))
			pubmsg.SetQoS(0)
			payload, err := json.Marshal(push.Data)
			if err != nil {
				log.Fatal(err)
			}
			pubmsg.SetPayload(payload)
			err = server.Publish(context, pubmsg, nil)
			if err != nil {
				logreport.Printf("[mqtt] %v", err)
			}
		}
	}()

	return mqtt
}

func (m *MQTT) Authenticate(id string, cred interface{}) (fmt.Stringer, error) {
	username := strings.Split(id, ",")
	if len(username) != 4 && len(username) != 5 {
		return nil, errors.New("user name should have format: '<emai>,<api name>,<remote endpoint codename>,<push platform codename>[,<environment name>]'")
	}
	environmentName := ""
	if len(username) == 5 {
		environmentName = username[4]
	}

	user, err := model.FindUserByEmail(m.DB, username[0])
	if err != nil {
		return nil, err
	}

	endpoints, err := model.FindRemoteEndpointForCodenameAndAPINameAndAccountID(m.DB, username[2], username[1], user.AccountID)
	if err != nil {
		return nil, err
	}
	if len(endpoints) != 1 {
		return nil, errors.New("invalid credentials")
	}

	found, password, endpoint, codename := false, "", endpoints[0], username[3]
	if environmentName != "" {
		for _, environment := range endpoint.EnvironmentData {
			if environment.Name == environmentName {
				push := &re.Push{}
				err = json.Unmarshal(environment.Data, push)
				if err != nil {
					return nil, err
				}
				for _, platform := range push.PushPlatforms {
					if platform.Codename == codename {
						found = true
						password = platform.Password
						break
					}
				}
				break
			}
		}
	}
	if !found {
		push := &re.Push{}
		err = json.Unmarshal(endpoint.Data, push)
		if err != nil {
			return nil, err
		}
		for _, platform := range push.PushPlatforms {
			if platform.Codename == codename {
				found = true
				password = platform.Password
				break
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("%v is not a valid platform", codename)
	}

	if password != "" {
		if cred.(string) != password {
			return nil, errors.New("invalid credentials")
		}
	}

	context := &Context{
		RemoteEndpointID:     endpoint.ID,
		PushPlatformCodename: codename,
	}

	return context, nil
}
