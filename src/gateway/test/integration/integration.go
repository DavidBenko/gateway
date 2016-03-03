package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/vincent-petithory/dataurl"
)

var IntegrationTest string
var IsIntegrationTest = false
var ApiImportDirectory string

const createAPIJSON = `{
  "api": {
    "name": "LDAP Test API",
    "description": "An API for testing LDAP remote endpoints",
    "cors_allow_origin": "*",
    "cors_allow_headers": "content-type, accept",
    "cors_allow_credentials": true,
    "cors_request_headers": "*",
    "cors_max_age": 600,
    "export": "@EXPORT@"
  }
}`

func init() {
	if value, err := strconv.ParseBool(IntegrationTest); err == nil {
		IsIntegrationTest = value
	}
}

func ImportAPI(apiName string, h *HttpHelper) (string, error) {
	var (
		status int
		body   string
		err    error
	)

	filename := filepath.Join(ApiImportDirectory, fmt.Sprintf("%s.json", apiName))
	importBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("Unable to read API file %s: %v", filename, err)
	}

	encoded := dataurl.EncodeBytes(importBytes)

	createAPI := strings.Replace(createAPIJSON, "@EXPORT@", string(encoded), -1)

	status, _, body, err = h.Post("http://127.0.0.1:5000/admin/apis", createAPI)
	if err != nil {
		return "", fmt.Errorf("Failed due to : %v", err)
	}

	if status != 200 {
		return "", fmt.Errorf("Expected a 200 status code from create API call and instead got %d", status)
	}

	var api struct {
		API struct {
			BaseURL string `json:"base_url"`
		} `json:"api"`
	}

	err = json.Unmarshal([]byte(body), &api)
	if err != nil {
		return "", fmt.Errorf("Unexpected error attempting to unmarshal JSON %v", err)
	}
	host := api.API.BaseURL

	return host, nil
}

// TestMain serves as the main testing entry point for the integration package
/*func TestMain(m *testing.M) {
	fmt.Println("\n\n\n***isIntegrationTest ", isIntegrationTest)
	if !isIntegrationTest {
		log.Println("Integration flag not set.  Skipping integration tests.")
		os.Exit(0)
		//return
	}

	os.Exit(m.Run())
}*/
