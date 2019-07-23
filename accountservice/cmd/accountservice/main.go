package main

import (
	"github.com/callistaenterprise/goblog/accountservice/cmd"
	"net/http"
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
	mc := initializeMessaging(cfg)
	cb.ConfigureHystrix([]string{"account-to-data", "account-to-image", "account-to-quotes"}, mc)

	client := &http.Client{}
	var transport http.RoundTripper = &http.Transport{
		DisableKeepAlives: true,
	}
	client.Transport = transport
	cb.Client = client
	h := service.NewHandler(mc, client)
	qlResolvers := service.NewLiveGraphQLResolvers(h)

	s := service.NewServer(cfg, h, qlResolvers)
	s.SetupRoutes()

	handleSigterm(func() {
		cb.Deregister(mc)
		mc.Close()
		s.Close()
	})
	s.Start()
}
func initializeTracing(cfg *cmd.Config) {
	tracing.InitTracing(cfg.ZipkinServerUrl, appName)
}

func initializeMessaging(cfg *cmd.Config) *messaging.AmqpClient {
	if cfg.AmqpConfig.ServerUrl == "" {
		panic("No 'amqp_server_url' set in configuration, cannot start")
	}

	mc := &messaging.AmqpClient{}
	mc.ConnectToBroker(cfg.AmqpConfig.ServerUrl)
	return mc
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
