package importer

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/PrasadG193/kubectl2go/pkg/kube"
	"github.com/PrasadG193/kubectl2go/pkg/stack"
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

//func (i *ImportManager) ListUsedPackages() [][]string {
// Find all external package structs used in having format "pkg.StructName"
//}

// RenamePackages renames packages having same name and modifies the kubeobject accordingly
// e.g        v1 -> corev1
//       apps/v1 -> appsv1
// extensions/v1 -> extensionsv1
func (i *ImportManager) RenamePackages() {
	// Split kubeobject to read it line by line
	kubeObject := strings.Split(i.Object, "\n")
	// Remove Creation TimeStamp
	// Refactor kubeobject
	i.Object = ""
	s := stack.New()
	vp := stack.New()
	for _, line := range kubeObject {
		//fmt.Printf("\n line: %s\n", line)
		//fmt.Printf("STACK=%+v\nVSTACK=%+v\n", s, vp)
		var re = regexp.MustCompile(PACKAGE_FORMAT)
		matched := re.FindAllStringSubmatch(line, -1)

		// Remove nil fields
		//if ifContainsNilFields(line) {
		//	continue
		//}

		// Line has external package
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

		// Line does not container external package
		if len(matched) != 1 || len(matched[0]) != 3 {
			i.Object += line + "\n"
			continue
		}

		_, version, kind := matched[0][0], matched[0][1], matched[0][2]
		//fmt.Printf("%+v %+v %+v\n", verkind, version, kind)
		// New struct started
		// if kubernetes resource, e.g v1.ObjectMeta
		if _, found := kube.ApiVersions[version]; !found {
			if strings.Contains(line, "{") {
				s.Push(version)
			}
			i.Imports[version] = ""
			i.Object += line + "\n"
			continue
		}

		// If valid kind e.g Pod
		// If not use parent package
		for k, _ := range(kube.KindApiMap) {
			if strings.Contains(kind, k) {
				kind = k
			}
		}
		group, found := kube.KindApiMap[kind]
		if !found {
			//if strings.Contains(line, "{") {
			//	s.Push(version)
			//}
			// If group is nil, use parent package
			if importAs, ok := vp.Top(); ok {
				// Modify imported struct name
				i.Object += strings.Replace(line, version, importAs.(string), 1) + "\n"
			} else {
				i.Object += line + "\n"
			}
			continue
		}

		//s.Push(kube.ApiPkgMap[group])
		//importAs := version
		//i.Imports[kube.ApiPkgMap[group]] = version
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

