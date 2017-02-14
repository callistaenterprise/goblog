package main

import (
        "fmt"
        "github.com/callistaenterprise/goblog/accountservice/service"
        "github.com/callistaenterprise/goblog/accountservice/dbclient"
        "flag"
        "github.com/spf13/viper"
        "github.com/callistaenterprise/goblog/accountservice/config"
        "github.com/callistaenterprise/goblog/accountservice/messaging"
)

var appName = "accountservice"

func main() {
        fmt.Printf("Starting %v\n", appName)
        parseFlags()
        config.LoadConfiguration(viper.GetString("configServerUrl"), appName, viper.GetString("profile"))
        initializeBoltClient()
        initializeMessaging()
        service.StartWebServer(viper.GetString("server_port"))
}

func parseFlags() {
        profile := flag.String("profile", "test", "Environment profile, something similar to spring profiles")
        configServerUrl := flag.String("configServerUrl", "http://configserver:8888", "Address to config server")

        flag.Parse()
        viper.Set("profile", *profile)
        viper.Set("configServerUrl", *configServerUrl)
}

func initializeMessaging() {
        if !viper.IsSet("broker_url") {
                panic("No 'broker_url' set in configuration, cannot start")
        }
        service.MessagingClient = &messaging.MessagingClient{}
        service.MessagingClient.ConnectToBroker(viper.GetString("broker_url"))
}

func initializeBoltClient() {
        service.DBClient = &dbclient.BoltClient{}
        service.DBClient.OpenBoltDb()
        service.DBClient.Seed()
}
