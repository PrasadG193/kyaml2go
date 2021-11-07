//const URL = "http://localhost:8080/v1/convert"
let BASE_URL = "https://kyaml2go-oymillvkxq-uc.a.run.app/v1/convert?method=";

let go = document.getElementById("goGenerator");
let codecopied = document.getElementById("codecopied");
let editor = "";
let CRDetailValid = true;
let changeTheme = document.getElementById("changeTheme");
let settheme = document.body.getAttribute("dark-theme");
let codeMirrorMainClass = document.getElementsByClassName(
	"CodeMirror cm-s-default"
);
const MODE_LIGHT  = "light"
const MODE_DARK = "dark"

// set theme
function applyTheme() {
	settheme = document.body.getAttribute("dark-theme");
	settheme == MODE_LIGHT
		? document.body.setAttribute("dark-theme", MODE_DARK)
		: document.body.setAttribute("dark-theme", MODE_LIGHT);
	let value = settheme == MODE_LIGHT ? MODE_DARK : MODE_LIGHT
	localStorage.setItem("camelTheme", `${value}`);
	changeEditorMode();
}

// apply mode
changeTheme.addEventListener("change", (event) => {
	applyTheme();
});

// Codemirror modes
editor = CodeMirror.fromTextArea(document.getElementById("yamlspecs"), {
	lineNumbers: true,
	mode: "text/x-yaml",
});

go = CodeMirror.fromTextArea(document.getElementById("goGenerator"), {
	lineNumbers: true,
	mode: "text/x-go",
});

window.generatorCall = function (client, action, query) {
	URL = formGooleFuncURL(action) + query;
	if (client == "type_dynamic") {
		URL = URL + "&dynamic=true";
	}

	let yamlData = document.getElementById("yamlspecs").value;
	document.getElementById("yamlspecs").style.border = "1px solid #ced4da";
	yamlData = editor.getValue();
	$.ajax({
		url: `${URL}`,
		type: "POST",
		data: yamlData,
		success: function (data) {
			document.getElementById("error").style.display = "none";
			document.getElementById("err-span").innerHTML = "";
			go.setValue(data);
		},
		error: function (jqXHR, request, error) {
			document.getElementById("yamlspecs").style.border = "1px solid red";
			if (jqXHR.status == 400) {
				if (isCRChecked()) {
					go.setOption("lineWrapping", true);
					displayError(
						"Invalid Kubernetes resource spec or package names. Please check the spec and try again."
					);
					go.setValue(jqXHR.responseText);
				} else {
					displayError(
						"Invalid Kubernetes resource spec. Please check the spec and try again."
					);
				}
			} else {
				displayError(
					"Something went wrong! Please report this to https://github.com/PrasadG193/kyaml2go/issues"
				);
			}
		},
	});
};

function formGooleFuncURL(action) {
	return BASE_URL + action;
}

// checks if cr checkbox is checked
function isCRChecked() {
	return document.getElementById("cr_check").checked;
}

function getValue(id) {
	v = document.getElementById(id).value.trim();
	return v;
}

client = document.getElementById("selectclient");

//Convert
dropDown = document.getElementById("selectaction");
document.getElementById("convert").addEventListener("click", () => {
	action = dropDown.value;
	if (action != "select") {
		hideError();
		query = "";
		if (isCRChecked()) {
			CRDetailValid = true;
			scheme = getValue("scheme");
			if (scheme.length == 0) {
				CRDetailValid = false;
			}
			apis = getValue("apis");
			clients = getValue("client");
			query =
				"&cr=true&scheme=" + scheme + "&apis=" + apis + "&client=" + clients;
		}

		if (CRDetailValid) {
			go.setValue("Generating...");
			hideError();
			generatorCall(client.value, action, query);
		} else {
			displayError("Please enter correct CR details. scheme pkg is required.");
		}
	} else {
		displayError("Please select the method.");
	}
});

//Clear YAML
document.getElementById("clearYaml").addEventListener("click", () => {
	editor.setValue("");
});

//Clear Go
document.getElementById("clearGo").addEventListener("click", () => {
	go.setValue("");
});

// Set sample specs and code
$(document).ready(setDeploymentSample());

function displayError(err) {
	document.getElementById("err-span").innerHTML = err;
	document.getElementById("error").style.display = "block";
}

function hideError() {
	document.getElementById("err-span").innerHTML = "";
	document.getElementById("error").style.display = "none";
	go.setOption("lineWrapping", false);
}

document.getElementById("copybutton").addEventListener("click", function () {
	if (window.navigator) {
		navigator.clipboard.writeText(go.getValue());
		codecopied.style.display = "flex";
		window.setTimeout(function () {
			codecopied.style.display = "none";
		}, 700);
	}
});

document.getElementById("cr_check").addEventListener("change", function () {
	hideError();
	if (isCRChecked()) {
		document.getElementById("cr_params").style.display = "block";
		setCRSample();
	} else {
		document.getElementById("cr_params").style.display = "none";
		setDeploymentSample();
	}
});

document.getElementById("selectclient").addEventListener("change", function () {
	if (client.value == "type_dynamic") {
		document.getElementById("cr_box").style.display = "none";
		document.getElementById("cr_params").style.display = "none";
		setDynamicClientDeploySample()
		go.setValue("");
	} else {
		document.getElementById("cr_box").style.display = "block";
		if (isCRChecked()) {
			document.getElementById("cr_params").style.display = "block";
		}
	}
});

