package module_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volvlabs/nebularcore/core/module"
	"github.com/volvlabs/nebularcore/core/module/mocks"
)

func TestRegistry_Register(t *testing.T) {
	tests := []struct {
		name    string
		module  func() module.Module
		wantErr bool
	}{
		{
			name: "register public module successfully",
			module: func() module.Module {
				testModule := mocks.NewModule(t)
				testModule.On("Name").Return("mod1")
				testModule.On("Dependencies").Return([]string{})
				testModule.On("Namespace").Return(module.PublicNamespace)
				return testModule
			},
			wantErr: false,
		},
		{
			name: "register tenant module successfully",
			module: func() module.Module {
				testModule := mocks.NewModule(t)
				testModule.On("Name").Return("mod2")
				testModule.On("Dependencies").Return([]string{})
				testModule.On("Namespace").Return(module.TenantNamespace)
				return testModule
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := module.NewRegistry()
			testModule := tt.module()
			err := r.Register(testModule)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify module was registered
				m, exists := r.GetByNamespace(testModule.Name(), testModule.Namespace())
				assert.True(t, exists)
				assert.Equal(t, testModule.Name(), m.Name())
			}
		})
	}
}

func TestRegistry_DuplicateRegistration(t *testing.T) {
	r := module.NewRegistry()
	mod := mocks.NewModule(t)
	mod.On("Name").Return("mod1")
	mod.On("Namespace").Return(module.PublicNamespace)
	mod.On("Dependencies").Return([]string{})

	// First registration should succeed
	assert.NoError(t, r.Register(mod))

	// Second registration should fail
	err := r.Register(mod)
	assert.Error(t, err)
}

func TestRegistry_DependencyResolution(t *testing.T) {
	r := module.NewRegistry()

	// Create modules with dependencies
	mod1 := mocks.NewModule(t)
	mod1.On("Name").Return("mod1")
	mod1.On("Dependencies").Return([]string{})
	mod1.On("Namespace").Return(module.PublicNamespace)
	mod2 := mocks.NewModule(t)
	mod2.On("Name").Return("mod2")
	mod2.On("Dependencies").Return([]string{"mod1"})
	mod2.On("Namespace").Return(module.PublicNamespace)
	mod3 := mocks.NewModule(t)
	mod3.On("Name").Return("mod3")
	mod3.On("Dependencies").Return([]string{"mod2"})
	mod3.On("Namespace").Return(module.PublicNamespace)

	// Register in reverse order to test dependency resolution
	assert.NoError(t, r.Register(mod1))
	assert.NoError(t, r.Register(mod2))
	assert.NoError(t, r.Register(mod3))

	// Check initialization order
	order := r.InitOrder()
	assert.Equal(t, []string{"mod1", "mod2", "mod3"}, order)
}

func TestRegistry_GetModules(t *testing.T) {
	r := module.NewRegistry()

	// Create and register modules in different namespaces
	pubMod := mocks.NewModule(t)
	pubMod.On("Name").Return("pub1")
	pubMod.On("Namespace").Return(module.PublicNamespace)
	pubMod.On("Dependencies").Return([]string{})
	tenantMod := mocks.NewModule(t)
	tenantMod.On("Name").Return("tenant1")
	tenantMod.On("Namespace").Return(module.TenantNamespace)
	tenantMod.On("Dependencies").Return([]string{})

	assert.NoError(t, r.Register(pubMod))
	assert.NoError(t, r.Register(tenantMod))

	// Test GetModules
	allMods := r.GetModules()
	assert.Len(t, allMods, 2)
	assert.Contains(t, allMods, "pub1")
	assert.Contains(t, allMods, "tenant1")

	// Test GetModulesByNamespace
	pubMods := r.GetModulesByNamespace(module.PublicNamespace)
	assert.Len(t, pubMods, 1)
	assert.Contains(t, pubMods, "pub1")

	tenantMods := r.GetModulesByNamespace(module.TenantNamespace)
	assert.Len(t, tenantMods, 1)
	assert.Contains(t, tenantMods, "tenant1")
}

