package ottocrypto

import (
	"errors"
	"fmt"
	"gateway/logreport"

	"github.com/robertkrimen/otto"
)

var (
	undefined = otto.Value{}

	// DefaultHashAlgorithm is used if nothing is supplied in the options.
	DefaultHashAlgorithm = "sha256"
	// DefaultPaddingScheme is  used if nothing is supplied in the options.
	DefaultPaddingScheme = "pkcs1v15"
)

// KeyDataSource is a source/cache for crypto keys.
type KeyDataSource interface {
	GetKey(int64, string) (interface{}, bool)
}

// GetKeyFromSource returns the crypto key given a key option in the options map and an accountID from the supplied
// KeyDataSource.
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

// GetOptionString gets the supplied key value from the options map. If optional is true, will not
// return an error if nothing is found.
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

func toOttoObjectValue(vm *otto.Otto, s string) otto.Value {
	obj, err := vm.Object(fmt.Sprintf("(%s)", string(s)))

	if err != nil {
		logreport.Print(err)
		return undefined
	}
	result, err := vm.ToValue(obj)
	if err != nil {
		logreport.Print(err)
		return undefined
	}
	return result

}

func getArgument(call otto.FunctionCall, index int) (interface{}, error) {
	arg := call.Argument(index)
	if arg == undefined {
		return nil, errors.New("undefined argument")
	}

	return arg.Export()
}
