package gwf

import (
	"fmt"
	"net/http"
)

// DefaultNotFoundHandler 默认404handler
var DefaultNotFoundHandler HandlerFunc = func(c *Context) {
	c.Bytes(http.StatusNotFound, []byte("资源不存在"))
}

// DefaultInternalServerErrorHandler 默认500handler
var DefaultInternalServerErrorHandler HandlerFunc = func(c *Context) {
	if GetConfig().IsOnlineEnvironment() || GetConfig().IsPreEnvironment() {
		c.Bytes(http.StatusInternalServerError, []byte("服务器内部错误"))
	} else {
		body := fmt.Sprintf("%s \n%s", c.errInternal.Err, c.errInternal.Stack)
		c.String(http.StatusInternalServerError, body)
	}
}
