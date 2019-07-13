package dbclient

import (
	"fmt"
	"strconv"

	"context"
	"github.com/sirupsen/logrus"
	"github.com/callistaenterprise/goblog/common/model"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/mock"
	"github.com/twinj/uuid"
)

type IGormClient interface {
	UpdateAccountImage(ctx context.Context, accountImage model.AccountImage) (model.AccountImage, error)
	StoreAccountImage(ctx context.Context, accountImage model.AccountImage) (model.AccountImage, error)
	QueryAccountImage(ctx context.Context, accountId string) (model.AccountImage, error)
	SetupDB(addr string)
	SeedAccountImages() error
	Check() bool
	Close()
}

func (gc *GormClient) StoreAccountImage(ctx context.Context, accountImage model.AccountImage) (model.AccountImage, error) {
	tx := gc.crDB.Begin()
	accountImage.ID = uuid.NewV4().String()
	tx = tx.Create(&accountImage)
	if tx.Error != nil {
		logrus.Errorf("Error storing account image: %v", tx.Error.Error())
		return model.AccountImage{}, tx.Error
	}
	tx = tx.Commit()
	return accountImage, nil
}

func (gc *GormClient) UpdateAccountImage(ctx context.Context, accountImage model.AccountImage) (model.AccountImage, error) {
	tx := gc.crDB.Begin()
	tx = tx.Save(&accountImage)
	if tx.Error != nil {
		logrus.Errorf("Error updating account image: %v", tx.Error.Error())
		return model.AccountImage{}, tx.Error
	}
	tx = tx.Commit()
	accountImage, _ = gc.QueryAccountImage(ctx, accountImage.ID)
	return accountImage, nil
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

func (gc *GormClient) QueryAccountImage(ctx context.Context, accountId string) (model.AccountImage, error) {
	span := tracing.StartChildSpanFromContext(ctx, "GormClient.QueryAccountImage")
	defer span.Finish()

	if gc.crDB == nil {
		return model.AccountImage{}, fmt.Errorf("Connection to DB not established!")
	}
	acc := model.AccountImage{}
	gc.crDB.First(&acc, "ID = ?", accountId)
	if acc.ID == "" {
		return acc, fmt.Errorf("")
	}
	return acc, nil
}

func (gc *GormClient) SetupDB(addr string) {
	logrus.Infof("Connecting with connection string: '%v'", addr)
	var err error
	gc.crDB, err = gorm.Open("postgres", addr)
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// Migrate the schema
	gc.crDB.AutoMigrate(&model.AccountImage{})
}

func (gc *GormClient) SeedAccountImages() error {
	if gc.crDB == nil {
		return fmt.Errorf("Connection to DB not established!")
	}
	gc.crDB.Delete(&model.AccountImage{})

	total := 100
	for i := 0; i < total; i++ {

		// Generate a key 10000 or larger
		key := strconv.Itoa(10000 + i)

		// Create an instance of our Account struct
		acc := model.AccountImage{
			ID:  key,
			URL: "http://path.to.some.image/" + key + ".png",
		}

		gc.crDB.Create(&acc)
		logrus.Infoln("Successfully created AccountImage instance.")
	}
	return nil
}

// MockGormClient is a mock implementation of a datastore client for testing purposes
type MockGormClient struct {
	mock.Mock
}

func (m *MockGormClient) UpdateAccountImage(ctx context.Context, accountImage model.AccountImage) (model.AccountImage, error) {
	args := m.Mock.Called(ctx, accountImage)
	return args.Get(0).(model.AccountImage), args.Error(1)
}
func (m *MockGormClient) StoreAccountImage(ctx context.Context, accountImage model.AccountImage) (model.AccountImage, error) {
	args := m.Mock.Called(ctx, accountImage)
	return args.Get(0).(model.AccountImage), args.Error(1)
}

func (m *MockGormClient) QueryAccountImage(ctx context.Context, accountId string) (model.AccountImage, error) {
	args := m.Mock.Called(ctx, accountId)
	return args.Get(0).(model.AccountImage), args.Error(1)
}

func (m *MockGormClient) SetupDB(addr string) {
	// Does nothing
}

func (m *MockGormClient) SeedAccountImages() error {
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
