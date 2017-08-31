package circuitbreaker

import (
	"encoding/json"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/spf13/viper"
	"net"
	"net/http"
	"time"
	"github.com/callistaenterprise/goblog/common/messaging"
	"github.com/Sirupsen/logrus"
	"github.com/callistaenterprise/goblog/common/util"
	"github.com/eapache/go-resiliency/retrier"
	"io/ioutil"
        "fmt"
)

var client http.Client

func Call(breakerName string, url string, method string) ([]byte, error) {
	output := make(chan []byte, 1)
	errors := hystrix.Go(breakerName, func() error {
		// talk to other services
		req, _ := http.NewRequest(method, url, nil)

		r := retrier.New(retrier.ConstantBackoff(3, 100*time.Millisecond), nil)
		times := 0
		err := r.Run(func() error {
                        logrus.Infof("Calling %v\n", times)
			resp, err := client.Do(req)
                        logrus.Infof("Returning %v\n", resp.StatusCode)
			if err == nil && resp.StatusCode < 299 {
				responseBody, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					output <- responseBody
					return nil
				}
                                return err
			} else if err != nil {
                                times++
                                logrus.Errorf("Attempt failed. Retrier: %v", times)
                        } else {
                                err = fmt.Errorf("Status was %v", resp.StatusCode)
                        }

			return err
		})

		// For hystrix, forward the err from the retrier. It's nil if OK.
		return err
	}, func(err error) error {
		logrus.Errorf("Breaker %v opened, error: %v", breakerName, err.Error())
		return err
	})

	select {
	case out := <-output:
		return out, nil

	case err := <-errors:
		return nil, err
	}
	return nil, nil
}


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
	logrus.Println("Launched hystrixStreamHandler at 8181")

	// Publish presence on RabbitMQ
	publishDiscoveryToken(amqpClient)
}

func publishDiscoveryToken(amqpClient messaging.IMessagingClient) {
	token := DiscoveryToken{
		State:   "UP",
		Address: util.GetIP(),
	}
	json, _ := json.Marshal(token)
	go func() {
		for {
			amqpClient.PublishOnQueue(json, "discovery")  //SendMessage(string(json), "application/json", "discovery")
			time.Sleep(time.Second * 30)
		}
	}()
}

func resolveProperty(command string, prop string) int {
	if viper.IsSet("hystrix.command." + command + "." + prop) {
		return viper.GetInt("hystrix.command." + command + "." + prop)
	} else {
		return getDefaultHystrixConfigPropertyValue(prop)
	}
}
func getDefaultHystrixConfigPropertyValue(prop string) int {
	switch prop {
	case "Timeout":
		return hystrix.DefaultTimeout
	case "MaxConcurrentRequests":
		return hystrix.DefaultMaxConcurrent
	case "RequestVolumeThreshold":
		return hystrix.DefaultVolumeThreshold
	case "SleepWindow":
		return hystrix.DefaultSleepWindow
	case "ErrorPercentThreshold":
		return hystrix.DefaultErrorPercentThreshold
	}
	panic("Got unknown hystrix property: " + prop + ". Panicing!")
}

type DiscoveryToken struct {
	State string `json:"state"`   // UP, RUNNING, DOWN ??
	Address string `json:"address"`
}
