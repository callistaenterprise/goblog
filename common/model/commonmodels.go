package model

import (
    "strings"
    "github.com/twinj/uuid"
    "github.com/jinzhu/gorm"
    "github.com/callistaenterprise/goblog/common/util"
)

// All models that needs to be known by more than one microservice goes in here to avoid violating DRY.

type AccountData struct {
    ID     string         `json:"" gorm:"primary_key"`
    Name   string         `json:"name"`
    Events []AccountEvent `json:"events" gorm:"ForeignKey:AccountID"`
}

// AccountEvent defines a single event on an Account. Uses GORM.
type AccountEvent struct {
    ID        string `json:"" gorm:"primary_key"`
    AccountID string `json:"-" gorm:"index"` // Don't serialize + index which is very important for performance.
    EventName string `json:"eventName"`
    Created   string `json:"created"`
}

// AccountImage with GORM ID
type AccountImage struct {
    ID       string `json:"id" gorm:"primary_key"`
    URL      string `json:"url"`
    ServedBy string `json:"servedBy"`
}

// Before creating, we append a CREATED event
func (ad *AccountData) BeforeCreate() (err error) {
    event := AccountEvent{ID: uuid.NewV4().String(), AccountID: ad.ID, EventName: "CREATED", Created: util.NowStr()}
    ad.Events = append(ad.Events, event)
    return
}

// BeforeUpdate uses the SetColumn method to append a UPDATED event
func (ad *AccountData) BeforeUpdate(scope *gorm.Scope) (err error) {
    event := AccountEvent{ID: uuid.NewV4().String(), AccountID: ad.ID, EventName: "UPDATED", Created: util.NowStr()}
    ad.Events = append(ad.Events, event)
    scope.SetColumn("events", ad.Events)
    return
}

// EmailAddress is just a little experiment with go types.
type EmailAddress string

// IsValid is just a sample method.
func (e EmailAddress) IsValid() bool {
    return strings.Contains(string(e), "@")
}
