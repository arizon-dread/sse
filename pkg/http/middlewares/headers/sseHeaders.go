package headers

import (
	"log"
	"net/http"
	"strings"

	"github.com/arizon-dread/sse/internal/config"
)

func SseHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conf := config.Get()
		log.Printf("Setting middleware headers")
		w.Header().Set("Access-Control-Allow-Origin", conf.Cors.Url)
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(conf.Cors.Headers, ","))
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(conf.Cors.Methods, ","))

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		next.ServeHTTP(w, r)
	})
}
