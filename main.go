package main

import _ "github.com/lib/pq"

import(
	"net/http"
	"fmt"
	"sync/atomic"
	"github.com/joho/godotenv"
	"os"
	"database/sql"

	"github.com/F0hor/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
}

func main() {
	fmt.Println("Stating server")

	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Failed to connect to DB:\n %v\n", err)
		return
	}
	dbQueries := database.New(db)

	cfg := apiConfig{
		db: dbQueries,
	}

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
