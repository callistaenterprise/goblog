package service

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

// StartWebServer starts a web server at the designated port.
func StartWebServer(serviceName, port string) {

	r := NewRouter(serviceName)
	http.Handle("/", r)

	logrus.Infof("Starting HTTP service at %v", port)
	err := http.ListenAndServe(":"+port, nil)

	if err != nil {
		logrus.Errorln("An error occured starting HTTP listener at port " + port)
		logrus.Errorln("Error: " + err.Error())
	}
}
