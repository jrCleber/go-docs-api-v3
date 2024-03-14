package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codechat.dev/internal"
	"github.com/sirupsen/logrus"
)

func main() {
	provider := internal.Provider{}
	internal.App(&provider)

	defer provider.Ctx.Storage.Client.Close()
	defer provider.Ctx.Amqp.Conn.Close()

	logger := provider.Ctx.Logger

	srvPort := provider.Ctx.Cfg.Server.Port

	srv := &http.Server{
		Addr:    ":" + srvPort,
		Handler: provider.Ctx.Router,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.WithFields(logrus.Fields{"error": err.Error()}).Fatal("server stopped")
		}

	logger.WithFields(logrus.Fields{"port": srvPort}).Info("Server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.WithFields(logrus.Fields{"signal": sig}).Warn("Server shuting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithFields(logrus.Fields{"error": err.Error()}).Fatal("Server forced to shutdown")
	} else {
		logger.Info("Server exited properly")
	}
}
