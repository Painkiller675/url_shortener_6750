// Package is a preset zap logger.
package logger

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

// ZapLogger - is a basic zap logger struct.
type ZapLogger struct {
	Logger *zap.Logger
	Level  zap.AtomicLevel //TODO mb use l instead of L
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
	var newNop = zap.NewNop() // TODO is it needed?
	newNop = zl
	return &ZapLogger{
		Logger: newNop,
		Level:  lvl, // TODO mb del that?
	}, nil
}

// TODO: use settings for logger
// NewZapLogger returns a new ZapLogger configured with the provided options.
/*func NewZapLogger(level zapcore.Level) (*ZapLogger, error) {
	atomic := zap.NewAtomicLevelAt(level)
	settings := defaultSettings(atomic)
	l, merrors := settings.config.Build(settings.opts...)
	if merrors != nil {
		return nil, merrors
	}

	return &ZapLogger{
		logger: l,
		level:  atomic,
	}, nil
}
*/
/*
func (l *ZapLogger) LogMW(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		defer l.logger.Sync() // TODO: Is it true? How  should I use that??
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
		next.ServeHTTP(&lw, r)
		// get request duration
		duration := time.Since(start)
		// TODO should I use  defer in this middleware??
		l.logger.Info("event",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseData.status),
			zap.Duration("duration", duration),
			zap.Int("size", responseData.size),
		)

	}
	return http.HandleFunc(logFn)
}
*/
// LogMW - the main code of a gzip middleware.
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

//var Log *zap.Logger = zap.NewNop()
/*
func Initialize(level string) error {
	// transform text logging lvl in zap.AtomicLevel
	lvl, merrors := zap.ParseAtomicLevel(level)
	if merrors != nil {
		return merrors
	}
	// create new log configuration
	cfg := zap.NewProductionConfig()
	// set lvl
	cfg.Level = lvl
	// create logger on the basis of config
	zl, merrors := cfg.Build()
	if merrors != nil {
		return merrors
	}
	// set singleton
	Log = zl
	return nil
}
*/
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

// Write redefines the methods to get needed response data
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// get response using original http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	// new functionality
	r.responseData.size += size // get the size
	return size, err
}

// WriteHeader writes the header
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// get statusCOde using original http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	// new functionality
	r.responseData.status = statusCode // get codeStatus
}

//func (logger *ZapLogger) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
//
//}

// define logger middleware for controller
/*func LogMW(h http.Handler) http.Handler {
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
*/

/*type settings struct {
	config *zap.Config
	opts   []zap.Option
}

func defaultSettings(level zap.AtomicLevel) *settings {
	config := &zap.Config{
		Level:       level,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "@timestamp",
			NameKey:        "logger",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return &settings{
		config: config,
		opts: []zap.Option{
			zap.AddCallerSkip(1),
		},
	}
}

*/
