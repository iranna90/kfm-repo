package testing

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"fmt"
)

func TestServer(t *testing.T) {
	var status = http.StatusNotFound;
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(int(status))
	}))

	res, err := http.Head(ts.URL)

	if err != nil {
		t.Log("Erro whil calling url ", ts.URL)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("Excpected 404 but it is not")
	}

	status = 200
	res, err = http.Head(ts.URL)

	if res.StatusCode != http.StatusOK {
		t.Error("Expected status code 200 but response has status code of ", res.StatusCode)
	}
}

func TestGetValueStubbing(t *testing.T) {
	result := checking(2, 3)

	// default get value 2
	if result != 7 {
		t.Error("excpted output is 7 but real output is ", result)
	}

	// stub the value
	bringValue = func() int {
		fmt.Println("returning frm stub value")
		return 10
	}

	result = checking(2, 3)

	if result != 15 {
		t.Error("Expected is 15 but real is ", result)
	}

}
