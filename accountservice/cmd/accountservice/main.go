package main

import (
	"github.com/callistaenterprise/goblog/accountservice/cmd"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexflint/go-arg"
	"github.com/callistaenterprise/goblog/accountservice/internal/pkg/service"
	cb "github.com/callistaenterprise/goblog/common/circuitbreaker"
	"github.com/callistaenterprise/goblog/common/messaging"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/sirupsen/logrus"
)

var appName = "accountservice"

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Infof("Starting %v\n", appName)

	// Initialize config struct and populate it froms env vars and flags.
	cfg := cmd.DefaultConfiguration()
	arg.MustParse(cfg)

	initializeTracing(cfg)
	initializeMessaging(cfg)
	cb.ConfigureHystrix([]string{"account-to-data", "account-to-image", "account-to-quotes"}, service.MessagingClient)

	handleSigterm(func() {
		cb.Deregister(service.MessagingClient)
		service.MessagingClient.Close()
	})
	service.StartWebServer(cfg.Name, cfg.ServerConfig.Port)
}
func initializeTracing(cfg *cmd.Config) {
	tracing.InitTracing(cfg.ZipkinServerUrl, appName)
}

func initializeMessaging(cfg *cmd.Config) {
	if cfg.AmqpConfig.ServerUrl == "" {
		panic("No 'amqp_server_url' set in configuration, cannot start")
	}

	service.MessagingClient = &messaging.AmqpClient{}
	service.MessagingClient.ConnectToBroker(cfg.AmqpConfig.ServerUrl)
}

// Handles Ctrl+C or most other means of "controlled" shutdown gracefully. Invokes the supplied func before exiting.
func handleSigterm(handleExit func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		handleExit()
		os.Exit(1)
	}()
}
