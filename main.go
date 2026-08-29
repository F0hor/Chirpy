package main

import(
	"net/http"
	"fmt"
)

func main() {
	fmt.Println("Stating server")
	cfg := apiConfig{}

	mux := http.NewServeMux()
	mux.Handle(
		"/app/", 
		cfg.middlewareMetricsInc(
			http.StripPrefix("/app", http.FileServer(http.Dir("./testFiles/"))),
		),
	)

	mux.HandleFunc("GET /api/healthz", func( w http.ResponseWriter, r *http.Request){
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handlerMetricsReset)

	serv := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	serv.ListenAndServe()
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
