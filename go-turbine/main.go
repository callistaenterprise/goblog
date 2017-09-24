package main

import (
        msg "github.com/callistaenterprise/goblog/common/messaging"
        "github.com/spf13/viper"
        "flag"
        "github.com/streadway/amqp"
        "github.com/callistaenterprise/goblog/go-turbine/model"
        "encoding/json"
        "github.com/Sirupsen/logrus"
        "time"
        "net/http"
        "bufio"
        "sync"
        "github.com/buger/jsonparser"
)

var msgClient msg.IMessagingClient

var registry map[string]model.Instance

var producers []model.HystrixProducer

var hystrixChan = make(chan []byte, 10000)

func init() {
        amqpBrokerUrl := flag.String("amqp.connection.url", "amqp://guest:guest@192.168.99.100:5672/", "Address to rabbitmq")

        flag.Parse()
        viper.Set("amqp.connection.url", *amqpBrokerUrl)
}

func main() {
        msgClient = &msg.MessagingClient{}
        msgClient.ConnectToBroker(viper.GetString("amqp.connection.url"))
        msgClient.SubscribeToQueue("discovery", "go-turbine", handleToken)



        // Scan func, makes sure all "producers" are connected.
        //go func() {
        //   for {
        //           for _, producer := range producers {
        //
        //           }
        //           time.Sleep(time.Second)
        //   }
        //}()

        go func() {
                for {
                        select {
                        case out := <-hystrixChan:

                                val, _, _, _ := jsonparser.Get(out, "type")
                                if string(val) == "HystrixCommand" {
                                        hc := model.CircuitBreaker{}
                                        err := json.Unmarshal(out[5:], &hc)
                                        if err != nil {
                                                logrus.Errorf("Error parsing JSON: %v", err)
                                        }
                                        logrus.Infof("Message: Name: %v Group: %v", hc.Name, hc.Group)
                                } else {
                                        logrus.Infof("Got other (%v)", string(val))
                                }

                        default:
                                time.Sleep(time.Millisecond * 100)
                        }
                }
        }()

        // TEMP
        connect(model.HystrixProducer{Ip: "192.168.99.100", State: "UP"})

        logrus.Infoln("Started go-turbine")
        // Block indefinitely
        wg := sync.WaitGroup{}
        wg.Add(1)
        wg.Wait()
}

func handleToken(d amqp.Delivery) {
        logrus.Infoln("Got delivery in go-turbine. Body is %v", string(d.Body))
        var token model.DiscoveryToken
        err := json.Unmarshal(d.Body, &token)
        if err != nil {
                logrus.Errorln("Failed to unmarshal token, error: %v", err.Error())
                return
        }
        if token.State == "UP" {
                var found bool = false
                for _, item := range producers {
                        if item.Ip == token.Address && item.State != "UP" {
                                producer := model.HystrixProducer{State: "UP", Ip: token.Address}
                                producers = append(producers, producer)
                                logrus.Infof("Added existing HystrixProducer %v", token.Address)
                                connect(producer)
                                found = true
                                break
                        }
                }

                if !found {
                        producer := model.HystrixProducer{State: "UP", Ip: token.Address}
                        producers = append(producers, producer)
                        logrus.Infof("Added new HystrixProducer %v", token.Address)
                        connect(producer)
                }
        } else if token.State == "DOWN" {
                for index, item := range producers {
                        if item.Ip == token.Address {
                                producers = remove(producers, index)
                                logrus.Infof("Removed HystrixProducer %v", token.Address)
                                break
                        }
                }
        }
}

func connect(producer model.HystrixProducer) {
        // Try to connect to the hystrix stream...
        url := "http://" + producer.Ip + ":8181/hystrix.stream"
        logrus.Infof("Connecting to %v", url)
        resp, err := http.Get(url)
        if err == nil {
                scanner := bufio.NewScanner(resp.Body)

                for {
                        if scanner.Scan() {
                                data := scanner.Bytes()
                                hystrixChan <- data
                        } else {
                                logrus.Infoln("No data, sleeping")
                                time.Sleep(time.Second * 1)
                        }
                }
        }  else {
                logrus.Errorf("Unable to connect to %v, msg: %v", url, err.Error())
        }

}

func remove(slice []model.HystrixProducer, s int) []model.HystrixProducer {
        return append(slice[:s], slice[s + 1:]...)
}
