package service

import (
        "net/http"
        log "github.com/Sirupsen/logrus"
)

func StartWebServer(port string) {

        r := NewRouter()
        http.Handle("/", r)

        log.Infof("Starting HTTP service at %v", port)
        err := http.ListenAndServe(":" + port, nil)

        if err != nil {
                log.Errorln("An error occured starting HTTP listener at port " + port)
                log.Errorln("Error: " + err.Error())
        }
}
