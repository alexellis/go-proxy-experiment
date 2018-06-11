package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {

	var port string
	if val, exists := os.LookupEnv("port"); exists {
		port = val
	} else {
		port = "8080"
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	c := http.Client{}
	http.HandleFunc("/client-post", clientPost(&c))

	http.HandleFunc("/http-post", httpPost())

	http.HandleFunc("/faas-post", faasPost())

	log.Fatal(s.ListenAndServe())
}

func httpPost() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		fn := query.Get("fn")

		defer r.Body.Close()

		res, err := http.Post(fmt.Sprintf("http://%s/", fn), r.Header.Get("Content-Type"), r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(res.StatusCode)

		if _, err := io.CopyBuffer(w, res.Body, nil); err != nil {
			log.Println(err)
		}

		defer res.Body.Close()

	}
}

func clientPost(client *http.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		fn := query.Get("fn")

		defer r.Body.Close()

		req, _ := http.NewRequest(r.Method, fmt.Sprintf("http://%s/", fn), r.Body)

		res, err := client.Do(req)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			return
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		w.WriteHeader(res.StatusCode)
		if _, err := io.CopyBuffer(w, res.Body, nil); err != nil {
			log.Println(err)
		}

	}
}
