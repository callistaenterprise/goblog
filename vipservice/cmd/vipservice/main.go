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
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/callistaenterprise/goblog/common/messaging"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/callistaenterprise/goblog/vipservice/cmd"
	"github.com/callistaenterprise/goblog/vipservice/internal/app/service"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var appName = "vipservice"

var messagingClient messaging.IMessagingClient

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Println("Starting " + appName + "...")

	// Initialize config struct and populate it froms env vars and flags.
	cfg := cmd.DefaultConfiguration()
	arg.MustParse(cfg)

	srv := service.NewServer(cfg)
	srv.SetupRoutes()

	initializeTracing(cfg)
	initializeMessaging(cfg)

	// Makes sure connection is closed when service exits.
	handleSigterm(func() {
		if messagingClient != nil {
			messagingClient.Close()
		}
	})
	srv.Start()
}

func initializeTracing(cfg *cmd.Config) {
	tracing.InitTracing(cfg.ZipkinServerUrl, appName)
}

func onMessage(delivery amqp.Delivery) {
	logrus.Infof("Got a message: %v\n", string(delivery.Body))

	defer tracing.StartTraceFromCarrier(delivery.Headers, "vipservice#onMessage").Finish()

	// Experimental!
	//carrier := make(opentracing.HTTPHeadersCarrier)
	//for k, v := range delivery.Headers {
	//        carrier.Set(k, v.(string))
	//}
	//
	//clientContext, err := tracing.Tracer.Extract(opentracing.HTTPHeaders, carrier)
	//var span opentracing.Span
	//if err == nil {
	//        span = tracing.Tracer.StartSpan(
	//                "vipservice onMessage", ext.RPCServerOption(clientContext))
	//} else {
	//        span = tracing.Tracer.StartSpan("vipservice onMessage")
	//}
	time.Sleep(time.Millisecond * 10)
}

func initializeMessaging(cfg *cmd.Config) {
	if cfg.AmqpConfig.ServerUrl == "" {
		panic("No 'broker_url' set in configuration, cannot start")
	}
	messagingClient = &messaging.AmqpClient{}
	messagingClient.ConnectToBroker(cfg.AmqpConfig.ServerUrl)

	// Call the subscribe method with queue name and callback function
	err := messagingClient.SubscribeToQueue("vip_queue", appName, onMessage)
	failOnError(err, "Could not start subscribe to vip_queue")

	logrus.Infoln("Successfully initialized messaging for vipservice")
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

func failOnError(err error, msg string) {
	if err != nil {
		logrus.Errorf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
