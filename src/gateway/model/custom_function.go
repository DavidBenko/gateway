package model

import (
	"bytes"
	"errors"
	"fmt"

	"gateway/docker"
	aperrors "gateway/errors"
	apsql "gateway/sql"

	dockerclient "github.com/fsouza/go-dockerclient"
)

const (
	CustomFunctionLanguageJava   = "java"
	CustomFunctionLanguageNode   = "node"
	CustomFunctionLanguageCSharp = "csharp"
	CustomFunctionLanguagePython = "python"
	CustomFunctionLanguagePHP    = "php"
	CustomFunctionLanguageOther  = "other"
)

type CustomFunction struct {
	AccountID int64 `json:"-"`
	UserID    int64 `json:"-"`
	APIID     int64 `json:"api_id,omitempty" db:"api_id" path:"apiID"`

	ID          int64  `json:"id,omitempty" path:"id"`
	Name        string `json:"name"`
	Language    string `json:"language"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	Memory      int64  `json:"memory"`
	CPUShares   int64  `json:"cpu_shares" db:"cpu_shares"`
	Timeout     int64  `json:"timeout"`
}

func (c *CustomFunction) ImageName() string {
	return fmt.Sprintf("%v_%v/%v", c.AccountID, c.APIID, c.ID)
}

func (c *CustomFunction) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if c.Name == "" {
		errors.Add("name", "must have a name")
	}
	if c.Memory < 8 {
		errors.Add("memory", "minimum memory limit allowed is 8MB")
	}
	if c.CPUShares < 0 {
		errors.Add("cpu_shares", "the minimum allowed cpu-shares is 0")
	}
	if c.Timeout < 0 || c.Timeout > 600 {
		errors.Add("timeout", "timeout should be between 0 and 600 seconds")
	}
	if isInsert {
		switch c.Language {
		case CustomFunctionLanguageJava:
		case CustomFunctionLanguageNode:
		case CustomFunctionLanguageCSharp:
		case CustomFunctionLanguagePython:
		case CustomFunctionLanguagePHP:
		case CustomFunctionLanguageOther:
		default:
			errors.Add("language", "invalid language")
		}
	}
	return errors
}

func (c *CustomFunction) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "custom_functions", "api_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

func (c *CustomFunction) All(db *apsql.DB) ([]*CustomFunction, error) {
	functions := []*CustomFunction{}
	err := db.Select(&functions, db.SQL("custom_functions/all"), c.APIID, c.AccountID)
	for _, function := range functions {
		function.AccountID = c.AccountID
		function.UserID = c.UserID
	}
	return functions, err
}

func (c *CustomFunction) Find(db *apsql.DB) (*CustomFunction, error) {
	function := CustomFunction{
		AccountID: c.AccountID,
		UserID:    c.UserID,
	}
	var err error
	if c.ID > 0 {
		err = db.Get(&function, db.SQL("custom_functions/find"), c.ID, c.APIID, c.AccountID)
	} else {
		err = db.Get(&function, db.SQL("custom_functions/find_name"), c.Name, c.APIID, c.AccountID)
	}
	return &function, err
}

func (c *CustomFunction) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("custom_functions/delete"), c.ID, c.APIID, c.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("custom_functions", c.AccountID, c.UserID, c.APIID, 0, c.ID, apsql.Delete)
}

func (c *CustomFunction) Insert(tx *apsql.Tx) error {
	var err error
	c.ID, err = tx.InsertOne(tx.SQL("custom_functions/insert"),
		c.APIID, c.AccountID, c.Name, c.Description, c.Active, c.Memory, c.CPUShares,
		c.Timeout)
	if err != nil {
		return err
	}

	return tx.Notify("custom_functions", c.AccountID, c.UserID, c.APIID, 0, c.ID, apsql.Insert)
}

func (c *CustomFunction) AfterInsert(tx *apsql.Tx) error {
	if c.Language == CustomFunctionLanguageOther {
		return nil
	}

	dir := "custom_function/" + c.Language
	files, err := AssetDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		file := CustomFunctionFile{
			AccountID:        c.AccountID,
			UserID:           c.UserID,
			APIID:            c.APIID,
			CustomFunctionID: c.ID,
			Name:             f,
			Body:             string(MustAsset(dir + "/" + f)),
		}
		err = file.Insert(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CustomFunction) Update(tx *apsql.Tx) error {
	err := tx.UpdateOne(tx.SQL("custom_functions/update"),
		c.Name, c.Description, c.Active, c.Memory, c.CPUShares, c.Timeout,
		c.ID, c.APIID, c.AccountID)
	if err != nil {
		return err
	}

	return tx.Notify("custom_functions", c.AccountID, c.UserID, c.APIID, 0, c.ID, apsql.Update)
}

func ExecuteCustomFunction(db *apsql.DB, accountID, apiID, customFunctionID int64,
	name string, input interface{}, checkActive bool) (*docker.RunOutput, error) {
	function := &CustomFunction{
		AccountID: accountID,
		APIID:     apiID,
		ID:        customFunctionID,
		Name:      name,
	}
	function, err := function.Find(db)
	if err != nil {
		return nil, err
	}

	if checkActive && !function.Active {
		return nil, errors.New("Custom function is not active")
	}

	stale := false

	file := CustomFunctionFile{
		AccountID:        function.AccountID,
		APIID:            function.APIID,
		CustomFunctionID: function.ID,
	}
	files, err := file.All(db)
	if err != nil {
		return nil, err
	}

	image, err := docker.InspectImage(function.ImageName())
	if err == dockerclient.ErrNoSuchImage {
		stale = true
	} else if err != nil {
		return nil, err
	} else {
		for _, file := range files {
			updated := file.UpdatedAt
			if updated == nil {
				updated = file.CreatedAt
			}
			if updated.After(image.Created) {
				stale = true
				break
			}
		}
	}

	if stale {
		input, err := files.Tar()
		if err != nil {
			return nil, err
		}

		output := &bytes.Buffer{}
		options := dockerclient.BuildImageOptions{
			Name:         function.ImageName(),
			NoCache:      true,
			InputStream:  input,
			OutputStream: output,
		}

		docker.BuildImage(options)
	}

	return docker.ExecuteImage(function.ImageName(), function.Memory,
		function.CPUShares, function.Timeout, input)
}
