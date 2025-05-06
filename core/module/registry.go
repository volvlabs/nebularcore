package module

import "fmt"

// Registry manages module registration and dependencies
type Registry struct {
	// Separate maps for public and tenant modules
	publicModules map[string]Module
	tenantModules map[string]Module
	// Initialization order for each namespace
	publicOrder []string
	tenantOrder []string
}

func NewRegistry() *Registry {
	return &Registry{
		publicModules: make(map[string]Module),
		tenantModules: make(map[string]Module),
	}
}

func (r *Registry) Register(m Module) error {
	// Get the appropriate module map based on namespace
	moduleMap := r.getModuleMap(m.Namespace())
	if _, exists := moduleMap[m.Name()]; exists {
		return fmt.Errorf("module already registered in %s namespace: %s", m.Namespace(), m.Name())
	}

	// Register the module in its namespace
	moduleMap[m.Name()] = m
	return r.updateInitOrder(m.Namespace())
}

// Get retrieves a module by name from either namespace
func (r *Registry) Get(name string) (Module, bool) {
	// Try public namespace first
	if m, ok := r.publicModules[name]; ok {
		return m, true
	}
	// Try tenant namespace
	m, ok := r.tenantModules[name]
	return m, ok
}

// GetByNamespace retrieves a module by name from a specific namespace
func (r *Registry) GetByNamespace(name string, namespace ModuleNamespace) (Module, bool) {
	moduleMap := r.getModuleMap(namespace)
	m, ok := moduleMap[name]
	return m, ok
}

// getModuleMap returns the appropriate module map for a given namespace
func (r *Registry) getModuleMap(namespace ModuleNamespace) map[string]Module {
	switch namespace {
	case PublicNamespace:
		return r.publicModules
	case TenantNamespace:
		return r.tenantModules
	default:
		return r.publicModules // Default to public for safety
	}
}

func (r *Registry) InitOrder() []string {
	// Combine both orders, with public modules first
	return append(r.publicOrder, r.tenantOrder...)
}

// GetModules returns all modules from both namespaces
func (r *Registry) GetModules() map[string]Module {
	all := make(map[string]Module)
	// Add public modules
	for k, v := range r.publicModules {
		all[k] = v
	}
	// Add tenant modules
	for k, v := range r.tenantModules {
		all[k] = v
	}
	return all
}

// GetModulesByNamespace returns modules for a specific namespace
func (r *Registry) GetModulesByNamespace(namespace ModuleNamespace) map[string]Module {
	return r.getModuleMap(namespace)
}

// updateInitOrder updates the initialization order based on dependencies for a specific namespace
func (r *Registry) updateInitOrder(namespace ModuleNamespace) error {
	// Get the appropriate module map and order slice
	moduleMap := r.getModuleMap(namespace)
	var order *[]string
	switch namespace {
	case PublicNamespace:
		order = &r.publicOrder
	case TenantNamespace:
		order = &r.tenantOrder
	default:
		return fmt.Errorf("invalid namespace: %s", namespace)
	}

	// Reset order for this namespace
	*order = nil

	// Track visited and temp marks for cycle detection
	visited := make(map[string]bool)
	temp := make(map[string]bool)

	// Visit all modules in this namespace
	for name := range moduleMap {
		if !visited[name] {
			if err := r.visitInNamespace(name, namespace, visited, temp, order); err != nil {
				return err
			}
		}
	}

	return nil
}

// visitInNamespace performs a topological sort using depth-first search within a namespace
func (r *Registry) visitInNamespace(name string, namespace ModuleNamespace, visited, temp map[string]bool, order *[]string) error {
	if temp[name] {
		return fmt.Errorf("cyclic dependency detected involving module %s in namespace %s", name, namespace)
	}
	if visited[name] {
		return nil
	}

	temp[name] = true

	// Visit dependencies
	moduleMap := r.getModuleMap(namespace)
	module := moduleMap[name]
	for _, dep := range module.Dependencies() {
		// Check if dependency exists in the same namespace
		if _, ok := moduleMap[dep]; !ok {
			return fmt.Errorf("module %s in namespace %s depends on missing module %s", name, namespace, dep)
		}
		if err := r.visitInNamespace(dep, namespace, visited, temp, order); err != nil {
			return err
		}
	}

	temp[name] = false
	visited[name] = true
	*order = append(*order, name)

	return nil
}
