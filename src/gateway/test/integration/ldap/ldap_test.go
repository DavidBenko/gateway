package ldap

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gateway/proxy/request/ldap"
	"gateway/test/integration"
	"log"
	"net/url"
	"os"
	"os/exec"
	"sync"
	"testing"
)

var (
	ldapSetupFile string
	once          sync.Once
	host          string

	h = integration.NewHTTPHelper()
)

var searchTests = []struct {
	description               string
	url                       string
	expectedLDAPStatusCode    int
	expectedResultCount       int
	expectedIncludeByteValues bool
	hasDistinguishedNames     []string
	expectTypesOnly           bool
	expectOnlyAttributes      []string
}{
	{
		description: "Plain search",
		url:         "/ldap_search",
		expectedLDAPStatusCode:    0,
		expectedResultCount:       7,
		expectedIncludeByteValues: false,
		hasDistinguishedNames: []string{
			"cn=Rick Snyder,ou=people,dc=anypresence,dc=com",
			"cn=Matt Cumello,ou=people,dc=anypresence,dc=com",
			"cn=Jeff Bozek,ou=people,dc=anypresence,dc=com",
			"cn=Heather Stein,ou=people,dc=anypresence,dc=com",
		},
	},
	{
		description: "Search for single object",
		url:         "/ldap_search?baseDistinguishedName=cn%3DRick%20Snyder,ou%3Dpeople,dc%3Danypresence,dc%3Dcom",
		expectedLDAPStatusCode:    0,
		expectedResultCount:       1,
		expectedIncludeByteValues: false,
	},
	{
		description: "Search for non-existent object",
		url:         "/ldap_search?baseDistinguishedName=dc%3Dmoveon,dc=%3Dorg",
		expectedLDAPStatusCode:    10,
		expectedResultCount:       0,
		expectedIncludeByteValues: false,
	},
	{
		description: "Search with base scope",
		url:         "/ldap_search?scope=base",
		expectedLDAPStatusCode:    0,
		expectedResultCount:       1,
		expectedIncludeByteValues: false,
	},
	{
		description: "Search with single scope",
		url:         "/ldap_search?scope=single",
		expectedLDAPStatusCode:    0,
		expectedResultCount:       2,
		expectedIncludeByteValues: false,
	},
	{
		description: "Search with subtree scope",
		url:         "/ldap_search?scope=subtree",
		expectedLDAPStatusCode:    0,
		expectedResultCount:       7,
		expectedIncludeByteValues: false,
	},
	{
		description: "Search with base64 byte values included",
		url:         "/ldap_search?includeByteValues=true",
		expectedLDAPStatusCode:    0,
		expectedResultCount:       7,
		expectedIncludeByteValues: true,
	},
	{
		description: "Search with smaller size limit than result set size",
		url:         "/ldap_search?sizeLimit=5",
		expectedLDAPStatusCode:    4,
		expectedResultCount:       5,
		expectedIncludeByteValues: false,
	},
	{
		description: "Search with larger size limit than result set size",
		url:         "/ldap_search?sizeLimit=8",
		expectedLDAPStatusCode:    0,
		expectedResultCount:       7,
		expectedIncludeByteValues: false,
	},
	{
		description: "Search with equal size limit to result set size",
		url:         "/ldap_search?sizeLimit=7",
		expectedLDAPStatusCode:    0,
		expectedResultCount:       7,
		expectedIncludeByteValues: false,
	},
	{
		description: "Search with one second timeLimit",
		url:         "/ldap_search?timeLimit=1",
		expectedLDAPStatusCode:    0,
		expectedResultCount:       7,
		expectedIncludeByteValues: false,
	},
	{
		description: "Search with references that are always dealiased",
		url:         fmt.Sprintf("/ldap_search?dereferenceAliases=always&baseDistinguishedName=%s", url.QueryEscape("ou=formerEmployees,dc=anypresence,dc=com")),
		expectedLDAPStatusCode:    0,
		expectedResultCount:       2,
		expectedIncludeByteValues: false,
		hasDistinguishedNames:     []string{"cn=Heather Stein,ou=people,dc=anypresence,dc=com"},
	},
	{
		description: "Search with references that are never dealiased",
		url:         fmt.Sprintf("/ldap_search?dereferenceAliases=never&baseDistinguishedName=%s", url.QueryEscape("ou=formerEmployees,dc=anypresence,dc=com")),
		expectedLDAPStatusCode:    0,
		expectedResultCount:       2,
		expectedIncludeByteValues: false,
		hasDistinguishedNames:     []string{"cn=Heather Stein,ou=formerEmployees,dc=anypresence,dc=com"},
	},
	{
		description: "Search with references that are dealiased on search",
		url:         fmt.Sprintf("/ldap_search?dereferenceAliases=search&baseDistinguishedName=%s", url.QueryEscape("ou=formerEmployees,dc=anypresence,dc=com")),
		expectedLDAPStatusCode:    0,
		expectedResultCount:       2,
		expectedIncludeByteValues: false,
		hasDistinguishedNames:     []string{"cn=Heather Stein,ou=people,dc=anypresence,dc=com"},
	},
	{
		description: "Search with references that are dealiased on find",
		url:         fmt.Sprintf("/ldap_search?dereferenceAliases=find&baseDistinguishedName=%s", url.QueryEscape("ou=formerEmployees,dc=anypresence,dc=com")),
		expectedLDAPStatusCode:    0,
		expectedResultCount:       2,
		expectedIncludeByteValues: false,
		hasDistinguishedNames:     []string{"cn=Heather Stein,ou=formerEmployees,dc=anypresence,dc=com"},
	},
	{
		description: "Search with typesOnly",
		url:         "/ldap_search?typesOnly=true",
		expectedLDAPStatusCode:    0,
		expectedResultCount:       7,
		expectedIncludeByteValues: false,
		expectTypesOnly:           true,
	},
	{
		description: "Search with an additional filter applied",
		url:         fmt.Sprintf("/ldap_search?filter=%s", url.QueryEscape("(objectclass=inetOrgPerson)")),
		expectedLDAPStatusCode:    0,
		expectedResultCount:       4,
		expectedIncludeByteValues: false,
	},
	{
		description: "Search for subset of attributes",
		url:         fmt.Sprintf("/ldap_search?attributes=%s", url.QueryEscape("uid,cn")),
		expectedLDAPStatusCode:    0,
		expectedResultCount:       7,
		expectedIncludeByteValues: false,
		expectOnlyAttributes:      []string{"uid", "cn"},
	},
}

