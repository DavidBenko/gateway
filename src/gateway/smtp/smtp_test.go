package smtp_test

import (
	"gateway/smtp"
	"testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type SmtpSuite struct{}

var _ = gc.Suite(&SmtpSuite{})

func smtpConfigs() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"simple": map[string]interface{}{
			"user":     "admin",
			"password": "abc123",
			"host":     "mail.server.com",
			"port":     80,
			"sender":   "admin@mail.server.com",
		},
		"no-user": map[string]interface{}{
			"user":     "",
			"password": "abc123",
			"host":     "mail.server.com",
			"port":     80,
			"sender":   "admin@mail.server.com",
		},
		"no-password": map[string]interface{}{
			"user":     "admin",
			"password": "",
			"host":     "mail.server.com",
			"port":     80,
			"sender":   "admin@mail.server.com",
		},
	}
}

func toSpec(s map[string]interface{}) *smtp.Spec {
	spec := &smtp.Spec{Host: s["host"].(string), Port: s["port"].(int), User: s["user"].(string), Password: s["password"].(string), Sender: s["sender"].(string)}
	spec.CreateAuth()
	return spec
}

func (s *SmtpSuite) TestNewServer(c *gc.C) {
	for i, t := range []struct {
		should      string
		given       map[string]interface{}
		expectError string
	}{{
		should: "work with a simple config",
		given:  smtpConfigs()["simple"],
	}} {
		c.Logf("Test %d: should %s", i, t.should)

		spec := toSpec(t.given)

		c.Check(spec, gc.NotNil)
	}
}

func (s *SmtpSuite) TestConnectionString(c *gc.C) {
	for i, t := range []struct {
		should string
		given  map[string]interface{}
		expect string
	}{{
		should: "return expected connection string",
		given:  smtpConfigs()["simple"],
		expect: "{\"host\":\"mail.server.com\",\"port\":80,\"user\":\"admin\",\"password\":\"abc123\",\"sender\":\"admin@mail.server.com\",\"Auth\":{}}",
	}} {
		c.Logf("Test %d: should %s", i, t.should)

		spec := toSpec(t.given)

		c.Check(spec.ConnectionString(), gc.Equals, t.expect)
	}
}

func (s *SmtpSuite) TestSmtpPool(c *gc.C) {
	for i, t := range []struct {
		should      string
		given       map[string]interface{}
		pool        *smtp.SmtpPool
		expectError string
	}{{
		should: "return correct connection",
		pool:   smtp.NewSmtpPool(),
		given:  smtpConfigs()["simple"],
	}} {
		c.Logf("Test %d: should %s", i, t.should)

		spec := toSpec(t.given)

		connection, err := t.pool.Connection(spec)

		c.Assert(err, jc.ErrorIsNil)

		// Should return the supplied spec
		c.Assert(connection, jc.DeepEquals, spec)
	}
}
