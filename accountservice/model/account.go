package model

type Account struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	ServedBy string `json:"servedBy"`
	Quote    Quote  `json:"quote"`
}

type Quote struct {
	Text     string `json:"quote"`
	ServedBy string `json:"ipAddress"`
	Language string `json:"language"`
}

func (a *Account) ToString() string {
	return a.Id + " " + a.Name
}
