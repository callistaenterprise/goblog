package dbclient

import (
        "github.com/jinzhu/gorm"
        _ "github.com/jinzhu/gorm/dialects/postgres"
        "github.com/callistaenterprise/goblog/accountservice/model"
        "strconv"
        "fmt"
        "github.com/stretchr/testify/mock"
)

type IGormClient interface {
        QueryAccount(accountId string) (model.Account, error)
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
        gc.crDB.Close()
}

func (gc *GormClient) QueryAccount(accountId string) (model.Account, error) {
        if gc.crDB == nil {
                return model.Account{}, fmt.Errorf("Connection to DB not established!")
        }
        acc := model.Account{}
        gc.crDB.Preload("Quote").First(&acc, "ID = ?", accountId)

        return acc, nil
}




func (gc *GormClient) SetupDB(addr string)  {
        var err error
        gc.crDB, err = gorm.Open("postgres", addr)
        if err != nil {
                panic("failed to connect database: " + err.Error())
        }

        // Migrate the schema
        gc.crDB.AutoMigrate(&model.Account{}, &model.Quote{})
}

func (gc *GormClient) SeedAccounts() error {
        if gc.crDB == nil {
                return fmt.Errorf("Connection to DB not established!")
        }
        gc.crDB.Delete(&model.Account{})
        gc.crDB.Delete(&model.Quote{})
        total := 100
        for i := 0; i < total; i++ {

                // Generate a key 10000 or larger
                key := strconv.Itoa(10000 + i)

                quote := model.Quote{
                        Text: "Testar..." + key,
                        Language: "sv",
                        ServedBy: "localhost",
                        ID: "q-" + key,
                }

                // Create an instance of our Account struct
                acc := model.Account{
                        ID: key,
                        Name: "Person_" + strconv.Itoa(i),
                        Quote: quote,
                        QuoteID: quote.ID,
                }

                gc.crDB.Create(&acc)
        }
        return nil
}








// MockGormClient is a mock implementation of a datastore client for testing purposes
type MockGormClient struct {
        mock.Mock
}

func (m *MockGormClient) QueryAccount(accountId string) (model.Account, error) {
        args := m.Mock.Called(accountId)
        return args.Get(0).(model.Account), args.Error(1)
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