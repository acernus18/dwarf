package requestutil

import (
	"bytes"
	"crypto/tls"
	"github.com/acernus18/dwarlf/serializeutil"
	"io/ioutil"
	"net/http"
)

func Get[T any](api string) (result T, err error) {
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
	result, err = serializeutil.Deserialize[T](responseBody)
	if err = response.Body.Close(); err != nil {
		return result, err
	}
	return result, err
}

func Post[T any](api string, data any) (result T, err error) {
	// [BugFix]: Avoid -> x509: certificate signed by unknown authority
	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	response, err := http.Post(api, "application/json", bytes.NewBuffer(serializeutil.Serialize(data)))
	if err != nil {
		return result, err
	}
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return result, err
	}
	result, err = serializeutil.Deserialize[T](responseBody)
	if err = response.Body.Close(); err != nil {
		return result, err
	}
	return result, err
}
