package jwt

import (
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	_DEFAULT_EXPIRATION_TIME = 15 * time.Minute
)

type JWThandler struct {
	secret         string
	issuer         string
	expirationTime time.Duration
}

type JWTClaims struct {
	jwt.StandardClaims
	UserEmail string `json:"user_email"`
}

func NewJWThandler(secret, issuer string, expirationTime ...time.Duration) *JWThandler {
	expTime := _DEFAULT_EXPIRATION_TIME
	if len(expirationTime) > 0 {
		expTime = expirationTime[0]
	}
	return &JWThandler{
		secret:         secret,
		issuer:         issuer,
		expirationTime: expTime,
	}
}

func (j *JWThandler) GenerateToken(userEmail string) (string, error) {
	claims := JWTClaims{
		UserEmail: userEmail,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(j.expirationTime).Unix(),
			Issuer:    j.issuer,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

func (j *JWThandler) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(j.secret), nil
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
		return false, err
	}
	return j.IsTokenExpiredFromClaims(claims), nil
}

func (j *JWThandler) IsTokenExpiredFromClaims(claims *JWTClaims) bool {
	return claims.ExpiresAt < time.Now().Unix()
}
