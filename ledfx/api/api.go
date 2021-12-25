package api

import (
	"fmt"
	"io"
	"ledfx/ledfx/color"
	"log"
	"net/http"
)

func InitApi(port int) error {
	// Hello world, the web server

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, LedFx Go!!\n")
		io.WriteString(w, "Have a good life!\n")
	}

	c := "#FF55FF"
	log.Println(color.NewColor(c))

	http.HandleFunc("/hello", helloHandler)
	log.Println("Listing for requests at http://localhost:8000/hello")
	return http.ListenAndServe(fmt.Sprint(":", port), nil)
}