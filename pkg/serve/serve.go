package serve

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// RequestHandler implement http request handlers
type RequestHandler struct {
}

// NewHandler returns new instance of RequestHandler
func NewHandler() *RequestHandler {
	return &RequestHandler{}
}

// HandleConvert parses http request to get K8s resource specs and return generated Go code
// for valid resource specs
func (rh *RequestHandler) HandleConvert(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("Bad Request. Error: %s", err.Error()), http.StatusBadRequest)
		return
	}

	urlPQ, _ := url.ParseQuery(r.URL.RawQuery)
	method := urlPQ.Get("method")
	cr := urlPQ.Get("cr")
	if method == "" {
		method = "create"
	}
	args := []string{method}
	if cr != "" {
		args = append(args, []string{"--cr", "--apis", urlPQ.Get("apis"), "--client", urlPQ.Get("client"), "--scheme", urlPQ.Get("scheme")}...)
	}

	log.Printf("Incoming request. %s/bin/kyaml2go %v %s", os.Getenv("GOPATH"), args, string(body))
	code, err := execute(fmt.Sprintf("%s/bin/kyaml2go", os.Getenv("GOPATH")), args, string(body))
	if err != nil {
		log.Printf("Failed to generate code. %s %v", code, err)
		http.Error(w, fmt.Sprintf("Bad Request. Error: %s %s", code, err.Error()), http.StatusBadRequest)
		return
	}

	io.WriteString(w, code)
}

func execute(cmd string, args []string, data string) (string, error) {
	fmt.Println(cmd, args)
	c := exec.Command(cmd, args...)
	c.Stdin = strings.NewReader(data)
	out, err := c.CombinedOutput()
	return string(out), err
}
