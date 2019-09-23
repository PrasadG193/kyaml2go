package generator

import (
	//"bufio"
	//"encoding/json"
	"fmt"
	"regexp"
	"strings"
	//	"io/ioutil"
	"go/format"
	//"os"

	"github.com/PrasadG193/kubectl2go/pkg/importer"
	"k8s.io/client-go/kubernetes/scheme"
	//"k8s.io/apimachinery/pkg/runtime"
)

type CodeGen struct {
	Raw          []byte
	Kind         string
	Group        string
	Version      string
	ReplicaCount string
	Imports      string
	KubeClient   string
	KubeObject   string
	KubeManage   string
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

	c.CleanupObject()
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
	if len(c.Group) == 0 {
		c.Group = "Core"
	}
	c.Version = strings.Title(objMeta.Version)

	// Add replica pointer function
	var re = regexp.MustCompile(`(?m)replicas: ([0-9]+)`)
	matched := re.FindAllStringSubmatch(string(c.Raw), -1)
	if len(matched) == 1 && len(matched[0]) == 2 {
		c.ReplicaCount = matched[0][1]
	}

	// Pretty struct
	c.KubeObject = prettyStruct(fmt.Sprintf("%#v", obj))
	return nil
}

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
        _, err = kubeclient.Create(object)
        if err != nil {
                panic(err)
        }
	`, c.Kind)
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

	// Add replica pointer function
	if len(c.ReplicaCount) > 0 {
		main += addReplicaFunc()
	}

	// Run gofmt
	goFormat, err := format.Source([]byte(main))
	if err != nil {
		return code, fmt.Errorf("go fmt error: %s", err.Error())
	}
	return string(goFormat), nil
}

func (c *CodeGen) CleanupObject() {
	kubeObject := strings.Split(c.KubeObject, "\n")
	kubeObject = RemoveSubObject(kubeObject, "CreationTimestamp")
	kubeObject = RemoveSubObject(kubeObject, "Status:")
	kubeObject = RemoveNilFields(kubeObject)
	kubeObject = c.MatchLabelsStruct(kubeObject)
	if len(c.ReplicaCount) > 0 {
		kubeObject = AddReplicaPointer(c.ReplicaCount, kubeObject)
	}
	c.KubeObject = ""
	for _, l := range kubeObject {
		if len(l) != 0 {
			c.KubeObject += l + "\n"
		}
	}
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

func addReplicaFunc() string {
	return fmt.Sprintf(`func int32Ptr(i int32) *int32 { return &i }`)
}

func RemoveNilFields(kubeobject []string) []string {
	nilFields := []string{"nil", "\"\","}
	for i, line := range kubeobject {
		for _, n := range nilFields {
			if strings.Contains(line, n) {
				kubeobject[i] = ""
			}
		}
	}
	return kubeobject
}

func RemoveSubObject(object []string, objectName string) []string {
	depth := 0
	for i, line := range object {
		if strings.Contains(line, objectName) {
			object[i] = ""
			depth = 1
			continue
		}
		if depth > 0 {
			object[i] = ""
			if strings.Contains(line, "{") {
				depth += 1
				continue
			}
			if strings.Contains(line, "}") {
				depth -= 1
				continue
			}
		}
	}
	return object
}

func AddReplicaPointer(replicaCount string, kubeobject []string) []string {
	for i, _ := range kubeobject {
		if strings.Contains(kubeobject[i], "Replicas") {
			kubeobject[i] = fmt.Sprintf("Replicas: int32Ptr(%s),", replicaCount)
		}
	}
	return kubeobject
}

func (c CodeGen) MatchLabelsStruct(kubeobject []string) []string {
	var re = regexp.MustCompile(`(?ms)matchLabels:(?:[\s]*([a-zA-Z]+):\s?([a-zA-Z]+))*`)
	matched := re.FindAllStringSubmatch(string(c.Raw), -1)
	if len(matched) != 1 || len(matched[0]) != 3 {
		return kubeobject
	}

	labels := ""
	for i, _ := range matched {
		labels += fmt.Sprintf("\"%s\": \"%s\",\n", matched[i][1], matched[i][2])
	}

	for i, _ := range kubeobject {
		if strings.Contains(kubeobject[i], "Selector:") {
			kubeobject[i] = fmt.Sprintf(`Selector: &v1.LabelSelector{
                                MatchLabels: map[string]string{
                                        %s
                                },
                        },`, labels)
		}
	}
	return kubeobject
}
