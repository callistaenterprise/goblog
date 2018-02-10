package main

import (
	"crypto/tls"
	"flag"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
)

var Log = logrus.New()

var baseAddr string
var zuul = false

var startindex = 0

func main() {

	usersPtr := flag.Int("users", 10, "Number of users")
	delayPtr := flag.Int("delay", 1000, "Delay between calls per user")
	zuulPtr := flag.Bool("zuul", false, "Route traffic through zuul")
	baseAddrPtr := flag.String("baseAddr", "192.168.99.100", "Base address of your Swarm cluster")

	flag.Parse()

	baseAddr = *baseAddrPtr
	zuul = *zuulPtr
	users := *usersPtr
	var _ int = *delayPtr
	wg2 := sync.WaitGroup{} // Use a WaitGroup to block main() exit
	wg2.Add(8)
	for i := 0; i < 8; i++ {

		go func(globalIdx int, waitGroup *sync.WaitGroup) {
			for j := 0; j < 12; j++ {
				createAccount((globalIdx * 1250) + j)
			}
			waitGroup.Done()
		}(i, &wg2)
	}
	wg2.Wait()

	for i := 0; i < users; i++ {
		//go securedTest()
		go readStuff()
		go writeStuff()
	}

	// Block...
	wg := sync.WaitGroup{} // Use a WaitGroup to block main() exit
	wg.Add(1)
	wg.Wait()

}

var defaultTransport http.RoundTripper = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var ids = make([]string, 0)

func createAccount(index int) {
	jsonstr := []byte("{\"name\":\"Firstname-" + strconv.Itoa(index) + " Lastname-" + strconv.Itoa(index) + "\"}")
	req, _ := http.NewRequest("POST", "http://192.168.99.100:6767/accounts", bytes.NewReader(jsonstr))
	resp, err := defaultTransport.RoundTrip(req)
	if err != nil {
		panic(err.Error())
	}
	body, err := ioutil.ReadAll(resp.Body)
	respMap := make(map[string]string)
	json.Unmarshal(body, &respMap)

	ids = append(ids, respMap["ID"])
	fmt.Println("Created account " + strconv.Itoa(index))
	startindex = index
}

func writeStuff() {
	for {
		startindex += 1
		createAccount(startindex)
		time.Sleep(time.Second * 4)
	}
}

func readStuff() {
	var url string
	if zuul {
		Log.Println("Using HTTPS through ZUUL")
		url = "https://" + baseAddr + ":8765/api/accounts/"
	} else {
		url = "http://" + baseAddr + ":6767/accounts/"
	}
	m := make(map[string]interface{})
	length := len(ids)
	for {
		accountIndex := rand.Intn(length)
		accountId := ids[accountIndex]
		serviceUrl := url + accountId

		var DefaultTransport http.RoundTripper = &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: false,
		}
		req, _ := http.NewRequest("GET", serviceUrl, nil)
		resp, err := DefaultTransport.RoundTrip(req)

		if err != nil {
			panic(err)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		printPretty(body, m)
		time.Sleep(time.Second * 1)
	}

}
func printPretty(body []byte, m map[string]interface{}) {
	if body == nil {
		return
	}
	err := json.Unmarshal(body, &m)
	if err != nil {
		return
	}
	quote := m["quote"].(map[string]interface{})["quote"].(string)
	quoteIp := m["quote"].(map[string]interface{})["ipAddress"].(string)
	quoteIp = quoteIp[strings.IndexRune(quoteIp, '/')+1:]

	imageUrl := m["imageData"].(map[string]interface{})["url"].(string)
	imageServedBy := m["imageData"].(map[string]interface{})["servedBy"].(string)

	fmt.Print("|" + m["name"].(string) + "\t|" + m["servedBy"].(string) + "\t|")
	fmt.Print(PadRight(quote, " ", 32) + "\t|" + quoteIp + "\t|")
	fmt.Println(PadRight(imageUrl, " ", 28) + "\t|" + imageServedBy + "\t|")

}

func PadRight(str, pad string, lenght int) string {
	for {
		str += pad
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}
