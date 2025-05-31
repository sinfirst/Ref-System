package logging

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Logger struct {
	Logger zap.SugaredLogger
}

type ResponseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *ResponseData
}

func NewLogger() *Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	sugar := *logger.Sugar()

	return &Logger{
		Logger: sugar,
	}
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
}

func (l *Logger) WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &ResponseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)
		l.Logger.Infoln(
			"\n",
			"-----REQUEST-----\n",
			"URI:", r.RequestURI, "\n",
			"Method:", r.Method, "\n",
			"Duration:", duration, "\n",
			"-----RESPONSE-----\n",
			"Status:", responseData.status, "\n",
			"Size:", responseData.size, "\n",
		)
	})
}
