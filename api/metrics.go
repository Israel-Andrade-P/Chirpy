package api

import (
	"fmt"
	"net/http"
)

func (cfg *Apiconfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *Apiconfig) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	hits := cfg.FileserverHits.Load()
	html := fmt.Sprintf(`
	        <html>
				<body>
					<h1>Welcome, Chirpy Admin</h1>
					<p>Chirpy has been visited %d times!</p>
				</body>
			</html>
	    `, hits)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func Readiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}

func NoCacheFileServer(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("If-Modified-Since")
		r.Header.Del("If-None-Match")
		h.ServeHTTP(w, r)
	})
}
