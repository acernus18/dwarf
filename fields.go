package dwarf

import (
	"fmt"
	"github.com/bluele/gcache"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type LoggerConfig struct {
	LogLevel int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	DBName   string
	Username string
	Password string
}

type ServerConfig struct {
	Host string
	Port int
}

type CacheConfig struct {
	Size int
}

type Config struct {
	Logger   LoggerConfig
	Database DatabaseConfig
	Server   ServerConfig
	Cache    CacheConfig
}

type ResponseType struct {
	SerialNum string
	Code      int64
	Message   string
	Data      any
}

type Page struct {
	PageSize  int
	PageIndex int
	Total     int64
	Data      any
}

type Credential struct {
	ID       string
	Subject  string
	Audience []string
}

type RequestContext[T any] struct {
	Context    *gin.Context
	SerialNum  string
	Resource   string
	Credential Credential
	Body       T
}

type RequestHandler[Req, Res any] func(RequestContext[Req]) (Res, error)

type CredentialProvider func(string) (Credential, error)

type Filter struct {
	Key    string
	Action int
	Value  any
}

type Order struct {
	Order string
	Desc  bool
}

type Pagination struct {
	PageSize  int
	PageIndex int
}

type ScopesParams struct {
	Filters    []Filter
	Orders     []Order
	Pagination Pagination
}

func (s ScopesParams) Where() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		tx := db
		validActions := map[int]string{
			LIKE: "LIKE",
			EQ:   "=",
			NEQ:  "!=",
			GT:   ">",
			LT:   "<",
			GTE:  ">=",
			LTE:  "<=",
			IN:   "IN",
			NIN:  "NOT IN",
		}
		if len(s.Filters) > 0 {
			for i := range s.Filters {
				action, valid := validActions[s.Filters[i].Action]
				if !valid {
					continue
				}
				filter := fmt.Sprintf("`%s` %s ?", caseToCamel(s.Filters[i].Key), action)
				tx = tx.Where(filter, s.Filters[i].Value)
			}
		}
		return tx
	}
}

func (s ScopesParams) Order() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		tx := db
		if len(s.Orders) > 0 {
			orders := make([]string, len(s.Orders))
			for i := range s.Orders {
				if s.Orders[i].Desc {
					orders = append(orders, fmt.Sprintf("`%s` DESC", caseToCamel(s.Orders[i].Order)))
				} else {
					orders = append(orders, fmt.Sprintf("`%s`", caseToCamel(s.Orders[i].Order)))
				}
			}
			tx = tx.Order(orders)
		}
		return tx
	}
}

func (s ScopesParams) Page() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (s.Pagination.PageIndex - 1) * s.Pagination.PageSize
		return db.Limit(s.Pagination.PageSize).Offset(offset)
	}
}

// KeyValue Schema
//create table key_values
//(
//	`id`          int unsigned auto_increment,
//	`created_at`  datetime     not null,
//	`updated_at`  datetime     not null,
//	`deleted_at`  datetime,
//	`key`   	  varchar(32)  not null,
//	`value` 	  json         not null,
//	primary key (`id`),
//	key `idx_deleted_id` (`deleted_at`) using btree,
//	unique key `u_idx_key` (`key`) using btree
//) Engine innodb
//default charset utf8mb4;
type KeyValue struct {
	gorm.Model
	Key   string
	Value datatypes.JSON
}

type Application struct {
	config   Config
	database *gorm.DB
	Engine   *gin.Engine
	Cache    gcache.Cache
}

func (application *Application) Start() error {
	address := fmt.Sprintf("%s:%d", application.config.Server.Host, application.config.Server.Port)
	return application.Engine.Run(address)
}

func (application *Application) GetDatabase() *gorm.DB {
	return application.database.Session(&gorm.Session{NewDB: true})
}
