package service

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"gateway/config"
	"gateway/core"
	"gateway/logreport"
	"gateway/model"
	"gateway/sql"
)

var (
	stopJobsService int32
	jobsCount       int64
)

func JobsService(conf config.Configuration, warp *core.Core) {
	ticker := time.NewTicker(time.Second * 30)
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
				go func(timer *model.Timer, now int64) {
					atomic.AddInt64(&jobsCount, 1)
					defer func() {
						atomic.AddInt64(&jobsCount, -1)
					}()
					logPrefix := fmt.Sprintf("%s [act %d] [timer %d] [api %d] [end %d]", config.Job,
						timer.AccountID, timer.ID, timer.APIID, timer.JobID)
					err = executeJob(timer, now, logPrefix, warp)
					if err != nil {
						logreport.Printf("%s %v", logPrefix, err)
					}
				}(timers[t], now.Unix())
				timers[t], timers[length-1] = timers[length-1], timers[t]
				timers = timers[:length-1]
			}
			if atomic.LoadInt32(&stopJobsService) > 0 {
				return
			}
		}
	}()
}

func StopJobsService() {
	atomic.StoreInt32(&stopJobsService, 1)
	for atomic.LoadInt64(&jobsCount) > 0 {
		time.Sleep(time.Second)
	}
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
