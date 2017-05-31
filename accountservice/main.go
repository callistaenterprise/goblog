package main

import (
        "flag"
        "fmt"
        "github.com/callistaenterprise/goblog/common/config"
        "github.com/callistaenterprise/goblog/common/messaging"
        "github.com/callistaenterprise/goblog/accountservice/dbclient"
        "github.com/callistaenterprise/goblog/accountservice/service"
        "github.com/spf13/viper"
        "os"
        "os/signal"
        "syscall"
)

var appName = "accountservice"

func init() {
        profile := flag.String("profile", "test", "Environment profile, something similar to spring profiles")
        configServerUrl := flag.String("configServerUrl", "http://configserver:8888", "Address to config server")
        configBranch := flag.String("configBranch", "master", "git branch to fetch configuration from")

        flag.Parse()

        fmt.Println("Specified configBranch is " + *configBranch)

        viper.Set("profile", *profile)
        viper.Set("configServerUrl", *configServerUrl)
        viper.Set("configBranch", *configBranch)
}

func main() {
        fmt.Printf("Starting %v\n", appName)

        config.LoadConfigurationFromBranch(
                viper.GetString("configServerUrl"),
                appName,
                viper.GetString("profile"),
                viper.GetString("configBranch"))
        initializeBoltClient()
        initializeMessaging()
        handleSigterm(func() {

        })
        service.MessagingConsumer.Subscribe(viper.GetString("config_event_bus"), "topic", appName, config.HandleRefreshEvent)
        //go config.StartListener(appName, viper.GetString("amqp_server_url"), viper.GetString("config_event_bus"))
        service.StartWebServer(viper.GetString("server_port"))
}

func initializeMessaging() {
        if !viper.IsSet("amqp_server_url") {
                panic("No 'amqp_server_url' set in configuration, cannot start")
        }
        service.MessagingClient = &messaging.MessagingClient{}
        service.MessagingClient.ConnectToBroker(viper.GetString("amqp_server_url"))

        service.MessagingConsumer = &messaging.MessagingConsumer{}
        service.MessagingConsumer.ConnectToBroker(viper.GetString("amqp_server_url"))
}

func initializeBoltClient() {
        service.DBClient = &dbclient.BoltClient{}
        service.DBClient.OpenBoltDb()
        service.DBClient.Seed()
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
