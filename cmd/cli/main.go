package main

import (
	//"encoding/json"
	"bufio"
	"os"
	"fmt"
	//	"io/ioutil"
	"log"
	//	"bytes"

	//	"gopkg.in/yaml.v2"
	//"k8s.io/client-go/pkg/api"
	//"k8s.io/apimachinery/pkg/runtime/serializer"
	//	"k8s.io/apimachinery/pkg/util/yaml"
	//"github.com/davecgh/go-spew/spew"
	//"k8s.io/api/core/v1"
	//"k8s.io/api/extensions/v1beta1"
	//appsv1 "k8s.io/api/apps/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/PrasadG193/kubectl2go/pkg/generator"
)

func main() {
	if len(os.Args) > 1 {
		//printHelp(os.Args[1])
		os.Exit(0)
	}

	// Read input from the console
	var data string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		data += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("Error while reading input:", err)
	}

	gen := generator.New([]byte(data), "create")
	code, err := gen.Generate()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(code)
}


