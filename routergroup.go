package gwf

import (
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
)

const AppDefaultRouterGroupName = "app"

// HandlerFunc 自定义handler类型
type HandlerFunc func(c *Context)

// HandlersChain handler链
type HandlersChain []HandlerFunc

// RouterGroup 路由组
type RouterGroup struct {
	name string
	app  *Application
	// Handlers 当前路由组下所有路由公共Handlers(middleware)
	Handlers HandlersChain
	// Routes 路由Map
	Routes        map[string]map[string]RouteInfo
	appNamePrefix string
}

// RouteInfo 路由详细信息
// handlers 包含RouterGroup中Handlers和局部middleware(如果有的话)以及主逻辑handler
type RouteInfo struct {
	handlers HandlersChain
	path     string
	method   string
}

// IRoutes Router接口
type IRoutes interface {
	GET(path string, handlers ...HandlerFunc)
	POST(path string, handlers ...HandlerFunc)
	PUT(path string, handlers ...HandlerFunc)
	PATCH(path string, handlers ...HandlerFunc)
	HEAD(path string, handlers ...HandlerFunc)
	DELETE(path string, handlers ...HandlerFunc)
	OPTIONS(path string, handlers ...HandlerFunc)
}

// NewRouterGroup 初始化
// 默认自动添加Recovery中间件
func NewRouterGroup(app *Application, name string) *RouterGroup {
	return &RouterGroup{app: app, name: name}
}

// EnableAppNameAsPathPrefix 开启appname作为path前缀
func (rg *RouterGroup) EnableAppNameAsPathPrefix() {
	rg.appNamePrefix = "/" + rg.app.config.Name
}

// AddMiddleware 添加中间件，整个路由组中的路由共享中间件
func (rg *RouterGroup) AddMiddleware(handlers ...HandlerFunc) {
	if handlers == nil {
		panic("nil handlers")
	}
	rg.Handlers = append(rg.Handlers, handlers...)
}

// GET method get
func (rg *RouterGroup) GET(path string, handlers ...HandlerFunc) {
	rg.addRoute(http.MethodGet, path, handlers...)
}

// POST method post
func (rg *RouterGroup) POST(path string, handlers ...HandlerFunc) {
	rg.addRoute(http.MethodPost, path, handlers...)
}

// PUT method put
func (rg *RouterGroup) PUT(path string, handlers ...HandlerFunc) {
	rg.addRoute(http.MethodPut, path, handlers...)
}

// PATCH method patch
func (rg *RouterGroup) PATCH(path string, handlers ...HandlerFunc) {
	rg.addRoute(http.MethodPatch, path, handlers...)
}

// HEAD method head
func (rg *RouterGroup) HEAD(path string, handlers ...HandlerFunc) {
	rg.addRoute(http.MethodHead, path, handlers...)
}

// DELETE method delete
func (rg *RouterGroup) DELETE(path string, handlers ...HandlerFunc) {
	rg.addRoute(http.MethodDelete, path, handlers...)
}

// OPTIONS method options
func (rg *RouterGroup) OPTIONS(path string, handlers ...HandlerFunc) {
	rg.addRoute(http.MethodOptions, path, handlers...)
}

func (rg *RouterGroup) addRoute(method, path string, handlers ...HandlerFunc) {

	if path != "/" {
		path = strings.TrimRight(path, "/")
	}
	if path[0] != '/' {
		panic("path must begin with '/'")
	}
	if path == "" {
		panic("invalid path " + path)
	}

	if handlers == nil {
		panic("nil handlers")
	}

	path = rg.getUrlPath(path)
	//rg.app.Logger.Debugf("addroute method: %s, path: %s", method, path)
	if rg.Routes == nil {
		rg.Routes = make(map[string]map[string]RouteInfo)
	}

	if _, ok := rg.Routes[method]; !ok {
		rg.Routes[method] = make(map[string]RouteInfo)
	}

	if _, exists := rg.Routes[method][path]; exists {
		if exists {
			panic("multiple registrations for " + path)
		}
	}

	totalHandlers := rg.combineHandlers(handlers)
	rg.Routes[method][path] = RouteInfo{handlers: totalHandlers, path: path, method: method}
}

func (rg *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(rg.Handlers) + len(handlers)
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, rg.Handlers)
	copy(mergedHandlers[len(rg.Handlers):], handlers)
	return mergedHandlers
}

func (rg *RouterGroup) handleRequest(c *Context) bool {
	path := c.Request.URL.Path
	method := c.Request.Method
	routeInfo := rg.match(method, path)
	if routeInfo.handlers != nil {
		c.handlers = routeInfo.handlers
		c.Next()
		return true
	}
	return false
}

func (rg *RouterGroup) getUrlPath(path string) string {
	if rg.appNamePrefix != "" {
		path = rg.appNamePrefix + path
	}
	return path
}

func (rg *RouterGroup) match(method, path string) (routeInfo RouteInfo) {
	if _, ok := rg.Routes[method]; !ok {
		return
	}
	for k, v := range rg.Routes[method] {
		if pathMatch(k, path) {
			routeInfo = v
			break
		}
	}
	return
}