function setDeploymentSample() {
	// Add sample input
	editor.setValue(`# Paste your Kubernetes yaml spec here...
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: demo
  template:
    metadata:
      labels:
        app: demo
    spec:
      containers:
      - image: nginx:1.12
        imagePullPolicy: IfNotPresent
        name: web
        ports:
        - containerPort: 80
          name: http
          protocol: TCP
`);
	if (client.value == "type_dynamic") {
		setDynamicClientDeploySample()
	} else {
		setTypedClientDeploySample()
	}
}

function setTypedClientDeploySample() {
	// Add sample output
	go.setValue(`// Auto-generated by kyaml2go - https://github.com/PrasadG193/kyaml2go
package main

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

func main() {
	// Create client
	var kubeconfig string
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
	kubeclient := clientset.AppsV1().Deployments("default")

	// Create resource object
	object := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptrint32(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									Name:          "http",
									HostPort:      0,
									ContainerPort: 80,
									Protocol:      corev1.Protocol("TCP"),
								},
							},
							Resources:       corev1.ResourceRequirements{},
							ImagePullPolicy: corev1.PullPolicy("IfNotPresent"),
						},
					},
				},
			},
			Strategy:        appsv1.DeploymentStrategy{},
			MinReadySeconds: 0,
		},
	}

	// Manage resource
	_, err = kubeclient.Create(object)
	if err != nil {
		panic(err)
	}
	fmt.Println("Deployment Created successfully!")
}

func ptrint32(p int32) *int32 {
	return &p
}
  `);

}

function setDynamicClientDeploySample() {
	// Add sample output
	go.setValue(`// Auto-generated by kyaml2go - https://github.com/PrasadG193/kyaml2go
package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

func main() {
	// Create client
	var kubeconfig string
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
	gvr := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

	// Create resource object
	object := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test-deployment",
				"namespace": "mysql",
			},
			"spec": map[string]interface{}{
				"replicas": 2,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "demo",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app": "demo",
						},
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"image":           "nginx:1.12",
								"imagePullPolicy": "IfNotPresent",
								"name":            "web",
								"ports": []interface{}{
									map[string]interface{}{
										"containerPort": 80,
										"name":          "http",
										"protocol":      "TCP",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Manage resource
	_, err = client.Resource(gvr).Namespace("mysql").Create(context.TODO(), object, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Deployment Created successfully!")
}
  `);

}

function setCRSample() {
	document.getElementById("scheme").value =
		"k8s.io/sample-controller/pkg/generated/clientset/versioned/scheme";
	document.getElementById("apis").value =
		"k8s.io/sample-controller/pkg/apis/samplecontroller";
	document.getElementById("client").value =
		"k8s.io/sample-controller/pkg/generated/clientset/versioned";

	// Add sample input
	editor.setValue(`# Paste your Kubernetes yaml spec here...
apiVersion: samplecontroller.k8s.io/v1alpha1
kind: Foo
metadata:
  name: example-foo
spec:
  deploymentName: example-foo
  replicas: 1
   `);
	if (client.value == "type_dynamic") {
		setDynamicClientCRSample()
	} else {
		setTypedClientCRSample()
	}
}

function setTypedClientCRSample() {
	// Add sample output
	go.setValue(`// Auto-generated by kyaml2go - https://github.com/PrasadG193/kyaml2go
package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	samplecontrollerv1alpha1 "k8s.io/sample-controller/pkg/apis/samplecontroller/v1alpha1"
	clientset "k8s.io/sample-controller/pkg/generated/clientset/versioned"
	"os"
	"path/filepath"
)

func main() {
	// Create client
	var kubeconfig string
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
	kubeclient := client.SamplecontrollerV1alpha1().Foos("default")

	// Create resource object
	object := &samplecontrollerv1alpha1.Foo{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Foo",
			APIVersion: "samplecontroller.k8s.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-foo",
		},
		Spec: samplecontrollerv1alpha1.FooSpec{
			DeploymentName: "example-foo",
			Replicas:       ptrint32(1),
		},
	}

	// Manage resource
	_, err = kubeclient.Create(context.TODO(), object, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Foo Created successfully!")
}

func ptrint32(p int32) *int32 {
	return &p
}
    `);
}

function setDynamicClientCRSample() {
	// Add sample output
	go.setValue(`// Auto-generated by kyaml2go - https://github.com/PrasadG193/kyaml2go
// Auto-generated by kyaml2go - https://github.com/PrasadG193/kyaml2go
package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

func main() {
	// Create client
	var kubeconfig string
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
	gvr := schema.GroupVersionResource{Group: "samplecontroller.k8s.io", Version: "v1alpha1", Resource: "foos"}

	// Create resource object
	object := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "samplecontroller.k8s.io/v1alpha1",
			"kind":       "Foo",
			"metadata": map[string]interface{}{
				"name": "example-foo",
			},
			"spec": map[string]interface{}{
				"deploymentName": "example-foo",
				"replicas":       1,
			},
		},
	}

	// Manage resource
	_, err = client.Resource(gvr).Namespace("default").Create(context.TODO(), object, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Foo Created successfully!")
}
    `);
}

function changeEditorMode() {
	for (let i = 0; i < codeMirrorMainClass.length; i++) {
		document
			.getElementsByClassName("CodeMirror")
		[i].classList.toggle("cm-s-material-darker");
	}
}

// Check if theme mode is already set
if (localStorage.getItem("camelTheme")) {
	const value = localStorage.getItem("camelTheme");
	document.body.setAttribute("dark-theme", value);
	value !== MODE_LIGHT && document.getElementById("theme-check").setAttribute("checked", "checked");
	value !== MODE_LIGHT && changeEditorMode();
}
