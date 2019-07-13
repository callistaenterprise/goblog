package circuitbreaker

import (
	"context"
	"encoding/json"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/callistaenterprise/goblog/common/messaging"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/callistaenterprise/goblog/common/util"
	"github.com/eapache/go-resiliency/retrier"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

// Client to do http requests with
var Client http.Client

// RETRIES is the number of retries to do in the retrier.
var retries = 3

// CallUsingCircuitBreaker performs a HTTP call inside a circuit breaker.
func CallUsingCircuitBreaker(ctx context.Context, breakerName string, url string, method string) ([]byte, error) {
	output := make(chan []byte, 1)
	errors := hystrix.Go(breakerName, func() error {

		req, _ := http.NewRequest(method, url, nil)
		tracing.AddTracingToReqFromContext(ctx, req)
		err := callWithRetries(req, output)

		return err // For hystrix, forward the err from the retrier. It's nil if OK.
	}, func(err error) error {
		logrus.Errorf("In fallback function for breaker %v, error: %v", breakerName, err.Error())
		circuit, _, _ := hystrix.GetCircuit(breakerName)
		logrus.Errorf("Circuit state is: %v", circuit.IsOpen())
		return err
	})

	select {
	case out := <-output:
		logrus.Debugf("Call in breaker %v successful", breakerName)
		return out, nil

	case err := <-errors:
		logrus.Debugf("Got error on channel in breaker %v. Msg: %v", breakerName, err.Error())
		return nil, err
	}
}

// PerformHTTPRequestCircuitBreaker performs the supplied http.Request within a circuit breaker.
func PerformHTTPRequestCircuitBreaker(ctx context.Context, breakerName string, req *http.Request) ([]byte, error) {
	output := make(chan []byte, 1)
	errors := hystrix.Go(breakerName, func() error {
		tracing.AddTracingToReqFromContext(ctx, req)
		err := callWithRetries(req, output)
		return err // For hystrix, forward the err from the retrier. It's nil if OK.
	}, func(err error) error {
		logrus.Errorf("In fallback function for breaker %v, error: %v", breakerName, err.Error())
		return err
	})

	select {
	case out := <-output:
		logrus.Debugf("Call in breaker %v successful", breakerName)
		return out, nil

	case err := <-errors:
		logrus.Errorf("Got error on channel in breaker %v. Msg: %v", breakerName, err.Error())
		return nil, err
	}
}

func callWithRetries(req *http.Request, output chan []byte) error {

	r := retrier.New(retrier.ConstantBackoff(retries, 100*time.Millisecond), nil)
	attempt := 0
	err := r.Run(func() error {
		attempt++
		resp, err := Client.Do(req)
		if err == nil && resp.StatusCode < 299 {
			responseBody, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				output <- responseBody
				return nil
			}
			return err
		} else if err == nil {
			err = fmt.Errorf("Status was %v", resp.StatusCode)
		}

		logrus.Errorf("Retrier failed, attempt %v", attempt)

		return err
	})
	return err
}

// ConfigureHystrix sets up hystrix circuit breakers.
func ConfigureHystrix(commands []string, amqpClient messaging.IMessagingClient) {

	for _, command := range commands {
		hystrix.ConfigureCommand(command, hystrix.CommandConfig{
			Timeout:                resolveProperty(command, "Timeout"),
			MaxConcurrentRequests:  resolveProperty(command, "MaxConcurrentRequests"),
			ErrorPercentThreshold:  resolveProperty(command, "ErrorPercentThreshold"),
			RequestVolumeThreshold: resolveProperty(command, "RequestVolumeThreshold"),
			SleepWindow:            resolveProperty(command, "SleepWindow"),
		})
		logrus.Printf("Circuit %v settings: %v", command, hystrix.GetCircuitSettings()[command])
	}

	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	go http.ListenAndServe(net.JoinHostPort("", "8181"), hystrixStreamHandler)
	logrus.Infoln("Launched hystrixStreamHandler at 8181")

	// Publish presence on RabbitMQ
	publishDiscoveryToken(amqpClient)
}

// Deregister publishes a Deregister token to Hystrix/Turbine
func Deregister(amqpClient messaging.IMessagingClient) {
	ip, err := util.ResolveIPFromHostsFile()
	if err != nil {
		ip = util.GetIPWithPrefix("10.0.")
	}
	token := DiscoveryToken{
		State:   "DOWN",
		Address: ip,
	}
	bytes, _ := json.Marshal(token)
	amqpClient.PublishOnQueue(bytes, "discovery")
	logrus.Infoln("Sent deregistration token over SpringCloudBus")
}

func publishDiscoveryToken(amqpClient messaging.IMessagingClient) {
	ip, err := util.ResolveIPFromHostsFile()
	if err != nil {
		ip = util.GetIPWithPrefix("10.0.")
	}
	token := DiscoveryToken{
		State:   "UP",
		Address: ip,
	}
	bytes, _ := json.Marshal(token)
	go func() {
		for {
			amqpClient.PublishOnQueue(bytes, "discovery")
			amqpClient.PublishOnQueue(bytes, "discovery")
			time.Sleep(time.Second * 30)
		}
	}()
}

func resolveProperty(command string, prop string) int {
	if viper.IsSet("hystrix.command." + command + "." + prop) {
		return viper.GetInt("hystrix.command." + command + "." + prop)
	}
	return getDefaultHystrixConfigPropertyValue(prop)

}
func getDefaultHystrixConfigPropertyValue(prop string) int {
	switch prop {
	case "Timeout":
		return 1000 //hystrix.DefaultTimeout
	case "MaxConcurrentRequests":
		return 200 //hystrix.DefaultMaxConcurrent
	case "RequestVolumeThreshold":
		return hystrix.DefaultVolumeThreshold
	case "SleepWindow":
		return hystrix.DefaultSleepWindow
	case "ErrorPercentThreshold":
		return hystrix.DefaultErrorPercentThreshold
	}
	panic("Got unknown hystrix property: " + prop + ". Panicing!")
}

// DiscoveryToken defines a struct for transmitting the state of a hystrix stream producer.
type DiscoveryToken struct {
	State   string `json:"state"` // UP, RUNNING, DOWN ??
	Address string `json:"address"`
}
