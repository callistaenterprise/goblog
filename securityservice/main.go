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
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/callistaenterprise/goblog/common/circuitbreaker"
	"github.com/callistaenterprise/goblog/common/config"
	"github.com/callistaenterprise/goblog/common/messaging"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/callistaenterprise/goblog/securityservice/service"
	ct "github.com/eriklupander/cloudtoolkit"
	"github.com/spf13/viper"
)

var appName = "securityservice"

var authServerDefaultUrl = "https://auth-server:9999/uaa/user"

var amqpClient *messaging.AmqpClient

func init() {
	profile := flag.String("profile", "test", "Environment profile, something similar to spring profiles")
	configServerURL := flag.String("configServerUrl", "http://configserver:8888", "Address to config server")
	configBranch := flag.String("configBranch", "master", "git branch to fetch configuration from")

	flag.Parse()

	viper.Set("profile", *profile)
	viper.Set("configServerURL", *configServerURL)
	viper.Set("configBranch", *configBranch)
}

func main() {
	start := time.Now().UTC()

	logrus.Infoln("Starting " + appName + "...")

	config.LoadConfigurationFromBranch(
		viper.GetString("configServerURL"),
		appName,
		viper.GetString("profile"),
		viper.GetString("configBranch"))

	initializeTracing()
	initializeMessaging()
	ct.InitOAuth2HandlerUsingUrl(authServerDefaultUrl)
	circuitbreaker.ConfigureHystrix([]string{"get_account_secured"}, amqpClient)

	service.ConfigureHttpClient() // Disable keep-alives so Docker Swarm can round-robin for us.

	go service.StartWebServer(viper.GetString("server_port")) // Starts HTTP service  (async)

	logrus.Infof("Started %v in %v\n", appName, time.Now().UTC().Sub(start))

	// Block...
	wg := sync.WaitGroup{} // Use a WaitGroup to block main() exit
	wg.Add(1)
	wg.Wait()
}

func initializeTracing() {
	if !viper.IsSet("zipkin_service_url") {
		panic("No 'zipkin_service_url' set in configuration, cannot start")
	}
	tracing.InitTracing(viper.GetString("zipkin_service_url"), appName)
}

func initializeMessaging() {
	if !viper.IsSet("amqp_server_url") {
		panic("No 'amqp_server_url' set in configuration, cannot start")
	}

	amqpClient = &messaging.AmqpClient{}
	amqpClient.ConnectToBroker(viper.GetString("amqp_server_url"))
	amqpClient.Subscribe(viper.GetString("config_event_bus"), "topic", appName, config.HandleRefreshEvent)
}
