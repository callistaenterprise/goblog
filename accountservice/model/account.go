package model

type Account struct {
        Id string `json:"id"`
        Name string  `json:"name"`
}

func (a *Account) ToString() string {
    return a.Id + " " + a.Name
}