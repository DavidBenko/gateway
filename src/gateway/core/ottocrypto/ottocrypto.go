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

// getKeyFromSource returns the crypto key given a key option in the options map and an accountID from the supplied
// KeyDataSource.
func getKeyFromSource(options map[string]interface{}, keySource KeyDataSource, accountID int64) (interface{}, error) {
	var key interface{}
	k, err := getOptionString(options, "key", false)
	if err != nil {
		return key, err
	}

	if val, found := keySource.GetKey(accountID, k); found {
		return val, nil
	}
	return key, fmt.Errorf("key not found with name %s", k)
}

// getOptionString gets the supplied key value from the options map. If optional is true, will not
// return an error if nothing is found.
func getOptionString(options map[string]interface{}, key string, optional bool) (string, error) {
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

func getOptions(opts interface{}, keySource KeyDataSource, accountID int64) (key interface{}, algorithm string, tag string, err error) {
	options, ok := opts.(map[string]interface{})
	if !ok {
		err = errors.New("options should be an object")
		return
	}

	algorithm = DefaultHashAlgorithm
	key, err = getKeyFromSource(options, keySource, accountID)
	if err != nil {
		return
	}

	tag, err = getOptionString(options, "tag", true)
	if err != nil {
		return
	}

	a, err := getOptionString(options, "algorithm", true)
	if err != nil {
		return
	}
	if a != "" {
		algorithm = a
	}
	return
}

func getData(call otto.FunctionCall) (string, error) {
	d, err := getArgument(call, 0)
	if err != nil {
		return "", err
	}

	if ds, ok := d.(string); ok {
		return ds, nil
	}
	return "", errors.New("data should be a string")
}
