package main

import (
        "fmt"
        "net"
        "os"
        "encoding/json"
        "flag"
)

/*
{"version":"1.1","host":"swarm-manager-0","short_message":"2017/02/17 21:05:48 Starting HTTP service at 6767","timestamp":1.487365548315e+09,"level":3,"_command":"./accountservice-linux-amd64 -profile=test","_container_id":"74bfb81aef3bf43c9e4cefb1ea77e45811f11f655233df6d2b9efe9ea167b679","_container_name":"accountservice.3.1dbe2shvmmxe99bwm97nsjda1","_created":"2017-02-17T21:05:46.501990081Z","_image_id":"sha256:c3e9103dbf596a00264ee1a729548f0b7003a51f67c9cdab50b099d363bb5c41","_image_name":"someprefix/accountservice:latest","_tag":"74bfb81aef3b"}
 */
var levels = map[int]string {
        0: "DEBUG",
        1: "INFO",
        2: "WARN",
        3: "ERROR",
        4: "FATAL",
        5: "UNKNOWN",
}

func main() {
        fmt.Println("Starting Gelf-tail server")
        port := flag.String("port", "12202", "UDP port for the gelftail")
        flag.Parse()

        ServerAddr, err := net.ResolveUDPAddr("udp", ":" + *port)
        checkError(err)

        ServerConn, err := net.ListenUDP("udp", ServerAddr)
        checkError(err)
        defer ServerConn.Close()

        buf := make([]byte, 8192)
        var item map[string]interface{}
        for {
                n, _, err := ServerConn.ReadFromUDP(buf)
                fmt.Println(string(buf[0:n]))
                json.Unmarshal(buf[0:n], &item)
                level := int(item["level"].(float64))
                fmt.Println(item["_created"].(string)[0:23] + " | " + item["_container_name"].(string)[0:16] + " | " + levels[level] + "|" + item["short_message"].(string))

                if err != nil {
                        fmt.Println("Error: ", err)
                }
                item = nil
        }
}

func checkError(err error) {
        if err != nil {
                fmt.Println("Error: ", err)
                os.Exit(0)
        }
}