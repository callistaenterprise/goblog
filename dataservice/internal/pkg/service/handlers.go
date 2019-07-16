package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/callistaenterprise/goblog/common/model"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (s *Server) GetAccountByNameWithCount(w http.ResponseWriter, r *http.Request) {
	var accountName = mux.Vars(r)["accountName"]
	result, _ := s.dbClient.QueryAccountByNameWithCount(r.Context(), accountName)
	data, _ := json.Marshal(&result)
	writeJSONResponse(w, http.StatusOK, data)
}

func (s *Server) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	accountData := model.AccountData{}
	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &accountData)

	accountData, err = s.dbClient.UpdateAccount(r.Context(), accountData)

	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}
	data, err := json.Marshal(&accountData)
	writeJSONResponse(w, http.StatusOK, data)
}

func (s *Server) StoreAccount(w http.ResponseWriter, r *http.Request) {
	accountData := model.AccountData{}
	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &accountData)
	if err != nil {
		logrus.Errorf("Problem unmarshalling AccountData JSON: %v", err.Error())
		writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	accountData, err = s.dbClient.StoreAccount(r.Context(), accountData)
	if err != nil {
		logrus.Errorf("Problem storing AccountData: %v", err.Error())
		writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	data, err := json.Marshal(&accountData)
	writeJSONResponse(w, http.StatusCreated, data)
}

// GetAccount loads an account instance, including a quote and an image URL using sub-services.
func (s *Server) GetAccount(w http.ResponseWriter, r *http.Request) {

	// Read the 'accountId' path parameter from the mux map
	var accountID = mux.Vars(r)["accountId"]

	account, err := s.dbClient.QueryAccount(r.Context(), accountID)

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
func (s *Server) RandomAccount(w http.ResponseWriter, r *http.Request) {
	account, err := s.dbClient.GetRandomAccount(r.Context())
	if err == nil {
		// If found, marshal into JSON, write headers and content
		data, _ := json.Marshal(account)
		writeJSONResponse(w, http.StatusOK, data)
	} else {
		writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
	}
}

func (s *Server) SeedAccounts(w http.ResponseWriter, r *http.Request) {
	err := s.dbClient.SeedAccounts()
	if err == nil {
		writeJSONResponse(w, http.StatusOK, []byte("{'result':'OK'}"))
	} else {
		writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
	}
}

func writeJSONResponse(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Connection", "close")
	w.WriteHeader(status)
	w.Write(data)
}
