package configutil

import (
	"fmt"
	"github.com/acernus18/dwarlf/ginutil"
	"github.com/bluele/gcache"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DEBUG = iota
	PRODUCT
)

type Config struct {
	Logger struct {
		LogLevel int
	}
	Database struct {
		Host     string
		Port     int
		DBName   string
		Username string
		Password string
	}
	Server struct {
		Host string
		Port int
	}
	Cache struct {
		Size int
	}
}

type Application struct {
	Config   Config
	Engine   *gin.Engine
	Database *gorm.DB
	Cache    gcache.Cache
}

func NewApplication(config Config) (*Application, error) {
	result := new(Application)
	result.Config = config
	result.Cache = gcache.New(config.Cache.Size).Simple().Build()
	user := config.Database.Username
	password := config.Database.Password
	database := config.Database.DBName
	address := fmt.Sprintf("%s:%d", config.Database.Host, config.Database.Port)
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, address, database)
	logLevel := logrus.DebugLevel
	gormLevel := logger.Info
	if config.Logger.LogLevel == PRODUCT {
		logLevel = logrus.InfoLevel
		gormLevel = logger.Silent
	}
	logrus.SetLevel(logLevel)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(logrus.StandardLogger(), logger.Config{
			LogLevel: gormLevel,
		}),
	})
	if err != nil {
		return nil, err
	}
	result.Database = db
	result.Engine = gin.Default()
	if err = result.Engine.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	result.Engine.GET("/health", ginutil.Wrap(func(ctx ginutil.RequestContext[any]) (string, error) {
		return "Health", nil
	}))
	return result, nil
}

func (application *Application) Start() error {
	address := fmt.Sprintf("%s:%d", application.Config.Server.Host, application.Config.Server.Port)
	return application.Engine.Run(address)
}
