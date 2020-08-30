package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	errLoadConfig   = errors.New("fatal error config file")
	errDecodeStruct = errors.New("unable to decode into struct")
)

type Config struct {
	AppName string `mapstructure:"app_name"`
	Log     struct {
		Format string `mapstructure:"format"`
		Level  string `mapstructure:"level"`
	}
	Listen  string `mapstructure:"listen"`
	Metrics struct {
		Enable bool   `mapstructure:"enable"`
		Port   string `mapstructure:"port"`
	}
	Web struct {
		Enable bool   `mapstructure:"enable"`
		Port   string `mapstructure:"port"`
	}
}

type Server struct {
	Config *Config
	Log    *logrus.Logger
}

func (s *Server) InitLogger() error {
	if s.Config == nil {
		return errors.New("config file not init")
	}

	logger := logrus.New()
	lvl, err := logrus.ParseLevel(s.Config.Log.Level)

	if err == nil {
		logger.SetLevel(lvl)
	}

	if s.Config.Log.Format == "JSON" {
		logger.SetFormatter(new(logrus.JSONFormatter))
	} else {
		logger.SetFormatter(&logrus.TextFormatter{TimestampFormat: "2006-01-02 15:04:05", FullTimestamp: true})
	}

	s.Log = logger

	return nil
}

func (s *Server) LoadConfig(configName string) error {
	viper.SetConfigName(configName)
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return errLoadConfig
	}

	var c Config

	err = viper.Unmarshal(&c)
	if err != nil {
		return errors.Wrapf(errDecodeStruct, "struct: %v", c)
	}

	s.Config = &c

	return nil
}

func (s *Server) Run(ctx context.Context) {
	if s.Config.Metrics.Enable {
		s.Log.Info("Prometheus server is enabled")

		go s.RegisterPrometheus(ctx)
	} else {
		s.Log.Info("Prometheus server is disabled")
	}

	if s.Config.Web.Enable {
		s.Log.Info("Web server is enabled")

		go s.RegisterWebServer(ctx)
	} else {
		s.Log.Info("Web server is disabled")
	}

	s.RegisterShutdown(ctx)
}

func (s *Server) Stop(ctx context.Context) error {
	s.Log.Info("stopping application server")

	return nil
}

func (s *Server) RegisterPrometheus(ctx context.Context) {
	http.Handle("/metrics", promhttp.Handler())

	err := http.ListenAndServe(s.Config.Metrics.Port, nil)
	if err != nil {
		s.Log.Error(err)
	}
}

func (s *Server) RegisterWebServer(ctx context.Context) {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, "Hello world")
	})

	err := http.ListenAndServe(s.Config.Web.Port, nil)
	if err != nil {
		s.Log.Error(err)
	}
}

func (s *Server) RegisterShutdown(ctx context.Context) {
	ch := make(chan os.Signal, 1)

	var err error

	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	<-ch
	s.Log.Warning("Interrupt signal fetched")

	err = s.Stop(ctx)
	if err != nil {
		s.Log.Error(errors.Wrap(err, "Very sad to see error on shutdown"))
	}
}
