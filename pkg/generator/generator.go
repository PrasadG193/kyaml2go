package generator

import (
	"fmt"
	"go/format"
	"regexp"
	"strings"

	"github.com/gdexlab/go-render/render"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/PrasadG193/kgoclient-gen/pkg/importer"
	"github.com/PrasadG193/kgoclient-gen/pkg/kube"
)

type KubeMethod string

const (
	MethodCreate = "create"
	MethodGet    = "get"
	MethodUpdate = "update"
	MethodDelete = "delete"
)

type CodeGen struct {
	raw             []byte
	method          KubeMethod
	name            string
	namespace       string
	kind            string
	group           string
	version         string
	replicaCount    string
	termGracePeriod string
	imports         string
	kubeClient      string
	runtimeObject   runtime.Object
	kubeObject      string
	kubeManage      string
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
	// Remove unnecessary fields
	c.cleanupObject()
	// Replace Data with StringData for secret object types
	c.secretStringData()

	if c.method != MethodDelete && c.method != MethodGet {
		i := importer.New(c.kind, c.group, c.version, c.kubeObject)
		c.imports, c.kubeObject = i.FindImports()
	}

	return c.prettyCode()
}

func (c *CodeGen) addKubeObject() error {
	var err error
	decode := scheme.Codecs.UniversalDeserializer().Decode
	c.runtimeObject, _, err = decode(c.raw, nil, nil)
	if err != nil {
		return fmt.Errorf("Error while decoding YAML object. Err was: %s", err)
	}

	// Find group and version
	objMeta := c.runtimeObject.GetObjectKind().GroupVersionKind()
	c.kind = strings.Title(objMeta.Kind)
	c.group = strings.Title(objMeta.Group)
	if len(c.group) == 0 {
		c.group = "Core"
	}
	c.version = strings.Title(objMeta.Version)

	// Find replica count
	var re = regexp.MustCompile(`replicas:\s?([0-9]+)`)
	matched := re.FindAllStringSubmatch(string(c.raw), -1)
	if len(matched) == 1 && len(matched[0]) == 2 {
		c.replicaCount = matched[0][1]
	}

	// Add terminationGracePeriodSeconds
	re = regexp.MustCompile(`terminationGracePeriodSeconds:\s?([0-9]+)`)
	matched = re.FindAllStringSubmatch(string(c.raw), -1)
	if len(matched) == 1 && len(matched[0]) == 2 {
		c.termGracePeriod = matched[0][1]
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
	c.kubeObject = prettyStruct(render.AsCode(c.runtimeObject))
	//fmt.Printf("%s\n\n", c.kubeObject)
	return nil
}

func (c *CodeGen) addKubeClient() {
	// TODO: dynamic namespace
	c.kubeClient = fmt.Sprintf(`var kubeconfig string
	kubeconfig, ok := os.LookupEnv("KUBECONFIG")
	if !ok {
		kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

        config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
                panic(err)
        }
        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
                panic(err)
        }
	`)

	method := fmt.Sprintf("kubeclient := clientset.%s%s().%ss()\n", c.group, c.version, c.kind)
	if _, ok := kube.KindNamespaced[c.kind]; ok {
		method = fmt.Sprintf("kubeclient := clientset.%s%s().%ss(\"%s\")", c.group, c.version, c.kind, c.namespace)
	}
	c.kubeClient += method
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
	if c.method != MethodDelete && c.method != MethodGet {
		if len(c.replicaCount) != 0 {
			main += addIntptrFunc("32")
		}
		if len(c.termGracePeriod) != 0 {
			main += addIntptrFunc("64")
		}
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
	kubeObject = replaceSubObject(kubeObject, "CreationTimestamp", "", -1)
	kubeObject = replaceSubObject(kubeObject, "Status", "", -1)
	kubeObject = removeNilFields(kubeObject)
	kubeObject = updateResources(kubeObject)
	if len(c.replicaCount) > 0 {
		kubeObject = replaceSubObject(kubeObject, "Replicas:", fmt.Sprintf("Replicas: int32Ptr(%s),", c.replicaCount), -1)
	}
	if len(c.termGracePeriod) > 0 {
		kubeObject = replaceSubObject(kubeObject, "TerminationGracePeriodSeconds:", fmt.Sprintf("TerminationGracePeriodSeconds: int64Ptr(%s),", c.termGracePeriod), -1)
	}

	// Remove binary secret data
	if c.kind == "Secret" {
		kubeObject = replaceSubObject(kubeObject, "CreationTimestamp", "", -1)
		kubeObject = replaceSubObject(kubeObject, "Data: map[string][]uint8", "", -1)
	}

	c.kubeObject = ""
	for _, l := range kubeObject {
		if len(l) != 0 {
			c.kubeObject += l + "\n"
		}
	}
}

func (c *CodeGen) secretStringData() {
	if c.kind != "Secret" {
		return
	}

	secretObject, ok := c.runtimeObject.(*v1.Secret)
	if !ok {
		return
	}
	secretObject.StringData = make(map[string]string)
	for key, val := range secretObject.Data {
		secretObject.StringData[key] = string(val)
	}
	c.runtimeObject = secretObject
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

func addIntptrFunc(bytes string) string {
	return fmt.Sprintf(`func int%sPtr(i int%s) *int%s { return &i }`, bytes, bytes, bytes)
}

func removeNilFields(kubeobject []string) []string {
	nilFields := []string{"nil", "\"\"", "false"}
	for i, line := range kubeobject {
		for _, n := range nilFields {
			if strings.Contains(line, n) {
				kubeobject[i] = ""
			}
		}
	}
	return kubeobject
}

// replace struct field and all sub fields
// n stands for no. of occurances you want to replace
// n < 0 = all occurances
func replaceSubObject(object []string, objectName, newObject string, n int) []string {
	depth := 0

	for i, line := range object {
		if n == 0 {
			break
		}
		if !strings.Contains(line, objectName) && depth == 0 {
			continue
		}
		if strings.Contains(line, "{") {
			depth += 1
		}
		if strings.Contains(line, "}") {
			depth -= 1
		}
		if strings.Contains(line, objectName) {
			object[i] = newObject
		} else {
			object[i] = ""
		}
		// Replace n occurances
		if depth == 0 {
			n--
		}
	}
	return object
}

func parseResourceValue(object []string) (string, string) {
	var value, format string
	for _, line := range object {
		// parse value
		re := regexp.MustCompile(`(?m)value:\s([0-9]*)`)
		matched := re.FindAllStringSubmatch(line, -1)
		if len(matched) >= 1 && len(matched[0]) == 2 {
			value = matched[0][1]
		}

		// Parse unit
		re = regexp.MustCompile(`(?m)resource\.Format\("([a-z-A-Z]*)"\)`)
		matched = re.FindAllStringSubmatch(line, -1)
		if len(matched) >= 1 && len(matched[0]) == 2 {
			format = matched[0][1]
			break
		}
	}
	return value, format
}

func updateResources(object []string) []string {
	cpu := "\"cpu\": resource.Quantity"
	mem := "\"memory\": resource.Quantity"
	storage := "\"storage\": resource.Quantity"
	for i, line := range object {
		if strings.Contains(line, cpu) {
			value, format := parseResourceValue(object[i+1:])
			replaceSubObject(object[i:], cpu, fmt.Sprintf("\"cpu\": *resource.NewQuantity(%s, resource.%s),", value, format), 1)
		}

		if strings.Contains(line, mem) {
			value, format := parseResourceValue(object[i+1:])
			replaceSubObject(object[i:], mem, fmt.Sprintf("\"memory\": *resource.NewQuantity(%s, resource.%s),", value, format), 1)
		}

		if strings.Contains(line, storage) {
			value, format := parseResourceValue(object[i+1:])
			replaceSubObject(object[i:], storage, fmt.Sprintf("\"storage\": *resource.NewQuantity(%s, resource.%s),", value, format), 1)
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