func ldapSetup(t *testing.T) error {
	var apiSetupErr error
	once.Do(func() {
		host, apiSetupErr = integration.ImportAPI("ldap_test_api", h)
	})

	if apiSetupErr != nil {
		return fmt.Errorf("Failed to import LDAP test API: %v", apiSetupErr)
	}

	out, err := exec.Command(
		"ldapadd",
		"-x",
		"-D", "cn=anypresence.com, dc=anypresence, dc=com",
		"-h", "192.168.99.100",
		"-w", "password",
		"-f", ldapSetupFile,
	).Output()

	if err != nil {
		fmt.Println(string(out))
		return fmt.Errorf("Failed to run ldapadd command due to %v", err)
	}

	return nil
}

func ldapTeardown(t *testing.T) error {
	c := exec.Command(
		"ldapdelete",
		"-x",
		"-D", "cn=anypresence.com, dc=anypresence, dc=com",
		"-h", "192.168.99.100",
		"-w", "password",
		"-e", "manageDSAit",
		"-r",
		"dc=us,dc=anypresence,dc=com",
	)
	c.Stderr = os.Stderr
	out, err := c.Output()
	if err != nil {
		fmt.Println(string(out))
		return fmt.Errorf("Failed to run ldapdelete command due to %v", err)
	}

	c = exec.Command(
		"ldapdelete",
		"-x",
		"-D", "cn=anypresence.com, dc=anypresence, dc=com",
		"-h", "192.168.99.100",
		"-w", "password",
		"-r",
		"dc=anypresence,dc=com",
	)
	c.Stderr = os.Stderr
	out, err = c.Output()
	if err != nil {
		fmt.Println(string(out))
		return fmt.Errorf("Failed to run ldapdelete command due to %v", err)
	}

	return nil
}

