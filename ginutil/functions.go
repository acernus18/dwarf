package ginutil

import (
	"fmt"
	"github.com/acernus18/dwarlf/gormutil"
	"github.com/acernus18/dwarlf/jwtutil"
	"github.com/duke-git/lancet/v2/random"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
	"regexp"
	"strconv"
	"time"
)

const resourceName = "Resource"

func parseError(err error) (int64, string) {
	pattern := regexp.MustCompile(`^\[(\d+)]: (.+)$`)
	if !pattern.Match([]byte(err.Error())) {
		return 100000, err.Error()
	}
	result := pattern.FindAllStringSubmatch(err.Error(), -1)
	code, e := strconv.ParseInt(result[0][1], 10, 64)
	if e != nil {
		code = 100000
	}
	message := result[0][2]
	return code, message
}

func IssueCredential(secret string, expire time.Duration, provider CredentialProvider) gin.HandlerFunc {
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
		token, err := jwtutil.GetToken(secret, jwt.RegisteredClaims{
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

func ParseCredential() gin.HandlerFunc {
	return Wrap(func(ctx RequestContext[any]) (Credential, error) {
		return ctx.Credential, nil
	})
}

func CredentialMiddleware(secret string) gin.HandlerFunc {
	return func(context *gin.Context) {
		claims, err := jwtutil.ParseToken(secret, context.GetHeader("Authorization"))
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

func Wrap[Req, Res any](handler RequestHandler[Req, Res]) gin.HandlerFunc {
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

func Query[T any](tx *gorm.DB) gin.HandlerFunc {
	return Wrap(func(ctx RequestContext[gormutil.ScopesParams]) ([]T, error) {
		return gormutil.Query[T](tx, ctx.Body)
	})
}

func QueryPage[T any](tx *gorm.DB) gin.HandlerFunc {
	return Wrap(func(ctx RequestContext[gormutil.ScopesParams]) (Page, error) {
		data, total, err := gormutil.QueryWithPage[T](tx, ctx.Body)
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
