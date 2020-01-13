package gwf

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddRoute(t *testing.T) {
	rg := NewRouterGroup(nil, "testrg")
	rg.GET("/", func(_ *Context) {})
	assert.Equal(t, len(rg.Routes), 1)
	assert.Equal(t, len(rg.Routes["GET"]), 1)
	assert.Empty(t, len(rg.Routes["POST"]))

	rg.POST("/", func(_ *Context) {})
	assert.Equal(t, len(rg.Routes), 2)
	assert.Equal(t, len(rg.Routes["POST"]), 1)
	assert.NotEmpty(t, len(rg.Routes["POST"]))

	rg.POST("/post", func(_ *Context) {})
	assert.Equal(t, len(rg.Routes["POST"]), 2)

	rg.PATCH("/patch", func(_ *Context) {})
	rg.HEAD("/head", func(_ *Context) {})
	rg.OPTIONS("/options", func(_ *Context) {})
	rg.DELETE("/delete", func(_ *Context) {})
}

func TestAddRouteFail(t *testing.T) {
	rg := NewRouterGroup(nil, "testrgfail")
	assert.Panics(t, func() {
		rg.addRoute("GET", "", func(_ *Context) {})
	})
	assert.Panics(t, func() {
		rg.addRoute("GET", "a", func(_ *Context) {})
	})

	rg.addRoute("POST", "/post", func(_ *Context) {})

	assert.Panics(t, func() {
		rg.addRoute("POST", "/post/", func(_ *Context) {})
	})
}

func TestAddMiddleware(t *testing.T) {
	rg := NewRouterGroup(nil, "testrgmiddleware")
	middleware0 := func(_ *Context) {}
	rg.AddMiddleware(middleware0)
	assert.Equal(t, len(rg.Handlers), 1)
	assert.Empty(t, len(rg.Routes))
	rg.addRoute("GET", "/", func(_ *Context) {})
	assert.Equal(t, len(rg.Routes["GET"]["/"].handlers), 2)
	compareFunc(t, middleware0, rg.Routes["GET"]["/"].handlers[0])
}

func TestEnableAppNameAsPathPrefix(t *testing.T) {
	appConfig := &applicationConfig{
		Addr:           "127.0.0.1",
		Name:           "goapptest",
		Version:        "v0.0.1",
		RestartTimeout: 5 * time.Second,
	}
	app := &Application{
		config: appConfig,
	}
	rg := NewRouterGroup(app, "testrgmiddleware")
	rg.EnableAppNameAsPathPrefix()
	assert.Equal(t, rg.appNamePrefix, "/goapptest")
}

type testController struct {
	*Controller
}

func (cc *testController) MyAction(c *Context) {
	c.Success(nil)
}

func TestRegisterDynamicRouter(t *testing.T) {
	rg := NewRouterGroup(nil, "test_control")
	rg.RegisterDynamicRouter(&testController{})
	curPkgPath := "/github.com/panda-win/gwf"
	_, ok1 := rg.Routes["GET"][curPkgPath+"/test/my"]
	_, ok2 := rg.Routes["POST"][curPkgPath+"/test/my"]
	assert.Equal(t, ok1, true)
	assert.Equal(t, ok2, true)
}

func TestRegisterDynamicRouterWithOptions(t *testing.T) {
	rg := NewRouterGroup(nil, "test_control")
	rg.RegisterDynamicRouter(&testController{}, []string{"GET"})
	curPkgPath := "/github.com/panda-win/gwf"
	_, ok1 := rg.Routes["GET"][curPkgPath+"/test/my"]
	_, ok2 := rg.Routes["POST"][curPkgPath+"/test/my"]
	assert.Equal(t, ok1, true)
	assert.Equal(t, ok2, false)
}

func compareFunc(t *testing.T, a, b interface{}) {
	sf1 := reflect.ValueOf(a)
	sf2 := reflect.ValueOf(b)
	if sf1.Pointer() != sf2.Pointer() {
		t.Error("different functions")
	}
}
