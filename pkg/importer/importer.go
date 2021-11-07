package importer

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/PrasadG193/kyaml2go/pkg/kube"
	"github.com/PrasadG193/kyaml2go/pkg/stack"
	"github.com/PrasadG193/kyaml2go/pkg/types"
)

// CommonImports contains common packages required for all the resources
var CommonImports = []string{
	"os",
	"fmt",
	"context",
	"path/filepath",
	"k8s.io/client-go/tools/clientcmd",
	"k8s.io/client-go/util/homedir",
}

const packageFormat = `(?m)[*&\]\s]([A-Za-z0-9]*?)\.?([A-Za-z0-9]+)[{|()]`

// ImportManager holds resource Kind, Group, Versions to figure out imports
type ImportManager struct {
	Kind    string
	Group   string
	Version string
	Object  string
	Imports map[string]string
}

// New returns an instance of ImportManager
func New(kind, group, version, obj, cliPkg string, clientType types.Client) ImportManager {
	im := ImportManager{
		Kind:    kind,
		Group:   group,
		Version: version,
		Object:  obj,
	}
	im.Imports = make(map[string]string)
	return im
}

// FindImports figures out imports required to compile generated Go code
func (i *ImportManager) FindImports() (string, string) {
	i.importAndRename()

	var imports string
	for k, v := range i.Imports {
		imports += fmt.Sprintf("%s \"%s\"\n", v, k)
	}

	return imports, i.Object
}

// importAndRename finds out if the external package needs to be imported.
// Renames packages having same name and modifies the kubeobject accordingly
// e.g        v1 -> corev1
//       apps/v1 -> appsv1
// extensions/v1 -> extensionsv1
func (i *ImportManager) importAndRename() {
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
		if strings.HasSuffix(line, "},") {
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
		var re = regexp.MustCompile(packageFormat)
		matched := re.FindAllStringSubmatch(line, -1)
		if len(matched) < 1 || len(matched[0]) < 3 {
			i.Object += line + "\n"
			continue
		}

		_, version, kind := matched[0][0], matched[0][1], matched[0][2]
		// if builtin type
		if len(version) == 0 {
			if strings.Contains(line, "{") {
				s.Push(kind)
			}
			i.Object += line + "\n"
			continue
		}

		// New struct started
		// if kubernetes resource, e.g v1.ObjectMeta
		if _, found := kube.APIVersions[version]; !found {
			if strings.Contains(line, "{") {
				s.Push(version)
			}

			i.Object += line + "\n"
			// Extract package name from full path
			if v, ok := kube.APIPkgMap[version]; ok {
				version = v
			}
			i.Imports[version] = ""
			continue
		}

		// If kind specs
		if _, found := kube.KindAPIMap[strings.TrimSuffix(kind, "Spec")]; found {
			kind = strings.TrimSuffix(kind, "Spec")
		}

		// If part of valid kind e.g DeploymentStrategy
		if _, found := kube.KindAPIMap[kind]; !found {
			for parent := range kube.KindAPIMap {
				if strings.HasPrefix(kind, parent) {
					kind = parent
				}
			}
		}

		// If not valid kind
		group, found := kube.KindAPIMap[kind]
		if !found {
			// If group is nil, use parent package
			if importAs, ok := vp.Top(); ok {
				// Modify imported struct name
				i.Object += strings.Replace(line, version, importAs.(string), 1) + "\n"
				if strings.Contains(line, "{") {
					vp.Push(importAs)
					s.Push(importAs.(string))
				}
			} else {
				i.Object += line + "\n"
			}
			continue
		}

		// Extract package name from complete package path
		p := strings.Split(kube.APIPkgMap[group], "/")
		importAs := p[len(p)-1] + version
		i.Imports[kube.APIPkgMap[group]+"/"+version] = importAs
		if strings.Contains(line, "{") {
			vp.Push(importAs)
			s.Push(importAs)
		}

		// Modify imported struct name
		i.Object += strings.Replace(line, version, importAs, 1) + "\n"
	}
}
