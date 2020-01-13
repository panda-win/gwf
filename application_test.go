package gwf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var panicMsg = "This is a panic."

var testAppConfig = `
appName: gwf
listen: :8888
appVersion: 0.0.0

`

func TestAppServer(t *testing.T) {
	initTestAppConfig(t)
	app := GetApplication()
	app.GET("/", func(c *Context) {
		c.Success(nil)
	})
	app.addHealthProfiling()
	app.addPprof()
	assert.Equal(t, len(app.Routes), 1)
	assert.Equal(t, len(app.otherRouterGroups), 2)
	assert.Equal(t, app.otherRouterGroups[0].name, "health_profiling_for_docker")
	assert.Equal(t, app.otherRouterGroups[1].name, "pprof")

	ts := createTestServer(app.ServeHTTP)

	r, _ := http.NewRequest("GET", ts.URL+"/", nil)
	b, _ := req(t, r)
	var res DefaultJson
	err := json.Unmarshal(b, &res)
	if err != nil {
		t.Fatalf("json unmarshal err:%v", err)
	}
	assert.Equal(t, res.Status, SUCCESS)

	r, _ = http.NewRequest("GET", ts.URL+"/healthz", nil)
	b, _ = req(t, r)
	assert.Equal(t, "200", string(b))

}

func TestNonLogin(t *testing.T) {
	ts := createAppServer(t)
	request, _ := http.NewRequest("GET", ts.URL+"/nonLogin", nil)
	b, _ := req(t, request)
	var res DefaultJson
	err := json.Unmarshal(b, &res)
	if err != nil {
		t.Fatalf("json unmarshal err: %v", err)
	}
	assert.Equal(t, res.Status, NONE_LOGIN)
}

func TestSuccess(t *testing.T) {
	ts := createAppServer(t)
	request, _ := http.NewRequest("GET", ts.URL+"/success", nil)
	b, _ := req(t, request)
	var res DefaultJson
	err := json.Unmarshal(b, &res)
	if err != nil {
		t.Fatalf("json unmarshal err: %v", err)
	}
	assert.Equal(t, res.Status, SUCCESS)
}

func TestFail(t *testing.T) {
	ts := createAppServer(t)
	request, _ := http.NewRequest("GET", ts.URL+"/fail", nil)
	b, _ := req(t, request)
	var res DefaultJson
	err := json.Unmarshal(b, &res)
	if err != nil {
		t.Fatalf("json unmarshal err: %v", err)
	}
	assert.Equal(t, res.Status, FAIL)
	assert.Equal(t, res.Message, "fail req")

}

func TestError(t *testing.T) {
	ts := createAppServer(t)
	request, _ := http.NewRequest("GET", ts.URL+"/error", nil)
	b, _ := req(t, request)
	var res DefaultJson
	err := json.Unmarshal(b, &res)
	if err != nil {
		t.Fatalf("json unmarshal err: %v", err)
	}
	assert.Equal(t, res.Status, FAIL)
	assert.Equal(t, res.Message, "err req")

}

func TestNotFound(t *testing.T) {
	ts := createAppServer(t)
	request, _ := http.NewRequest("GET", ts.URL+"/notfound", nil)
	b, _ := req(t, request)
	assert.Equal(t, "资源不存在", string(b))
}

func createAppServer(t *testing.T) *httptest.Server {

	ts := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		rg := NewRouterGroup(nil, "rg_test")
		rg.GET("/nonLogin", func(c *Context) {
			c.NotLogin()
		})

		rg.GET("/success", func(c *Context) {
			c.Success(nil)
		})

		rg.GET("/fail", func(c *Context) {
			c.Fail(nil, "fail req")
		})

		rg.GET("/error", func(c *Context) {
			c.Error(fmt.Errorf("err req"))
		})

		rg.GET("/auth", ApiAuth(), func(c *Context) {
			c.Writer.Write([]byte("success"))
		})

		rg.GET("/not_recovery", func(c *Context) {
			panic(panicMsg)
		})

		rg.GET("/recovery", Recovery(), func(c *Context) {
			panic(panicMsg)
		})

		rg.GET("/format_response", FormatJson(), func(c *Context) {
			res := map[string]interface{}{
				"msg": nil,
				"list": []interface{}{
					12,
					true,
					1.3,
					nil,
				},
			}
			b, _ := json.Marshal(res)
			c.Writer.Write(b)
		})

		ctx := newCtx(nil, r)
		ctx.Writer = NewResponseWriter(w, nil, nil, nil)
		if r.URL.Path == "/not_recovery" {
			assert.PanicsWithValue(t, panicMsg, func() { rg.handleRequest(ctx) })
		} else {
			if !rg.handleRequest(ctx) {
				DefaultNotFoundHandler(ctx)
			}
		}
	})
	return ts
}

func createTestServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func req(t *testing.T, r *http.Request) (b []byte, err error) {
	hc := http.Client{Timeout: 3 * time.Second}
	resp, err := hc.Do(r)
	if err != nil {
		t.Fatalf("http request err: %v", err)
	}
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read resp body err: %v", err)
	}
	return
}

func initTestAppConfig(t *testing.T) {
	rootPath, err := getDeployRootPath(true)
	if err != nil {
		t.Fatal(err)
	}
	confFilename := fmt.Sprintf("%s/config/app.yaml", rootPath)
	err := os.MkdirAll(path.Dir(confFilename), 0777)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(configFilename, []byte(testAppConfig), 0777)
	if err != nil {
		t.Error(err)
	}
}
