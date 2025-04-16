package main

import (
	"context"
	"fmt"
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/controller"
	gzipMW "github.com/Painkiller675/url_shortener_6750/internal/middleware/gzip"
	"github.com/Painkiller675/url_shortener_6750/internal/middleware/logger"
	"github.com/Painkiller675/url_shortener_6750/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

// @title My_URL_Shortener
// @version 1.0
// @description backend to short URLs

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// to print global values
// go run -ldflags "-X main.buildVersion=v1.0.1 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')'
// -X main.buildCommit=iter20" main.go
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func init() {
	// is used for profiling and documentation
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
}

func main() {
	// init config
	config.SetConfig()

	// init logger
	l, err := logger.NewZapLogger(config.StartOptions.LogLvl)
	if err != nil {
		log.Panic(err)
	}
	l.Logger.Info("Starting server", zap.String("ConString: ", config.StartOptions.DBConStr), zap.String("BaseURL:", config.StartOptions.BaseURL))
	//render logger for gzip
	//gzipMW.NewGzipLogger(l.Logger)

	//init the context
	ctx := context.Background()

	// init storage
	s, err := repository.ChooseStorage(ctx, l.Logger)
	if err != nil {
		panic(err) // TODO: [MENTOR] is it good to panic here or I could handle it miles better?
	}

	// init jobs for deleting
	chanJobs := make(chan controller.JobToDelete, 100)
	defer close(chanJobs)

	// launch the delete goroutine
	go deleteURL(s, chanJobs)

	// create a wait group
	//var wg sync.WaitGroup // TODO bring it to controller
	// init controller
	c := controller.New(l.Logger, s, chanJobs) //

	// init router
	r := chi.NewRouter()

	// set logger for chi router
	r.Use(l.LogMW)
	r.Use(gzipMW.GzMW)

	// routing
	r.Route("/", func(r chi.Router) {
		r.Post("/", c.CreateShortURLHandler())
		r.Get("/ping", c.PingDB())
		r.Get("/{id}", c.GetLongURLHandler())
		r.Post("/api/shorten", c.CreateShortURLJSONHandler())
		r.Post("/api/shorten/batch", c.CreateShortURLJSONBatchHandler())
		r.Get("/api/user/urls", c.GetUserURLSHandler())
		r.Delete("/api/user/urls", c.DeleteURLSHandler())
		r.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		r.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		r.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		r.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		r.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	})
	// graceful shutdown
	var srv = http.Server{Addr: config.StartOptions.HTTPServer.Address}
	sigChan := make(chan os.Signal, 2)
	idleConnsClosed := make(chan struct{}) // to notice the main thread that connections were closed
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	go func() {
		<-sigChan
		// TODO : или так gracefully идея: просто поставить таймер на 10 секунд дав удалиться до конца всему
		if err := srv.Shutdown(ctx); err != nil {
			// ошибки закрытия Listener
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	fmt.Printf("Build version: %s\n Build date: %s\n Build commit: %s\n\n", buildVersion, buildDate, buildCommit)
	//start server, choose http or https
	if config.StartOptions.HTTPSEnabled {
		l.Logger.Info("Running HTTPS server", zap.String("address", config.StartOptions.HTTPServer.Address))
		if err := http.ListenAndServeTLS(":8080", "../../internal/cert/localhost.pem", "../../internal/cert/localhost-key.pem", r); err != nil {
			panic(err) // TODO: Как решить вопрос с путями к сертификатам
		}
	} else { // start http server
		l.Logger.Info("Running HTTP server", zap.String("address", config.StartOptions.HTTPServer.Address))
		if err := http.ListenAndServe(config.StartOptions.HTTPServer.Address, r); err != nil {
			panic(err)
		}
	}
	<-idleConnsClosed
	// TODO: close files, pg connections etc
	fmt.Println("Shutting down server GRACEFULLY")
	// TODO: ещё идея, пробросить wg в controller и там делать add
	//wg.Wait() // gracefull shutdown TODO: wg.Add???!
}

// TODO: либо в горутине запускать сервак и после этого ловить сигнал ..
func deleteURL(s repository.URLStorage, jobs chan controller.JobToDelete) {
	for job := range jobs { // waiting for data in buffered channel
		if err := s.DeleteURLsByUserID(context.Background(), job.UserID, job.LsURL); err != nil {
			fmt.Println("[ERROR]", zap.Error(err)) // TODO [MENTOR]: how to go it up? is it necessary?
		}
		//TODO: for GS Сделать тут wg.Done()

	}
}
