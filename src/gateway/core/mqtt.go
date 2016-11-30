package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"

	"gateway/config"
	"gateway/core/vm"
	"gateway/http"
	"gateway/logreport"
	"gateway/model"
	"gateway/push"

	"github.com/nanoscaleio/surgemq/message"
	"github.com/nanoscaleio/surgemq/service"
	stripe "github.com/stripe/stripe-go"
)

var blacklistedResponseHeaders = []string{
	// Proxy request's responses may be encoded;
	// we don't want to pass that through
	"Content-Encoding",
}

type mqttRequest struct {
	Method   string `json:"method"`
	Host     string `json:"host"`
	URI      string `json:"uri"`
	Path     string `json:"path"`
	RawQuery string `json:"rawQuery"`
	Body     string `json:"body"`

	RemoteAddress string `json:"remoteAddress"`
	ContentLength int64  `json:"contentLength"`

	Headers map[string]interface{} `json:"headers"`
	Form    map[string]interface{} `json:"form"`
	Query   map[string]interface{} `json:"query"`
	Vars    map[string]string      `json:"vars"`
	Params  map[string]interface{} `json:"params"`

	ID string `json:"id"`
}

type ProxyResponse struct {
	StatusCode int                    `json:"statusCode"`
	Body       string                 `json:"body"`
	Headers    map[string]interface{} `json:"headers"`
}

func (c *Core) ExecuteMQTT(context fmt.Stringer, logPrint logreport.Logf, msg *message.PublishMessage, remote net.Addr, onpub service.OnPublishFunc) error {
	ctx, db := context.(*push.Context), c.OwnDb
	channel := &model.ProxyEndpointChannel{
		AccountID:        ctx.RemoteEndpoint.AccountID,
		APIID:            ctx.RemoteEndpoint.APIID,
		RemoteEndpointID: ctx.RemoteEndpoint.ID,
		Name:             strings.TrimLeft(string(msg.Topic()), "/"),
	}
	var err error
	channel, err = channel.FindByName(db)
	if err != nil {
		return err
	}

	logPrefix := config.Proxy
	uuid, err := http.NewUUID()
	if err != nil {
		logreport.Printf("%s Could not generate request UUID", logPrefix)
		uuid = "x"
	}
	logPrefix = fmt.Sprintf("%s [act %d] [api %d] [end %d] [req %s]", logPrefix,
		channel.AccountID, channel.APIID, channel.ProxyEndpointID, uuid)

	logreport.Printf("%s [access] %s", logPrefix, string(msg.Topic()))

	endpoint, err := model.FindProxyEndpointForProxy(db, channel.ProxyEndpointID, model.ProxyEndpointTypeHTTP)
	if err != nil {
		return err
	}
	libraries, err := model.AllLibrariesForProxy(db, channel.APIID)
	if err != nil {
		return err
	}

	codeTimeout := c.Conf.Proxy.CodeTimeout
	if stripe.Key != "" {
		plan, err := model.FindPlanByAccountID(db, channel.AccountID)
		if err != nil {
			return err
		}
		if plan.JavascriptTimeout < codeTimeout {
			codeTimeout = plan.JavascriptTimeout
		}
	}

	logPrint("%s [route] %s", logPrefix, endpoint.Name)

	request := mqttRequest{
		Method:        "mqtt",
		Host:          c.Conf.Proxy.Domain,
		URI:           c.Conf.Push.MQTTURI,
		Path:          string(msg.Topic()),
		RemoteAddress: remote.String(),
		ID:            uuid,
	}
	err = json.Unmarshal(msg.Payload(), &request)
	if err != nil {
		return err
	}

	if schema := endpoint.Schema; schema != nil && schema.RequestSchema != "" {
		err := c.ProcessSchema(endpoint.Schema.RequestSchema, string(msg.Payload()))
		if err != nil {
			if err.Error() == "EOF" {
				return errors.New("a json document is required in the request")
			}
			return err
		}
	}

	vm := &vm.CoreVM{}
	vm.InitCoreVM(VMCopy(channel.AccountID, c.KeyStore), logPrint, logPrefix, &c.Conf.Proxy, endpoint, libraries, codeTimeout)

	incomingJSON, err := json.Marshal(&request)
	if err != nil {
		return err
	}
	vm.Set("__ap_proxyRequestJSON", string(incomingJSON))
	scripts := []interface{}{
		"var request = JSON.parse(__ap_proxyRequestJSON);",
		"var response = new AP.HTTP.Response();",
	}

	if _, err := vm.RunAll(scripts); err != nil {
		return err
	}

	if err = c.RunComponents(vm, endpoint.Components); err != nil {
		if err.Error() == "JavaScript took too long to execute" {
			logPrint("%s [timeout] JavaScript execution exceeded %ds timeout threshold", logPrefix, codeTimeout)
		}
		return err
	}

	responseObject, err := vm.Run("response;")
	if err != nil {
		return err
	}
	responseJSON, err := c.ObjectJSON(vm, responseObject)
	if err != nil {
		return err
	}
	response, err := ProxyResponseFromJSON(responseJSON)
	if err != nil {
		return err
	}

	if schema := endpoint.Schema; schema != nil &&
		(schema.ResponseSchema != "" ||
			(schema.ResponseSameAsRequest && schema.RequestSchema != "")) {
		responseSchema := schema.ResponseSchema
		if schema.ResponseSameAsRequest {
			responseSchema = schema.RequestSchema
		}
		err := c.ProcessSchema(responseSchema, response.Body)
		if err != nil {
			if err.Error() == "EOF" {
				return errors.New("a json document is required in the response")
			}
			return err
		}
	}

	response.Headers["Content-Length"] = len(response.Body)

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}

	responseMessage := message.NewPublishMessage()
	responseMessage.SetTopic(msg.Topic())
	responseMessage.SetPayload(responseBytes)

	return onpub(responseMessage)
}

func ProxyResponseFromJSON(responseJSON string) (*ProxyResponse, error) {
	response := ProxyResponse{}
	err := json.Unmarshal([]byte(responseJSON), &response)
	if err == nil {
		if response.Headers == nil {
			response.Headers = make(map[string]interface{})
		}
	}
	for _, header := range blacklistedResponseHeaders {
		delete(response.Headers, header)
	}
	return &response, err
}
