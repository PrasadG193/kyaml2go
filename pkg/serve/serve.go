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

	//"github.com/PrasadG193/kyaml2go/pkg/generator"
	"github.com/julienschmidt/httprouter"
)

const path = "./manifest.yaml"

// HandleConvert parses http request to get K8s resource specs and return generated Go code
// for valid resource specs
func HandleConvert(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("Bad Request. Error: %s", err.Error()), http.StatusBadRequest)
		return
	}

	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		// create file if not exists
		if os.IsNotExist(err) {
			file, err = os.Create(path)
			if err != nil {
				log.Println(err)
				http.Error(w, fmt.Sprintf("Bad Request. Error: %s", err.Error()), http.StatusBadRequest)
				return
			}
		}
		return
	}
	defer file.Close()
	//defer os.Remove(path)
	_, err = file.WriteString(string(body))
	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("Bad Request. Error: %s", err.Error()), http.StatusBadRequest)
		return
	}

	urlPQ, _ := url.ParseQuery(r.URL.RawQuery)
	method := urlPQ.Get("method")
	cr := urlPQ.Get("cr")
	if method != "" {
		method = "create"
	}
	args := []string{method, "-f", path}
	if cr != "" {
		args = append(args, []string{"--cr", "--apis", urlPQ.Get("apis"), "--client", urlPQ.Get("client"), "--schema", urlPQ.Get("schema")}...)
	}

	code, err := execute(fmt.Sprintf("%s/bin/kyaml2go", os.Getenv("GOPATH")), args)
	if err != nil {
		log.Printf("Failed to generate code. %s %v", code, err)
		http.Error(w, fmt.Sprintf("Bad Request. Error: %s %s", code, err.Error()), http.StatusBadRequest)
		return
	}

	io.WriteString(w, code)
}

func execute(cmd string, args []string) (string, error) {
	fmt.Println(cmd, args)
	c := exec.Command(cmd, args...)
	out, err := c.CombinedOutput()
	return string(out), err
}
