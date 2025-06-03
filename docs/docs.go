package docs

import (
	_ "embed"
	"net/http"
)

//go:embed swagger.json
var swaggerSpec []byte

// Handler возвращает HTTP-хендлер, который отдаёт встроенную OpenAPI 3.0 спецификацию.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(swaggerSpec)
	})
}
