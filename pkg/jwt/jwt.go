package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

// claimsKey — имя поля в claim, где лежит ID пользователя.
const claimsKey = "user_id"

// ServiceJWT — конкретная реализация Service.
type ServiceJWT struct {
	secret     []byte
	ttlInHours time.Duration
}

// NewJWT принимает секрет и TTL (в часах) из конфига и возвращает реализацию.
func NewJWT(secret string, ttlHours time.Duration) *ServiceJWT {
	return &ServiceJWT{
		secret:     []byte(secret),
		ttlInHours: ttlHours,
	}
}
func (j *ServiceJWT) OnStart(_ context.Context) error { return nil }
func (j *ServiceJWT) OnStop(_ context.Context) error  { return nil }

// Generate формирует новый токен с полем claimsKey = userID и exp = now + ttlInHours.
func (j *ServiceJWT) Generate(userID int32) (string, error) {
	claims := jwt.MapClaims{
		claimsKey: userID,
		"exp":     time.Now().Add(j.ttlInHours * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// Validate парсит токен, проверяет подпись и возвращает userID.
func (j *ServiceJWT) Validate(tokenString string) (int32, error) {
	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что метод подписи — HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil || !parsed.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims format")
	}

	// Извлекаем userID из claims
	raw, exists := claims[claimsKey]
	if !exists {
		return 0, errors.New("user_id not found in token")
	}

	// jwt.MapClaims всегда кладёт цифры как float64
	uidFloat, ok := raw.(float64)
	if !ok {
		return 0, errors.New("user_id claim has unexpected type")
	}

	return int32(uidFloat), nil
}
