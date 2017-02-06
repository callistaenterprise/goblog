package model

type Account struct {
        Id string `json:"id"`
        Name string  `json:"name"`
        ServedBy string `json:"servedBy"`
}

func (a *Account) ToString() string {
    return a.Id + " " + a.Name
}