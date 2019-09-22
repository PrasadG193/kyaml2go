package generator

import (
	//"bufio"
	//"encoding/json"
	"fmt"
	"strings"
	//	"io/ioutil"
	"go/format"
	//"os"

	"github.com/PrasadG193/kubectl2go/pkg/importer"
	"k8s.io/client-go/kubernetes/scheme"
	//"k8s.io/apimachinery/pkg/runtime"
)

type CodeGen struct {
	Raw        []byte
	Kind       string
	Group      string
	Version    string
	Imports    string
	KubeClient string
	KubeObject string
	KubeManage string
}

func New(raw []byte) CodeGen {
	return CodeGen{
		Raw: raw,
	}
}

func (c *CodeGen) Generate() (code string, err error) {
	// Convert yaml specs to runtime object
	if err = c.AddKubeObject(); err != nil {
		return code, err
	}

	// Create kubeclient
	c.AddKubeClient()
	// Add methods to kubeclient
	c.AddKubeManage()

	i := importer.New(c.Kind, c.Group, c.Version, c.KubeObject)
	c.Imports, c.KubeObject = i.FindImports()

	return c.PrettyCode()

	// ---------------------------------------------------
	// trim object type
	//specs := strings.SplitN(fun.kubeobject, "{", 2)[1]
	//specs = "{"+specs

	//deploy := (*appsv1.Deployment).fun.kubeobject

	//jsonData, err := json.Marshal(obj)
	//fmt.Printf("%+s %+v\n", string(jsonData), err)

	//var mapper appsv1.Deployment
	//err = json.Unmarshal([]byte(jsonData), &mapper)
	//if err != nil {
	//	log.Fatalf("cannot unmarshal data: %v", err)
	//}
	//fmt.Printf("aunmarshal data::\n%+v\n", mapper)

	//	var yamlData appsv1.Deployment
	//	err = yaml.Unmarshal([]byte(data), &yamlData)
	//	if err != nil {
	//		log.Fatalf("cannot unmarshal data: %v", err)
	//	}
	//	//fmt.Printf("%s\n unmarshal data::\n%#v\n", data, um)
	//
	//
	//	//var um map[string]interface{}
	//	jsonData, err := json.Marshal(yamlData)
	//	if err != nil {
	//		log.Fatalf("cannot marshal data: %v", err)
	//	}
	//	fmt.Printf("jsonData::\n%#v\n", string(jsonData))
	//	//yamlData := []byte(data)
	//
	//	var um appsv1.Deployment
	//	err = yaml.Unmarshal(jsonData, &um)
	//	if err != nil {
	//		log.Fatalf("cannot unmarshal data: %v", err)
	//	}
	//	fmt.Printf("%s\n unmarshal data::\n%#v\n", data, um)

	return code, nil
}

func (c *CodeGen) AddKubeObject() error {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(c.Raw, nil, nil)
	if err != nil {
		return fmt.Errorf("Error while decoding YAML object. Err was: %s", err)
	}

	// Find group and version
	objMeta := obj.GetObjectKind().GroupVersionKind()
	c.Kind = strings.Title(objMeta.Kind)
	c.Group = strings.Title(objMeta.Group)
	c.Version = strings.Title(objMeta.Version)

	// Pretty struct
	c.KubeObject = prettyStruct(fmt.Sprintf("%#v", obj))
	return nil
}

//func (f *function) AddImports(obj runtime.Object) {
//	f.imports += fmt.Sprintf(`
//		"fmt"
//		"os"
//
//		"k8s.io/client-go/kubernetes"
//		"k8s.io/client-go/tools/clientcmd"
//	`)
//}

func (c *CodeGen) AddKubeClient() {
	// TODO: dynamic namespace
	c.KubeClient = fmt.Sprintf(`var kubeconfig = os.Getenv("KUBECONFIG")

        config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
                panic(err)
        }
        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
                panic(err)
        }
        kubeclient := clientset.%s%s().%ss("default")
	`, c.Group, c.Version, c.Kind)
}

func (c *CodeGen) AddKubeManage() {
	// TODO: dynamic methods
	c.KubeManage = fmt.Sprintf(`fmt.Println("Creating %s...")
        result, err := kubeclient.Create(object)
        if err != nil {
                panic(err)
        }
	`)
}

func (c *CodeGen) PrettyCode() (code string, err error) {
	main := fmt.Sprintf(`
	package main

	import (
		%s
	)

	func main() {
	// Create client
	%s

	// Create resource object
	object := %s

	// Manage resource
	%s
	}
	`, c.Imports, c.KubeClient, c.KubeObject, c.KubeManage)

	// Run gofmt
	goFormat, err := format.Source([]byte(main))
	if err != nil {
		return code, fmt.Errorf("go fmt error: %s", err.Error())
	}
	return string(goFormat), nil
}

func prettyStruct(obj string) string {
	obj = strings.ReplaceAll(obj, ", ", ",\n")
	obj = strings.ReplaceAll(obj, "{", " {\n")
	obj = strings.ReplaceAll(obj, "}", ",\n}")

	// Run gofmt
	goFormat, err := format.Source([]byte(obj))
	if err != nil {
		fmt.Println("gofmt error", err)
	}
	return string(goFormat)
}
