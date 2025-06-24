package module

// Resource represents a module's resource that can be protected by permissions
type Resource struct {
	// Name is the unique identifier for this resource
	Name string

	// Actions are the available operations on this resource
	Actions []string

	// Description provides details about what this resource represents
	Description string

	// Metadata contains any additional information about the resource
	Metadata map[string]interface{}
}

// ResourceProvider is an interface that modules can implement to provide their resources
type ResourceProvider interface {
	// GetResources returns a list of resources provided by the module
	GetResources() []Resource
}
