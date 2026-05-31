package httpapi

import "net/http"

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	if lrw.wroteHeader {
		return
	}

	lrw.statusCode = code
	lrw.wroteHeader = true
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LoggingResponseWriter) Write(data []byte) (int, error) {
	if !lrw.wroteHeader {
		lrw.WriteHeader(http.StatusOK)
	}

	return lrw.ResponseWriter.Write(data)
}

func (lrw *LoggingResponseWriter) StatusCode() int {
	return lrw.statusCode
}
