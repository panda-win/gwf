package gwf

import (
	"log"
	"reflect"
	"runtime"
	"strings"
)

const INIT_METHOD_NAME = "Init"

const (
	controllerTypeApi = iota
	controllerTypeAdmin
)

// IController Controller 接口
type IController interface {
	Init()
	GetApplication() *Application
}

// Controller，api的controller使用这个
type Controller struct {
	*log.Logger
	app *Application
}

// Init Controller 初始化
func (c *Controller) Init() {
	c.app = GetApplication()
	c.Logger = c.app.Logger
}

// GetApplication 获取app数据
func (c *Controller) GetApplication() *Application {
	return c.app
}

// AdminController 后台专用controller
type AdminController struct {
	*log.Logger
	app *Application
}

// Init AdminController 初始化
func (c *AdminController) Init() {
	c.app = GetApplication()
	c.Logger = c.app.Logger
}

// GetApplication 获取app数据
func (c *AdminController) GetApplication() *Application {
	return c.app
}

// 判断是否是Controller
func isController(controller interface{}) bool {
	if controller == nil {
		return false
	}

	if reflect.TypeOf(controller).Kind() != reflect.Ptr {
		return false
	}

	controllerType := reflect.ValueOf(controller).Elem().Type()

	//是否embeding了*gwf.Controller或者*gwf.AdminController
	baseController := reflect.ValueOf(controller).Elem().FieldByName("Controller")
	if !baseController.IsValid() || baseController.Kind() != reflect.Ptr ||
		baseController.Type().Elem().Name() != "Controller" {

		baseController := reflect.ValueOf(controller).Elem().FieldByName("AdminController")
		if !baseController.IsValid() || baseController.Kind() != reflect.Ptr ||
			baseController.Type().Elem().Name() != "AdminController" {
			return false
		}
	}

	//是否是XXXController
	if !strings.HasSuffix(controllerType.Name(), "Controller") {
		return false
	}

	return true
}

func isControllerActionMethod(actionMethod interface{}) bool {
	actionFuncType := reflect.TypeOf(actionMethod)
	actionMethodName := runtime.FuncForPC(reflect.ValueOf(actionMethod).Pointer()).Name()
	if !strings.HasSuffix(actionMethodName, "Action") {
		return false
	}
	if actionFuncType.NumIn() != 2 {
		return false
	}

	param := actionFuncType.In(1)
	k := param.Kind()
	if k != reflect.Ptr {
		return false
	}

	param = param.Elem()
	if param.Name() != "Context" || !strings.HasSuffix(param.PkgPath(), "github.com/gwf") {
		return false
	}

	return true
}

func callActionMethod(c *Context, actionMethod interface{}) {
	actionMethodType := reflect.TypeOf(actionMethod)
	actionMethodName := runtime.FuncForPC(reflect.ValueOf(actionMethod).Pointer()).Name()
	actionMethodNameSlice := strings.Split(actionMethodName, ".")
	actionMethodName = actionMethodNameSlice[len(actionMethodNameSlice)-1]
	param := actionMethodType.In(0)
	k := param.Kind()
	if k != reflect.Ptr {
		panic("Action定义错误")
	}
	controllerType := param.Elem()
	controller := reflect.New(controllerType)

	baseControllerValue := controller.Elem().FieldByName("Controller")
	baseControllerType := controllerTypeApi
	if !baseControllerValue.IsValid() {
		baseControllerValue = controller.Elem().FieldByName("AdminController")
		if !baseControllerValue.IsValid() {
			panic("Controller定义错误，目前只支持Controller和AdminController")
		}
		baseControllerType = controllerTypeAdmin
	}

	var baseController IController
	switch baseControllerType {
	case controllerTypeApi:
		baseController = &Controller{}
	case controllerTypeAdmin:
		baseController = &AdminController{}
	default:
		panic("BaseController错误")
	}

	baseControllerValue.Set(reflect.ValueOf(baseController))
	initMethod := controller.MethodByName(INIT_METHOD_NAME)
	if initMethod.IsValid() {
		//定义了Init方法
		initMethod.Call([]reflect.Value{})
	}

	m := controller.MethodByName(actionMethodName)
	if m.IsValid() {
		m.Call([]reflect.Value{reflect.ValueOf(c)})
	} else {
		//不会走到这里来
		panic("未定义")
	}
}
