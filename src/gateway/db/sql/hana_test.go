package sql_test

import (
	"encoding/json"

	sql "gateway/db/sql"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func hanaConfigs() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"simple": map[string]interface{}{
			"user":     "SYSTEM",
			"password": "pass",
			"host":     "hanaserver",
			"port":     30015,
		},
		"bad": map[string]interface{}{
			"user": "SYSTEM",
			"port": 30015,
		},
		"badport": map[string]interface{}{
			"user":     "SYSTEM",
			"password": "pass",
			"port":     -1234,
		},
	}
}

func hanaSpecs() map[string]*sql.HanaSpec {
	specs := make(map[string]*sql.HanaSpec)
	for name, spc := range hanaConfigs() {
		hanaSpec := &sql.HanaSpec{}
		vals, err := json.Marshal(spc)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(vals, hanaSpec)

		specs[name] = hanaSpec
	}
	return specs
}

func (s *SQLSuite) TestHanaConfig(c *gc.C) {
	for i, t := range []struct {
		should       string
		given        *sql.HanaSpec
		expectString string
		expectUnique string
		expectError  string
	}{{
		should:       "work with a simple config",
		given:        hanaSpecs()["simple"],
		expectString: "hdb://SYSTEM:pass@hanaserver:30015",
		expectUnique: "hdb://SYSTEM:pass@hanaserver:30015",
	}, {
		should:      "not work without a host and password",
		given:       hanaSpecs()["bad"],
		expectError: `hdb config errors: bad value "" for "password"; bad value "" for "host"`,
	}, {
		should:      "not work with a bad config",
		given:       hanaSpecs()["badport"],
		expectError: `hdb config errors: bad value "" for "host"; bad value -1234 for "port"`,
	}} {
		c.Logf("Test %d: should %s", i, t.should)

		obtained, err := sql.Config(sql.Connection(t.given))
		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			continue
		}
		c.Assert(err, jc.ErrorIsNil)
		c.Check(obtained.ConnectionString(), gc.Equals, t.expectString)
		c.Check(obtained.UniqueServer(), gc.Equals, t.expectUnique)
	}
}
