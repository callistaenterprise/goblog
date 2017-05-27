package config

import "testing"
import (
        . "github.com/smartystreets/goconvey/convey"
        "gopkg.in/h2non/gock.v1"
        "github.com/spf13/viper"
)

var SERVICE_NAME = "accountservice"

func TestHandleRefreshEvent(t *testing.T) {
        // Configure initial viper values
        viper.Set("configServerUrl", "http://configserver:8888")
        viper.Set("profile", "test")
        viper.Set("configBranch", "master")

        // Mock the expected outgoing request for new config
        defer gock.Off()
        gock.New("http://configserver:8888").
                Get("/accountservice/test/master").
                Reply(200).
                BodyString(`{"name":"accountservice-test","profiles":["test"],"label":null,"version":null,"propertySources":[{"name":"file:/config-repo/accountservice-test.yml","source":{"server_port":6767,"server_name":"Accountservice RELOADED"}}]}`)

        Convey("Given a refresh event received, targeting our application", t, func() {
                var body = `{"type":"RefreshRemoteApplicationEvent","timestamp":1494514362123,"originService":"config-server:docker:8888","destinationService":"accountservice:**","id":"53e61c71-cbae-4b6d-84bb-d0dcc0aeb4dc"}
`
                Convey("When handled", func() {
                        handleRefreshEvent([]byte(body), SERVICE_NAME)

                        Convey("Then Viper should have been re-populated with values from Source", func() {
                              So(viper.GetString("server_name"), ShouldEqual, "Accountservice RELOADED")
                        })
                })
        })
}

func TestHandleRefreshEventForOtherApplication(t *testing.T) {

        gock.Intercept()
        defer gock.Off()
        Convey("Given a refresh event received, targeting another application", t, func() {
                var body = `{"type":"RefreshRemoteApplicationEvent","timestamp":1494514362123,"originService":"config-server:docker:8888","destinationService":"vipservice:**","id":"53e61c71-cbae-4b6d-84bb-d0dcc0aeb4dc"}
`
                Convey("When parsed", func() {
                        handleRefreshEvent([]byte(body), SERVICE_NAME)

                        Convey("Then no outgoing HTTP requests should have been intercepted", func() {
                                So(gock.HasUnmatchedRequest(), ShouldBeFalse)
                        })
                })
        })
}
