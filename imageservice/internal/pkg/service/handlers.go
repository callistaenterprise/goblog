package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/callistaenterprise/goblog/imageservice/internal/pkg/dbclient"
	"image"
	"net/http"
	"os"
	"strconv"

	"github.com/callistaenterprise/goblog/common/model"
	"github.com/callistaenterprise/goblog/common/util"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

type Handler struct {
	dbClient  dbclient.IGormClient
	myIP      string
	isHealthy bool
}

func NewHandler(dbClient dbclient.IGormClient) *Handler {
	myIP, err := util.ResolveIPFromHostsFile()
	if err != nil {
		myIP = util.GetIP()
	}
	return &Handler{dbClient: dbClient, myIP: myIP, isHealthy: true}
}

func (h *Handler) CreateAccountImage(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	accountImage := model.AccountImage{}
	err = json.Unmarshal(body, &accountImage)
	if err != nil {
		writeServerError(w, err.Error())
		return
	}

	accountImage, err = h.dbClient.StoreAccountImage(r.Context(), accountImage)
	if err != nil {
		writeServerError(w, err.Error())
		return
	}
	respData, _ := json.Marshal(&accountImage)
	writeResponse(w, respData)
}

func (h *Handler) UpdateAccountImage(w http.ResponseWriter, r *http.Request) {
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

	accountImage, err = h.dbClient.StoreAccountImage(r.Context(), accountImage)
	if err != nil {
		writeServerError(w, err.Error())
		return
	}
	respData, _ := json.Marshal(&accountImage)
	writeResponse(w, respData)
}

func (h *Handler) GetAccountImage(w http.ResponseWriter, r *http.Request) {
	accountImage, err := h.dbClient.QueryAccountImage(r.Context(), mux.Vars(r)["accountId"])
	accountImage.ServedBy = h.myIP
	data, err := json.Marshal(&accountImage)
	if err != nil {
		writeServerError(w, err.Error())
	} else {
		writeResponse(w, data)
	}
}

/**
 * Takes the filename and tries to decode an image from /testimages/{filename}. Used for testing.
 */
func (h *Handler) ProcessImageFromFile(w http.ResponseWriter, r *http.Request) {

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
	writeResponse(w, outputData)
}

func (h *Handler) Close() {
	h.dbClient.Close()
}

func writeResponse(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func writeServerError(w http.ResponseWriter, msg string) {
	logrus.Error(msg)
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}
