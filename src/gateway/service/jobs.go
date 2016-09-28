package service

import (
	"fmt"
	"math/rand"
	"time"

	"gateway/config"
	"gateway/core"
	"gateway/logreport"
	"gateway/model"
	"gateway/sql"
)

func JobsService(conf config.Configuration, warp *core.Core) {
	ticker := time.NewTicker(time.Minute)
	source := rand.New(rand.NewSource(time.Now().Unix()))
	go func() {
		for now := range ticker.C {
			timer := model.Timer{}
			timers, err := timer.AllReady(warp.OwnDb, now.Unix())
			if err != nil {
				logreport.Printf("%s %v", config.Job, err)
			}
			for len(timers) > 0 {
				length := len(timers)
				t := source.Intn(length)
				logPrefix := fmt.Sprintf("%s [act %d] [timer %d] [api %d] [end %d]", config.Job,
					timers[t].AccountID, timers[t].ID, timers[t].APIID, timers[t].JobID)
				err = executeJob(timers[t], now.Unix(), logPrefix, warp)
				if err != nil {
					logreport.Printf("%s %v", logPrefix, err)
				}
				timers[t], timers[length-1] = timers[length-1], timers[t]
				timers = timers[:length-1]
			}
		}
	}()
}

func executeJob(timer *model.Timer, now int64, logPrefix string, warp *core.Core) (err error) {
	db := warp.OwnDb

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
		return nil
	}
	if fresh.Next > now {
		return nil
	}

	logreport.Printf("%s %s", logPrefix, fresh.Name)

	err = db.DoInTransaction(func(tx *sql.Tx) error {
		if fresh.Once {
			return fresh.Delete(tx)
		}
		return fresh.Update(tx)
	})
	if err != nil {
		return err
	}

	attributesJSON, err := timer.Attributes.MarshalJSON()
	if err != nil {
		return err
	}

	return warp.ExecuteJob(timer.JobID, timer.APIID, logPrefix, string(attributesJSON))
}
