package model

import (
    "github.com/callistaenterprise/goblog/common/model"
)

// The accountservice defines types only it knows about. The AccountData and AccountImage types are pulled
// from the common/model repo.

// Account defines ...    gorm:"ForeignKey:QuoteID"
type Account struct {
    ID            string               `json:"id"`
    Name          string               `json:"name"`
    ServedBy      string               `json:"servedBy"`
    Quote         Quote                `json:"quote"`
    ImageData     model.AccountImage   `json:"imageData"`
    AccountEvents []model.AccountEvent `json:"accountEvents"`
}

// Quote defines a Quote as provided by the quotes-service
type Quote struct {
    Text     string `json:"quote"`
    ServedBy string `json:"l"`
    Language string `json:"language"`
}

// ToString is a somewhat generic ToString method.
func (a *Account) ToString() string {
    return a.ID + " " + a.Name
}
