package service

import (
    "encoding/json"
    "io/ioutil"
    "net/http"
    "strconv"

    "github.com/Sirupsen/logrus"
    "github.com/callistaenterprise/goblog/common/model"
    "github.com/callistaenterprise/goblog/dataservice/dbclient"
    "github.com/gorilla/mux"
)

// DBClient is our GORM instance.
var DBClient dbclient.IGormClient

func GetAccountByNameWithCount(w http.ResponseWriter, r *http.Request) {
    var accountName = mux.Vars(r)["accountName"]
    result, _ := DBClient.QueryAccountByNameWithCount(r.Context(), accountName)
    data, _ := json.Marshal(&result)
    writeJSONResponse(w, http.StatusOK, data)
}

func UpdateAccount(w http.ResponseWriter, r *http.Request) {
    accountData := model.AccountData{}
    body, err := ioutil.ReadAll(r.Body)
    err = json.Unmarshal(body, &accountData)

    accountData, err = DBClient.UpdateAccount(r.Context(), accountData)

    if err != nil {
        writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
        return
    }
    data, err := json.Marshal(&accountData)
    writeJSONResponse(w, http.StatusOK, data)
}

func StoreAccount(w http.ResponseWriter, r *http.Request) {
    accountData := model.AccountData{}
    body, err := ioutil.ReadAll(r.Body)
    err = json.Unmarshal(body, &accountData)
    if err != nil {
        logrus.Errorf("Problem unmarshalling AccountData JSON: %v", err.Error())
        writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
        return
    }

    accountData, err = DBClient.StoreAccount(r.Context(), accountData)
    if err != nil {
        logrus.Errorf("Problem storing AccountData: %v", err.Error())
        writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
        return
    }

    data, err := json.Marshal(&accountData)
    writeJSONResponse(w, http.StatusCreated, data)
}

// GetAccount loads an account instance, including a quote and an image URL using sub-services.
func GetAccount(w http.ResponseWriter, r *http.Request) {

    // Read the 'accountId' path parameter from the mux map
    var accountID = mux.Vars(r)["accountId"]

    account, err := DBClient.QueryAccount(r.Context(), accountID)

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

func writeJSONResponse(w http.ResponseWriter, status int, data []byte) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Content-Length", strconv.Itoa(len(data)))
    w.Header().Set("Connection", "close")
    w.WriteHeader(status)
    w.Write(data)
}
