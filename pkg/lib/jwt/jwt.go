package auth

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type RouteRuleEnum string

const (
	Admin    RouteRuleEnum = "a"
	Business RouteRuleEnum = "b"
	Instance RouteRuleEnum = "i"
)

const (
	AUD = "codechat_v2.users"
	AZP = "codechat_v2"
	ISS = "ms_db_v2"
	TYP = "Bearer"
)

type Jwt struct {
	ID        string // token id
	Expires   bool
	ExpiresIn int
	Rules     []RouteRuleEnum
}

type CustomClaims struct {
	Exp        int64         `json:"exp"`
	Iat        int64         `json:"iat"`
	Jti        string        `json:"jti"`
	Sub        string        `json:"sub"`
	Typ        string        `json:"typ"`
	Scope      string        `json:"scope"`
	Profile    string        `json:"profile"`
	Aud        string        `json:"aud"`
	Azp        string        `json:"azp"`
	Iss        string        `json:"iss"`
	RoleAccess RouteRuleEnum `json:"roles_access"`
	Z          bool          `json:"z"`
	jwt.StandardClaims
}

func readPublicKey(path string) (any, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode(bytes)
	publicKey, err := x509.ParsePKIXPublicKey(keyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

func NewJwt() *Jwt {
	return &Jwt{
		ID: uuid.NewString(),
	}
}

func (j *Jwt) Read(accessToken *string) (*CustomClaims, error) {
	publicKey, err := readPublicKey("./.keys/jwt/public_key.pem")
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(*accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)

	if !ok {
		return nil, errors.New("invalid token")
	}

	if !token.Valid && !claims.Z {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (j *Jwt) IsScope(scope RouteRuleEnum) bool {
	return slices.Contains(j.Rules, scope)
}

func (j *Jwt) VerifyAttributes(decode *CustomClaims) bool {
	if decode.Aud != AUD || decode.Azp != AZP || decode.Iss != ISS || decode.Typ != TYP {
		return false
	}

	return true
}
