package main

import (
	"bufio"
	//"encoding/json"
	"fmt"
	"strings"
	//	"io/ioutil"
	"go/format"
	"log"
	"os"
	//	"bytes"

	//	"gopkg.in/yaml.v2"
	//"k8s.io/client-go/pkg/api"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/runtime/serializer"
	//	"k8s.io/apimachinery/pkg/util/yaml"
	//"github.com/davecgh/go-spew/spew"
	//"k8s.io/api/core/v1"
	//"k8s.io/api/extensions/v1beta1"
	//appsv1 "k8s.io/api/apps/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var ApiPkgMap = map[string]string{
	"admissionregistration.k8s.io": "k8s.io/api/admissionregistration",
	"apiextensions.k8s.io":         "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions",
	"apiregistration.k8s.io":       "k8s.io/kube-aggregator/pkg/apis/apiregistration",
	"apps":                         "k8s.io/api/apps",
	"authentication.k8s.io":        "k8s.io/api/authentication",
	"autoscaling":                  "k8s.io/api/autoscaling",
	"batch":                        "k8s.io/api/batch",
	"certificates.k8s.io":          "k8s.io/api/certificates",
	"coordination.k8s.io":          "k8s.io/api/coordination",
	"events.k8s.io":                "k8s.io/api/events",
	"extensions":                   "k8s.io/api/extensions",
	"networking.k8s.io":            "k8s.io/api/networking",
	"node.k8s.io":                  "k8s.io/api/node",
	"policy":                       "k8s.io/api/policy",
	"rbac.authorization.k8s.io":    "k8s.io/api/authorization",
	"scheduling.k8s.io":            "k8s.io/api/scheduling",
	"storage.k8s.io":               "k8s.io/api/storage",
}

type function struct {
	kind       string
	group      string
	version    string
	imports    string
	kubeconfig string
	kubeclient string
	kubeobject string
	kubemanage string
	//body string
}

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

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(data), nil, nil)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
	}
	//fmt.Printf("%#v\n", obj.(*appsv1.Deployment))

	// Find group and version
	objKind := obj.GetObjectKind().GroupVersionKind()

	fun := function{kind: strings.Title(objKind.Kind), group: strings.Title(objKind.Group), version: strings.Title(objKind.Version)}
	fun.AddImports(obj)
	fun.AddClient(obj)
	//fun.AddBody()
	fun.AddObject(obj)
	fun.AddKubeManage(obj)
	fun.printfunc()

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
}

func (f *function) AddImports(obj runtime.Object) {
	f.imports += fmt.Sprintf(`
		"fmt"
		"os"

		"k8s.io/client-go/kubernetes"
		"k8s.io/client-go/tools/clientcmd"
	`)
}

func (f *function) AddClient(obj runtime.Object) {
	f.kubeclient = fmt.Sprintf(`var kubeconfig = os.Getenv("KUBECONFIG")

        config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
                panic(err)
        }
        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
                panic(err)
        }
        kubeclient := clientset.%s%s().%ss("default")
	`, f.group, f.version, f.kind)
}

func (f *function) AddObject(obj runtime.Object) {
	f.kubeobject = fmt.Sprintf("%#v", obj)
}

func (f *function) AddKubeManage(obj runtime.Object) {
	f.kubemanage = fmt.Sprintf(`fmt.Println("Creating deployment...")
        result, err := kubeclient.Create(deployment)
        if err != nil {
                panic(err)
        }
	`)
}

func (f *function) printfunc() {

	main_go := fmt.Sprintf(`
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
	`, f.imports, f.kubeclient, prettyStruct(f.kubeobject), f.kubemanage)

	// Convert result into go format
	goFormat, err := format.Source([]byte(main_go))
	if err != nil {
		log.Fatal("go fmt error:", err)
	}
	fmt.Printf(string(goFormat))
}

func prettyStruct(obj string) string {
	obj = strings.ReplaceAll(obj, ", ", ",\n")
	obj = strings.ReplaceAll(obj, "{", " {\n")
	obj = strings.ReplaceAll(obj, "}", ",\n}")
	return obj
}
