package dbclient

import (
    "context"
    "fmt"
    "strconv"
    "github.com/Sirupsen/logrus"
    "github.com/callistaenterprise/goblog/common/model"
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/postgres"
    "github.com/stretchr/testify/mock"
    "github.com/twinj/uuid"
    "github.com/callistaenterprise/goblog/common/tracing"
)

type IGormClient interface {
    UpdateAccount(ctx context.Context, accountData model.AccountData) (model.AccountData, error)
    StoreAccount(ctx context.Context, accountData model.AccountData) (model.AccountData, error)
    QueryAccount(ctx context.Context, accountId string) (model.AccountData, error)
    GetRandomAccount(ctx context.Context) (model.AccountData, error)
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

// StoreAccount uses ACID TX.
func (gc *GormClient) StoreAccount(ctx context.Context, accountData model.AccountData) (model.AccountData, error) {
    span := tracing.StartChildSpanFromContext(ctx, "GormClient.StoreAccount")
    defer span.Finish()

    if gc.crDB == nil {
        return model.AccountData{}, fmt.Errorf("Connection to DB not established!")
    }
    accountData.ID = uuid.NewV4().String()

    tx := gc.crDB.Begin()
    tx = tx.Create(&accountData)
    if tx.Error != nil {

        logrus.Errorf("Error creating AccountData: %v", tx.Error.Error())
        return model.AccountData{}, tx.Error
    }
    tx = tx.Commit()
    if tx.Error != nil {
        logrus.Errorf("Error committing AccountData: %v", tx.Error.Error())
        return model.AccountData{}, tx.Error
    }
    logrus.Infoln("Successfully created AccountData instance")
    return accountData, nil
}

// UpdateAccount uses ACID TX.
func (gc *GormClient) UpdateAccount(ctx context.Context, accountData model.AccountData) (model.AccountData, error) {
    span := tracing.StartChildSpanFromContext(ctx, "GormClient.UpdateAccount")
    defer span.Finish()

    if gc.crDB == nil {
        return model.AccountData{}, fmt.Errorf("Connection to DB not established!")
    }
    tx := gc.crDB.Begin()
    tx = tx.Save(&accountData)
    if tx.Error != nil {
        logrus.Errorf("Error updating AccountData: %v", tx.Error.Error())
        return model.AccountData{}, tx.Error
    }
    tx.Commit()
    if tx.Error != nil {
        logrus.Errorf("Error committing AccountData: %v", tx.Error.Error())
        return model.AccountData{}, tx.Error
    }
    logrus.Infoln("Successfully updated AccountData instance")

    // Read object from DB before return.
    accountData, _ = gc.QueryAccount(ctx, accountData.ID)
    return accountData, nil
}

func (gc *GormClient) QueryAccount(ctx context.Context, accountId string) (model.AccountData, error) {
    span := tracing.StartChildSpanFromContext(ctx, "GormClient.QueryAccount")
    defer span.Finish()

    if gc.crDB == nil {
        return model.AccountData{}, fmt.Errorf("connection to DB not established!")
    }
    tx := gc.crDB.Begin()
    acc := model.AccountData{}
    tx = tx.Preload("Events").First(&acc, "ID = ?", accountId)
    if tx.Error != nil {
        return acc, tx.Error
    }
    if acc.ID == "" {
        return acc, fmt.Errorf("no account found having ID %v", accountId)
    }
    tx.Commit()
    return acc, nil
}

func (gc *GormClient) GetRandomAccount(ctx context.Context) (model.AccountData, error) {
    span := tracing.StartChildSpanFromContext(ctx, "GormClient.GetRandomAccount")
    defer span.Finish()

    if gc.crDB == nil {
        return model.AccountData{}, fmt.Errorf("connection to DB not established!")
    }
    tx := gc.crDB.Begin()
    acc := model.AccountData{}
    tx = tx.Preload("Events").First(&acc)
    if tx.Error != nil {
        return acc, tx.Error
    }
    if acc.ID == "" {
        return acc, fmt.Errorf("no random account found")
    }
    tx.Commit()
    return acc, nil
}

func (gc *GormClient) QueryAccountByNameWithCount(ctx context.Context, name string) ([]Pair, error) {

    rows, err := gc.crDB.Table("account_data as ad").
        Select("name, count(ae.ID)").
        Joins("join account_events as ae on ae.account_id = ad.id").
        Where("name like ?", name+"%").
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
        return fmt.Errorf("connection to DB not established")
    }
    gc.crDB.Delete(&model.AccountData{})
    gc.crDB.Delete(&model.AccountEvent{})
    total := 100
    for i := 0; i < total; i++ {

        // Generate a key 10000 or larger
        key := strconv.Itoa(10000 + i)

        // Create an instance of our Account struct
        acc := model.AccountData{
            ID:   key,
            Name: "Person_" + strconv.Itoa(i),
        }

        gc.crDB.Create(&acc)
    }
    logrus.Infof("Successfully created %v account instances.", 100)
    return nil
}

type Pair struct {
    Name  string
    Count uint8
}

// MockGormClient is a mock implementation of a datastore client for testing purposes
type MockGormClient struct {
    mock.Mock
}

func (m *MockGormClient) StoreAccount(ctx context.Context, accountData model.AccountData) (model.AccountData, error) {
    args := m.Mock.Called(accountData)
    return args.Get(0).(model.AccountData), args.Error(1)
}

func (m *MockGormClient) UpdateAccount(ctx context.Context, accountData model.AccountData) (model.AccountData, error) {
    args := m.Mock.Called(accountData)
    return args.Get(0).(model.AccountData), args.Error(1)
}

func (m *MockGormClient) QueryAccount(ctx context.Context, accountId string) (model.AccountData, error) {
    args := m.Mock.Called(ctx, accountId)
    return args.Get(0).(model.AccountData), args.Error(1)
}

func (m *MockGormClient) GetRandomAccount(ctx context.Context) (model.AccountData, error) {
    args := m.Mock.Called(ctx)
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
