package jwt

// InterfaceJWT описывает методы для работы с JWT.
type InterfaceJWT interface {
	// Generate создаёт новый JWT на основе userID.
	Generate(userID int32) (string, error)

	// Validate разбирает токен, проверяет подпись и возвращает userID из claims.
	Validate(tokenString string) (int32, error)
}
