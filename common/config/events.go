package config

import (
	"encoding/json"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

func HandleRefreshEvent(d amqp.Delivery) {
	body := d.Body
	consumerTag := d.ConsumerTag
	updateToken := &UpdateToken{}
	err := json.Unmarshal(body, updateToken)
	if err != nil {
		logrus.Printf("Problem parsing UpdateToken: %v", err.Error())
	} else {
		if strings.Contains(updateToken.DestinationService, consumerTag) {
			logrus.Println("Reloading Viper config from Spring Cloud Config server")

			// Consumertag is same as application name.
			LoadConfigurationFromBranch(
				viper.GetString("configServerUrl"),
				consumerTag,
				viper.GetString("profile"),
				viper.GetString("configBranch"))
		}
	}
}

// {"type":"RefreshRemoteApplicationEvent","timestamp":1494514362123,"originService":"config-server:docker:8888","destinationService":"xxxaccoun:**","id":"53e61c71-cbae-4b6d-84bb-d0dcc0aeb4dc"}
type UpdateToken struct {
	Type               string `json:"type"`
	Timestamp          int    `json:"timestamp"`
	OriginService      string `json:"originService"`
	DestinationService string `json:"destinationService"`
	Id                 string `json:"id"`
}
