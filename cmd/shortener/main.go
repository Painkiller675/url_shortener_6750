package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Painkiller675/url_shortener_6750/internal/business"
	"github.com/Painkiller675/url_shortener_6750/internal/controller/grpc"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/controller"
	gzipMW "github.com/Painkiller675/url_shortener_6750/internal/middleware/gzip"
	"github.com/Painkiller675/url_shortener_6750/internal/middleware/logger"
	"github.com/Painkiller675/url_shortener_6750/internal/repository"
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
	if err := config.SetConfig(); err != nil {
		log.Fatal(err)
	}

	// init logger
	l, err := logger.NewZapLogger(config.StartOptions.LogLvl)
	if err != nil {
		log.Panic(err)
	}
	//l.Logger.Info("Starting server", zap.String("ConString: ", config.StartOptions.DBConStr), zap.String("BaseURL:", config.StartOptions.BaseURL.String()))
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
	var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup                         // TODO [MENTOR] надо ли создавать отдельный канал для gRPC под удаление?
	chanJobs := make(chan models.JobToDelete, 100) // смысла нет в БУФЕРЕ, если клиентов >100
	// ибо всё равно узкое горлышко - обращение к БД и там все эти горутины всё равно встанут в очередь
	//defer close(chanJobs)

	// launch the deleting goroutine
	wg2.Add(1)
	go deleteURL(&wg2, s, chanJobs)

	// create a wait group
	//var wg sync.WaitGroup // TODO bring it to controller

	// init business logic instance
	busLog := business.NewBusiness(s, l.Logger)

	// init controller
	c := controller.New(busLog, l.Logger, chanJobs, &wg1) //

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
		r.Post("/api/shorten", c.CreateShortURLJSONHandler()) // TODO: gRPC??
		r.Post("/api/shorten/batch", c.CreateShortURLJSONBatchHandler())
		r.Get("/api/user/urls", c.GetUserURLSHandler())
		r.Delete("/api/user/urls", c.DeleteURLSHandler())
		r.Get("/api/internal/stats", c.GetStats())
		r.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		r.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		r.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		r.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		r.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	})
	// graceful shutdown
	var srv = http.Server{Addr: config.StartOptions.HTTPServer.Address, Handler: r}
	sigChan := make(chan os.Signal, 2)
	idleConnsClosed := make(chan struct{}) // to notice the main thread that connections were closed
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	go func() {
		<-sigChan
		fmt.Println("shutting down server gracefully..")
		if err := srv.Shutdown(ctx); err != nil { // дожидается отработки всех хэндлеров, перестаёт слушать на порту
			// ошибки закрытия Listener

			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	//grpcServer := &grpc.Server{}
	grpcServer := grpc.NewGRPCServer(busLog, l.Logger, &wg1, chanJobs)

	go func() {
		if err := grpc.Serve(grpcServer); err != nil {
			log.Println("gRPC server failed: %w", err)
		}
	}()

	fmt.Printf("Build version: %s\n Build date: %s\n Build commit: %s\n\n", buildVersion, buildDate, buildCommit)
	//start server, choose http or https
	go func() {
		//var srv = http.Server{Addr:config.StartOptions.HTTPServer.Address, Handler: r} // create the server

		if config.StartOptions.HTTPSEnabled {

			l.Logger.Info("Running HTTPS server", zap.String("address", config.StartOptions.HTTPServer.Address))
			l.Logger.Info("Server options:", zap.String("ConString: ", config.StartOptions.DBConStr), zap.String("BaseURL:", config.StartOptions.BaseURL.String()))
			if err := srv.ListenAndServeTLS(config.StartOptions.CertFile, config.StartOptions.KeyFile); err != nil {
				if !errors.Is(err, http.ErrServerClosed) { // если завершился не по штатному шатдауну
					panic(err)
				}
			}
		} else { // start http server
			l.Logger.Info("Running HTTP server", zap.String("address", config.StartOptions.HTTPServer.Address))
			if err := srv.ListenAndServe(); err != nil {

				if !errors.Is(err, http.ErrServerClosed) { // если завершился не по штатному шатдауну
					panic(err)
				}
			}
		}
	}()

	<-idleConnsClosed

	wg1.Wait() // ВСЕ ГОРУТИНЫ ХЭНДЛЕРОВ ОТРАБОТАЛИ
	close(chanJobs)
	wg2.Wait() // ждём отработки горутины непосредственного удаления (если что-то ещё осталось в буфере)
	// TODO:  how to handle error here????
	if err := s.Close(); err != nil {
		log.Fatal("cannot close storage connection", zap.Error(err))
	}
	fmt.Println("server was gracefully stopped")

}

// 1 гасим веб сервер
// в хэндлер wg пробросить и wg.Add и ждать
// далее после wg.wait закрыть канал но в нём могут быть данные => вторую wg ждём для go deleteURL(s, chanJobs)
// TODO: либо в горутине запускать сервак и после этого ловить сигнал ..
func deleteURL(wg *sync.WaitGroup, s repository.URLStorage, jobs chan models.JobToDelete) {
	defer wg.Done()
	for job := range jobs { // waiting for data in buffered channel
		if err := s.DeleteURLsByUserID(context.Background(), job.UserID, job.LsURL); err != nil {
			fmt.Println("[ERROR]", zap.Error(err)) // TODO [MENTOR]: how to go it up? is it necessary?
		}
		//TODO: for GS Сделать тут wg.Done()

	}
}
