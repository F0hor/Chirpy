package main

import _ "github.com/lib/pq"

import(
	"net/http"
	"fmt"
	"sync/atomic"
	"os"
	"database/sql"
	
	"github.com/joho/godotenv"

	"github.com/F0hor/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	isDev bool
	secret string
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

	var isDev bool
	if os.Getenv("PLATFORM") == "dev" {
		isDev = true
	} else {
		isDev = false
	}

	cfg := apiConfig{
		db: dbQueries,
		isDev: isDev,
		secret: os.Getenv("SECRET"),
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
	mux.HandleFunc("POST /api/chirps", cfg.handlerCreateChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.handlerDeleteChirp)
	mux.HandleFunc("GET /api/chirps", cfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerGetChirp)

	mux.HandleFunc("POST /api/users", cfg.handlerCreateUser)
	mux.HandleFunc("PUT /api/users", cfg.handlerUpdateUser)
	mux.HandleFunc("POST /api/login", cfg.handlerLoginUser)
	mux.HandleFunc("POST /api/refresh", cfg.handlerRefreshUser)
	mux.HandleFunc("POST /api/revoke", cfg.handlerRevokeUser)

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
