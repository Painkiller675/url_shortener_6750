package logger

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

var Log *zap.Logger = zap.NewNop()

func Initialize(level string) error {
	// transform text logging lvl in zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	// create new log configuration
	cfg := zap.NewProductionConfig()
	// set lvl
	cfg.Level = lvl
	// create logger on the basis of config
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	// set singleton
	Log = zl
	return nil
}

// structure to save response  info
type (
	responseData struct {
		status int
		size   int
	}
	// add the realization of http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// redefine the methods to get needed response data
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// get response using original http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	// new functionality
	r.responseData.size += size // get the size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// get statusCOde using original http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	// new functionality
	r.responseData.status = statusCode // get codeStatus
}

// define logger middleware for handlers
func LogMW(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		defer Log.Sync() // TODO: Is it true? How  should I use that??
		// get current time
		start := time.Now()
		// mold collections to fill the structure
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		// fill the custom logging ResponseWriter
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		// serve an original request with custom ResponseWriter
		h.ServeHTTP(&lw, r)
		// get request duration
		duration := time.Since(start)
		// TODO should I use  defer in this middleware??
		Log.Info("event",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseData.status),
			zap.Duration("duration", duration),
			zap.Int("size", responseData.size),
		)

	}

	return http.HandlerFunc(logFn)
}
