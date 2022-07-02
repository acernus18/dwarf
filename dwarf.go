package dwarf

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/duke-git/lancet/v2/random"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io/ioutil"
	"net/http"
	"time"
)

func HTTPGet[T any](api string) (result T, err error) {
	// [BugFix]: Avoid -> x509: certificate signed by unknown authority
	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	response, err := http.Get(api)
	if err != nil {
		return result, err
	}
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return result, err
	}
	result, err = deserialize[T](responseBody)
	if err = response.Body.Close(); err != nil {
		return result, err
	}
	return result, err
}

func HTTPPost[T any](api string, data any) (result T, err error) {
	// [BugFix]: Avoid -> x509: certificate signed by unknown authority
	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	response, err := http.Post(api, "application/json", bytes.NewBuffer(serialize(data)))
	if err != nil {
		return result, err
	}
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return result, err
	}
	result, err = deserialize[T](responseBody)
	if err = response.Body.Close(); err != nil {
		return result, err
	}
	return result, err
}

func LoadCache[T any](cache gcache.Cache, key string, expire time.Duration, supplier func() (T, error)) (result T, err error) {
	value, err := cache.Get(key)
	if err == nil {
		return deserialize[T](value.([]byte))
	}
	if err != gcache.KeyNotFoundError {
		return result, err
	}
	result, err = supplier()
	if err != nil {
		return result, err
	}
	return result, cache.SetWithExpire(key, serialize(result), expire)
}

func PrettyJSON(v any) string {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(v)
	//content, _ := json.MarshalIndent(v, "", "  ")
	return buffer.String()
}

func HandlerWrap[Req, Res any](handler RequestHandler[Req, Res]) gin.HandlerFunc {
	return func(context *gin.Context) {
		var ctx RequestContext[Req]
		ctx.Context = context
		ctx.Resource = context.Param(resourceName)
		ctx.SerialNum = fmt.Sprintf("SERIAL-%d-%s", time.Now().Unix(), random.RandString(5))
		if err := context.Bind(&ctx.Body); err != nil {
			context.JSON(200, ResponseType{
				SerialNum: ctx.SerialNum,
				Code:      100000,
				Message:   err.Error(),
				Data:      nil,
			})
			return
		}
		credentialValue, exist := context.Get("Credential")
		if exist {
			if credential, ok := credentialValue.(Credential); ok {
				ctx.Credential = credential
			}
		}
		result, err := handler(ctx)
		if err != nil {
			code, message := parseError(err)
			context.JSON(200, ResponseType{
				SerialNum: ctx.SerialNum,
				Code:      code,
				Message:   message,
				Data:      nil,
			})
			return
		}
		response := ResponseType{
			SerialNum: ctx.SerialNum,
			Code:      0,
			Message:   "SUC",
			Data:      result,
		}
		context.JSON(200, response)
	}
}

func IssueCredentialRouter(secret string, expire time.Duration, provider CredentialProvider) gin.HandlerFunc {
	return func(context *gin.Context) {
		var requestBody struct {
			Code string `json:"code"`
		}
		if err := context.Bind(&requestBody); err != nil {
			context.JSON(200, ResponseType{
				SerialNum: fmt.Sprintf("SERIAL-%d-%s", time.Now().Unix(), random.RandString(5)),
				Code:      110001,
				Message:   err.Error(),
				Data:      nil,
			})
			return
		}
		credential, err := provider(requestBody.Code)
		if err != nil {
			context.JSON(200, ResponseType{
				SerialNum: fmt.Sprintf("SERIAL-%d-%s", time.Now().Unix(), random.RandString(5)),
				Code:      110002,
				Message:   err.Error(),
				Data:      nil,
			})
			return
		}
		token, err := getJWT(secret, jwt.RegisteredClaims{
			Issuer:    "JwtUtil",
			Subject:   credential.Subject,
			Audience:  credential.Audience,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        credential.ID,
		})
		if err != nil {
			context.JSON(200, ResponseType{
				SerialNum: fmt.Sprintf("SERIAL-%d-%s", time.Now().Unix(), random.RandString(5)),
				Code:      110003,
				Message:   err.Error(),
				Data:      nil,
			})
			return
		}
		response := ResponseType{
			SerialNum: fmt.Sprintf("SERIAL-%d-%s", time.Now().Unix(), random.RandString(5)),
			Code:      0,
			Message:   "SUC",
			Data:      token,
		}
		context.JSON(200, response)
	}
}

