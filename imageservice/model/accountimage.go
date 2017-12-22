package model

// AccountImage with GORM ID
type AccountImage struct {
    ID       string `json:"id" gorm:"primary_key"`
    URL      string `json:"url"`
    ServedBy string `json:"servedBy"`
}


// ToString is a somewhat generic ToString method.
func (a *AccountImage) ToString() string {
    return a.URL + " " + a.ServedBy
}
