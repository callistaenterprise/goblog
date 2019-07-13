package circuitbreaker

import (
	"testing"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
	"gopkg.in/h2non/gock.v1"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000",
	})
	logrus.SetLevel(logrus.DebugLevel)
}

func TestCallUsingResilienceAllFails(t *testing.T) {
	defer gock.Off()

	buildGockMatcherTimes(500, 4)
	hystrix.Flush()

	Convey("Given that we've mocked 4 requests to return 500 Server Error", t, func() {

		Convey("When ", func() {
			bytes, err := CallUsingCircuitBreaker(context.TODO(), "TEST", "http://quotes-service", "GET")

			Convey("Then", func() {
				So(err, ShouldNotBeNil)
				So(bytes, ShouldBeNil)
			})
		})
	})
}

func TestCallUsingResilienceLastSucceeds(t *testing.T) {
	defer gock.Off()
	retries = 3
	buildGockMatcherTimes(500, 2)
	body := []byte("Some response")
	buildGockMatcherWithBody(200, string(body))
	hystrix.Flush()

	Convey("Given a Call request", t, func() {

		Convey("When", func() {
			bytes, err := CallUsingCircuitBreaker(context.TODO(), "TEST", "http://quotes-service", "GET")

			Convey("Then", func() {
				So(err, ShouldBeNil)
				So(bytes, ShouldNotBeNil)
				So(string(bytes), ShouldEqual, string(body))
			})
		})
	})
}

func TestCallHystrixOpensAfterThresholdPassed(t *testing.T) {
	defer gock.Off()
	for a := 0; a < 6; a++ {
		buildGockMatcher(500)
	}
	hystrix.Flush()

	Convey("Given hystrix allows 5 failed requests with no retries", t, func() {
		retries = 0
		hystrix.ConfigureCommand("TEST", hystrix.CommandConfig{
			RequestVolumeThreshold: 5,
		})
		Convey("When 6 failed requests performed", func() {
			for a := 0; a < 6; a++ {
				CallUsingCircuitBreaker(context.TODO(), "TEST", "http://quotes-service", "GET")
			}

			Convey("Then make sure the circuit has been opened", func() {
				cb, _, _ := hystrix.GetCircuit("TEST")
				So(cb.IsOpen(), ShouldBeTrue)
			})
		})
	})
}

func buildGockMatcherTimes(status int, times int) {
	for a := 0; a < times; a++ {
		buildGockMatcher(status)
	}
}

func buildGockMatcherWithBody(status int, body string) {
	gock.New("http://quotes-service").
		Reply(status).BodyString(body)
}

func buildGockMatcher(status int) {
	buildGockMatcherWithBody(status, "")
}
