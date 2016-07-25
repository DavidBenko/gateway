package OttoCrypto

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"gateway/logreport"

	"github.com/robertkrimen/otto"
	"golang.org/x/crypto/bcrypt"
)

// IncludeHashing extends the Otto VM and adds a hashing function.
func IncludeHashing(vm *otto.Otto) {
	setHashPassword(vm)
	setHash(vm)
}

func setHashPassword(vm *otto.Otto) {
	vm.Set("_hashPassword", func(call otto.FunctionCall) otto.Value {
		undefined := otto.Value{}

		passwordArg := call.Argument(0)
		if passwordArg == undefined {
			logreport.Print("password is undefined")
			return undefined
		}

		password, err := passwordArg.ToString()

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		iterationsArg := call.Argument(1)

		if iterationsArg == undefined {
			logreport.Print("iterations is undefined")
			return undefined
		}

		iterations, err := iterationsArg.ToInteger()

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		result, err := bcrypt.GenerateFromPassword([]byte(password), int(iterations))

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		val, err := vm.ToValue(string(result[:]))

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		return val
	})
}

func setHash(vm *otto.Otto) {
	vm.Set("_hash", func(call otto.FunctionCall) otto.Value {
		undefined := otto.Value{}

		tagArg := call.Argument(0)
		if tagArg == undefined {
			logreport.Print("tag is undefined")
			return undefined
		}

		tag, err := tagArg.ToString()

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		dataArg := call.Argument(1)
		if dataArg == undefined {
			logreport.Print("data is undefined")
			return undefined
		}

		data, err := dataArg.ToString()

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		result := hash(tag, data)

		val, err := vm.ToValue(result)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		return val
	})
}

func hash(tag string, data string) string {
	h := hmac.New(sha512.New512_256, []byte(tag))
	h.Write([]byte(data))
	val := h.Sum(nil)

	return base64.StdEncoding.EncodeToString(val)
}
