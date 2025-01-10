package logger

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

type ZapLogger struct {
	Logger *zap.Logger
	Level  zap.AtomicLevel //TODO mb use l instead of L? + create logger with settings
}

// NewZapLogger returns a new ZapLogger configured with the provided options.
func NewZapLogger(level string) (*ZapLogger, error) {

	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	// create new log configuration
	cfg := zap.NewProductionConfig()
	// set lvl
	cfg.Level = lvl
	// create logger on the basis of config
	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	var newNop = zap.NewNop()
	newNop = zl
	return &ZapLogger{
		Logger: newNop,
		Level:  lvl,
	}, nil
}

func (l *ZapLogger) LogMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer l.Logger.Sync() // TODO: Is it true? How  should I use that??
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
		next.ServeHTTP(&lw, req)
		// get request duration
		duration := time.Since(start)
		// TODO should I use  defer in this middleware??
		l.Logger.Info("event",
			zap.String("uri", req.RequestURI),
			zap.String("method", req.Method),
			zap.Int("status", responseData.status),
			zap.Duration("duration", duration),
			zap.Int("size", responseData.size),
		)
	})
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
