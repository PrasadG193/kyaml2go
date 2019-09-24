package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PrasadG193/kubectl2go/pkg/serve"
	"github.com/julienschmidt/httprouter"
)

const API_VERSION = "v1"
const PORT = "8080"

func main() {
	router := httprouter.New()
	log.Printf("server started accepting requests on port=%s..\n", PORT)
	router.POST("/v1/convert", serve.HandleConvert)
	router.POST(fmt.Sprintf("/convert", API_VERSION), serve.HandleConvert)
	log.Fatal(http.ListenAndServe(":8080", router))
}
