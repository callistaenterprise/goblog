package service

import (
    "github.com/gorilla/mux"
    "encoding/json"
    "net/http"
    "github.com/callistaenterprise/goblog/dataservice/dbclient"
    "strconv"
)

// DBClient is our GORM instance.
var DBClient dbclient.IGormClient

func GetAccountByNameWithCount(w http.ResponseWriter, r *http.Request) {
    var accountName = mux.Vars(r)["accountName"]
    result, _ := DBClient.QueryAccountByNameWithCount(r.Context(), accountName)
    data, _ := json.Marshal(&result)
    writeJSONResponse(w, http.StatusOK, data)
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
