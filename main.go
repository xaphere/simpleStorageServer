package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var storage = kingpin.Flag("storage", "file directory to server").Envar("STORAGE").ExistingDir()
var port = kingpin.Flag("port", "port to server on").Envar("PORT").Default("3000").String()

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()

	log := logrus.New()
	ctx := context.Background()
	logEntry := log.WithFields(map[string]interface{}{
		"storage": *storage,
		"port":    *port,
	})
	logEntry.Info("Starting file server")
	stopSrvFn := startFileServer(ctx, ":"+*port, *storage, log)
	waitForSignal()
	logEntry.Info("Stopping file server")
	stopSrvFn()
}

func startFileServer(ctx context.Context, address string, storage string, log *logrus.Logger) func() {
	srv := &http.Server{
		Addr:    address,
		Handler: http.FileServer(http.Dir(storage)),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.WithError(err).Error("server closed unexpectedly")
		}
	}()
	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.WithError(err).Error("server failed to close properly")
		}
	}
}

func waitForSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
