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

	"github.com/sirupsen/logrus"
	"github.com/callistaenterprise/goblog/common/config"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/callistaenterprise/goblog/imageservice/dbclient"
	"github.com/callistaenterprise/goblog/imageservice/service"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

var appName = "imageservice"

func init() {
	profile := flag.String("profile", "test", "Environment profile, something similar to spring profiles")
	configServerURL := flag.String("configServerUrl", "http://configserver:8888", "Address to config server")
	configBranch := flag.String("configBranch", "master", "git branch to fetch configuration from")

	flag.Parse()

	viper.Set("service_name", appName)
	viper.Set("profile", *profile)
	viper.Set("configServerUrl", *configServerURL)
	viper.Set("configBranch", *configBranch)
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Infof("Starting %v", appName)

	start := time.Now().UTC()
	config.LoadConfigurationFromBranch(viper.GetString("configServerUrl"), appName, viper.GetString("profile"), viper.GetString("configBranch"))
	initializeTracing()
	service.DBClient = &dbclient.GormClient{}
	service.DBClient.SetupDB(viper.GetString("cockroachdb_conn_url"))

	if viper.GetString("profile") == "dev" {
		service.DBClient.SeedAccountImages()
	}

	go service.StartWebServer(viper.GetString("server_port")) // Starts HTTP service  (async)

	handleSigterm(func() {
		logrus.Infoln("Captured Ctrl+C")
		service.DBClient.Close()
	})

	logrus.Infof("Started %v in %v", appName, time.Now().UTC().Sub(start))
	// Block...
	wg := sync.WaitGroup{} // Use a WaitGroup to block main() exit
	wg.Add(1)
	wg.Wait()
}
func initializeTracing() {
	tracing.InitTracing(viper.GetString("zipkin_server_url"), appName)
}

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
