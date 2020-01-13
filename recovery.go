package gwf

import (
	"net"
	"net/http/httputil"
	"os"
	"strings"
)

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				// 检查连接断开的情况
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						errMsg := strings.ToLower(se.Error())
						if strings.Contains(errMsg, "broken pipe") ||
							strings.Contains(errMsg, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				stack := Stack(4)
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				// 如果是connect broken，无法写入status
				if brokenPipe {
					c.app.Logger.Printf("[Recovery] [BrokenPipe] panic recovery:%s dump:[%s]",
						err, string(httpRequest))
					c.Abort()
				} else {
					c.app.Logger.Printf("[Recovery] [Stack] panic recovery:%s stack:%s dump:[%s]",
						err, StackNoNewLine(0), string(httpRequest))
					c.Abort()
					c.errInternal = &Error{
						Err:   err,
						Type:  ErrorTypeInternal,
						Stack: string(stack),
					}
					DefaultInternalServerErrorHandler(c)
				}

			}
		}()

		c.Next()
	}
}
