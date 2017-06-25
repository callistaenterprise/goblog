package cloudtoolkit

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
)

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
