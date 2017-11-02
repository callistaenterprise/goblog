package util

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestResolveIp(t *testing.T) {

	Convey("Given a Call request", t, func() {

		Convey("When", func() {
			ipAddress, err := ResolveIPFromHostsFile()

			Convey("Then", func() {
				So(err, ShouldBeNil)
				So(ipAddress, ShouldNotBeNil)
				So(string(ipAddress), ShouldContainSubstring, ".")
			})
		})
	})

}
