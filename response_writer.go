package gwf

import (
	"io"
	"net/http"
)

const (
	// NoWritten 默认未写入响应时的size初始值
	NoWritten = -1
	// DefaultStatus 默认响应状态码200
	DefaultStatus = http.StatusOK
)

// ResponseStatusHandler 响应状态的装饰器handler
type ResponseStatusHandler func(int) int

// ResponseHeaderHandler 响应头的装饰器handler
type ResponseHeaderHandler func(http.Header)

// ResponseBodyHandler 响应body的装饰器handler
type ResponseBodyHandler func([]byte) []byte

type responseWriter struct {
	http.ResponseWriter
	size   int
	status int

	ResponseStatusHandler ResponseStatusHandler
	ResponseHeaderHandler ResponseHeaderHandler
	ResponseBodyHandler   ResponseBodyHandler
}

// NewResponseWriter 初始化
func NewResponseWriter(w http.ResponseWriter, respStatusHandler ResponseStatusHandler, respHeaderHandler ResponseHeaderHandler,
	respBodyHandler ResponseBodyHandler) *responseWriter {
	return &responseWriter{
		ResponseWriter:        w,
		size:                  NoWritten,
		status:                DefaultStatus,
		ResponseStatusHandler: respStatusHandler,
		ResponseHeaderHandler: respHeaderHandler,
		ResponseBodyHandler:   respBodyHandler,
	}
}

// Written 已写入响应返回true
func (w *responseWriter) Written() bool {
	return w.size != NoWritten
}

// WriteHeader 仅修改responseWriter中的status
func (w *responseWriter) WriteHeader(code int) {
	if code > 0 && w.status != code {
		w.status = code
	}
}

// WriteHeaderNow 将responseWriter中status写入响应头
func (w *responseWriter) WriteHeaderNow() {
	if !w.Written() {
		w.size = 0

		if w.ResponseHeaderHandler != nil {
			w.ResponseHeaderHandler(w.ResponseWriter.Header())
		}

		if w.ResponseStatusHandler != nil {
			w.status = w.ResponseStatusHandler(w.status)
		}
		w.ResponseWriter.WriteHeader(w.status)
	}
}

// Write 写入响应数据
func (w *responseWriter) Write(data []byte) (n int, err error) {
	w.WriteHeaderNow()
	if w.ResponseBodyHandler != nil {
		data = w.ResponseBodyHandler(data)
	}
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	return
}

// WriteString 写入响应数据
func (w *responseWriter) WriteString(s string) (n int, err error) {
	w.WriteHeaderNow()
	n, err = io.WriteString(w.ResponseWriter, s)
	w.size += n
	return
}

// Status 返回当前响应状态码
func (w *responseWriter) Status() int {
	return w.status
}

// Size 返回当前响应字节数
func (w *responseWriter) Size() int {
	return w.size
}
