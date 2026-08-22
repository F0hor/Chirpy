package main

import(
	"net/http"
	"time"
	"fmt"
)

func main() {
	fmt.Println("Stating server")

	mux := http.NewServeMux()
	mux.Handle(
		"/app/", 
		http.StripPrefix("/app", http.FileServer(http.Dir("./testFiles/"))),
	)
	mux.HandleFunc("/healthz", func( w http.ResponseWriter, r *http.Request){
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		
		w.Write([]byte("OK"))
})

	serv := &http.Server{
		Addr: ":8080",
		Handler: mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	serv.ListenAndServe()
}
