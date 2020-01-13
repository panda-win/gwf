package gwf

import (
	"net/http"
	"testing"
)

func TestNotRecovery(t *testing.T) {
	initTestAppConfig(t)
	ts := createAppServer(t)
	request, _ := http.NewRequest("GET", ts.URL+"/not_recovery", nil)
	req(t, request)
}

func TestRecovery(t *testing.T) {
	initTestAppConfig(t)
	ts := createAppServer(t)
	request, _ := http.NewRequest("GET", ts.URL+"/recovery", nil)
	req(t, request)
}