func TestLDAPSearch(t *testing.T) {
	defer ldapTeardown(t)
	err := ldapSetup(t)
	if err != nil {
		t.Error(err)
		return
	}

	for _, searchTest := range searchTests {
		result := struct {
			SearchResults struct {
				Entries []struct {
					DistinguishedName string `json:"distinguishedName"`
					Attributes        []struct {
						ByteValues []string `json:"byteValues"`
						Name       string   `json:"name"`
						Values     []string `json:"values"`
					} `json:"attributes"`
				} `json:"entries"`
			} `json:"searchResults"`
			StatusCode int `json:"statusCode"`
		}{}

		status, _, body, err := h.Get(fmt.Sprintf("%s%s", host, searchTest.url))
		if err != nil {
			t.Error(err)
			continue
		}

		if status != 200 {
			t.Errorf("[%s] Expected status to be 200, but was instead %d", searchTest.description, status)
			continue
		}

		if err := json.Unmarshal([]byte(body), &result); err != nil {
			t.Errorf("[%s] Expected to be able to unmarshal JSON but encountered error: %v", searchTest.description, err)
			continue
		}

		if searchTest.expectedLDAPStatusCode != result.StatusCode {
			t.Errorf("[%s] Expected statusCode of %d but instead got %d", searchTest.description, searchTest.expectedLDAPStatusCode, result.StatusCode)
			continue
		}

		if len(result.SearchResults.Entries) != searchTest.expectedResultCount {
			t.Errorf("[%s] Expected to have length %d but was instead %d", searchTest.description, searchTest.expectedResultCount, len(result.SearchResults.Entries))
			continue
		}

		distinguishedNames := map[string]bool{}
		for _, dn := range searchTest.hasDistinguishedNames {
			distinguishedNames[dn] = true
		}

	outer:
		for _, entry := range result.SearchResults.Entries {
			for _, attr := range entry.Attributes {
				if len(searchTest.expectOnlyAttributes) > 0 && !arrayContains(searchTest.expectOnlyAttributes, attr.Name) {
					t.Errorf("[%s] Didn't expect to receive attribute %s", searchTest.description, attr.Name)
					break outer
				}
				if searchTest.expectedIncludeByteValues {
					// TODO
					if len(attr.ByteValues) != len(attr.Values) {
						t.Errorf("[%s] Expected ByteValues and Values to have the same number of entries", searchTest.description)
						break outer
					}
					for idx, byteValue := range attr.ByteValues {
						if byteValue != base64.StdEncoding.EncodeToString([]byte(attr.Values[idx])) {
							t.Errorf("[%s] Expected byteValue to be the base64 encoding of value %s", searchTest.description, attr.Values[idx])
							break outer
						}
					}
				} else {
					if len(attr.ByteValues) > 0 {
						t.Errorf("[%s] Received byte values in attribute, but expected byte values not to be there", searchTest.description)
						break outer
					}
				}

				if searchTest.expectTypesOnly && len(attr.Values) > 0 {
					t.Errorf("[%s] Expected no values to be present in typesOnly search", searchTest.description)
					break outer
				}
			}
			delete(distinguishedNames, entry.DistinguishedName)
		}

		keys := []string{}
		for k := range distinguishedNames {
			keys = append(keys, k)
		}

		if len(keys) > 0 {
			t.Errorf("[%s] Expected results to contain distinguished names: %v", searchTest.description, keys)
		}

	}

}

func arrayContains(ary []string, value string) bool {
	for _, val := range ary {
		if val == value {
			return true
		}
	}
	return false
}

func TestLDAPAdd(t *testing.T) {
	defer ldapTeardown(t)
	err := ldapSetup(t)
	if err != nil {
		t.Error(err)
		return
	}

	addPayload := ldap.AddOperation{
		DistinguishedName: "cn=Rakesh Rao,ou=people,dc=anypresence,dc=com",
		Attributes: []*ldap.Attribute{
			&ldap.Attribute{Type: "objectclass", Values: []string{"inetOrgPerson"}},
			&ldap.Attribute{Type: "cn", Values: []string{"Rakesh Rao"}},
			&ldap.Attribute{Type: "sn", Values: []string{"Rao"}},
			&ldap.Attribute{Type: "uid", Values: []string{"rrao"}},
			&ldap.Attribute{Type: "userpassword", Values: []string{"secret"}},
			&ldap.Attribute{Type: "mail", Values: []string{"rrao@anypresence.com"}},
			&ldap.Attribute{Type: "description", Values: []string{"Founder and CTO"}},
			&ldap.Attribute{Type: "ou", Values: []string{"Executives"}},
		},
	}

	addJSON, err := json.Marshal(addPayload)
	if err != nil {
		t.Errorf("Unable to construct add operation request %v", err)
		return
	}

	status, _, body, err := h.Post(fmt.Sprintf("%s%s", host, "/ldap_add"), string(addJSON))
	if err != nil {
		t.Error(err)
	}

	if status != 200 {
		t.Errorf("Expected status to be 0, but was %d", status)
	}

	results := map[string]interface{}{}
	if err := json.Unmarshal([]byte(body), &results); err != nil {
		t.Error(err)
	}

	if sc, ok := results["statusCode"].(int); ok {
		if sc != 0 {
			t.Error(err)
		}
	}

	status, _, body, err = h.Get(fmt.Sprintf("%s%s", host, "/ldap_search"))
	if err != nil {
		t.Error(err)
	}

	result := struct {
		SearchResults struct {
			Entries []json.RawMessage `json:"entries"`
		} `json:"searchResults"`
		StatusCode int `json:"statusCode"`
	}{}

	if err := json.Unmarshal([]byte(body), &result); err != nil {
		t.Error(err)
	}

	if len(result.SearchResults.Entries) != 8 {
		t.Errorf("Entry was not added successfully. Expected 8 entries but found %d", len(result.SearchResults.Entries))
	}
}

