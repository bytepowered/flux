package listener

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

// HttpResponseMultiWriter 实现HttpBody写入响应数据时，同时向多个Writer写入数据
// 用以实现Http响应数据流复制
type HttpResponseMultiWriter struct {
	statusCode int
	io.Writer
	http.ResponseWriter
}

// StatusCode 返回当前Response写入的HttpStatusCode；如果未写入响应，返回0；
func (w *HttpResponseMultiWriter) StatusCode() int {
	return w.statusCode
}

// WriteHeader 写入Response的HttpStatusCode；每个请求仅允许写入一次；
func (w *HttpResponseMultiWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *HttpResponseMultiWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *HttpResponseMultiWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *HttpResponseMultiWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func NewEmptyHttpResponseMultiWriter(responseWriter http.ResponseWriter) *HttpResponseMultiWriter {
	return &HttpResponseMultiWriter{
		statusCode:     0, // 0 表示未写入响应码
		Writer:         responseWriter,
		ResponseWriter: responseWriter,
	}
}

func NewHttpMultiResponseWriter(copyWriter io.Writer, responseWriter http.ResponseWriter) *HttpResponseMultiWriter {
	return &HttpResponseMultiWriter{
		statusCode:     0, // 0 表示未写入响应码
		Writer:         io.MultiWriter(responseWriter, copyWriter),
		ResponseWriter: responseWriter,
	}
}
