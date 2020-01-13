// gwf 是一个精简的mvc框架
// 1、支持静态路由和动态路由，其中动态路由包含反射自某一controller中的所有action
// 2、支持pulic目录下静态文件路由
// 3、支持middleware
// 4、支持context
// 5、支持平滑重启
// 6、支持template快速开发
// 7、集成健康探针与pprof监控
package gwf

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// 对于上传数据的最大内存设置，如果超过该值则采用临时文件存储
const defaultMultipartMemory = 16 << 20 // 16 MB

// 静态文件目录
const staticFileDir = "public"

type applicationConfig struct {
	// Addr 服务地址
	Addr string
	// Name 应用名称
	Name string
	// Version 应用版本
	Version string
	// RestartTimeout 平滑重启的超时时间
	RestartTimeout time.Duration
}

// Application 是应用的抽象
type Application struct {
	config *applicationConfig
	// Logger 日志
	Logger *log.Logger

	// RouterGroup app级别的RouterGroup
	*RouterGroup

	//其他的RouterGroup
	otherRouterGroups []*RouterGroup

	// 404处理handler
	notFound HandlerFunc

	// 客户端上传数据的最大内存占用量
	maxMultipartMemory int64

	//是否开启静态资源路由
	enableStaticFileServer bool
	fileServer             http.Handler
}

var application *Application
var applicationOnce sync.Once

// 初始化
func New() *Application {
	applicationOnce.Do(func() {
		application = createApplication()
	})
	return application
}

func createApplication() *Application {
	logger := log.New(os.Stdout, "gwf: ", log.Lshortfile)
	appConfig := &applicationConfig{
		Addr:           GetConfig().Listen,
		Name:           GetConfig().AppName,
		Version:        GetConfig().AppVersion,
		RestartTimeout: 5 * time.Second,
	}

	app := &Application{
		config:             appConfig,
		Logger:             logger,
		notFound:           DefaultNotFoundHandler,
		maxMultipartMemory: defaultMultipartMemory,
	}
	app.RouterGroup = NewRouterGroup(app, APP_DEFAULT_ROUTER_GROUP_NAME)

	return app
}

// SetRestartTimeout 设置平滑重启超时时间
func (app *Application) SetRestartTimeout(timeout time.Duration) {
	app.config.RestartTimeout = timeout
}

// SetNotFound 设置404的处理器
func (app *Application) SetNotFound(handler HandlerFunc) {
	app.notFound = handler
}

// SetMaxMultipartMemory 设置客户端上传数据的最大内存占用量
func (app *Application) SetMaxMultipartMemory(n int64) {
	app.maxMultipartMemory = n
}

// ServeHttp实现了http.Handler接口
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	st := time.Now()
	context := newCtx(app, r)

	writer := NewResponseWriter(w, ResponseStatusHandler(func(status int) int {
		costTime := time.Since(st).Milliseconds()
		if status == http.StatusOK {
			return http.StatusOK
		} else {
			app.Logger.Printf("异常http状态码记录 status code:%d cost time(ms):%d url:%s", status, costTime, r.URL.Path)
		}
		return status
	}), nil, nil)

	context.Writer = writer

	if app.handleRequest(context) {
		return
	} else {
		for _, rg := range app.otherRouterGroups {
			if rg.handleRequest(context) {
				return
			}
		}
	}
	//静态资源
	if app.enableStaticFileServer && strings.HasPrefix(r.URL.Path, app.appNamePrefix+"/public/") {
		r2 := copyRequest(r)
		app.fileServer.ServeHTTP(newW, r2)
		return
	}
	app.notFound(context)
}

// AddRouterGroup 添加额外的路由组，app初始化时会默认添加app级别路由组
func (app *Application) AddRouterGroup(rg *RouterGroup) {
	app.otherRouterGroups = append(app.otherRouterGroups, rg)
}

// EnableStaticFileServer 开启静态文件服务器，可以基于public目录提供静态文件服务
func (app *Application) EnableStaticFileServer() {
	app.enableStaticFileServer = true
	root := fmt.Sprintf("%s/%s", rootPath, staticFileDir)
	app.fileServer = http.FileServer(http.Dir(root))
}

// Start启动App, 此方法会监听配置的端口，提供服务
// 此方法会阻塞主协程
// 路由设置一定要在此方法之前设定，否则不生效
func (app *Application) Start() {
	app.Logger.Printf("start app %s %s at %s ...", app.config.Name, app.config.Version, app.config.Addr)

	app.addHealthProfiling()

	app.addPprof()

	server := &http.Server{
		Addr:     app.config.Addr,
		Handler:  http.HandlerFunc(app.ServeHTTP),
		ErrorLog: app.Logger,
	}

	err := RunGrace(server, app.config.RestartTimeout)
	if err != nil {
		app.Logger.Printf("app异常退出 err:%s", err)
	}
}

// 增加监控探针
func (app *Application) addHealthProfiling() {
	rg := NewRouterGroup(app, "health_profiling_for_app")
	rg.GET("/healthz", func(c *Context) {
		c.String(200, "200")
	})
	app.AddRouterGroup(rg)
}

// 增加pprof监控
func (app *Application) addPprof() {
	rg := NewRouterGroup(app, "pprof")
	rg.EnableAppNameAsPathPrefix()
	rg.GET("/internal/debug/pprof/allocs", func(c *Context) {
		r2 := copyRequest(c.Request)
		r2.URL.Path = "/debug/pprof/allocs"
		pprof.Index(c.Writer, r2)
	})
	rg.GET("/internal/debug/pprof/heap", func(c *Context) {
		r2 := copyRequest(c.Request)
		r2.URL.Path = "/debug/pprof/heap"
		pprof.Index(c.Writer, r2)
	})
	rg.GET("/internal/debug/pprof/goroutine", func(c *Context) {
		r2 := copyRequest(c.Request)
		r2.URL.Path = "/debug/pprof/goroutine"
		pprof.Index(c.Writer, r2)
	})
	rg.GET("/internal/debug/pprof/mutex", func(c *Context) {
		r2 := copyRequest(c.Request)
		r2.URL.Path = "/debug/pprof/mutex"
		pprof.Index(c.Writer, r2)
	})
	rg.GET("/internal/debug/pprof/profile", func(c *Context) {
		pprof.Profile(c.Writer, c.Request)
	})
	app.AddRouterGroup(rg)
}

// 复制http.Request，来源于net/http包的StripPrefix方法
func copyRequest(r *http.Request) *http.Request {
	r2 := new(http.Request)
	*r2 = *r
	r2.URL = new(url.URL)
	*r2.URL = *r.URL
	return r2
}
