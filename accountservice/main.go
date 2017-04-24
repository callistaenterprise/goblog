package main

import (
	"fmt"
	"github.com/callistaenterprise/goblog/accountservice/config"
	"github.com/callistaenterprise/goblog/accountservice/dbclient"
	"github.com/callistaenterprise/goblog/accountservice/service"
	"github.com/spf13/viper"
	"flag"
)

var appName = "accountservice"

func main() {
	fmt.Printf("Starting %v\n", appName)
	parseFlags()
	config.LoadConfiguration(viper.GetString("configServerUrl"), appName, viper.GetString("profile"))
	initializeBoltClient()
	service.StartWebServer("6767")
}

func parseFlags() {
	profile := *flag.String("profile", "test", "Environment profile, something similar to spring profiles")
	configServerUrl := *flag.String("configServerUrl", "http://configserver:8888", "Address to config server")
	flag.Parse()
	viper.Set("profile", profile)
	viper.Set("configServerUrl", configServerUrl)
}

func initializeBoltClient() {
	service.DBClient = &dbclient.BoltClient{}
	service.DBClient.OpenBoltDb()
	service.DBClient.Seed()
}
