package model

import "strings"

// Account defines ...    gorm:"ForeignKey:QuoteID"
type Account struct {
    ID        string         `json:"id"`
    Name      string         `json:"name"`
    ServedBy  string         `json:"servedBy"`
    Quote     Quote          `json:"quote"`
    QuoteID   string         `json:"-"`
    ImageData AccountImage   `json:"imageData"`
    Events    []AccountEvent `json:"events" gorm:"ForeignKey:AccountID"`
}

// AccountEvent defines a single event on an Account. Uses GORM.
type AccountEvent struct {
    ID        string `json:"-" gorm:"primary_key"`
    AccountID string `json:"-"`
    EventName string `json:"eventName"`
    Created   string `json:"created"`
}

// Quote defines a Quote as provided by the quotes-service
type Quote struct {
    Text     string `json:"quote"`
    ServedBy string `json:"ipAddress"`
    Language string `json:"language"`
}

// AccountImage with GORM ID
type AccountImage struct {
    ID       string `json:"-" gorm:"primary_key"`
    URL      string `json:"url"`
    ServedBy string `json:"servedBy"`
}

// ToString is a somewhat generic ToString method.
func (a *Account) ToString() string {
    return a.ID + " " + a.Name
}

// ToString is a somewhat generic ToString method.
func (a *AccountImage) ToString() string {
    return a.URL + " " + a.ServedBy
}

// EmailAddress is just a little experiment with go types.
type EmailAddress string

// IsValid is just a sample method.
func (e EmailAddress) IsValid() bool {
    return strings.Contains(string(e), "@")
}
