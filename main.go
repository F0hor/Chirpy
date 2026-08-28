package main

import(
	"net/http"
	"fmt"
	"sync/atomic"
	"encoding/json"
	"log"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

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

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	type returnVals struct {
		Valid bool `json:"valid"`
	}
	ret := returnVals{
		Valid: true,
	}

	respondWithJSON(w, 200, ret)
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
		
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerMetricsReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
		
	w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
