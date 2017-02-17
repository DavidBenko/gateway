package service

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"gateway/admin"
	"gateway/config"
	"gateway/errors/report"
	"gateway/logreport"

	"github.com/jmoiron/sqlx"
)

const (
	postgresCurrentVersion = 1
)

var (
	ErrMigrate = errors.New("The log db is not up to date. Please migrate by invoking with the -postgres-logging-migrate flag.")
)

type PostgresMessage struct {
	ID         int64     `json:"id"`
	Text       string    `json:"text"`
	Time       time.Time `json:"time"`
	AccountID  int64     `json:"account_id" db:"account_id"`
	APIID      int64     `json:"api_id" db:"api_id"`
	EndpointID int64     `json:"endpoint_id" db:"endpoint_id"`
	TimerID    int64     `json:"timer_id" db:"timer_id"`
}

func NewPostgresMessage(message string) *PostgresMessage {
	logDate := TimeRegexp.FindStringSubmatch(message)
	var date time.Time
	if len(logDate) == 3 {
		var err error
		date, err = time.Parse(LogTimeFormat, logDate[1])
		if err != nil {
			logreport.Fatal(err)
		}
		seconds, err := strconv.Atoi(logDate[2])
		if err != nil {
			logreport.Fatal(err)
		}
		date = date.Add(time.Duration(seconds) * time.Microsecond)
	} else {
		return nil
	}
	var properties [4]int64
	for i, re := range []*regexp.Regexp{AccountRegexp, APIRegexp, EndpointRegexp, TimerRegexp} {
		matches := re.FindStringSubmatch(message)
		if len(matches) == 2 {
			properties[i], _ = strconv.ParseInt(matches[1], 10, 64)
		} else {
			properties[i] = 0
		}
	}

	return &PostgresMessage{
		Text:       message,
		Time:       date,
		AccountID:  properties[0],
		APIID:      properties[1],
		EndpointID: properties[2],
		TimerID:    properties[3],
	}
}

func PostgresLoggingService(conf config.PostgresLogging) {
	if !conf.Enable {
		return
	}

	logreport.Printf("%s Starting Postgres logging service", config.System)

	reportError := func(err error) {
		logreport.Fatalf("%s %v", config.System, err)
	}

	db, err := sqlx.Open("postgres", conf.ConnectionString)
	if err != nil {
		reportError(err)
	}

	db.SetMaxOpenConns(int(conf.MaxConnections))

	var currentVersion int64
	err = db.Get(&currentVersion, `SELECT version FROM schema LIMIT 1`)
	migrate := conf.Migrate
	if err != nil {
		tx := db.MustBegin()
		tx.MustExec(`
      CREATE TABLE IF NOT EXISTS schema (
        version integer
      );
    `)
		tx.MustExec(`INSERT INTO schema VALUES (0);`)
		err = tx.Commit()
		if err != nil {
			reportError(err)
		}

		migrate = true
	}

	if currentVersion < postgresCurrentVersion {
		if !migrate {
			reportError(ErrMigrate)
		}

		if currentVersion < 1 {
			err = migrateToV1(db)
			if err != nil {
				reportError(err)
			}
		}
	}

	admin.Postgres = db

	reportError = func(err error) {
		fmt.Printf("[postgres-logging] %v\n", err)
		report.Error(err, nil)
	}

	indexer := make(chan *PostgresMessage, 8192)
	go func() {
		var buffer [10]*PostgresMessage
		c := 0

		send := func() (err error) {
			if c <= 0 {
				return nil
			}
			defer func() {
				c = 0
			}()

			tx, err := db.Beginx()
			if err != nil {
				return err
			}
			defer func() {
				if err != nil {
					tx.Rollback()
					return
				}
				err = tx.Commit()
			}()
			for _, message := range buffer[:c] {
				_, err = tx.Exec(`INSERT into logs (
            text, time, account_id,
            api_id, endpoint_id, timer_id
          ) VALUES (
            $1, $2, $3,
            $4, $5, $6
          );`, message.Text, message.Time, message.AccountID,
					message.APIID, message.EndpointID, message.TimerID)
				if err != nil {
					return err
				}
			}

			return nil
		}

		for {
			write := time.Tick(time.Second)
			select {
			case message := <-indexer:
				if c >= len(buffer) {
					err := send()
					if err != nil {
						reportError(err)
					}
				}
				buffer[c] = message
				c++
			case <-write:
				err := send()
				if err != nil {
					reportError(err)
				}
			}
		}
	}()

	go func() {
		logs, unsubscribe := admin.Interceptor.Subscribe("PostgresLoggingService")
		defer unsubscribe()

		add := func(message string) {
			postgresMessage := NewPostgresMessage(message)
			if postgresMessage == nil {
				return
			}
			indexer <- postgresMessage
		}
		processLogs(logs, add)
	}()

	deleteTicker := time.NewTicker(24 * time.Hour)
	go func() {
		for _ = range deleteTicker.C {
			days := time.Duration(conf.DeleteAfter)
			end := time.Now().Add(-days * 24 * time.Hour)
			_, err := db.Exec(`DELETE FROM logs WHERE time < $1;`, &end)
			if err != nil {
				reportError(err)
			}
		}
	}()
}

func migrateToV1(db *sqlx.DB) error {
	tx := db.MustBegin()
	tx.MustExec(`
    CREATE TABLE IF NOT EXISTS "logs" (
      "id" SERIAL PRIMARY KEY,
      "text" TEXT NOT NULL,
      "time" TIMESTAMPTZ NOT NULL,
      "account_id" INTEGER NOT NULL,
      "api_id" INTEGER NOT NULL,
      "endpoint_id" INTEGER NOT NULL,
      "timer_id" INTEGER NOT NULL,
      "tsv" tsvector
    );
  `)
	tx.MustExec(`
    CREATE INDEX idx_logs_time ON logs USING btree(time);
    CREATE INDEX idx_logs_account_id ON logs USING btree(account_id);
    CREATE INDEX idx_logs_api_id ON logs USING btree(api_id);
    CREATE INDEX idx_logs_endpoint_id ON logs USING btree(endpoint_id);
    CREATE INDEX idx_logs_timer_id ON logs USING btree(timer_id);
    CREATE INDEX idx_logs_tsv ON logs USING gin(tsv);
  `)
	tx.MustExec(`
    CREATE TRIGGER logs_tsv_trigger BEFORE INSERT OR UPDATE
    ON logs FOR EACH ROW EXECUTE PROCEDURE
    tsvector_update_trigger(tsv, 'pg_catalog.english', text);
  `)
	tx.MustExec(`UPDATE schema SET version = 1;`)
	err := tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
