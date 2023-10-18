package jwt

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	DefaultExpirationTime = 15 * time.Minute
)

type JWThandler struct {
	secret         []byte
	issuer         string
	expirationTime time.Duration
}

type JWTClaims struct {
	jwt.StandardClaims
	UserCode string `json:"userCode"`
}

func NewJWThandler(secret, issuer string, expirationTime ...time.Duration) *JWThandler {
	expTime := DefaultExpirationTime
	if len(expirationTime) > 0 {
		expTime = expirationTime[0]
	}
	return &JWThandler{
		secret:         []byte(secret),
		issuer:         issuer,
		expirationTime: expTime,
	}
}

func (j *JWThandler) GenerateToken(userCode string) (string, time.Time, error) {
	claims := JWTClaims{
		UserCode: userCode,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(j.expirationTime).Unix(),
			Issuer:    j.issuer,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.secret))

	expires := time.Unix(claims.ExpiresAt, 0)
	return signedToken, expires, err
}

func (j *JWThandler) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return j.secret, nil
		})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, err
	}
	return claims, nil
}

func (j *JWThandler) IsTokenExpired(tokenString string) (bool, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired by") {
			return true, nil
		}
		return false, err
	}
	return j.IsTokenExpiredFromClaims(claims), nil
}

func (j *JWThandler) IsTokenExpiredFromClaims(claims *JWTClaims) bool {
	return claims.ExpiresAt < time.Now().Unix()
}
