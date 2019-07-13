package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/callistaenterprise/goblog/common/config"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/callistaenterprise/goblog/dataservice/dbclient"
	"github.com/callistaenterprise/goblog/dataservice/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var appName = "dataservice"

func init() {
	profile := flag.String("profile", "test", "Environment profile, something similar to spring profiles")
	configServerURL := flag.String("configServerUrl", "http://configserver:8888", "Address to config server")
	configBranch := flag.String("configBranch", "master", "git branch to fetch configuration from")

	flag.Parse()

	viper.Set("service_name", appName)
	viper.Set("profile", *profile)
	viper.Set("configServerURL", *configServerURL)
	viper.Set("configBranch", *configBranch)
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Infof("Starting %v\n", appName)

	config.LoadConfigurationFromBranch(
		viper.GetString("configServerURL"),
		appName,
		viper.GetString("profile"),
		viper.GetString("configBranch"))

	service.DBClient = &dbclient.GormClient{}
	service.DBClient.SetupDB(viper.GetString("cockroachdb_conn_url"))
	service.DBClient.SeedAccounts()

	initializeTracing()

	handleSigterm(func() {
		logrus.Infoln("Captured Ctrl+C")
		service.DBClient.Close()
	})
	service.StartWebServer(viper.GetString("server_port"))
}
func initializeTracing() {
	tracing.InitTracing(viper.GetString("zipkin_server_url"), appName)
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
