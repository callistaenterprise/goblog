package model

import "strings"

type Account struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	ServedBy string `json:"servedBy"`
	Quote    Quote  `json:"quote"`
	ImageUrl string  `json:"imageUrl"`
}

type Quote struct {
	Text     string `json:"quote"`
	ServedBy string `json:"ipAddress"`
	Language string `json:"language"`
}

func (a *Account) ToString() string {
	return a.Id + " " + a.Name
}

type EmailAddress string

func (e EmailAddress) IsValid() bool {
	return strings.Contains(string(e), "@")
}
