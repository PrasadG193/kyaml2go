package serve

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/PrasadG193/kubectl2go/pkg/generator"
	"github.com/julienschmidt/httprouter"
)

func HandleConvert(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Printf("here")
	urlPQ, _ := url.ParseQuery(r.URL.RawQuery)
	method := generator.KubeMethod(urlPQ.Get("method"))
	if len(method) == 0 {
		method = generator.MethodCreate
	}
	body, err := ioutil.ReadAll(r.Body)
	gen := generator.New(body, method)
	code, err := gen.Generate()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(body), method, err)
	io.WriteString(w, code)
}
