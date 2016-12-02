package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"gateway/config"
	apvm "gateway/core/vm"
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
	Error      string                 `json:"error,omitempty"`
	Body       string                 `json:"body"`
	Headers    map[string]interface{} `json:"headers"`
}

func (c *Core) ExecuteMQTT(context fmt.Stringer, logPrint logreport.Logf, msg *message.PublishMessage, remote net.Addr, onpub service.OnPublishFunc) (err error) {
	start := time.Now()
	var (
		vm            *apvm.CoreVM
		httpErr       http.Error
		responseBytes []byte
	)
	defer func() {
		if httpErr != nil {
			err = httpErr.Error()
			response := &ProxyResponse{
				StatusCode: httpErr.Code(),
				Error:      httpErr.String(),
				Headers: map[string]interface{}{
					"Content-Length": 0,
				},
			}
			var errr error
			responseBytes, errr = json.Marshal(response)
			if errr != nil {
				err = errr
			}
		}
		if responseBytes != nil {
			responseMessage := message.NewPublishMessage()
			responseMessage.SetTopic(msg.Topic())
			responseMessage.SetQoS(msg.QoS())
			responseMessage.SetPayload(responseBytes)
			err = onpub(responseMessage)
		}
	}()
	httpError := func(err error) http.Error {
		if !c.DevMode {
			return http.DefaultServerError()
		}

		return http.NewServerError(err)
	}
	httpJavascriptError := func(err error, env *model.Environment) http.Error {
		if env == nil {
			return httpError(err)
		}

		if env.ShowJavascriptErrors {
			return http.NewServerError(err)
		}

		return http.DefaultServerError()
	}

	ctx, db := context.(*push.Context), c.OwnDb
	channel := &model.ProxyEndpointChannel{
		AccountID:        ctx.RemoteEndpoint.AccountID,
		APIID:            ctx.RemoteEndpoint.APIID,
		RemoteEndpointID: ctx.RemoteEndpoint.ID,
		Name:             strings.TrimLeft(string(msg.Topic()), "/"),
	}
	channel, err = channel.FindByName(db)
	if err != nil {
		httpErr = httpError(err)
		return
	}

	logPrefix := config.Proxy
	uuid, err := http.NewUUID()
	if err != nil {
		logreport.Printf("%s Could not generate request UUID", logPrefix)
		uuid = "x"
	}
	logPrefix = fmt.Sprintf("%s [act %d] [api %d] [end %d] [req %s]", logPrefix,
		channel.AccountID, channel.APIID, channel.ProxyEndpointID, uuid)
	defer func() {
		if httpErr != nil {
			errString := "Unknown Error"
			lines := strings.Split(httpErr.String(), "\n")
			if len(lines) > 0 {
				errString = lines[0]
			}
			logPrint("%s [error] %s", logPrefix, errString)
		}
		var proxiedRequestsDuration time.Duration
		if vm != nil {
			proxiedRequestsDuration = vm.ProxiedRequestsDuration
		}

		total := time.Since(start)
		processing := total - proxiedRequestsDuration
		logPrint("%s [time] %v (processing %v, requests %v)",
			logPrefix, total, processing, proxiedRequestsDuration)
	}()

	logPrint("%s [access] %s", logPrefix, string(msg.Topic()))

	endpoint, err := model.FindProxyEndpointForProxy(db, channel.ProxyEndpointID, model.ProxyEndpointTypeHTTP)
	if err != nil {
		httpErr = httpError(err)
		return
	}
	libraries, err := model.AllLibrariesForProxy(db, channel.APIID)
	if err != nil {
		httpErr = httpError(err)
		return
	}

	codeTimeout := c.Conf.Proxy.CodeTimeout
	if stripe.Key != "" {
		plan, err := model.FindPlanByAccountID(db, channel.AccountID)
		if err != nil {
			httpErr = httpError(err)
			return err
		}
		if plan.JavascriptTimeout < codeTimeout {
			codeTimeout = plan.JavascriptTimeout
		}
	}

	logPrint("%s [route] %s", logPrefix, endpoint.Name)

	request := mqttRequest{
		Method:        model.ProxyEndpointTestMethodGet,
		Host:          c.Conf.Proxy.Domain,
		URI:           c.Conf.Push.MQTTURI,
		Path:          string(msg.Topic()),
		RemoteAddress: remote.String(),
		ID:            uuid,
	}
	err = json.Unmarshal(msg.Payload(), &request)
	if err != nil {
		httpErr = httpError(err)
		return
	}
	if request.ContentLength == 0 {
		request.ContentLength = int64(len(request.Body))
	}

	if schema := endpoint.Schema; schema != nil && schema.RequestSchema != "" {
		err = c.ProcessSchema(endpoint.Schema.RequestSchema, string(msg.Payload()))
		if err != nil {
			if err.Error() == "EOF" {
				httpErr = http.NewError(errors.New("a json document is required in the request"), 422)
				return
			}
			httpErr = http.NewError(err, 400)
			return
		}
	}

	vm = &apvm.CoreVM{}
	vm.InitCoreVM(VMCopy(channel.AccountID, c.KeyStore), logPrint, logPrefix, &c.Conf.Proxy, endpoint, libraries, codeTimeout)

	incomingJSON, err := json.Marshal(&request)
	if err != nil {
		httpErr = httpError(err)
		return
	}
	vm.Set("__ap_proxyRequestJSON", string(incomingJSON))
	scripts := []interface{}{
		"var request = JSON.parse(__ap_proxyRequestJSON);",
		"var response = new AP.HTTP.Response();",
	}

	if _, err = vm.RunAll(scripts); err != nil {
		httpErr = httpError(err)
		return
	}

	if err = c.RunComponents(vm, endpoint.Components); err != nil {
		if err.Error() == "JavaScript took too long to execute" {
			logPrint("%s [timeout] JavaScript execution exceeded %ds timeout threshold", logPrefix, codeTimeout)
		}
		httpErr = httpJavascriptError(err, endpoint.Environment)
		return
	}

	responseObject, err := vm.Run("response;")
	if err != nil {
		httpErr = httpError(err)
		return
	}
	responseJSON, err := c.ObjectJSON(vm, responseObject)
	if err != nil {
		httpErr = httpError(err)
		return
	}
	response, err := ProxyResponseFromJSON(responseJSON)
	if err != nil {
		httpErr = httpError(err)
		return
	}

	if schema := endpoint.Schema; schema != nil &&
		(schema.ResponseSchema != "" ||
			(schema.ResponseSameAsRequest && schema.RequestSchema != "")) {
		responseSchema := schema.ResponseSchema
		if schema.ResponseSameAsRequest {
			responseSchema = schema.RequestSchema
		}
		err = c.ProcessSchema(responseSchema, response.Body)
		if err != nil {
			if err.Error() == "EOF" {
				httpErr = http.NewError(errors.New("a json document is required in the response"), 500)
				return
			}
			httpErr = http.NewError(err, 500)
			return
		}
	}

	response.Headers["Content-Length"] = len(response.Body)

	responseBytes, err = json.Marshal(response)
	if err != nil {
		httpErr = httpError(err)
		return
	}

	return nil
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
