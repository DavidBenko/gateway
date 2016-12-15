package core

import (
	"encoding/json"

	"gateway/core/vm"
	"gateway/logreport"
	"gateway/model"

	stripe "github.com/stripe/stripe-go"
)

func (c *Core) ExecuteJob(jobID, accountID, apiID int64, logPrint logreport.Logf, logPrefix, parameters string) (err error) {
	conf := &c.Conf.Job

	job, err := model.FindProxyEndpointForProxy(c.OwnDb, jobID, model.ProxyEndpointTypeJob)
	if err != nil {
		return err
	}
	libraries, err := model.AllLibrariesForProxy(c.OwnDb, apiID)
	if err != nil {
		return err
	}

	codeTimeout := conf.GetCodeTimeout()
	if stripe.Key != "" {
		plan, err := model.FindPlanByAccountID(c.OwnDb, accountID)
		if err != nil {
			return err
		}
		if plan.JobTimeout < codeTimeout {
			codeTimeout = plan.JobTimeout
		}
	}

	vm := &vm.CoreVM{}
	vm.InitCoreVM(VMCopy(accountID, c.VMKeyStore), logPrint, logPrefix, conf, job, libraries, codeTimeout)

	vm.Set("__ap_jobParametersJSON", parameters)
	scripts := []interface{}{
		"var parameters = JSON.parse(__ap_jobParametersJSON);",
	}
	if _, err = vm.RunAll(scripts); err != nil {
		return err
	}
	vm.Set("result", "done")

	if err = c.RunComponents(vm, job.Components); err != nil {
		if err.Error() == "JavaScript took too long to execute" {
			logreport.Printf("%s [timeout] JavaScript execution exceeded %ds timeout threshold", logPrefix, conf.GetCodeTimeout())
			return nil
		}
		return err
	}

	value, err := vm.Get("result")
	if err != nil {
		return err
	}
	export, err := value.Export()
	if err != nil {
		return err
	}
	result, err := json.Marshal(export)
	if err != nil {
		return err
	}
	logPrint("%s %s %s", logPrefix, job.Name, string(result))

	return nil
}
