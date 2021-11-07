package generator

import (
	"fmt"
	"go/format"
	"regexp"
	"strings"

	"github.com/gdexlab/go-render/render"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/PrasadG193/kyaml2go/pkg/importer"
	"github.com/PrasadG193/kyaml2go/pkg/kube"
	"github.com/PrasadG193/kyaml2go/pkg/types"
)

const (
	crdKind   = "CustomResourceDefinition"
	crdClient = "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

// CodeGen holds K8s resource object
type CodeGen struct {
	raw             []byte
	method          types.KubeMethod
	name            string
	namespace       string
	kind            string
	group           string
	version         string
	resource        string
	client          types.Client
	clientpkg       string
	crAPI           string
	isCR            bool
	namespaced      bool
	replicaCount    string
	termGracePeriod string
	imports         string
	kubeClient      string
	runtimeObject   runtime.Object
	kubeObject      string
	kubeManage      string
	extraFuncs      map[string]string
}

// New returns instance of CodeGen
func New(raw []byte, method types.KubeMethod, isCR, namespaced, isDynamic bool, clientpkg, crAPI string) CodeGen {
	client := types.TypedClient
	if isDynamic {
		client = types.DynamicClient
	}
	return CodeGen{
		raw:        raw,
		method:     method,
		clientpkg:  clientpkg,
		crAPI:      crAPI,
		isCR:       isCR,
		namespaced: namespaced,
		extraFuncs: make(map[string]string),
		client:     client,
	}
}

// Generate returns Go code for types.KubeMethod on a K8s resource
func (c *CodeGen) Generate() (code string, err error) {
	// Convert yaml specs to runtime object
	if err = c.addKubeObject(); err != nil {
		return code, err
	}

	if c.crAPI != "" {
		setCRImports(c.kind, c.group, c.version, c.crAPI, c.namespaced)
	}

	// Create kubeclient
	switch c.client {
	case types.DynamicClient:
		c.addDynamicKubeClient()
		c.addDynamicKubeManage()
	default:
		c.addTypedKubeClient()
		c.addTypedKubeManage()
		// Remove unnecessary fields
		c.cleanupObject()
	}

	if c.client == types.TypedClient && c.method != types.MethodDelete && c.method != types.MethodGet {
		var imports string
		i := importer.New(c.kind, c.group, c.version, c.kubeObject, c.clientpkg, c.client)
		imports, c.kubeObject = i.FindImports()
		c.imports += imports
		c.addPtrMethods()
	}
	return c.prettyCode()
}

// addKubeObject converts raw yaml specs to runtime object
func (c *CodeGen) addKubeObject() error {
	var err error
	var objMeta *schema.GroupVersionKind
	setScheme()

	switch c.client {
	case types.DynamicClient:
		obj := &unstructured.Unstructured{}
		_, objMeta, err = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(c.raw, nil, obj)
		c.runtimeObject = obj
	default:
		c.runtimeObject, objMeta, err = scheme.Codecs.UniversalDeserializer().Decode(c.raw, nil, nil)
	}
	if err != nil || objMeta == nil {
		return fmt.Errorf("Error while decoding YAML object. Err was: %s", err)
	}

	// Find group, kind, resource and version
	c.kind = strings.Title(objMeta.Kind)
	c.group = strings.Title(objMeta.Group)
	if len(c.group) == 0 {
		c.group = "Core"
	}
	c.version = strings.Title(objMeta.Version)
	// Pod => Pods
	c.resource = fmt.Sprintf("%ss", c.kind)
	// Ingress => Ingresses
	if strings.HasSuffix(c.kind, "ss") {
		c.resource = fmt.Sprintf("%ses", c.kind)
	}
	// PodSecurityPolicy => PodSecurityPolicies
	if strings.HasSuffix(c.kind, "y") {
		// Ingress => Ingresses
		c.resource = fmt.Sprintf("%sies", strings.TrimRight(c.kind, "y"))
	}

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
	re = regexp.MustCompile(`name:\s?"?([-a-zA-Z\d\.]+)`)
	matched = re.FindAllStringSubmatch(string(c.raw), -1)
	if len(matched) >= 1 && len(matched[0]) == 2 {
		c.name = matched[0][1]
	}

	// Add namespace
	if nsScoped, ok := kube.KindNamespaced[c.kind]; nsScoped || (!ok && c.namespaced) {
		c.namespace = "default"
	}

	if c.kind != "MutatingWebhookConfiguration" && c.kind != "ValidatingWebhookConfiguration" {
		re = regexp.MustCompile(`namespace:\s?"?([-a-zA-Z]+)`)
		matched = re.FindAllStringSubmatch(string(c.raw), -1)
		if len(matched) >= 1 && len(matched[0]) == 2 {
			c.namespace = matched[0][1]
		}
	}

	// Replace Data with StringData for secret object types
	c.secretStringData()

	// Pretty struct
	switch c.client {
	case types.DynamicClient:
		c.kubeObject = prettyStruct(fmt.Sprintf("%#v", c.runtimeObject))
	default:
		c.kubeObject = prettyStruct(render.AsCode(c.runtimeObject))
	}
	return nil
}

// addTypedKubeClient adds code to create typed kube client
func (c *CodeGen) addTypedKubeClient() {
	c.kubeClient = fmt.Sprintf(`var kubeconfig string
	kubeconfig, ok := os.LookupEnv("KUBECONFIG")
	if !ok {
		kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

        config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
                panic(err)
        }
        client, err := clientset.NewForConfig(config)
        if err != nil {
                panic(err)
        }
	`)

	method := fmt.Sprintf("kubeclient := client.%s%s().%s()", strings.Split(c.group, ".")[0], c.version, c.resource)
	if kube.KindNamespaced[c.kind] || c.namespace != "" {
		method = fmt.Sprintf("kubeclient := client.%s%s().%s(\"%s\")", strings.Split(c.group, ".")[0], c.version, c.resource, c.namespace)
	}
	c.kubeClient += method
}

// addDynamicKubeClient adds code to create dynamic kube client
func (c *CodeGen) addDynamicKubeClient() {
	c.kubeClient = fmt.Sprintf(`var kubeconfig string
	kubeconfig, ok := os.LookupEnv("KUBECONFIG")
	if !ok {
		kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

        config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
                panic(err)
        }
        client, err := dynamic.NewForConfig(config)
        if err != nil {
                panic(err)
        }
	`)

	group := strings.ToLower(c.group)
	if group == "core" {
		group = ""
	}
	c.kubeClient += fmt.Sprintf("gvr := schema.GroupVersionResource{Group: \"%s\", Version: \"%s\", Resource: \"%s\"}\n", group, strings.ToLower(c.version), strings.ToLower(c.resource))

	method := "kubeclient := client.Resource(gvr)"
	if kube.KindNamespaced[c.kind] || c.namespace != "" {
		method = fmt.Sprintf("kubeclient := client.Resource(gvr).Namespace(\"%s\")", c.namespace)
	}
	c.kubeClient += method
}

// addTypedKubeManage add methods to manage resource
func (c *CodeGen) addTypedKubeManage() {
	var method string

	// Add imports
	for _, i := range importer.CommonImports {
		c.imports += fmt.Sprintf("\"%s\"\n", i)
	}
	clientPkg := "k8s.io/client-go/kubernetes"
	if c.kind == crdKind {
		clientPkg = crdClient
	}
	if c.clientpkg != "" {
		clientPkg = c.clientpkg
	}
	c.imports += fmt.Sprintf("clientset \"%s\"\n", clientPkg)
	c.imports += "metav1 \"k8s.io/apimachinery/pkg/apis/meta/v1\"\n"

	methodStr := strings.Title(c.method.String())
	switch c.method {
	case types.MethodDelete:
		method = fmt.Sprintf("err = kubeclient.%s(context.TODO(), \"%s\", metav1.%sOptions{})", methodStr, c.name, methodStr)
	case types.MethodGet:
		method = fmt.Sprintf("found, err := kubeclient.%s(context.TODO(), \"%s\", metav1.%sOptions{})", methodStr, c.name, methodStr)
	default:
		method = fmt.Sprintf("_, err = kubeclient.%s(context.TODO(), object, metav1.%sOptions{})", methodStr, methodStr)
	}

	c.kubeManage = fmt.Sprintf(`%s
        if err != nil {
                panic(err)
        }
	`, method)

	if c.method != types.MethodGet {
		c.kubeManage += fmt.Sprintf(`fmt.Println("%s %sd successfully!")`, c.kind, methodStr)
		return
	}

	c.kubeManage += fmt.Sprintf(`fmt.Printf("Found object : %s", found)`, "%+v")
}

// addDynamicKubeManage add methods to manage resource
func (c *CodeGen) addDynamicKubeManage() {
	var method string

	// Add imports
	for _, i := range importer.CommonImports {
		c.imports += fmt.Sprintf("\"%s\"\n", i)
	}
	c.imports += "\"k8s.io/client-go/dynamic\"\n"
	c.imports += "\"k8s.io/apimachinery/pkg/runtime/schema\"\n"
	c.imports += "metav1 \"k8s.io/apimachinery/pkg/apis/meta/v1\"\n"

	methodStr := strings.Title(c.method.String())
	switch c.method {
	case types.MethodDelete:
		method = fmt.Sprintf("err = kubeclient.%s(context.TODO(), \"%s\", metav1.%sOptions{})", methodStr, c.name, methodStr)
	case types.MethodGet:
		method = fmt.Sprintf("found, err := kubeclient.%s(context.TODO(), \"%s\", metav1.%sOptions{})", methodStr, c.name, methodStr)
	default:
		c.imports += "\"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured\"\n"
		method = fmt.Sprintf("_, err = kubeclient.%s(context.TODO(), object, metav1.%sOptions{})", methodStr, methodStr)
	}

	c.kubeManage = fmt.Sprintf(`%s
        if err != nil {
                panic(err)
        }
	`, method)

	if c.method != types.MethodGet {
		c.kubeManage += fmt.Sprintf(`fmt.Println("%s %sd successfully!")`, c.kind, methodStr)
		return
	}

	c.kubeManage += fmt.Sprintf(`fmt.Printf("Found object : %s", found)`, "%+v")
}

// prettyCode generates final go code well indented by gofmt
func (c *CodeGen) prettyCode() (code string, err error) {
	kubeobject := fmt.Sprintf(`// Create resource object
	object := %s`, c.kubeObject)

	if c.method == types.MethodDelete || c.method == types.MethodGet {
		kubeobject = ""
	}

	main := fmt.Sprintf(`
	// Auto-generated by kyaml2go - https://github.com/PrasadG193/kyaml2go
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

	// Add pointer methods
	for _, f := range c.extraFuncs {
		main += f
	}
	// Run gofmt
	goFormat, err := format.Source([]byte(main))
	if err != nil {
		return code, fmt.Errorf("go fmt error: %s", err.Error())
	}
	return string(goFormat), nil

}

// cleanupObject removes fields with nil values
func (c *CodeGen) cleanupObject() {
	if c.method == types.MethodDelete || c.method == types.MethodGet {
		c.kubeObject = ""
	}
	kubeObject := strings.Split(c.kubeObject, "\n")
	kubeObject = replaceSubObject(kubeObject, "CreationTimestamp", "", -1)
	kubeObject = replaceSubObject(kubeObject, "Status", "", -1)
	kubeObject = replaceSubObject(kubeObject, "Generation", "", -1)
	kubeObject = removeNilFields(kubeObject)
	kubeObject = updateResources(kubeObject)

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

// secretStringData replaces binary data in resource object to readable string data
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
	// Preserve {}, do not replace them with `{\n,}`
	// The hack is to replace {} with special chars, and let the replacement happen
	// where it is required, then revert back the replacement
	obj = strings.ReplaceAll(obj, "{}", "!!")
	obj = strings.ReplaceAll(obj, ", ", ",\n")
	obj = strings.ReplaceAll(obj, "{", " {\n")
	obj = strings.ReplaceAll(obj, "}", ",\n}")
	obj = strings.ReplaceAll(obj, "!!", "{}")

	// Run gofmt
	goFormat, err := format.Source([]byte(obj))
	if err != nil {
		fmt.Println("gofmt error", err)
	}
	return string(goFormat)
}

func removeNilFields(kubeobject []string) []string {
	nilFields := []string{"nil", "\"\"", "false", "{}"}
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
			depth++
		}
		if strings.Contains(line, "}") {
			depth--
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

func (c *CodeGen) addPtrMethods() {
	object := strings.Split(c.kubeObject, "\n")
	for i, line := range object {
		var typeName, funcName, param string
		re := regexp.MustCompile(`(?m)\(&([a-zA-Z0-9]*.([a-zA-Z]*))\)\(([a-zA-Z0-9-\/"]*)\)`)
		matched := re.FindAllStringSubmatch(line, -1)
		if len(matched) == 1 && len(matched[0]) == 4 {
			typeName = matched[0][1]
			funcName = "ptr" + matched[0][2]
			if len(matched[0][2]) == 0 {
				funcName = "ptr" + matched[0][1]
			}
			param = matched[0][3]
			object[i] = strings.Replace(object[i], matched[0][0], fmt.Sprintf("%s(%s)", funcName, param), 1)
			c.extraFuncs[funcName] = fmt.Sprintf(`
			func %s(p %s) *%s { 
				return &p 
			}
			`, funcName, typeName, typeName)
			// func int%sPtr(i int%s) *int%s { return &i }
		}

		// Fix "&" => "*" values altered by go-render
		re = regexp.MustCompile(`(?m)".*[&].*"`)
		matched = re.FindAllStringSubmatch(line, -1)
		if len(matched) == 1 {
			object[i] = strings.Replace(object[i], "&", "*", -1)
		}
	}
	c.kubeObject = ""
	for _, l := range object {
		if len(l) != 0 {
			c.kubeObject += l + "\n"
		}
	}
}

// e.g "cpu": resource.Quantity(1Gi)" => "cpu": *resource.NewQuantity(700, resource.DecimalSI)"
// TODO: Use resource.MustParse() method instead
func updateResources(object []string) []string {
	resources := []string{"cpu", "memory", "storage", "pods"}
	for i, line := range object {
		for _, res := range resources {
			s := fmt.Sprintf("\"%s\": resource.Quantity", res)
			if strings.Contains(line, s) {
				value, format := parseResourceValue(object[i+1:])
				replaceSubObject(object[i:], s, fmt.Sprintf("\"%s\": *resource.NewQuantity(%s, resource.%s),", res, value, format), 1)
			}
		}
	}
	return object
}

func setCRImports(kind, group, version, pkg string, namespaced bool) {
	kube.APIPkgMap[group] = pkg
	kube.KindAPIMap[kind] = group
	kube.APIVersions[version] = true
	kube.KindNamespaced[kind] = namespaced
}
