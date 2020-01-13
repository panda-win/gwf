package gwf

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	// 平滑重启时的环境变量
	GRACE_ENV = "GWF_GRACE=true"
)

var graceLogger = log.New(os.Stdout, "gwf-grace: ", log.Lshortfile)

// Grace接口实现平滑重启
type Grace interface {
	Run(*http.Server) error
}

type grace struct {
	srv      *http.Server
	listener net.Listener
	timeout  time.Duration
	err      error
}

func (g *grace) reload() *grace {
	f, err := g.listener.(*net.TCPListener).File()
	if err != nil {
		g.err = err
		return g
	}
	defer f.Close()

	var args []string
	if len(os.Args) > 1 {
		args = append(args, os.Args[1:]...)
	}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), GRACE_ENV)
	cmd.ExtraFiles = []*os.File{f}

	g.err = cmd.Start()
	return g
}

func (g *grace) stop() *grace {
	if g.err != nil {
		return g
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	if err := g.srv.Shutdown(ctx); err != nil {
		g.err = err
	}
	return g
}

func (g *grace) run() (err error) {
	if _, ok := syscall.Getenv(strings.Split(GRACE_ENV, "=")[0]); ok {
		// 传入的文件描述符从3开始，0,1,2分别是标准输入，输出，错误输出
		f := os.NewFile(3, "")
		if g.listener, err = net.FileListener(f); err != nil {
			graceLogger.Error("err:" + err.Error())
			return
		}
	} else if g.listener, err = net.Listen("tcp", g.srv.Addr); err != nil {
		graceLogger.Error("err:" + err.Error())
		return
	}

	terminate := make(chan error)
	go func() {
		if err := g.srv.Serve(g.listener); err != nil {
			terminate <- err
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit)

	graceLogger.Debug("开始等待信号...")
	for {
		select {
		case s := <-quit:
			switch s {
			case syscall.SIGINT, syscall.SIGTERM:
				graceLogger.Warn("停止应用")
				signal.Stop(quit)
				return g.stop().err
			case syscall.SIGUSR2:
				graceLogger.Warn("重启应用")
				return g.reload().stop().err
			}
		case err = <-terminate:
			graceLogger.Error("错误:" + err.Error())
			return
		}
	}
}

func newGrace(timeout time.Duration) Grace {
	return &grace{timeout: timeout}
}

func (g *grace) Run(srv *http.Server) error {
	g.srv = srv
	return g.run()
}

// RunGrace方法平滑的运行某个http.Server，平滑重启时，需要停止的进程在timeout后退出
func RunGrace(srv *http.Server, timeout time.Duration) error {
	return newGrace(timeout).Run(srv)
}
