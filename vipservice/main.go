/**
The MIT License (MIT)

Copyright (c) 2016 Callista Enterprise

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package main

import (
	"flag"
	"fmt"
	"github.com/callistaenterprise/goblog/vipservice/config"
	"github.com/callistaenterprise/goblog/vipservice/messaging"
	"github.com/callistaenterprise/goblog/vipservice/service"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"os"
	"os/signal"
	"syscall"
)

var appName = "vipservice"

var consumer messaging.IMessagingConsumer

func init() {
        configServerUrl := flag.String("configServerUrl", "http://configserver:8888", "Address to config server")
        profile := flag.String("profile", "test", "Environment profile, something similar to spring profiles")
	configBranch := flag.String("configBranch", "master", "git branch to fetch configuration from")
        flag.Parse()
        
        viper.Set("profile", *profile)
	viper.Set("configServerUrl", *configServerUrl)
	viper.Set("configBranch", *configBranch)
}

func main() {
	fmt.Println("Starting " + appName + "...")

	config.LoadConfigurationFromBranch(viper.GetString("configServerUrl"), appName, viper.GetString("profile"), viper.GetString("configBranch"))
	initializeMessaging()

	// Call the subscribe method with queue name and callback function
	go consumer.Subscribe("vipQueue", onMessage)

	// Makes sure connection is closed when service exits.
	handleSigterm(func() {
		if consumer != nil {
			consumer.Close()
		}
	})
        go config.StartListener(appName, viper.GetString("amqp_server_url"), viper.GetString("config_event_bus"))

	service.StartWebServer(viper.GetString("server_port"))
}

func onMessage(delivery amqp.Delivery) {
	fmt.Printf("Got a message: %v\n", string(delivery.Body))
}

func initializeMessaging() {
	if !viper.IsSet("amqp_server_url") {
		panic("No 'broker_url' set in configuration, cannot start")
	}
	consumer = &messaging.MessagingConsumer{}
	consumer.ConnectToBroker(viper.GetString("amqp_server_url"))
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
