package push

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gateway/logreport"
	"gateway/model"
	re "gateway/model/remote_endpoint"
	apsql "gateway/sql"

	"github.com/AnyPresence/surgemq/auth"
	"github.com/AnyPresence/surgemq/service"
)

type MQTT struct {
	DB     *apsql.DB
	Server *service.Server
}

type Context struct {
	RemoteEndpointID     int64
	PushPlatformCodename string
	EnvironmentName      string
}

func (c *Context) String() string {
	return fmt.Sprintf("%v,%v,%v", c.RemoteEndpointID, c.PushPlatformCodename, c.EnvironmentName)
}

func SetupMQTT(db *apsql.DB) *MQTT {
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
		EnvironmentName:      environmentName,
	}

	return context, nil
}
