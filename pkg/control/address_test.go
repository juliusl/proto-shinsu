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

	test, err := EmptyAddress().FromString("https://localhost/library/ubuntu#blob/uploads")
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	u, err := test.URI()
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if u.String() != "https://empty://?label=reference@localhost/library/ubuntu#blob/uploads" {
		t.Error("unexpected uri", u.String())
		t.Fail()
	}

	ep, err := test.GetEmptyParameters()
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	for _, e := range ep {
		q := e.Query()

		if !q.Has("label") {
			t.Error("expected empty parameter to have a label")
			t.Fail()
		}
	}
}
