package model

import (
    "strings"
)

// All models that needs to be known by more than one microservice goes in here to avoid violating DRY.

type AccountData struct {
    ID        string `json:"" gorm:"primary_key"`
    Name      string         `json:"name"`
    Events    []AccountEvent `json:"events" gorm:"ForeignKey:AccountID"`
}

// AccountEvent defines a single event on an Account. Uses GORM.
type AccountEvent struct {
    ID        string `json:"" gorm:"primary_key"`
    AccountID string `json:"-"`
    EventName string `json:"eventName"`
    Created   string `json:"created"`
}

// AccountImage with GORM ID
type AccountImage struct {
    ID       string `json:"id" gorm:"primary_key"`
    URL      string `json:"url"`
    ServedBy string `json:"servedBy"`
}

// EmailAddress is just a little experiment with go types.
type EmailAddress string

// IsValid is just a sample method.
func (e EmailAddress) IsValid() bool {
    return strings.Contains(string(e), "@")
}
