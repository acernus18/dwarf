package dwarf

import (
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func serialize[T any](object T) []byte {
	result, err := json.Marshal(object)
	if err != nil {
		return []byte("")
	}
	return result
}

func deserialize[T any](data []byte) (T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}
	return result, nil
}

func caseToCamel(name string) string {
	characters := []rune(name)
	builder := strings.Builder{}
	continueFlag := false
	for i, r := range characters {
		if unicode.IsUpper(r) && i != 0 {
			if !continueFlag {
				builder.WriteRune('_')
				continueFlag = true
			}
		} else {
			continueFlag = false
		}
		builder.WriteRune(unicode.ToLower(r))
	}
	return builder.String()
}

func getJWT(secret string, claims jwt.RegisteredClaims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func parseJWT(secret, token string) (*jwt.RegisteredClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if tokenClaims == nil {
		return nil, errors.New("[110000]: parse token fail")
	}
	if result, ok := tokenClaims.Claims.(*jwt.RegisteredClaims); ok && tokenClaims.Valid {
		return result, nil
	}
	return nil, errors.New("[110000]: parse token fail")
}

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
