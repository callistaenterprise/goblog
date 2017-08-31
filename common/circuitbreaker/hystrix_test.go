package circuitbreaker

import (
        "gopkg.in/h2non/gock.v1"
        . "github.com/smartystreets/goconvey/convey"
        "testing"
)



func TestCallUsingResilienceAllFails(t *testing.T) {
        defer gock.Off()

        for a := 0 ; a < 4; a++ {
                gock.New("http://quotes-service:8080").
                        Get("/api/quote").
                        MatchParam("strength", "4").
                        Reply(500)
        }

        Convey("Given a Call request", t, func() {

                Convey("When", func() {
                        bytes, err := Call("TEST", "http://quotes-service:8080/api/quote?strength=4", "GET")

                        Convey("Then", func() {
                                So(err, ShouldNotBeNil)
                                So(bytes, ShouldBeNil)
                        })
                })
        })
}


func TestCallUsingResilienceLastSucceedsFails(t *testing.T) {
        defer gock.Off()

        for a := 0 ; a < 3; a++ {
                gock.New("http://quotes-service:8080").
                        Get("/api/quote").
                        MatchParam("strength", "4").
                        Reply(500)
        }
        gock.New("http://quotes-service:8080").
                Get("/api/quote").
                MatchParam("strength", "4").
                Reply(200).
                BodyString(`{"quote":"May the source be with you. Always.","ipAddress":"10.0.0.5:8080","language":"en"}`)


        Convey("Given a Call request", t, func() {

                Convey("When", func() {
                        bytes, err := Call("TEST", "http://quotes-service:8080/api/quote?strength=4", "GET")

                        Convey("Then", func() {
                                So(err, ShouldBeNil)
                                So(bytes, ShouldNotBeNil)
                        })
                })
        })
}
