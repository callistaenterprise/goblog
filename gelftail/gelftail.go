package main

import (
	"encoding/json"
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/callistaenterprise/goblog/gelftail/aggregator"
	"github.com/callistaenterprise/goblog/gelftail/transformer"
	"io/ioutil"
	"net"
	"os"
	"sync"
)

var authToken = ""
var port *string

func init() {
	data, err := ioutil.ReadFile("token.txt")
	if err != nil {
		msg := "Cannot find token.txt that should contain our Loggly token"
		logrus.Errorln(msg)
		panic(msg)
	}
	authToken = string(data)

	port = flag.String("port", "12202", "UDP port for the gelftail")
	flag.Parse()
}

func main() {
	logrus.Println("Starting Gelf-tail server...")

	ServerConn := startUDPServer(*port) // Remember to dereference the pointer for our "port" flag
	defer ServerConn.Close()

	var bulkQueue = make(chan []byte, 1) // Buffered channel to put log statements ready for LaaS upload into

	go aggregator.Start(bulkQueue, authToken)        // Start goroutine that'll collect and then upload batches of log statements
	go listenForLogStatements(ServerConn, bulkQueue) // Start listening for UDP traffic

	logrus.Infoln("Started Gelf-tail server")
	// Block indefinitely
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func startUDPServer(port string) *net.UDPConn {
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+port)
	checkError(err)

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	checkError(err)

	return ServerConn
}

func checkError(err error) {
	if err != nil {
		logrus.Println("Error: ", err)
		os.Exit(0)
	}
}

func listenForLogStatements(ServerConn *net.UDPConn, bulkQueue chan []byte) {
	buf := make([]byte, 8192)
	var item map[string]interface{}
	for {
		n, _, err := ServerConn.ReadFromUDP(buf)
		if err != nil {
			logrus.Errorln("Problem reading UDP message into buffer: %v\n", err.Error())
			continue
		}

		err = json.Unmarshal(buf[0:n], &item)
		if err != nil {
			logrus.Errorln("Problem unmarshalling log message into JSON: " + err.Error())
			continue
		}
		processedLogMessage, err := transformer.ProcessLogStatement(item)
		if err != nil {
			logrus.Printf("Problem parsing message: %v", string(buf[0:n]))
		} else {
			bulkQueue <- processedLogMessage
		}
		item = nil
	}
}
