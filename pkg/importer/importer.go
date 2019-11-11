package importer

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/PrasadG193/kgoclient-gen/pkg/kube"
	"github.com/PrasadG193/kgoclient-gen/pkg/stack"
)

var COMMON_IMPORTS = []string{
	"os",
	"fmt",
	"k8s.io/client-go/kubernetes",
	"k8s.io/client-go/tools/clientcmd",
}

const PACKAGE_FORMAT = `(?m)([A-Za-z0-9]*)\.([A-Za-z]+)`

type ImportManager struct {
	Kind    string
	Group   string
	Version string
	Object  string
	Imports map[string]string
}

func New(kind, group, version, obj string) ImportManager {
	im := ImportManager{
		Kind:    kind,
		Group:   group,
		Version: version,
		Object:  obj,
	}
	im.Imports = make(map[string]string)
	// Add default imports
	for _, i := range COMMON_IMPORTS {
		im.Imports[i] = ""
	}
	return im
}

func (i *ImportManager) FindImports() (string, string) {
	//	pkgList := i.ListUsedPackages()
	i.RenamePackages()

	var imports string
	for k, v := range i.Imports {
		imports += fmt.Sprintf("%s \"%s\"\n", v, k)
	}

	return imports, i.Object
}

// RenamePackages renames packages having same name and modifies the kubeobject accordingly
// e.g        v1 -> corev1
//       apps/v1 -> appsv1
// extensions/v1 -> extensionsv1
func (i *ImportManager) RenamePackages() {
	// Split kubeobject to read it line by line
	kubeObject := strings.Split(i.Object, "\n")
	i.Object = ""
	// Stack to store imported struct's pkg name
	s := stack.New()
	// Stack to store imported valid k8s struct's pkg name
	vp := stack.New()
	for _, line := range kubeObject {
		// Check if end of the struct
		// Pop package names from stack
		if strings.EqualFold(line, "},") {
			if p, ok := s.Pop(); ok {
				parentPkg, _ := vp.Top()
				if reflect.DeepEqual(parentPkg, p) {
					vp.Pop()
				}
			}
			i.Object += line + "\n"
			continue
		}

		// Line has external package
		var re = regexp.MustCompile(PACKAGE_FORMAT)
		matched := re.FindAllStringSubmatch(line, -1)
		if len(matched) != 1 || len(matched[0]) != 3 {
			i.Object += line + "\n"
			continue
		}

		_, version, kind := matched[0][0], matched[0][1], matched[0][2]
		// New struct started
		// if kubernetes resource, e.g v1.ObjectMeta
		if _, found := kube.ApiVersions[version]; !found {
			if strings.Contains(line, "{") {
				s.Push(version)
			}

			i.Object += line + "\n"
			// Extract package name from full path
			if v, ok := kube.ApiPkgMap[version]; ok {
				version = v
			}
			i.Imports[version] = ""
			continue
		}

		// If part of valid kind e.g PodSpec
		for k, _ := range kube.KindApiMap {
			if strings.Contains(kind, k) {
				kind = k
			}
		}

		// If not valid kind
		group, found := kube.KindApiMap[kind]
		if !found {
			// If group is nil, use parent package
			if importAs, ok := vp.Top(); ok {
				// Modify imported struct name
				i.Object += strings.Replace(line, version, importAs.(string), 1) + "\n"
			} else {
				i.Object += line + "\n"
			}
			continue
		}

		// Extract package name from complete package path
		p := strings.Split(kube.ApiPkgMap[group], "/")
		importAs := p[len(p)-1] + version
		if strings.Contains(line, "{") {
			vp.Push(importAs)
			s.Push(importAs)
			i.Imports[kube.ApiPkgMap[group]+"/"+version] = importAs
		}

		// Modify imported struct name
		i.Object += strings.Replace(line, version, importAs, 1) + "\n"
	}
}
