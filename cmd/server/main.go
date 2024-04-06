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

	"github.com/vzveiteskostrami/goph-keeper/internal/server/adb"
	"github.com/vzveiteskostrami/goph-keeper/internal/server/auth"
	"github.com/vzveiteskostrami/goph-keeper/internal/server/compressing"
	"github.com/vzveiteskostrami/goph-keeper/internal/server/config"
	"github.com/vzveiteskostrami/goph-keeper/internal/server/logging"
	"github.com/vzveiteskostrami/goph-keeper/internal/server/routes"
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

	r.Route("/", func(r chi.Router) {
		r.Use(compressing.GZIPHandle)
		r.Use(logging.WithLogging)
		r.Use(auth.AuthHandle)
		r.Post("/del", routes.DeleteDataWritef)
		r.Post("/wlist", routes.UserDataWritef)
		r.Post("/list", routes.UserDataListf)
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

	https := strings.HasPrefix(srv.Addr, "https://")
	if https {
		srv.Addr = srv.Addr[8:]
	} else {
		if strings.HasPrefix(srv.Addr, "http://") {
			srv.Addr = srv.Addr[7:]
		} else {
			https = !(strings.Contains(srv.Addr, "localhost") || strings.Contains(srv.Addr, "127.0.0.1"))
		}
	}

	if https {
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
			HostPolicy: autocert.HostWhitelist(*config.Get().CertAddresses...),
		}
		srv.TLSConfig = manager.TLSConfig()
		if err := srv.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			logging.S().Fatal(err)
		}
	} else {
		logging.S().Infow(
			"Starting server",
			"addr", srv.Addr,
		)
		srv.ListenAndServe()
	}
	logging.S().Infoln("Major thread go home")
}
