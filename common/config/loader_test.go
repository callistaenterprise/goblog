package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
)

func TestParseConfiguration(t *testing.T) {

	Convey("Given a JSON configuration response body", t, func() {
		var body = `{"name":"accountservice-dev","profiles":["dev"],"label":null,"version":null,"propertySources":[{"name":"file:/config-repo/accountservice-dev.yml","source":{"server_port":6767,"server_name":"Accountservice DEV","zipkin.service.url":"http://zipkin:9411"}}]}`

		Convey("When parsed", func() {
			parseConfiguration([]byte(body))

			Convey("Then Viper should have been populated with values from Source", func() {
				So(viper.GetString("server_name"), ShouldEqual, "Accountservice DEV")
				So(viper.GetString("server_port"), ShouldEqual, "6767")
			})
		})
	})
}
