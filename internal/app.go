package internal

import (
	"context"
	"os"

	handler "codechat.dev/api/handlers"
	"codechat.dev/api/routers"
	"codechat.dev/guards"
	"codechat.dev/internal/domain/instance"
	sendmessage "codechat.dev/internal/domain/send_message"
	"codechat.dev/internal/whatsapp"
	middle "codechat.dev/middlewares"
	"codechat.dev/pkg/config"
	"codechat.dev/pkg/messaging"
	"codechat.dev/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

type appContext struct {
	Logger  *logrus.Entry
	Cfg     *config.AppConfig
	Router  *chi.Mux
	Storage *whatsapp.Store
	Amqp    *messaging.Amqp
}

type Provider struct {
	Ctx appContext
}

func App(provider *Provider) {
	ctx := context.Background()

	build := logrus.New()
	build.SetFormatter(&logrus.JSONFormatter{})
	logger := build.WithFields(logrus.Fields{"name": "internal-app"})
	logger.Logger.Out = os.Stdout

	logger.Info("Starting application...")

	logger.Info("Loading environment variables...")
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Panic("Cannot load '.env': ", err)
	}
	logger.Info("Env loaded: success")

	logger.Info("Starting storage...")
	dbUrl, err := config.DatabaseUrl(*cfg.DbPath)
	if err != nil {
		logger.Panic("Problems with the sqlite database url: ", err)
	}
	storage, err := whatsapp.StoreConnect(dbUrl)
	if err != nil {
		logger.Panic("Failed to initialize the storage: ", err)
	}
	logger.Info("Storage initialized")

	utils.DbUrl = dbUrl

	msgClient, err := messaging.NewConnection(cfg.Messaging)
	if err != nil {
		logger.Panic("amqp connection failed: ", err)
	}

	manager := instance.NewInstancesManager(storage, msgClient)
	logger.Info("Instance Manager initialized")

	instanceService := instance.NewService(storage, manager, msgClient, cfg.Routes.MsManager)
	logger.Info("InstanceService loaded.")

	go func() {
		msgClient.SetQueues()
		logger.Info("All queues loaded.")
		manager.Load(instanceService)
		logger.Info("All instances loaded.")
	}()

	instanceRoutes := routers.NewInstanceRouter(
		handler.NewInstance(instanceService),
	)
	logger.WithFields(
		logrus.Fields{"service": "ok", "handler": "ok"},
	).Info("Instance routers initialized")

	whatsAppRoutes := routers.NewWhatsAppRouter(
		handler.NewInstance(instanceService),
	)
	logger.WithFields(
		logrus.Fields{"service": "ok", "handler": "ok"},
	).Info("WhatsApp routers initialized")

	sendmessageRoutes := routers.NewSendMessageRouter(
		handler.NewSendMessage(
			sendmessage.NewService(storage, manager, msgClient, ctx),
		),
	)
	logger.WithFields(
		logrus.Fields{"service": "ok", "handler": "ok"},
	).Info("SendMessage routers initialized")

	msManagerRoutes := routers.NewMsManagerRouter(cfg.Routes, instanceService)
	logger.WithFields(
		logrus.Fields{"ms": "ok"},
	).Info("Microservice routes routers initialized")

	r := chi.NewRouter()
	logger.Info("Global route handler initialized.")

	r.Use(middle.LoggerMiddleware(logger))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	logger.Info("Loaded middlewares")

	authGuard := guards.NewAuthGuard(storage, cfg.GlobalToken)
	instanceGuard := guards.NewInstanceGuard(&manager.Wa, storage, cfg.GlobalToken)

	logger.Info("Grouped router initialized")
	instanceRouter := instanceRoutes.
		Auth(authGuard).
		GlobalMiddleware(instanceGuard.IsAnInstance).
		RootPath("/instance").
		RootParam("/{instance}").
		Routers(
			routers.InnerRouters{
				Path:               "/whatsapp",
				Router:             whatsAppRoutes,
				InstanceMiddleware: []handler.InstanceMiddleware{instanceGuard.IsLoggedIn},
			},
			routers.InnerRouters{
				Path:               "/send",
				Router:             sendmessageRoutes,
				InstanceMiddleware: []handler.InstanceMiddleware{instanceGuard.IsLoggedIn},
			},
			routers.InnerRouters{
				Path:               "/",
				Router:             msManagerRoutes,
			},
		)

	routers.PingOptions(r)
	routers.NotFound(r)

	r.Mount("/api/v3", instanceRouter)

	provider.Ctx = appContext{
		Logger:  logger,
		Cfg:     cfg,
		Storage: storage,
		Router:  r,
		Amqp:    msgClient,
	}
}
