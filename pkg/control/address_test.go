package control

import (
	"net/http"
	"testing"
)

func TestAddressAPI(t *testing.T) {
	add := &Address{
		host:      "test.com",
		root:      "v2",
		namespace: "test",
		term:      "term",
		reference: "testreference",
	}

	a, err := add.API(http.MethodHead, http.MethodGet)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if a.String() != "api://test.com/v2/term?method=HEAD&method=GET" {
		t.Error("unexpected value", a.String())
		t.Fail()
	}
}
