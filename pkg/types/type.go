package types

// Client defines client type for Kubernetes APIs
type Client int

// KubeMethod define methods to manage K8s resource
type KubeMethod string

const (
	// MethodCreate to create K8s resource
	MethodCreate = "create"
	// MethodGet to get K8s resource
	MethodGet = "get"
	// MethodUpdate to update K8s resource
	MethodUpdate = "update"
	// MethodDelete to delete K8s resource
	MethodDelete = "delete"

	// Client types
	TypedClient Client = iota
	DynamicClient
)

func (m KubeMethod) String() string {
	return string(m)
}