func TestModify(t *testing.T) {
	defer ldapTeardown(t)
	err := ldapSetup(t)
	if err != nil {
		t.Error(err)
		return
	}

	modPayload := ldap.ModifyOperation{
		DistinguishedName: "cn=Rick Snyder,ou=people,dc=anypresence,dc=com",
		AddAttributes: []ldap.Attribute{
			ldap.Attribute{Type: "displayName", Values: []string{"Rick Snyder"}},
			ldap.Attribute{Type: "departmentNumber", Values: []string{"38849"}},
		},
		DeleteAttributes: []ldap.Attribute{
			ldap.Attribute{Type: "mail", Values: []string{"rsnyder@anypresence.com"}},
			ldap.Attribute{Type: "description", Values: []string{"That guy over there"}},
			ldap.Attribute{Type: "cn", Values: []string{"Uncle Rockford"}},
		},
		ReplaceAttributes: []ldap.Attribute{
			ldap.Attribute{Type: "ou", Values: []string{"Product Development"}},
			ldap.Attribute{Type: "uid", Values: []string{"rickford"}},
		},
	}

	modJSON, err := json.Marshal(modPayload)
	if err != nil {
		t.Errorf("Unable to construct modify operation request %v", err)
		return
	}

	status, _, body, err := h.Post(fmt.Sprintf("%s%s", host, "/ldap_modify"), string(modJSON))
	if err != nil {
		t.Error(err)
	}

	if status != 200 {
		t.Errorf("Expected status to be 200, but was %d", status)
		return
	}

	results := map[string]interface{}{}
	if err := json.Unmarshal([]byte(body), &results); err != nil {
		t.Error(err)
		return
	}

	if sc, ok := results["statusCode"].(int); ok {
		if sc != 0 {
			t.Error(err)
			return
		}
	}

	status, _, body, err = h.Get(fmt.Sprintf("%s%s?baseDistinguishedName=%s", host, "/ldap_search", url.QueryEscape("cn=Rick Snyder,ou=people,dc=anypresence,dc=com")))
	if err != nil {
		t.Error(err)
		return
	}

	if status != 200 {
		t.Errorf("Expected status to be 200, but was %d", status)
		return
	}

	result := struct {
		SearchResults struct {
			Entries []struct {
				DistinguishedName string `json:"distinguishedName"`
				Attributes        []struct {
					Name   string   `json:"name"`
					Values []string `json:"values"`
				} `json:"attributes"`
			} `json:"entries"`
		} `json:"searchResults"`
		StatusCode int `json:"statusCode"`
	}{}

	if err := json.Unmarshal([]byte(body), &result); err != nil {
		t.Error(err)
		return
	}

	if result.StatusCode != 0 {
		t.Errorf("Expected status code to be 0, but was instead %d", result.StatusCode)
		return
	}

	if len(result.SearchResults.Entries) != 1 {
		t.Errorf("Expected only one search result but instead found %d", len(result.SearchResults.Entries))
	}

	expectedAttributes := map[string]string{
		"objectClass":      "inetOrgPerson",
		"cn":               "Rick Snyder",
		"sn":               "Snyder",
		"userPassword":     "p4s5w0rD",
		"displayName":      "Rick Snyder",
		"departmentNumber": "38849",
		"ou":               "Product Development",
		"uid":              "rickford",
	}

	entry := result.SearchResults.Entries[0]

	for _, a := range entry.Attributes {
		if _, ok := expectedAttributes[a.Name]; !ok {
			t.Errorf("Attribute found with unexpected name: %s", a.Name)
			continue
		}

		if len(a.Values) > 1 {
			t.Errorf("All attributes are expected to have only one value, but found one with %d", len(a.Values))
			continue
		}

		if a.Values[0] != expectedAttributes[a.Name] {
			t.Errorf("Expected attribute with name %s to have value %s, but instead found %s", a.Name, expectedAttributes[a.Name], a.Values[0])
			continue
		}

		delete(expectedAttributes, a.Name)
	}

	if len(expectedAttributes) > 0 {
		t.Errorf("Not all expected attributes were present! Expected but not present: %v", expectedAttributes)
	}
}

