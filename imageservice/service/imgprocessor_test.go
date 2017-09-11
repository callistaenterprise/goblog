package service

import (
        "testing"
        . "github.com/smartystreets/goconvey/convey"
        "os"
        "image"
        "bytes"
)

func TestUtilSpec(t *testing.T) {

        Convey("Given you have the cake image", t, func() {

                fImg1, err := os.Open("../testimages/cake.jpg")
                defer fImg1.Close()
                if err != nil {
                        panic(err.Error())
                }
                sourceImage, _, err := image.Decode(fImg1)
                buffer := new(bytes.Buffer)

                Convey("Apply the Sepia filter", func() {
                        Sepia(sourceImage, buffer)

                        Convey("There should be at least 10 kb in the buffer", func() {
                                So(len(buffer.Bytes()), ShouldBeGreaterThan, 10000)
                        })
                })
        })
}
