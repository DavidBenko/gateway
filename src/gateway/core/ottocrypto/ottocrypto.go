package ottocrypto

import "fmt"

// DefaultHashAlgorithm is used if nothing is supplied in the options.
var DefaultHashAlgorithm = "sha256"

// Default padding scheme used if nothing is supplied in the options.
var DefaultPaddingScheme = "pkcs1v15"

func GetKeyFromSource(options map[string]interface{}, keySource KeyDataSource, accountID int64) (interface{}, error) {
	var key interface{}
	k, err := GetOptionString(options, "key", false)
	if err != nil {
		return key, err
	}

	if val, found := keySource.GetKey(accountID, k); found {
		return val, nil
	}
	return key, fmt.Errorf("key not found with name %s", k)
}

func GetOptionString(options map[string]interface{}, key string, optional bool) (string, error) {
	if k, ok := options[key]; ok {
		if s, ok := k.(string); ok {
			return s, nil
		}
		return "", fmt.Errorf("%s should be a string", key)
	}
	if optional {
		return "", nil
	}
	return "", fmt.Errorf("option not found with name %s", key)
}
