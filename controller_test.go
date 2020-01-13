package gwf

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/iancoleman/strcase"
)

type NotControllerA struct {
	*Controller
}

type SomeController struct {
	*Controller
}

func (controller *SomeController) SomeAction(c *Context) {

}

func (controller *SomeController) NotActionA(c *Context) {

}

func (controller *SomeController) NotAction(c *context.Context) {

}

func TestIsController(t *testing.T) {
	if isController(&NotControllerA{}) {
		t.Fatal("NotControllerA expect not, suffix must be Controller")
	}

	if isController(SomeController{}) {
		t.Fatal("SomeController expect not, must ptr")
	}

	if !isController(&SomeController{}) {
		t.Fatal("SomeController expect yes")
	}
}

func TestIsControllerActionMethod(t *testing.T) {
	if isControllerActionMethod((*SomeController).NotActionA) {
		t.Fatal("NotActionA expect not, suffix must be Action")
	}

	if isControllerActionMethod((*SomeController).NotAction) {
		t.Fatal("NotAction params error")
	}

	if !isControllerActionMethod((*SomeController).SomeAction) {
		t.Fatal("SomeAction expect ok")
	}

	controllerPointerType := reflect.ValueOf(&SomeController{}).Type()
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
		if actionName != "some" {
			t.Fatal("expect actionName equal some")
		}
	}
}

func TestCallActionMethod(t *testing.T) {
	initTestAppConfig(t)
	callActionMethod(nil, (*SomeController).SomeAction)
}
