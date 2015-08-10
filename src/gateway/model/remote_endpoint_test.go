package model_test

import (
	"encoding/json"
	"testing"

	"github.com/jmoiron/sqlx/types"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"gateway/db"
	sqls "gateway/db/sqlserver"
	"gateway/model"
	re "gateway/model/remote_endpoint"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type RemoteEndpointSuite struct{}

var _ = gc.Suite(&RemoteEndpointSuite{})

func data() map[string]interface{} {
	return map[string]interface{}{
		"simple": map[string]interface{}{
			"config": map[string]interface{}{
				"server":   "some.url.net",
				"port":     1234,
				"user id":  "user",
				"password": "pass",
				"database": "db",
				"schema":   "dbschema",
			},
		},
		"complicated": map[string]interface{}{
			"config": map[string]interface{}{
				"server":             "some.url.net",
				"port":               1234,
				"user id":            "user",
				"password":           "pass",
				"database":           "db",
				"schema":             "dbschema",
				"connection timeout": 30,
			},
			"maxOpenConn": 80,
			"maxIdleConn": 100,
		},
		"badConfig": map[string]interface{}{
			"config": map[string]interface{}{
				"server": "some.url.net",
			},
		},
		"badConfigType": map[string]interface{}{
			"config": 8,
		},
		"badMaxIdleType": map[string]interface{}{
			"config": map[string]interface{}{
				"server":   "some.url.net",
				"port":     1234,
				"user id":  "user",
				"password": "pass",
				"database": "db",
				"schema":   "dbschema",
			},
			"maxOpenConn": "hello",
		},
	}
}

func specs() map[string]db.Specifier {
	specs := make(map[string]db.Specifier)
	for _, which := range []struct {
		name string
		kind string
	}{{
		"simple", model.RemoteEndpointTypeSQLServer,
	}, {
		"complicated", model.RemoteEndpointTypeSQLServer,
	}} {
		switch which.kind {
		case model.RemoteEndpointTypeSQLServer:
			d := data()[which.name].(map[string]interface{})
			js, err := json.Marshal(d)
			if err != nil {
				panic(err)
			}
			var conf re.SQLServer
			err = json.Unmarshal(js, &conf)
			if err != nil {
				panic(err)
			}
			s, err := sqls.Config(
				sqls.Connection(conf.Config),
				sqls.MaxOpenIdle(conf.MaxOpenConn, conf.MaxIdleConn),
			)
			if err != nil {
				panic(err)
			}
			specs[which.name] = s
		default:
		}
	}
	return specs
}

func (s *RemoteEndpointSuite) TestDBConfig(c *gc.C) {
	for i, t := range []struct {
		should      string
		givenConfig string
		givenType   string
		expectSpec  string
		expectError string
	}{{
		should:      "work with a simple config",
		givenConfig: "simple",
		givenType:   model.RemoteEndpointTypeSQLServer,
		expectSpec:  "simple",
	}, {
		should:      "work with a complex config",
		givenConfig: "complicated",
		givenType:   model.RemoteEndpointTypeSQLServer,
		expectSpec:  "complicated",
	}, {
		should:      "fail with a bad config",
		givenConfig: "badConfig",
		givenType:   model.RemoteEndpointTypeSQLServer,
		expectError: `SQL Config missing "port" key`,
	}, {
		should:      "fail with a bad config type",
		givenConfig: "badConfigType",
		givenType:   model.RemoteEndpointTypeSQLServer,
		expectError: `bad JSON for SQL Server config: json: cannot unmarshal number into Go value of type sqlserver.Conn`,
	}, {
		should:      "fail with a bad max idle type",
		givenConfig: "badMaxIdleType",
		givenType:   model.RemoteEndpointTypeSQLServer,
		expectError: `bad JSON for SQL Server config: json: cannot unmarshal string into Go value of type int`,
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		data := data()[t.givenConfig]
		dataJSON, err := json.Marshal(data)
		endpoint := &model.RemoteEndpoint{
			Type: t.givenType,
			Data: types.JsonText(json.RawMessage(dataJSON)),
		}
		spec, err := endpoint.DBConfig()
		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			continue
		}
		c.Assert(err, jc.ErrorIsNil)
		expectSpec := specs()[t.expectSpec]
		c.Check(spec, jc.DeepEquals, expectSpec)
	}
}
