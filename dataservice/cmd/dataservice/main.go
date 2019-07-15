package main

import (
	"github.com/alexflint/go-arg"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/callistaenterprise/goblog/dataservice/cmd"
	"github.com/callistaenterprise/goblog/dataservice/internal/pkg/dbclient"
	"github.com/callistaenterprise/goblog/dataservice/internal/pkg/service"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var appName = "dataservice"

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Infof("Starting %v\n", appName)

	// Initialize config struct and populate it froms env vars and flags.
	cfg := cmd.DefaultConfiguration()
	arg.MustParse(cfg)

	service.DBClient = &dbclient.GormClient{}
	service.DBClient.SetupDB(cfg.CockroachdbConnUrl)
	service.DBClient.SeedAccounts()

	initializeTracing(cfg)

	handleSigterm(func() {
		logrus.Infoln("Captured Ctrl+C")
		service.DBClient.Close()
	})
	service.StartWebServer(cfg.Name, cfg.ServerConfig.Port)
}
func initializeTracing(cfg *cmd.Config) {
	tracing.InitTracing(cfg.ZipkinServerUrl, appName)
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
