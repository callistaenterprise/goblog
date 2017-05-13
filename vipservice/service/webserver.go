package service

import (
        "net/http"
        "github.com/Sirupsen/logrus"
)

func StartWebServer(port string) {
        r := NewRouter()
        http.Handle("/", r)

        logrus.Println("Starting HTTP service at " + port)
        err := http.ListenAndServe(":" + port, nil)

        if err != nil {
                logrus.Println("An error occured starting HTTP listener at port " + port)
                logrus.Println("Error: " + err.Error())
        }
}
