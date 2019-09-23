package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/PrasadG193/kubectl2go/pkg/serve"
)

const API_VERSION = "v1"

func main() {
	router := httprouter.New()
	router.POST("/v1/convert", serve.HandleConvert)
	router.POST(fmt.Sprintf("/convert", API_VERSION), serve.HandleConvert)
	log.Fatal(http.ListenAndServe(":8080", router))
}
