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
	"github.com/Sirupsen/logrus"
	"net/url"
	"strings"
)

var Log = logrus.New()

var baseAddr string
var zuul bool

func main() {

	usersPtr := flag.Int("users", 10, "Number of users")
	delayPtr := flag.Int("delay", 1000, "Delay between calls per user")
	zuulPtr := flag.Bool("zuul", true, "Route traffic through zuul")
	baseAddrPtr := flag.String("baseAddr", "192.168.99.100", "Base address of your Swarm cluster")

	flag.Parse()

	baseAddr = *baseAddrPtr
	zuul = *zuulPtr
	users := *usersPtr
	var _ int = *delayPtr

	for i := 0; i < users; i++ {
		// go securedTest()
		go standardTest()
	}

	// Block...
	wg := sync.WaitGroup{} // Use a WaitGroup to block main() exit
	wg.Add(1)
	wg.Wait()

}

func getToken() string {

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Add("client_id", "acme")
	data.Add("scope", "webshop")
	data.Add("username", "user")
	data.Add("password", "password")
	req, err := http.NewRequest("POST", "https://"+baseAddr+":9999/uaa/oauth/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		panic(err.Error())
	}
	var DefaultTransport http.RoundTripper = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	headers := make(map[string][]string)
	headers["Authorization"] = []string{"Basic YWNtZTphY21lc2VjcmV0"}
	headers["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	req.Header = headers

	resp, err := DefaultTransport.RoundTrip(req)
	if err != nil {
		panic(err.Error())
	}
	if resp.StatusCode > 299 {
		panic("Call to get auth token returned status " + resp.Status)
	}
	respdata, _ := ioutil.ReadAll(resp.Body)
	m := make(map[string]interface{})
	err = json.Unmarshal(respdata, &m)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Got TOKEN: " + string(m["access_token"].(string)))
	return string(m["access_token"].(string))
}

var defaultTransport http.RoundTripper = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

func securedTest() {

	var token = getToken()
	for {
		accountId := rand.Intn(99) + 10000
		url := "https://" + baseAddr + ":8765/api/secured/account/" + strconv.Itoa(accountId)

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Add("Connection", "keep-alive")
		req.Header.Add("Keep-Alive", "timeout=10, max=5")
		resp, err := defaultTransport.RoundTrip(req)
		if resp.StatusCode != 200 {
			fmt.Println("Status: " + resp.Status)
		}
		if err != nil {
			fmt.Println("Error: " + err.Error())
			panic(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			fmt.Println("Body: " + string(body))
			panic(err)
		}
		m := make(map[string]string)
		json.Unmarshal(body, &m)
		Log.Println("Account: " + m["accountServedBy"] + "\tComposite: " + m["imageServedBy"] + "\tQuote: " + m["quoteServedBy"])
		time.Sleep(time.Second * 1)
	}
}

func standardTest() {
	var url string
	if zuul {
		Log.Println("Using HTTPS through ZUUL")
		url = "https://" + baseAddr + ":8765/api/accounts/"
	} else {
		url = "http://" + baseAddr + ":6767/accounts/"
	}
	m := make(map[string]interface{})
	for {
		accountId := rand.Intn(99) + 10000
		serviceUrl := url + strconv.Itoa(accountId)

		var DefaultTransport http.RoundTripper = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: false,
		}
		req, _ := http.NewRequest("GET", serviceUrl, nil)
		resp, err := DefaultTransport.RoundTrip(req)

		if err != nil {
			panic(err)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(body, &m)
		quote :=  m["quote"].(map[string]interface{})["quote"].(string)
		quoteIp := m["quote"].(map[string]interface{})["ipAddress"].(string)
		quoteIp = quoteIp[strings.IndexRune(quoteIp, '/') + 1 :]

		imageUrl := m["imageData"].(map[string]interface{})["url"].(string)
		imageServedBy := m["imageData"].(map[string]interface{})["servedBy"].(string)

		fmt.Print("|" + m["name"].(string) + "\t|" + m["servedBy"].(string) + "\t|")
		fmt.Print(PadRight(quote, " ", 32) + "\t|" + quoteIp + "\t|")
		fmt.Println(PadRight(imageUrl, " ", 28) + "\t|" + imageServedBy + "\t|")
		time.Sleep(time.Second * 1)
	}

}

func PadRight(str, pad string, lenght int) string {
	for {
		str += pad
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}
