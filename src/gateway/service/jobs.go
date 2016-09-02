package service

import (
	"math"
	"math/rand"
	"time"

	"gateway/config"
	"gateway/logreport"
	"gateway/model"
	"gateway/sql"
)

func JobsService(conf config.Configuration, db *sql.DB) {
	ticker := time.NewTicker(time.Minute)
	source := rand.New(rand.NewSource(time.Now().Unix()))
	go func() {
		for _ = range ticker.C {
			timer, now := model.Timer{}, time.Now().Unix()
			timers, err := timer.AllReady(db, now)
			if err != nil {
				logreport.Printf("[jobs] %v", err)
			}
			for len(timers) > 0 {
				length := len(timers)
				t := source.Intn(length)
				locked, err := db.TryLock("timers", timers[t].ID)
				if err != nil {
					logreport.Printf("[jobs] %v", err)
				}
				if locked {
					fresh, err := timers[t].Find(db)
					if err != nil {
						logreport.Printf("[jobs] %v", err)
					}
					if fresh.Next < now {
						err = db.DoInTransaction(func(tx *sql.Tx) error {
							if fresh.Once {
								fresh.Next = math.MaxInt64
							}
							return fresh.Update(tx)
						})
						if err != nil {
							logreport.Printf("[jobs] %v", err)
						}
					}
					_, err = db.Unlock("timers", timers[t].ID)
					if err != nil {
						logreport.Printf("[jobs] %v", err)
					}
				}
				timers[t], timers[length-1] = timers[length-1], timers[t]
				timers = timers[:length-1]
			}
		}
	}()
}
