package service

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

func StartWebServer(port string) {

	r := NewRouter()
	http.Handle("/", r)

	logrus.Infof("Starting HTTP service at %v", port)
	err := http.ListenAndServe(":"+port, nil)

	if err != nil {
		logrus.Errorln("An error occured starting HTTP listener at port " + port)
		logrus.Errorln("Error: " + err.Error())
	}
}
