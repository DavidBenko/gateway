package vm

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/robertkrimen/otto"
)

var knownFuncs *regexp.Regexp

type apFunc string

const (
	prepareRequests apFunc = "prepareRequests"
)

func init() {
	funcs := []string{
		string(prepareRequests),
	}

	knownFuncs = regexp.MustCompile(strings.Join(funcs, "|"))
}

// makeCall ensures the given call is in the set of known AP functions defined
// for the VM, then makes the call with the given slice of string args.
//
// TODO(binary132): bring all known AP funcs into this.
// TODO(binary132): measure performance impact of regex
func (c *CoreVM) makeCall(fn apFunc, args []string) (otto.Value, error) {
	if !knownFuncs.MatchString(string(fn)) {
		return otto.UndefinedValue(), fmt.Errorf("No such AP function %q", fn)
	}
	script := fmt.Sprintf("AP.%s(%s);", fn, strings.Join(args, ","))

	return c.Run(script)
}
