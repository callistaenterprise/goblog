package model

type Account struct {
        ID string `json:"id"`
        Name string  `json:"name"`
        ServedBy string `json:"servedBy"`
        Quote Quote `json:"quote" gorm:"ForeignKey:QuoteID"`
        QuoteID string      `json:"-"`
}

type Quote struct {
        ID string `json:"-" gorm:"primary_key"`
        Text string `json:"quote"`
        ServedBy string `json:"ipAddress"`
        Language string `json:"language"`
}


func (a *Account) ToString() string {
    return a.ID + " " + a.Name
}
