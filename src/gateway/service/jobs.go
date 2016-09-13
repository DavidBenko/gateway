package service

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"gateway/config"
	"gateway/core"
	"gateway/core/vm"
	"gateway/logreport"
	"gateway/model"
	"gateway/sql"
)

func JobsService(conf config.Configuration, db *sql.DB, warp *core.Core) {
	ticker := time.NewTicker(time.Minute)
	source := rand.New(rand.NewSource(time.Now().Unix()))
	go func() {
		for now := range ticker.C {
			timer := model.Timer{}
			timers, err := timer.AllReady(db, now.Unix())
			if err != nil {
				logreport.Printf("%s %v", config.Job, err)
			}
			for len(timers) > 0 {
				length := len(timers)
				t := source.Intn(length)
				logPrefix := fmt.Sprintf("%s [act %d] [api %d] [end %d]", config.Job,
					timers[t].AccountID, timers[t].APIID, timers[t].JobID)
				err = executeJob(db, timers[t], now.Unix(), logPrefix, &conf.Job, warp)
				if err != nil {
					logreport.Printf("%s %v", logPrefix, err)
				}
				timers[t], timers[length-1] = timers[length-1], timers[t]
				timers = timers[:length-1]
			}
		}
	}()
}

func executeJob(db *sql.DB, timer *model.Timer, now int64, logPrefix string, conf vm.VMConfig, warp *core.Core) (err error) {
	locked, err := db.TryLock("timers", timer.ID)
	if err != nil {
		return err
	}
	if !locked {
		return nil
	}

	defer func() {
		_, errUnlock := db.Unlock("timers", timer.ID)
		if errUnlock != nil {
			err = errUnlock
		}
	}()

	fresh, err := timer.Find(db)
	if err != nil {
		return err
	}
	if fresh.Next > now {
		return nil
	}

	err = db.DoInTransaction(func(tx *sql.Tx) error {
		if fresh.Once {
			fresh.Next = math.MaxInt64
		}
		return fresh.Update(tx)
	})
	if err != nil {
		return err
	}

	job, err := model.FindProxyEndpointForProxy(db, timer.JobID, model.ProxyEndpointTypeJob)
	if err != nil {
		return err
	}
	libraries, err := model.AllLibrariesForProxy(db, timer.APIID)
	if err != nil {
		return err
	}

	vm := &vm.CoreVM{}
	vm.InitCoreVM(core.VMCopy(), logreport.Printf, logPrefix, conf, job, libraries, conf.GetCodeTimeout())

	attributesJSON, err := timer.Attributes.MarshalJSON()
	if err != nil {
		return err
	}
	vm.Set("__ap_jobAttributesJSON", attributesJSON)
	scripts := []interface{}{
		"var attributes = JSON.parse(__ap_jobAttributesJSON);",
	}
	if _, err = vm.RunAll(scripts); err != nil {
		return err
	}

	if err = warp.RunComponents(vm, job.Components); err != nil {
		if err.Error() == "JavaScript took too long to execute" {
			logreport.Printf("%s [timeout] JavaScript execution exceeded %ds timeout threshold", logPrefix, conf.GetCodeTimeout())
			return nil
		}
		return err
	}

	return nil
}
