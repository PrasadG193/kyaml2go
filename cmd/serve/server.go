package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PrasadG193/kyaml2go/pkg/serve"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

const apiVersion = "v1"
const port = "8080"

func main() {
	router := httprouter.New()
	log.Printf("server started accepting requests on port=%s..\n", port)
	router.POST(fmt.Sprintf("/%s/convert", apiVersion), serve.HandleConvert)

	handler := cors.Default().Handler(router)
	log.Fatal(http.ListenAndServe(":8080", handler))
}
