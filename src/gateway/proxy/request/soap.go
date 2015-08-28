package request

import (
	"encoding/json"
	"fmt"
	"gateway/model"
	"log"
)

type SoapRequest struct {
	WSDL string `json:"wsdl"`
}

type SoapResponse struct {
}

func (soapResponse *SoapResponse) JSON() ([]byte, error) {
	return []byte{}, nil
}

func (soapResponse *SoapResponse) Log() string {
	log.Println("TODO SoapResponse.Log")
	return ""
}

func (soapRequest *SoapRequest) Perform() Response {
	log.Printf("WSDL is soapRequest.wsdl %v", soapRequest.WSDL)

	// TODO - placeholder for SOAP

	return &SoapResponse{}
}

func (soapRequest *SoapRequest) Log(devMode bool) string {
	log.Println("TODO SoapRequest.Log")
	return ""
}

func NewSoapRequest(endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := new(SoapRequest)

	if err := json.Unmarshal(endpoint.Data, request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal request json: %v", err)
	}

	// TODO  : make this look like other NewXXXRequest methods so that environment data is used in addition to what comes in from endpoint.Data
	return request, nil
}
