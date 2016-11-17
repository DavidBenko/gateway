package model

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	aperrors "gateway/errors"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

type Timer struct {
	AccountID int64 `json:"-" db:"account_id"`
	UserID    int64 `json:"-"`

	ID         int64          `json:"id,omitempty" path:"id"`
	APIID      int64          `json:"api_id" db:"api_id"`
	JobID      int64          `json:"job_id" db:"job_id"`
	Name       string         `json:"name"`
	Once       bool           `json:"once"`
	TimeZone   int64          `json:"time_zone" db:"time_zone"`
	Minute     string         `json:"minute"`
	Hour       string         `json:"hour"`
	DayOfMonth string         `json:"day_of_month" db:"day_of_month"`
	Month      string         `json:"month"`
	DayOfWeek  string         `json:"day_of_week" db:"day_of_week"`
	Next       int64          `json:"next"`
	Parameters types.JsonText `json:"parameters"`
	Data       types.JsonText `json:"-"`
}

type field struct {
	name     string
	min, max int
	names    map[string]int
}

const starBit = 1 << 63

func (f *field) Parse(a string) (bits uint64, err error) {
	switch a {
	case "*":
		for b := f.min; b <= f.max; b++ {
			bits |= 1 << uint(b)
		}
		bits |= starBit
	default:
		parts := strings.Split(a, ",")
		for _, part := range parts {
			if b, err := strconv.Atoi(part); err == nil {
				if b < f.min || b > f.max {
					return 0, fmt.Errorf("%v is out of range %v-%v for %v", b, f.min, f.max, f.name)
				}
				bits |= 1 << uint(b)
			} else if c, ok := f.names[part]; ok {
				bits |= 1 << uint(c)
			} else {
				return 0, fmt.Errorf("%v is invalid for %v", part, f.name)
			}
		}
	}

	return bits, nil
}

var (
	fieldMinute     = field{"minutes", 0, 59, nil}
	fieldHour       = field{"hours", 0, 23, nil}
	fieldDayOfMonth = field{"day of month", 1, 31, nil}
	fieldMonth      = field{"months", 1, 12, map[string]int{
		"jan": 1,
		"feb": 2,
		"mar": 3,
		"apr": 4,
		"may": 5,
		"jun": 6,
		"jul": 7,
		"aug": 8,
		"sep": 9,
		"oct": 10,
		"nov": 11,
		"dec": 12,
	}}
	fieldDayOfWeek = field{"day of week", 0, 6, map[string]int{
		"sun": 0,
		"mon": 1,
		"tue": 2,
		"wed": 3,
		"thu": 4,
		"fri": 5,
		"sat": 6,
	}}
)

func (t *Timer) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if t.Name == "" || strings.TrimSpace(t.Name) == "" {
		errors.Add("name", "must not be blank")
	}

	if t.Once {
		return errors
	}

	if t.Minute == "" || strings.TrimSpace(t.Minute) == "" {
		errors.Add("minute", "must not be blank")
	}
	if _, err := fieldMinute.Parse(t.Minute); err != nil {
		errors.Add("minute", err.Error())
	}

	if t.Hour == "" || strings.TrimSpace(t.Hour) == "" {
		errors.Add("hour", "must not be blank")
	}
	if _, err := fieldHour.Parse(t.Hour); err != nil {
		errors.Add("hour", err.Error())
	}

	if t.DayOfMonth == "" || strings.TrimSpace(t.DayOfMonth) == "" {
		errors.Add("day_of_month", "must not be blank")
	}
	if _, err := fieldDayOfMonth.Parse(t.DayOfMonth); err != nil {
		errors.Add("day_of_month", err.Error())
	}

	if t.Month == "" || strings.TrimSpace(t.Month) == "" {
		errors.Add("month", "must not be blank")
	}
	if _, err := fieldMonth.Parse(t.Month); err != nil {
		errors.Add("month", err.Error())
	}

	if t.DayOfWeek == "" || strings.TrimSpace(t.DayOfWeek) == "" {
		errors.Add("day_of_week", "must not be blank")
	}
	if _, err := fieldDayOfWeek.Parse(t.DayOfWeek); err != nil {
		errors.Add("day_of_week", err.Error())
	}

	return errors
}

