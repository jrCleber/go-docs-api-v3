package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"codechat.dev/pkg/utils"
	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type AMQP struct {
	Url    string
	Queues []string
}

type JwtConfig struct {
	Expires bool
}

type DbPath struct {
	Path     string
	FileName string
	Query    string
}

type Route struct {
	MsManager string
}

type Container struct {
	ID   string
	Name string
}

type AppConfig struct {
	Server      *ServerConfig
	Messaging   *AMQP
	Jwt         *JwtConfig
	GlobalToken string
	Container   *Container
	Queues      []string
	DbPath      *DbPath
	Routes      *Route
}

var GlobalWebhook string
var LicenseKey string

func DatabaseUrl(env DbPath) (string, error) {
	dbPath := env.Path
	if !strings.HasSuffix(dbPath, "/") {
		dbPath = dbPath + "/"
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		e := os.MkdirAll(dbPath, 0755)
		if err != nil {
			return "", e
		}

		_, err := os.Create(dbPath)
		if err != nil {
			return "", err
		}
	}

	dbPath = dbPath + env.FileName

	if env.FileName == "" {
		dbPath = dbPath + "codechat.db"
	}

	if env.Query != "" {
		dbPath = dbPath + "?" + env.Query
	}

	return dbPath, nil
}

// log.Fatal("Unable to load '.env' file: %v", err)
func LoadConfig() (*AppConfig, error) {

	dockerEnv := os.Getenv("DOCKER_ENV") == "true"

	if !dockerEnv {
		err := godotenv.Load()
		if err != nil {
			return nil, err
		}
	}

	readTimeout, err := time.ParseDuration(os.Getenv("SERVER_READ_TIMEOUT"))
	if err != nil {
		return nil, err
	}

	writeTimeout, err := time.ParseDuration(os.Getenv("SERVER_WRITE_TIMEOUT"))
	if err != nil {
		return nil, err
	}

	expires, err := strconv.ParseBool(os.Getenv("JWT_EXPIRES"))
	if err != nil {
		return nil, err
	}

	queues := strings.Split(os.Getenv("QUEUES"), ":")
	if len(queues) == 0 {
		return nil, errors.New("failed queue read")
	}

	dbPath := os.Getenv("DB_PATH")
	fileName := os.Getenv("CONTAINER_NAME") + ".db"
	if dbPath == "" || fileName == "" {
		return nil, errors.New("data file path not defined")
	}

	dbQuery := os.Getenv("DB_QUERY_URL")

	amqpUrl := utils.StringJoin("/", os.Getenv("AMQP_URL"), os.Getenv("AMQP_VHOST"))

	GlobalWebhook = os.Getenv("GLOBAL_WEBHOOK")
	LicenseKey = os.Getenv("LICENSE_KEY")

	config := AppConfig{
		Server: &ServerConfig{
			Port:         os.Getenv("SERVER_PORT"),
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
		Messaging: &AMQP{
			Url:    amqpUrl,
			Queues: strings.Split(os.Getenv("AMQP_QUEUES"), ","),
		},
		Jwt:    &JwtConfig{Expires: expires},
		Queues: queues,
		DbPath: &DbPath{
			Path:     dbPath,
			FileName: fileName,
			Query:    dbQuery,
		},
		Routes: &Route{
			MsManager: os.Getenv("BASE_ROUTER_MS_MANAGER"),
		},
		GlobalToken: os.Getenv("GLOBAL_TOKEN"),
		Container: &Container{
			ID:   os.Getenv("CONTAINER_ID"),
			Name: os.Getenv("CONTAINER_NAME"),
		},
	}

	return &config, nil
}
