package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
)

var authToken = "13fed9fe-a64b-425c-8233-ebcc8bc79f57"

func main() {

	fmt.Println("Starting Gelf-tail server")
	port := flag.String("port", "12202", "UDP port for the gelftail")
	flag.Parse()

	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+*port)
	checkError(err)

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	checkError(err)
	defer ServerConn.Close()

	var bulkQueue = make(chan []byte, 1)
	go startCollector(bulkQueue)

	buf := make([]byte, 8192)
	var item map[string]interface{}
	for {
		n, _, err := ServerConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Errorf("Problem reading UDP message into buffer: %v", err.Error())
			continue
		}
		fmt.Println(string(buf[0:n]))
		json.Unmarshal(buf[0:n], &item)
		err = processLogStatement(item, bulkQueue)
		if err != nil {
			fmt.Printf("Problem parsing message: %v", string(buf[0:n]))
		}
		item = nil
	}
}

func processLogStatement(item map[string]interface{}, bulkQueue chan []byte) error {
	// Extract the short_message, print and parse it:
	if item["short_message"] != nil {
		shortMessageString := item["short_message"].(string)
		var shortMessage map[string]interface{}
		err := json.Unmarshal([]byte(shortMessageString), &shortMessage)
		if err != nil {
			fmt.Printf("Error parsing short_message: %v\n", err.Error())
			return fmt.Errorf("Error parsing short_message.")
		}

		// Add the level and msg fields to the "main" one. Remove short_message
		if shortMessage != nil {
			item["msg"] = shortMessage["msg"].(string)
			item["level"] = shortMessage["level"].(string)
			delete(item, "short_message")
		} else {
			fmt.Println("Found log item with unparsable short_message: " + shortMessageString)
			return fmt.Errorf("Found log item with unparsable short_message.")
		}

		finalMessage, err := json.Marshal(item)
		bulkQueue <- finalMessage
		return nil
	} else {

		return fmt.Errorf("Could not process log statement.")
	}
}

func startCollector(bulkQueue chan []byte) {
	buf := new(bytes.Buffer)
	for {
		msg := <-bulkQueue
		buf.Write(msg)
		buf.WriteString("\n")

		size := buf.Len()
		if size > 1024 {
			sendBulk(*buf)
			buf.Reset()
		} else {
			//fmt.Printf("Buffer size not large enough yet (%v), waiting for more data.\n", size)
		}
	}
}

var client = &http.Client{}

// //"content-type:text/plain" -d '{ "message" : "hello" }' http://logs-01.loggly.com/inputs/13fed9fe-a64b-425c-8233-ebcc8bc79f57/tag/http/
func sendBulk(buffer bytes.Buffer) {
	req, err := http.NewRequest("POST", "http://logs-01.loggly.com/inputs/"+authToken+"/tag/http/", bytes.NewReader(buffer.Bytes()))
	if err != nil {
		fmt.Println("Error creating bulk request: " + err.Error())
	}
	req.Header.Add("context-type", "text/plain")
	resp, err := client.Do(req)

        if err != nil || resp.StatusCode != 200 {
		fmt.Println("Error sending bulk: " + err.Error())
                return
	}
	// fmt.Printf("Successfully sent batch of %v bytes to Loggly\n", buffer.Len())
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}
