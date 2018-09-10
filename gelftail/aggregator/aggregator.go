package aggregator

import (
	"bytes"
	"fmt"
	"github.com/Sirupsen/logrus"
	"net/http"
)

var client = &http.Client{}
var logglyUrl = "https://logs-01.loggly.com/inputs/%s/tag/http/"
var url string

func Start(bulkQueue chan []byte, authToken string) {
	url = fmt.Sprintf(logglyUrl, authToken)
	buf := new(bytes.Buffer)
	for {
		msg := <-bulkQueue
		buf.Write(msg)
		buf.WriteString("\n")

		size := buf.Len()
		if size > 1024 {
			sendBulk(*buf)
			buf.Reset()
		}
	}
}

func sendBulk(buffer bytes.Buffer) {

	req, err := http.NewRequest("POST", url,
		bytes.NewReader(buffer.Bytes()))

	if err != nil {
		logrus.Errorln("Error creating bulk upload HTTP request: " + err.Error())
		return
	}
	resp, err := client.Do(req)

	if err != nil || resp.StatusCode != 200 {
		logrus.Errorln("Error sending bulk: " + err.Error())
		return
	}
	logrus.Debugf("Successfully sent batch of %v bytes to Loggly\n", buffer.Len())
}
