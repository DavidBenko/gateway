package admin

import "testing"

var pe = proxyEndpoint{}

func TestName(t *testing.T) {
	if pe.Name() != "proxy_endpoints" {
		t.Error("Expected name to be 'proxy_endpoints'")
	}
}
