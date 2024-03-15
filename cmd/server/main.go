package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codechat.dev/internal"
	"codechat.dev/pkg/messaging"
	"github.com/sirupsen/logrus"
)

func main() {
	provider := internal.NewProvider()
	internal.App(provider)
	
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
	}()

	provider.Ctx.Amqp.SendMessage(string(messaging.APP_STATUS), map[string]any{
		"Event": messaging.APP_STATUS,
		"Instance": map[string]string{
			"instanceId": provider.Ctx.Cfg.Container.ID,
			"name":       provider.Ctx.Cfg.Container.Name,
		},
		"Data": map[string]string{
			"status": "on",
		},
	})

	logger.WithFields(logrus.Fields{"port": srvPort}).Info("Server started")

	fmt.Println("START")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	os.Stdout.Sync()

	provider.Ctx.Amqp.SendMessage(string(messaging.APP_STATUS), map[string]any{
		"Event": messaging.APP_STATUS,
		"Instance": map[string]string{
			"instanceId": provider.Ctx.Cfg.Container.ID,
			"name":       provider.Ctx.Cfg.Container.Name,
		},
		"Data": map[string]string{
			"status": "off",
		},
	})

	logger.WithFields(logrus.Fields{"signal": sig}).Log(logrus.WarnLevel, "Server shuting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithFields(logrus.Fields{"error": err.Error()}).Fatal("Server forced to shutdown")
	} else {
		logger.Info("Server exited properly")
	}
}
