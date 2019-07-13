package model

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetAccountWrongPath(t *testing.T) {

	Convey("Given a an email address", t, func() {
		var email EmailAddress
		email = "erik@domain.nu"

		var email2 EmailAddress
		email2 = "erikdomain.nu"

		Convey("When validated", func() {
			valid := email.IsValid()
			Convey("Then result should be true", func() {
				So(valid, ShouldBeTrue)
			})
			invalid := email2.IsValid()
			Convey("Then result should be false", func() {
				So(invalid, ShouldBeFalse)
			})
		})
	})
}