func pathMatch(routePath, path string) bool {
	return routePath == path
}

func getControllerPath(controllerType reflect.Type) string {
	controllerName := controllerType.Name()
	controllerName = strings.Replace(controllerName, "Controller", "", 1)
	controllerName = strcase.ToSnake(controllerName)
	controllerPath := controllerType.PkgPath()
	p := regexp.MustCompile(`gopkg.babytree-inc.com/[\w]+/[\w]+/pkg/controller/`)
	controllerPath = p.ReplaceAllString(controllerPath, "")
	return fmt.Sprintf("/%s/%s", controllerPath, controllerName)
}

// 什么是动态路由？动态路由是不通过Routable中的方法注册的路由
// 动态路由需要在InitRouter中新增如下代码:
//  app.RegisterDynamicRouter(&api.MobileAskController{})
// 也可以注册到其他的RouterGroup中
// 注意：动态路由会注册GET/POST两种方式的请求
// 为了更好的满足RESTFUL定义，请使用静态路由，使用IRoutes中的方法注册路由
// options参数有以下调用方式:
//  // 注册到指定的方法中
//  RegisterDynamicRouter(&api.MobileAskController{}, []string{"get", "post"})
//  // 只设定path的前缀,前缀必须以/打头
//  RegisterDynamicRouter(&api.MobileAskController{}, []string{}, "/prefix")
func (rg *RouterGroup) RegisterDynamicRouter(controller interface{}, options ...interface{}) {
	if controller == nil {
		panic("controller不能是nil")
	}

	if !isController(controller) {
		panic("controller定义不满足规范")
	}

	if len(options) > 2 {
		panic("options参数错误")
	}

	var methodList = make([]string, 0)
	var pathPrefix = ""

	var getPathPrefixOption = func(i interface{}) {
		if prefix, ok := i.(string); !ok {
			panic("options参数错误，第一个选项是字符串")
		} else {
			if prefix != "" && !strings.HasPrefix(prefix, "/") {
				panic("options参数错误，path前缀必须以/开始")
			}
			pathPrefix = prefix
		}
	}

	var getMethodList = func(i interface{}) {
		if m, ok := i.([]string); !ok {
			panic("options参数错误, 第二个选项是[]string")
		} else if len(m) > 0 {
			methodList = m
		}
	}

	if len(options) == 1 {
		getMethodList(options[0])
	}

	if len(options) == 2 {
		getMethodList(options[1])
		getPathPrefixOption(options[0])
	}

	controllerType := reflect.ValueOf(controller).Elem().Type()
	controllerPointerType := reflect.ValueOf(controller).Type()
	controllerPath := getControllerPath(controllerType)
	for i := 0; i < controllerPointerType.NumMethod(); i++ {
		actionMethod := controllerPointerType.Method(i)
		if actionMethod.PkgPath != "" {
			// 未导出的方法，忽略
			continue
		}

		if !isControllerActionMethod(actionMethod.Func.Interface()) {
			continue
		}
		actionName := actionMethod.Name
		actionName = strings.Replace(actionName, "Action", "", 1)
		actionName = strcase.ToSnake(actionName)

		//是合法的action
		apiPath := pathPrefix + controllerPath + "/" + actionName
		//rg.app.Logger.Debugf("添加动态路由: path:%s controller:%s method:%s", apiPath, controllerType.Name(), actionMethod.Name)
		var action = func(c *Context) {
			callActionMethod(c, actionMethod.Func.Interface())
		}
		if len(methodList) == 0 {
			rg.GET(apiPath, action)
			rg.POST(apiPath, action)
		} else {
			for _, m := range methodList {
				m = strings.ToUpper(m)
				switch m {
				case http.MethodGet:
					rg.GET(apiPath, action)
				case http.MethodPost:
					rg.POST(apiPath, action)
				case http.MethodPut:
					rg.PUT(apiPath, action)
				case http.MethodPatch:
					rg.PATCH(apiPath, action)
				case http.MethodHead:
					rg.HEAD(apiPath, action)
				case http.MethodDelete:
					rg.DELETE(apiPath, action)
				case http.MethodOptions:
					rg.OPTIONS(apiPath, action)
				default:
					panic("未知的http方法")
				}
			}
		}

	}
}

// EnableDebugTemplate打开模板功能
// 在开发环境下是可以调试模板的
func (rg *RouterGroup) EnableTemplate() {
	SetFuncMap(template.FuncMap{
		"now": time.Now,
		"staticFileUrl": func(ctx *Context, url string) string {
			// 开发环境和测试使用/public下的文件
			if GetConfig().IsDevEnvironment() || GetConfig().IsTestEnvironment() {
				return rg.appNamePrefix + url
			}

			//其他环境使用CDN
			return rg.appNamePrefix + url
		},
	})

	if GetConfig().IsDevEnvironment() {
		//开发环境可以debug模板
		EnableDebug()
	}
	LoadTemplate()
}
