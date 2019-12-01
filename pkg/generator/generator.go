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
	MethodGet    = "get"
	MethodUpdate = "update"
	MethodDelete = "delete"
)

type CodeGen struct {
	raw          []byte
	method       KubeMethod
	name         string
	namespace    string
	kind         string
	group        string
	version      string
	replicaCount string
	imports      string
	kubeClient   string
	kubeObject   string
	kubeManage   string
}

func (m KubeMethod) String() string {
	return string(m)
}

func New(raw []byte, method KubeMethod) CodeGen {
	return CodeGen{
		raw:    raw,
		method: method,
	}
}

func (c *CodeGen) Generate() (code string, err error) {
	// Convert yaml specs to runtime object
	if err = c.addKubeObject(); err != nil {
		return code, err
	}

	// Create kubeclient
	c.addKubeClient()
	// Add methods to kubeclient
	c.addKubeManage()

	c.cleanupObject()

	if c.method != MethodDelete && c.method != MethodGet {
		i := importer.New(c.kind, c.group, c.version, c.kubeObject)
		c.imports, c.kubeObject = i.FindImports()
	}

	return c.prettyCode()
}

func (c *CodeGen) addKubeObject() error {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(c.raw, nil, nil)
	if err != nil {
		return fmt.Errorf("Error while decoding YAML object. Err was: %s", err)
	}

	// Find group and version
	objMeta := obj.GetObjectKind().GroupVersionKind()
	c.kind = strings.Title(objMeta.Kind)
	c.group = strings.Title(objMeta.Group)
	if len(c.group) == 0 {
		c.group = "Core"
	}
	c.version = strings.Title(objMeta.Version)

	// Add replica pointer function
	var re = regexp.MustCompile(`replicas:\s?([0-9]+)`)
	matched := re.FindAllStringSubmatch(string(c.raw), -1)
	if len(matched) == 1 && len(matched[0]) == 2 {
		c.replicaCount = matched[0][1]
	}

	// Add object name
	re = regexp.MustCompile(`name:\s?"?([-a-zA-Z]+)`)
	matched = re.FindAllStringSubmatch(string(c.raw), -1)
	if len(matched) >= 1 && len(matched[0]) == 2 {
		c.name = matched[0][1]
	}

	// Add namespace
	c.namespace = "default"
	re = regexp.MustCompile(`namespace:\s?"?([-a-zA-Z]+)`)
	matched = re.FindAllStringSubmatch(string(c.raw), -1)
	if len(matched) >= 1 && len(matched[0]) == 2 {
		c.namespace = matched[0][1]
	}

	// Pretty struct
	c.kubeObject = prettyStruct(fmt.Sprintf("%#v", obj))
	//fmt.Printf("%s\n\n", c.kubeObject)
	return nil
}

func (c *CodeGen) addKubeClient() {
	// TODO: dynamic namespace
	c.kubeClient = fmt.Sprintf(`var kubeconfig = os.Getenv("KUBECONFIG")

        config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
                panic(err)
        }
        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
                panic(err)
        }
        kubeclient := clientset.%s%s().%ss("%s")
	`, c.group, c.version, c.kind, c.namespace)
}

func (c *CodeGen) addKubeManage() {
	var method string
	switch c.method {
	case MethodDelete:
		// Add imports
		for _, i := range importer.COMMON_IMPORTS {
			c.imports += fmt.Sprintf("\"%s\"\n", i)
		}
		c.imports += "metav1 \"k8s.io/apimachinery/pkg/apis/meta/v1\"\n"

		param := fmt.Sprintf(`"%s", &metav1.DeleteOptions{}`, c.name)
		method = fmt.Sprintf("err = kubeclient.%s(%s)", strings.Title(c.method.String()), param)

	case MethodGet:
		// Add imports
		for _, i := range importer.COMMON_IMPORTS {
			c.imports += fmt.Sprintf("\"%s\"\n", i)
		}
		c.imports += "metav1 \"k8s.io/apimachinery/pkg/apis/meta/v1\"\n"

		param := fmt.Sprintf(`"%s", metav1.GetOptions{}`, c.name)
		method = fmt.Sprintf("found, err := kubeclient.%s(%s)\n", strings.Title(c.method.String()), param)
		// Add log
		method += fmt.Sprintf(`fmt.Printf("Found object : %s", found)`, "%+v")

	default:
		method = fmt.Sprintf("_, err = kubeclient.%s(object)", strings.Title(c.method.String()))
	}

	c.kubeManage = fmt.Sprintf(`fmt.Println("%s %s...")
	%s
        if err != nil {
                panic(err)
        }
	`, strings.Title(c.method.String()), c.kind, method)
}

func (c *CodeGen) prettyCode() (code string, err error) {
	kubeobject := fmt.Sprintf(`// Create resource object
	object := %s`, c.kubeObject)

	if c.method == MethodDelete || c.method == MethodGet {
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
	`, c.imports, c.kubeClient, kubeobject, c.kubeManage)

	// Add replica pointer function
	if len(c.replicaCount) > 0 && c.method != MethodDelete && c.method != MethodGet {
		main += addReplicaFunc()
	}

	// Run gofmt
	goFormat, err := format.Source([]byte(main))
	if err != nil {
		return code, fmt.Errorf("go fmt error: %s", err.Error())
	}
	return string(goFormat), nil
}

func (c *CodeGen) cleanupObject() {
	if c.method == MethodDelete || c.method == MethodGet {
		c.kubeObject = ""
	}
	kubeObject := strings.Split(c.kubeObject, "\n")
	kubeObject = removeSubObject(kubeObject, "CreationTimestamp")
	kubeObject = removeSubObject(kubeObject, "Status:")
	kubeObject = removeNilFields(kubeObject)
	kubeObject = c.matchLabelsStruct(kubeObject)
	if len(c.replicaCount) > 0 {
		kubeObject = addReplicaPointer(c.replicaCount, kubeObject)
	}
	c.kubeObject = ""
	for _, l := range kubeObject {
		if len(l) != 0 {
			c.kubeObject += l + "\n"
		}
	}
}

func prettyStruct(obj string) string {
	obj = strings.Replace(obj, ", ", ",\n", -1)
	obj = strings.Replace(obj, "{", " {\n", -1)
	obj = strings.Replace(obj, "}", ",\n}", -1)

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

func removeNilFields(kubeobject []string) []string {
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

func removeSubObject(object []string, objectName string) []string {
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

func addReplicaPointer(replicaCount string, kubeobject []string) []string {
	for i, _ := range kubeobject {
		if strings.Contains(kubeobject[i], "Replicas") {
			kubeobject[i] = fmt.Sprintf("Replicas: int32Ptr(%s),", replicaCount)
		}
	}
	return kubeobject
}

func (c CodeGen) matchLabelsStruct(kubeobject []string) []string {
	var re = regexp.MustCompile(`(?ms)matchLabels:(?:[\s]*([a-zA-Z]+):\s?([a-zA-Z]+))*`)
	matched := re.FindAllStringSubmatch(string(c.raw), -1)
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
