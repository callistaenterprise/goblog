package service

import (
	"encoding/json"
	"github.com/callistaenterprise/goblog/common/model"
	"github.com/callistaenterprise/goblog/common/util"
	"github.com/callistaenterprise/goblog/dataservice/internal/app/dbclient"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
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

func (h *Handler) GetAccountByNameWithCount(w http.ResponseWriter, r *http.Request) {
	var accountName = chi.URLParam(r, "accountName")
	result, _ := h.dbClient.QueryAccountByNameWithCount(r.Context(), accountName)
	data, _ := json.Marshal(&result)
	writeJSONResponse(w, http.StatusOK, data)
}

func (h *Handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	accountData := model.AccountData{}
	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &accountData)

	accountData, err = h.dbClient.UpdateAccount(r.Context(), accountData)

	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}
	data, err := json.Marshal(&accountData)
	writeJSONResponse(w, http.StatusOK, data)
}

func (h *Handler) StoreAccount(w http.ResponseWriter, r *http.Request) {
	accountData := model.AccountData{}
	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &accountData)
	if err != nil {
		logrus.Errorf("Problem unmarshalling AccountData JSON: %v", err.Error())
		writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	accountData, err = h.dbClient.StoreAccount(r.Context(), accountData)
	if err != nil {
		logrus.Errorf("Problem storing AccountData: %v", err.Error())
		writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	data, err := json.Marshal(&accountData)
	writeJSONResponse(w, http.StatusCreated, data)
}

// GetAccount loads an account instance, including a quote and an image URL using sub-services.
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {

	// Read the 'accountId' path parameter from the mux map
	var accountID = chi.URLParam(r, "accountId")

	account, err := h.dbClient.QueryAccount(r.Context(), accountID)

	if err == nil {
		// If found, marshal into JSON, write headers and content
		data, _ := json.Marshal(account)
		writeJSONResponse(w, http.StatusOK, data)
	} else {
		logrus.Errorf("Error reading accountID '%v' from DB: %v", accountID, err.Error())
		if err.Error() != "" {
			writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		} else {
			writeJSONResponse(w, http.StatusNotFound, []byte("Account not found"))
		}
	}
}

// RandomAccount is strictly used for easily getting hold of an account id for demo purposes.
func (h *Handler) RandomAccount(w http.ResponseWriter, r *http.Request) {
	account, err := h.dbClient.GetRandomAccount(r.Context())
	if err == nil {
		// If found, marshal into JSON, write headers and content
		data, _ := json.Marshal(account)
		writeJSONResponse(w, http.StatusOK, data)
	} else {
		writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
	}
}

func (h *Handler) SeedAccounts(w http.ResponseWriter, r *http.Request) {
	err := h.dbClient.SeedAccounts()
	if err == nil {
		writeJSONResponse(w, http.StatusOK, []byte("{'result':'OK'}"))
	} else {
		writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
	}
}

func (h *Handler) Close() {
	h.dbClient.Close()
}

func writeJSONResponse(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Connection", "close")
	w.WriteHeader(status)
	w.Write(data)
}
