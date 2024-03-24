package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"golang.org/x/crypto/acme/autocert"

	"github.com/vzveiteskostrami/goph-keeper/internal/adb"
	"github.com/vzveiteskostrami/goph-keeper/internal/auth"
	"github.com/vzveiteskostrami/goph-keeper/internal/compressing"
	"github.com/vzveiteskostrami/goph-keeper/internal/config"
	"github.com/vzveiteskostrami/goph-keeper/internal/logging"
	"github.com/vzveiteskostrami/goph-keeper/internal/routes"
)

var (
	srv          *http.Server
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)

	logging.LoggingInit()
	config.ReadData()
	if config.Get() == nil {
		os.Exit(11)
	}

	adb.Init()
	defer adb.Close()

	doHTTP()
}

func mainRouter() chi.Router {
	r := chi.NewRouter()

	r.Route("/ready", func(r chi.Router) {
		r.Use(compressing.GZIPHandle)
		r.Use(logging.WithLogging)
		r.Get("/", routes.Readyf)
	})

	r.Route("/register", func(r chi.Router) {
		r.Use(compressing.GZIPHandle)
		r.Use(logging.WithLogging)
		r.Post("/", routes.Registerf)
	})

	r.Route("/login", func(r chi.Router) {
		r.Use(compressing.GZIPHandle)
		r.Use(logging.WithLogging)
		r.Post("/", routes.Authf)
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(compressing.GZIPHandle)
		r.Use(logging.WithLogging)
		r.Use(auth.AuthHandle)
		//r.Post("/shorten", shorturl.SetJSONLinkf)
		//r.Post("/shorten/batch", shorturl.SetJSONBatchLinkf)
		//r.Get("/user/urls", shorturl.GetOwnerURLsListf)
		//r.Delete("/user/urls", shorturl.DeleteOwnerURLsListf)
	})

	return r
}

func doHTTP() {
	cfg := config.Get()
	srv = &http.Server{
		Addr:        *cfg.ServerAddress,
		Handler:     mainRouter(),
		IdleTimeout: time.Second * 1,
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	go func() {
		<-sigs
		if err := srv.Shutdown(context.Background()); err != nil {
			logging.S().Errorln("Server shutdown error", err)
		} else {
			logging.S().Infoln("Server has been closed succesfully")
		}
	}()

	if strings.Contains(srv.Addr, "localhost") || strings.Contains(srv.Addr, "127.0.0.1") {
		logging.S().Infow(
			"Starting server",
			"addr", srv.Addr,
		)
		srv.ListenAndServe()
	} else {
		logging.S().Infow(
			"Starting server with SSL/TLS",
			"addr", srv.Addr,
		)
		manager := &autocert.Manager{
			// директория для хранения сертификатов
			Cache: autocert.DirCache("cache-dir"),
			// функция, принимающая Terms of Service издателя сертификатов
			Prompt: autocert.AcceptTOS,
			// перечень доменов, для которых будут поддерживаться сертификаты
			HostPolicy: autocert.HostWhitelist(*config.Get().CertAddresses),
		}
		srv.TLSConfig = manager.TLSConfig()
		if err := srv.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			logging.S().Fatal(err)
		}
	}
	logging.S().Infoln("Major thread go home")
}