func TestRegistry_GetModulesByNamespace_Order(t *testing.T) {
	tests := []struct {
		name      string
		modules   []module.Module
		namespace module.ModuleNamespace
		expected  []string
	}{
		{
			name: "public namespace FIFO order",
			modules: func() []module.Module {
				m1 := mocks.NewModule(t)
				m1.On("Name").Return("mod1")
				m1.On("Dependencies").Return([]string{})
				m1.On("Namespace").Return(module.PublicNamespace)

				m2 := mocks.NewModule(t)
				m2.On("Name").Return("mod2")
				m2.On("Dependencies").Return([]string{"mod1"})
				m2.On("Namespace").Return(module.PublicNamespace)

				m3 := mocks.NewModule(t)
				m3.On("Name").Return("mod3")
				m3.On("Dependencies").Return([]string{"mod2"})
				m3.On("Namespace").Return(module.PublicNamespace)

				return []module.Module{m1, m2, m3}
			}(),
			namespace: module.PublicNamespace,
			expected:  []string{"mod1", "mod2", "mod3"},
		},
		{
			name: "tenant namespace FIFO order",
			modules: func() []module.Module {
				m1 := mocks.NewModule(t)
				m1.On("Name").Return("tenant1")
				m1.On("Dependencies").Return([]string{})
				m1.On("Namespace").Return(module.TenantNamespace)

				m2 := mocks.NewModule(t)
				m2.On("Name").Return("tenant2")
				m2.On("Dependencies").Return([]string{"tenant1"})
				m2.On("Namespace").Return(module.TenantNamespace)

				return []module.Module{m1, m2}
			}(),
			namespace: module.TenantNamespace,
			expected:  []string{"tenant1", "tenant2"},
		},
		{
			name: "mixed namespaces should not affect order",
			modules: func() []module.Module {
				m1 := mocks.NewModule(t)
				m1.On("Name").Return("pub1")
				m1.On("Dependencies").Return([]string{})
				m1.On("Namespace").Return(module.PublicNamespace)

				m2 := mocks.NewModule(t)
				m2.On("Name").Return("tenant1")
				m2.On("Dependencies").Return([]string{})
				m2.On("Namespace").Return(module.TenantNamespace)

				m3 := mocks.NewModule(t)
				m3.On("Name").Return("pub2")
				m3.On("Dependencies").Return([]string{"pub1"})
				m3.On("Namespace").Return(module.PublicNamespace)

				return []module.Module{m1, m2, m3}
			}(),
			namespace: module.PublicNamespace,
			expected:  []string{"pub1", "pub2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := module.NewRegistry()

			for _, m := range tt.modules {
				fmt.Println("Registering module:", m.Name(), "test name: ", tt.name)
				assert.NoError(t, r.Register(m))
			}

			modules := r.GetModulesInOrder(tt.namespace)

			var actualOrder []string
			for _, module := range modules {
				fmt.Println("Retrieved module:", module.Name, "test name: ", tt.name)
				actualOrder = append(actualOrder, module.Name)
			}

			assert.Equal(t, tt.expected, actualOrder, "module order should match registration order")

			assert.Equal(t, len(tt.expected), len(modules), "number of modules should match")
			for _, name := range tt.expected {
				exists := false
				for _, module := range modules {
					if module.Name == name {
						exists = true
						break
					}
				}

				assert.True(t, exists, "module %s should exist", name)
			}
		})
	}
}

func TestRegistry_MissingDependency(t *testing.T) {
	r := module.NewRegistry()

	// Create module with missing dependency
	mod := mocks.NewModule(t)
	mod.On("Name").Return("mod1")
	mod.On("Dependencies").Return([]string{"missing"})
	mod.On("Namespace").Return(module.PublicNamespace)

	// Registration should fail due to missing dependency
	err := r.Register(mod)
	assert.Error(t, err)
}
