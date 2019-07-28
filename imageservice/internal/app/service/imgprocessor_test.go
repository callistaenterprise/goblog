package service

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"image"
	"os"
	"testing"
)

func TestUtilSpec(t *testing.T) {

	fImg1, err := os.Open("../../../testimages/cake.jpg")
	defer fImg1.Close()
	if err != nil {
		panic(err.Error())
	}
	sourceImage, _, err := image.Decode(fImg1)
	buffer := new(bytes.Buffer)

	Sepia(sourceImage, buffer)

	assert.True(t, len(buffer.Bytes()) > 10000)
}
