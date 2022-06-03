package httpMiddleware

import (
	"github.com/rs/cors"
	"net/http"
)

var c = cors.New(cors.Options{
	AllowedOrigins:   []string{"*"},
	AllowCredentials: true,
	Debug:            false,
})

func HandleCors(h http.Handler) http.Handler {
	return c.Handler(h)
}
