package ginutil

import "github.com/gin-gonic/gin"

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
