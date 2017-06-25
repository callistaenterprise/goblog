package service

import (
        "net/http"
        "log"
        ct "github.com/eriklupander/cloudtoolkit"
)

func StartWebServer(port string) {

        r := NewRouter()
        http.Handle("/", r)
        ct.Log.Println("Starting HTTP service at " + port)
        err := http.ListenAndServe(":" + port, nil)
        if err != nil {
                log.Println("An error occured starting HTTP listener at port " + port)
                log.Println("Error: " + err.Error())
        }
}
