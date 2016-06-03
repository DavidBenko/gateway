package sql

import (
	"errors"
	"fmt"

	"gateway/db"

	_ "github.com/SAP/go-hdb/driver"
)

// HanaSpec implements db.Specifier for Hana connection parameters.
type HanaSpec struct {
	spec
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

func (h *HanaSpec) validate() error {
	return validate(h, []validation{
		{kw: "user", errCond: h.User == "", val: h.User},
		{kw: "password", errCond: h.Password == "", val: h.Password},
		{kw: "host", errCond: h.Host == "", val: h.Host},
		{kw: "port", errCond: h.Port < 0, val: h.Port},
	})
}

func (h *HanaSpec) driver() driver {
	return hana
}

func (h *HanaSpec) ConnectionString() string {
	return fmt.Sprintf("hdb://%s:%s@%s:%d",
		h.User,
		h.Password,
		h.Host,
		h.Port,
	)
}

func (h *HanaSpec) UniqueServer() string {
	return h.ConnectionString()
}

func (h *HanaSpec) NewDB() (db.DB, error) {
	return newDB(h)
}

// UpdateWith validates `hSpec` and updates `h` with its contents if it is
// valid.
func (h *HanaSpec) UpdateWith(hSpec *HanaSpec) error {
	if hSpec == nil {
		return errors.New("cannot update a HanaSpec with a nil Specifier")
	}
	if err := hSpec.validate(); err != nil {
		return err
	}
	*h = *hSpec
	return nil
}