func ParseCredentialRouter() gin.HandlerFunc {
	return HandlerWrap(func(ctx RequestContext[any]) (Credential, error) {
		return ctx.Credential, nil
	})
}

func CredentialMiddleware(secret string) gin.HandlerFunc {
	return func(context *gin.Context) {
		claims, err := parseJWT(secret, context.GetHeader("Authorization"))
		if err != nil {
			context.JSON(200, ResponseType{
				SerialNum: fmt.Sprintf("SERIAL-%d-%s", time.Now().Unix(), random.RandString(5)),
				Code:      110004,
				Message:   err.Error(),
				Data:      nil,
			})
			context.Abort()
			return
		}
		context.Set("Credential", Credential{
			ID:       claims.ID,
			Subject:  claims.Subject,
			Audience: claims.Audience,
		})
		context.Next()
	}
}

func GinQuery[T any](tx *gorm.DB) gin.HandlerFunc {
	return HandlerWrap(func(ctx RequestContext[ScopesParams]) ([]T, error) {
		return DBQuery[T](tx, ctx.Body)
	})
}

func GinQueryPage[T any](tx *gorm.DB) gin.HandlerFunc {
	return HandlerWrap(func(ctx RequestContext[ScopesParams]) (Page, error) {
		data, total, err := DBQueryWithPage[T](tx, ctx.Body)
		if err != nil {
			return Page{}, nil
		}
		return Page{
			PageSize:  ctx.Body.Pagination.PageSize,
			PageIndex: ctx.Body.Pagination.PageIndex,
			Total:     total,
			Data:      data,
		}, nil
	})
}

func DBQuery[T any](tx *gorm.DB, params ScopesParams) ([]T, error) {
	result := make([]T, 0)
	if err := tx.Scopes(params.Where(), params.Order()).Find(&result).Error; err != nil {
		return result, err
	}
	return result, nil
}

func DBQueryWithPage[T any](tx *gorm.DB, params ScopesParams) ([]T, int64, error) {
	var model T
	var count int64 = 0
	result := make([]T, 0)
	scopes := tx.Model(&model).Scopes(params.Where(), params.Order())
	if err := scopes.Count(&count).Error; err != nil {
		return result, count, err
	}
	if err := scopes.Scopes(params.Page()).Find(&result).Error; err != nil {
		return result, count, err
	}
	return result, count, nil
}

func DBTake[T any](tx *gorm.DB, filter Filter) (T, error) {
	var defaultResult T
	result, err := DBQuery[T](tx, ScopesParams{
		Filters: []Filter{filter},
		Orders:  nil,
	})
	if err != nil {
		return defaultResult, err
	}
	if len(result) < 1 {
		return defaultResult, gorm.ErrRecordNotFound
	}
	return result[0], nil
}

func DBUpdate[T any](tx *gorm.DB, filter Filter, operator func(T) (T, error)) error {
	record, err := DBTake[T](tx, filter)
	if err != nil {
		return err
	}
	destination, err := operator(record)
	if err != nil {
		return err
	}
	if err = tx.Model(&record).Updates(destination).Error; err != nil {
		return err
	}
	return nil
}

func DBSet[T any](tx *gorm.DB, key string, value T) error {
	var count int64
	if err := tx.Model(&KeyValue{}).Where("`key` = ?", key).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return tx.Create(&KeyValue{
			Key:   key,
			Value: serialize(value),
		}).Error
	}
	return tx.Model(&KeyValue{}).Where("`key` = ?", key).Update("`value`", serialize(value)).Error
}

func DBGet[T any](tx *gorm.DB, key string) (T, error) {
	var kv KeyValue
	if err := tx.Where("`key` = ?", key).Take(&kv).Error; err != nil {
		var defaultValue T
		return defaultValue, err
	}
	return deserialize[T](kv.Value)
}

func NewApplication(config Config) (*Application, error) {
	result := new(Application)
	result.config = config
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
	result.database = db
	result.Engine = gin.Default()
	if err = result.Engine.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	result.Engine.GET("/health", HandlerWrap(func(ctx RequestContext[any]) (string, error) {
		return "Health", nil
	}))
	return result, nil
}
