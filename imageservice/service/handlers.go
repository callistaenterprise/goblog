package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"net/http"
	"os"
	"strconv"

	"github.com/callistaenterprise/goblog/common/model"
	"github.com/callistaenterprise/goblog/common/util"
	"github.com/callistaenterprise/goblog/imageservice/dbclient"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

// DBClient is our GORM instance.
var DBClient dbclient.IGormClient

var myIp string

func init() {
	var err error
	myIp, err = util.ResolveIPFromHostsFile()
	if err != nil {
		myIp = util.GetIP()
	}
}

func CreateAccountImage(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	accountImage := model.AccountImage{}
	err = json.Unmarshal(body, &accountImage)
	if err != nil {
		writeServerError(w, err.Error())
		return
	}

	accountImage, err = DBClient.StoreAccountImage(r.Context(), accountImage)
	if err != nil {
		writeServerError(w, err.Error())
		return
	}
	respData, _ := json.Marshal(&accountImage)
	writeAndReturn(w, respData)
}

func UpdateAccountImage(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	accountImage := model.AccountImage{}
	err = json.Unmarshal(body, &accountImage)
	if err != nil {
		writeServerError(w, err.Error())
		return
	}
	if accountImage.ID == "" {
		writeServerError(w, "No ID supplied")
		return
	}

	accountImage, err = DBClient.StoreAccountImage(r.Context(), accountImage)
	if err != nil {
		writeServerError(w, err.Error())
		return
	}
	respData, _ := json.Marshal(&accountImage)
	writeAndReturn(w, respData)
}

func GetAccountImage(w http.ResponseWriter, r *http.Request) {
	accountImage, err := DBClient.QueryAccountImage(r.Context(), mux.Vars(r)["accountId"])
	accountImage.ServedBy = myIp
	data, err := json.Marshal(&accountImage)
	if err != nil {
		writeServerError(w, err.Error())
	} else {
		writeResponse(w, data)
	}

}
func writeResponse(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

/**
 * Takes the filename and tries to decode an image from /testimages/{filename}. Used for testing.
 */
func ProcessImageFromFile(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var filename = vars["filename"]
	logrus.Println("Serving image for account: " + filename)

	fImg1, err := os.Open("testimages/" + filename)
	defer fImg1.Close()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	sourceImage, _, err := image.Decode(fImg1)

	if err != nil {
		writeServerError(w, err.Error())
		return
	}
	buf := new(bytes.Buffer)
	err = Sepia(sourceImage, buf)

	if err != nil {
		fmt.Println(err.Error())
		writeServerError(w, err.Error())
		return
	}
	outputData := buf.Bytes()
	writeAndReturn(w, outputData)
}

func writeAndReturn(w http.ResponseWriter, outputData []byte) {

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(outputData)))
	w.WriteHeader(http.StatusOK)
	w.Write(outputData)
}

func writeServerError(w http.ResponseWriter, msg string) {
	logrus.Error(msg)
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}
