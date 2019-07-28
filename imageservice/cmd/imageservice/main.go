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
	"github.com/alexflint/go-arg"
	"github.com/callistaenterprise/goblog/imageservice/cmd"
	"sync"
	"time"

	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/callistaenterprise/goblog/imageservice/internal/app/dbclient"
	"github.com/callistaenterprise/goblog/imageservice/internal/app/service"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var appName = "imageservice"

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Infof("Starting %v", appName)

	start := time.Now().UTC()

	// Initialize config struct and populate it froms env vars and flags.
	cfg := cmd.DefaultConfiguration()
	arg.MustParse(cfg)

	server := service.NewServer(service.NewHandler(dbclient.NewGormClient(cfg)), cfg)
	server.SetupRoutes()

	initializeTracing(cfg)

	if cfg.Environment == "dev" {
		server.SeedAccountImages()
	}

	handleSigterm(func() {
		logrus.Infoln("Captured Ctrl+C")
		server.Close()
	})

	go server.Start()
	logrus.Infof("Started %v in %v", appName, time.Now().UTC().Sub(start))
	// Block...
	wg := sync.WaitGroup{} // Use a WaitGroup to block main() exit
	wg.Add(1)
	wg.Wait()
}
func initializeTracing(cfg *cmd.Config) {
	tracing.InitTracing(cfg.ZipkinServerUrl, appName)
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
