package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/stenehall/gosh/internal/config"
)

const (
	CSP = "default-src 'self'; style-src 'self' 'unsafe-inline'"
)

// Server is the main entry for the templates server functionality of Gosh.
func Server(gosh config.Gosh) error {
	log.Printf("Serving gosh on port %d\n\n", gosh.Port)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	data := struct {
		Title      string
		ShowTitle  bool
		Background string
		Sets       []config.Set
	}{
		Title:      gosh.Title,
		ShowTitle:  gosh.ShowTitle,
		Background: gosh.Background,
		Sets:       gosh.Sets,
	}
	ts, tmplError := template.ParseFiles("templates/index.gohtml")

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))
	http.Handle("/favicons/", http.StripPrefix("/favicons/", http.FileServer(http.Dir("favicons/"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if tmplError != nil {
			log.Println(tmplError.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}

		err := ts.Execute(w, data)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
		}
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Println(err)
		}
		return
	})

	err := http.ListenAndServe(":"+fmt.Sprint(gosh.Port), logRequest(http.DefaultServeMux))
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func logRequest(handler http.Handler) http.Handler {
	return gziphandler.GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

		header := w.Header()
		header.Set("Content-Security-Policy", CSP)

		handler.ServeHTTP(w, r)
	}))
}
