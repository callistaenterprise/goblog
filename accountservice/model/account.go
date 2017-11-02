package model

import "strings"

// Account defines ...
type Account struct {
        ID       string `json:"id"`
        Name     string `json:"name"`
        ServedBy string `json:"servedBy"`
        Quote    Quote  `json:"quote"`
        ImageData AccountImage  `json:"imageData"`
}

// Quote defines a Quote as provided by the quotes-service
type Quote struct {
	Text     string `json:"quote"`
	ServedBy string `json:"ipAddress"`
	Language string `json:"language"`
}

// AccountImage
type AccountImage struct {
        URL string `json:"url"`
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
