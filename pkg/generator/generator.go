package generator

import (
	"fmt"
	"go/format"
	"regexp"
	"strings"

	"github.com/PrasadG193/kgoclient-gen/pkg/importer"
	"k8s.io/client-go/kubernetes/scheme"
)

type KubeMethod string

const (
	MethodCreate = "create"
	MethodUpdate = "update"
	MethodDelete = "delete"
)

type CodeGen struct {
	Raw          []byte
	Method       KubeMethod
	Name         string
	Namespace    string
	Kind         string
	Group        string
	Version      string
	ReplicaCount string
	Imports      string
	KubeClient   string
	KubeObject   string
	KubeManage   string
}

func (m KubeMethod) String() string {
	return string(m)
}

func New(raw []byte, method KubeMethod) CodeGen {
	return CodeGen{
		Raw:    raw,
		Method: method,
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

	if c.Method != MethodDelete {
		i := importer.New(c.Kind, c.Group, c.Version, c.KubeObject)
		c.Imports, c.KubeObject = i.FindImports()
	}

	return c.PrettyCode()
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
	var re = regexp.MustCompile(`replicas:\s?([0-9]+)`)
	matched := re.FindAllStringSubmatch(string(c.Raw), -1)
	if len(matched) == 1 && len(matched[0]) == 2 {
		c.ReplicaCount = matched[0][1]
	}

	// Add object name
	re = regexp.MustCompile(`name:\s?"?([-a-zA-Z]+)`)
	matched = re.FindAllStringSubmatch(string(c.Raw), -1)
	if len(matched) >= 1 && len(matched[0]) == 2 {
		c.Name = matched[0][1]
	}

	// Add Namespace
	c.Namespace = "default"
	re = regexp.MustCompile(`namespace:\s?"?([-a-zA-Z]+)`)
	matched = re.FindAllStringSubmatch(string(c.Raw), -1)
	if len(matched) >= 1 && len(matched[0]) == 2 {
		c.Namespace = matched[0][1]
	}

	// Pretty struct
	c.KubeObject = prettyStruct(fmt.Sprintf("%#v", obj))
	//fmt.Printf("%s\n\n", c.KubeObject)
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
        kubeclient := clientset.%s%s().%ss("%s")
	`, c.Group, c.Version, c.Kind, c.Namespace)
}

func (c *CodeGen) AddKubeManage() {
	// if Delete method
	method := fmt.Sprintf("_, err = kubeclient.%s(object)", strings.Title(c.Method.String()))
	if c.Method == MethodDelete {
		// Add imports
		for _, i := range importer.COMMON_IMPORTS {
			c.Imports += fmt.Sprintf("\"%s\"\n", i)
		}
		c.Imports += "metav1 \"k8s.io/apimachinery/pkg/apis/meta/v1\"\n"

		param := fmt.Sprintf(`"%s", &metav1.DeleteOptions{}`, c.Name)
		method = fmt.Sprintf("err = kubeclient.%s(%s)", strings.Title(c.Method.String()), param)
	}

	c.KubeManage = fmt.Sprintf(`fmt.Println("%s %s...")
	%s
        if err != nil {
                panic(err)
        }
	`, strings.Title(c.Method.String()), c.Kind, method)
}

func (c *CodeGen) PrettyCode() (code string, err error) {
	kubeobject := fmt.Sprintf(`// Create resource object
	object := %s`, c.KubeObject)

	if c.Method == MethodDelete {
		kubeobject = ""
	}

	main := fmt.Sprintf(`
	package main

	import (
		%s
	)

	func main() {
	// Create client
	%s

	%s

	// Manage resource
	%s
	}
	`, c.Imports, c.KubeClient, kubeobject, c.KubeManage)

	// Add replica pointer function
	if len(c.ReplicaCount) > 0 && c.Method != MethodDelete {
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
	if c.Method == MethodDelete {
		c.KubeObject = ""
	}
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
