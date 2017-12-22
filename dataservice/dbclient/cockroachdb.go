package dbclient

import (
    "fmt"
    "strconv"

    "github.com/callistaenterprise/goblog/common/model"
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/postgres"
    "github.com/stretchr/testify/mock"
    "context"
    "github.com/Sirupsen/logrus"
    "time"
)

type IGormClient interface {
    QueryAccount(ctx context.Context, accountId string) (model.AccountData, error)
    QueryAccountByNameWithCount(ctx context.Context, name string) ([]Pair, error)
    SetupDB(addr string)
    SeedAccounts() error
    Check() bool
    Close()
}

type GormClient struct {
    crDB *gorm.DB
}

func (gc *GormClient) Check() bool {
    return gc.crDB != nil
}

func (gc *GormClient) Close() {
    logrus.Infoln("Closing connection to CockroachDB")
    gc.crDB.Close()
}

func (gc *GormClient) QueryAccount(ctx context.Context, accountId string) (model.AccountData, error) {
    if gc.crDB == nil {
        return model.AccountData{}, fmt.Errorf("Connection to DB not established!")
    }
    acc := model.AccountData{}
    gc.crDB.Preload("Events").First(&acc, "ID = ?", accountId)
    if acc.ID == "" {
        return acc, fmt.Errorf("")
    }
    return acc, nil
}

func (gc *GormClient) QueryAccountByNameWithCount(ctx context.Context, name string) ([]Pair, error) {

    rows, err := gc.crDB.Table("account_data as ad").
    Select("name, count(ae.ID)").
    Joins("join account_events as ae on ae.account_id = ad.id").
    Where("name like ?", name + "%").
    Group("name").Rows()

    result := make([]Pair, 0)
    for rows.Next() {
        pair := Pair{}
        rows.Scan(&pair.Name, &pair.Count)
        result = append(result, pair)
    }
    return result, err
}

func (gc *GormClient) SetupDB(addr string) {
    logrus.Infof("Connecting with connection string: '%v'", addr)
    var err error
    gc.crDB, err = gorm.Open("postgres", addr)
    if err != nil {
        panic("failed to connect database: " + err.Error())
    }

    // Migrate the schema
    gc.crDB.AutoMigrate(&model.AccountData{}, &model.AccountEvent{})
}

func (gc *GormClient) SeedAccounts() error {
    if gc.crDB == nil {
        return fmt.Errorf("Connection to DB not established!")
    }
    gc.crDB.Delete(&model.AccountData{})
    gc.crDB.Delete(&model.AccountEvent{})
    total := 100
    for i := 0; i < total; i++ {

        // Generate a key 10000 or larger
        key := strconv.Itoa(10000 + i)

        // Create an AccountEvent struct
        accountEvent := model.AccountEvent{
            ID:        "accountEvent-" + key,
            EventName: "CREATED",
            Created:   time.Now().Format("2006-01-02T15:04:05"),
        }

        accountEvents := make([]model.AccountEvent, 0)
        accountEvents = append(accountEvents, accountEvent)

        // Create an instance of our Account struct
        acc := model.AccountData{
            ID:     key,
            Name:   "Person_" + strconv.Itoa(i),
            Events: accountEvents,
        }

        gc.crDB.Create(&acc)
        logrus.Infoln("Successfully created account instance.")
    }
    return nil
}

type Pair struct {
    Name string
    Count uint8
}

// MockGormClient is a mock implementation of a datastore client for testing purposes
type MockGormClient struct {
    mock.Mock
}

func (m *MockGormClient) QueryAccount(ctx context.Context, accountId string) (model.AccountData, error) {
    args := m.Mock.Called(ctx, accountId)
    return args.Get(0).(model.AccountData), args.Error(1)
}

func (m *MockGormClient) QueryAccountByNameWithCount(ctx context.Context, name string) ([]Pair, error) {
    args := m.Mock.Called(ctx, name)
    return args.Get(0).([]Pair), args.Error(1)
}

func (m *MockGormClient) SetupDB(addr string) {
    // Does nothing
}

func (m *MockGormClient) SeedAccounts() error {
    args := m.Mock.Called()
    return args.Get(0).(error)
}

func (m *MockGormClient) Check() bool {
    args := m.Mock.Called()
    return args.Get(0).(bool)
}

func (m *MockGormClient) Close() {
    // Does nothing
}