func TestDelete(t *testing.T) {
	defer ldapTeardown(t)
	err := ldapSetup(t)
	if err != nil {
		t.Error(err)
		return
	}

	status, _, body, err := h.Get(fmt.Sprintf("%s%s?distinguishedName=%s", host, "/ldap_delete", url.QueryEscape("cn=Matt Cumello,ou=people,dc=anypresence,dc=com")))

	if err != nil {
		t.Error(err)
	}

	if status != 200 {
		t.Errorf("Expected status to be 0, but was %d", status)
	}

	status, _, body, err = h.Get(fmt.Sprintf("%s%s", host, "/ldap_search"))
	if err != nil {
		t.Error(err)
	}

	result := struct {
		SearchResults struct {
			Entries []json.RawMessage `json:"entries"`
		} `json:"searchResults"`
		StatusCode int `json:"statusCode"`
	}{}

	if err := json.Unmarshal([]byte(body), &result); err != nil {
		t.Error(err)
	}

	if len(result.SearchResults.Entries) != 6 {
		t.Errorf("Entry was not added successfully. Expected 6 entries but found %d", len(result.SearchResults.Entries))
	}
}

func TestBind(t *testing.T) {
	defer ldapTeardown(t)
	err := ldapSetup(t)
	if err != nil {
		t.Error(err)
		return
	}

	testCases := []struct {
		Description        string
		Username           string
		Password           string
		ExpectedStatusCode int
	}{
		{
			Description:        "Valid Password",
			Username:           "cn=Rick Snyder,ou=people,dc=anypresence,dc=com",
			Password:           "p4s5w0rD",
			ExpectedStatusCode: 0,
		},
		{
			Description:        "Invalid Password",
			Username:           "cn=Rick Snyder,ou=people,dc=anypresence,dc=com",
			Password:           "p4s5w0RD",
			ExpectedStatusCode: 49,
		},
	}

	for _, tc := range testCases {
		status, _, body, err := h.Get(
			fmt.Sprintf(
				"%s%s?username=%s&password=%s",
				host,
				"/ldap_bind",
				url.QueryEscape(tc.Username),
				tc.Password,
			),
		)

		if err != nil {
			t.Errorf("[%s] Encountered unexpected error: %v", tc.Description, err)
			continue
		}

		if status != 200 {
			t.Errorf("[%s] Expected status to be 0, but was %d", tc.Description, status)
			continue
		}

		result := map[string]interface{}{}

		err = json.Unmarshal([]byte(body), &result)
		if err != nil {
			t.Error(err)
			continue
		}

		if sc, ok := result["statusCode"].(float64); ok {
			if int(sc) != tc.ExpectedStatusCode {
				t.Errorf("[%s] Expected statusCode to be 0, but was instead %d", tc.Description, int(sc))
				continue
			}
		} else {
			t.Errorf("[%s] Expected statusCode to be an integer value, but was %T", tc.Description, result["statusCode"])
		}
	}
}

func TestCompare(t *testing.T) {
	defer ldapTeardown(t)
	err := ldapSetup(t)
	if err != nil {
		t.Error(err)
		return
	}

	status, _, body, err := h.Get(
		fmt.Sprintf(
			"%s%s?distinguishedName=%s&attribute=%s&value=%s",
			host,
			"/ldap_compare",
			url.QueryEscape("cn=Matt Cumello,ou=people,dc=anypresence,dc=com"),
			"objectclass",
			"inetOrgPerson",
		),
	)

	if err != nil {
		t.Error(err)
	}

	if status != 200 {
		t.Errorf("Expected status to be 200 but was instead %d", status)
	}

	result := struct {
		CompareResult struct {
			Matches bool `json:"matches"`
		} `json:"compareResult"`
		StatusCode int `json:"statusCode"`
	}{}

	if err := json.Unmarshal([]byte(body), &result); err != nil {
		t.Error(err)
	}

	if result.StatusCode != 0 {
		t.Errorf("Expected statusCode to be 0 but was instead %d", result.StatusCode)
	}

	if !result.CompareResult.Matches {
		t.Errorf("Expected matches to equal true but was intead %v", result.CompareResult.Matches)
	}
}

/*func TestTLS(t *testing.T) {
	t.Error("Not implemented yet")
}*/

// TODO - add additional tests
func TestMain(m *testing.M) {
	fmt.Println("\n\n\n***isIntegrationTest ", integration.IsIntegrationTest)
	if !integration.IsIntegrationTest {
		log.Println("Integration flag not set.  Skipping integration tests.")
		os.Exit(0)
		//return
	}

	os.Exit(m.Run())
}