func (t *Timer) FindNext(next time.Time) time.Time {
	next = next.Add(1*time.Minute - time.Duration(next.Second())*time.Second)
	added, limit := false, next.Year()+5

	minute, _ := fieldMinute.Parse(t.Minute)
	hour, _ := fieldHour.Parse(t.Hour)
	dayOfMonth, _ := fieldDayOfMonth.Parse(t.DayOfMonth)
	month, _ := fieldMonth.Parse(t.Month)
	dayOfWeek, _ := fieldDayOfWeek.Parse(t.DayOfWeek)

	dayMatches := func() bool {
		domMatch := 1<<uint(next.Day())&dayOfMonth > 0
		dowMatch := 1<<uint(next.Weekday())&dayOfWeek > 0

		if dayOfMonth&starBit > 0 || dayOfWeek&starBit > 0 {
			return domMatch && dowMatch
		}
		return domMatch || dowMatch
	}

NEXT:
	if next.Year() > limit {
		return time.Time{}
	}

	for 1<<uint(next.Month())&month == 0 {
		if !added {
			added = true
			next = time.Date(next.Year(), next.Month(), 1, 0, 0, 0, 0, next.Location())
		}
		next = next.AddDate(0, 1, 0)

		if next.Month() == time.January {
			goto NEXT
		}
	}

	for !dayMatches() {
		if !added {
			added = true
			next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
		}
		next = next.AddDate(0, 0, 1)

		if next.Day() == 1 {
			goto NEXT
		}
	}

	for 1<<uint(next.Hour())&hour == 0 {
		if !added {
			added = true
			next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), 0, 0, 0, next.Location())
		}
		next = next.Add(1 * time.Hour)

		if next.Hour() == 0 {
			goto NEXT
		}
	}

	for 1<<uint(next.Minute())&minute == 0 {
		if !added {
			added = true
			next = next.Truncate(time.Minute)
		}
		next = next.Add(1 * time.Minute)

		if next.Minute() == 0 {
			goto NEXT
		}
	}

	return next.Add(time.Duration(rand.Intn(16)) * time.Second)
}

func (t *Timer) Schedule(tx *apsql.Tx) error {
	if t.Once {
		return nil
	}

	current, err := tx.DB.CurrentTime()
	if err != nil {
		return err
	}
	location := time.FixedZone(fmt.Sprintf("fz%v", t.TimeZone), int(t.TimeZone)*60*60)
	next := t.FindNext(current.In(location))
	t.Next = next.Unix()

	return nil
}

func (t *Timer) ScheduleTime(now time.Time) {
	if t.Once {
		return
	}

	location := time.FixedZone(fmt.Sprintf("fz%v", t.TimeZone), int(t.TimeZone)*60*60)
	next := t.FindNext(now.In(location))
	t.Next = next.Unix()
}

func (t *Timer) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "timers", "api_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

func (t *Timer) All(db *apsql.DB) ([]*Timer, error) {
	timers := []*Timer{}
	err := db.Select(&timers, db.SQL("timers/all"), t.AccountID)
	if err != nil {
		return nil, err
	}
	for _, timer := range timers {
		timer.AccountID = t.AccountID
		timer.UserID = t.UserID
	}
	return timers, nil
}

func (t *Timer) AllReady(db *apsql.DB, now int64) ([]*Timer, error) {
	timers := []*Timer{}
	err := db.Select(&timers, db.SQL("timers/all_ready"), now)
	if err != nil {
		return nil, err
	}
	return timers, nil
}

func (t *Timer) Find(db *apsql.DB) (*Timer, error) {
	timer := Timer{
		AccountID: t.AccountID,
		UserID:    t.UserID,
	}
	err := db.Get(&timer, db.SQL("timers/find"), t.ID, t.AccountID)
	if err != nil {
		return nil, err
	}
	return &timer, nil
}

func (t *Timer) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("timers/delete"), t.ID, t.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("timers", t.AccountID, t.UserID, t.APIID, 0, t.ID, apsql.Delete)
}

func (t *Timer) Insert(tx *apsql.Tx) error {
	err := t.Schedule(tx)
	if err != nil {
		return err
	}

	parameters, err := marshaledForStorage(t.Parameters)
	if err != nil {
		return err
	}

	data, err := marshaledForStorage(t.Data)
	if err != nil {
		return err
	}

	t.ID, err = tx.InsertOne(tx.SQL("timers/insert"), t.APIID, t.AccountID,
		t.JobID, t.APIID,
		t.Name, t.Once, t.TimeZone,
		t.Minute, t.Hour, t.DayOfMonth, t.Month, t.DayOfWeek,
		t.Next, parameters, data)
	if err != nil {
		return err
	}
	return tx.Notify("timers", t.AccountID, t.UserID, t.APIID, 0, t.ID, apsql.Insert)
}

func (t *Timer) Update(tx *apsql.Tx) error {
	err := t.Schedule(tx)
	if err != nil {
		return err
	}

	parameters, err := marshaledForStorage(t.Parameters)
	if err != nil {
		return err
	}

	data, err := marshaledForStorage(t.Data)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(tx.SQL("timers/update"), t.APIID, t.AccountID,
		t.JobID, t.APIID,
		t.Name, t.Once, t.TimeZone,
		t.Minute, t.Hour, t.DayOfMonth, t.Month, t.DayOfWeek,
		t.Next, parameters, data,
		t.ID, t.APIID, t.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("timers", t.AccountID, t.UserID, t.APIID, 0, t.ID, apsql.Update)
}

func (t *Timer) UpdateTime(tx *apsql.Tx, now time.Time) error {
	t.ScheduleTime(now)

	parameters, err := marshaledForStorage(t.Parameters)
	if err != nil {
		return err
	}

	data, err := marshaledForStorage(t.Data)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(tx.SQL("timers/update"), t.APIID, t.AccountID,
		t.JobID, t.APIID,
		t.Name, t.Once, t.TimeZone,
		t.Minute, t.Hour, t.DayOfMonth, t.Month, t.DayOfWeek,
		t.Next, parameters, data,
		t.ID, t.APIID, t.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("timers", t.AccountID, t.UserID, t.APIID, 0, t.ID, apsql.Update)
}
