package auth

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-playground/assert/v2"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
)

func TestRegisterNewRoleWithValidPaths(t *testing.T) {
	// Arrange:
	acm := NewAccessControlManager()

	// Act:
	acm.Register("role1", "policy1", "conf1")

	// Assert:
	assert.Equal(t, "role1", acm.aclConfigs["role1"].Role)
	assert.Equal(t, "policy1", acm.aclConfigs["role1"].PolicyPath)
	assert.Equal(t, "conf1", acm.aclConfigs["role1"].ConfPath)
}

func TestRegisterMultipleRolesWithValidPaths(t *testing.T) {
	// Arrange:
	acm := NewAccessControlManager()

	aclConfigs := []AclConfig{
		{Role: "role1", PolicyPath: "policy1", ConfPath: "conf1"},
		{Role: "role2", PolicyPath: "policy2", ConfPath: "conf2"},
		{Role: "role3", PolicyPath: "policy3", ConfPath: "conf3"},
	}

	// Act:
	acm.RegisterAll(aclConfigs)

	// Assert:
	assert.Equal(t, "role1", acm.aclConfigs["role1"].Role)
	assert.Equal(t, "policy1", acm.aclConfigs["role1"].PolicyPath)
	assert.Equal(t, "conf1", acm.aclConfigs["role1"].ConfPath)

	assert.Equal(t, "role2", acm.aclConfigs["role2"].Role)
	assert.Equal(t, "policy2", acm.aclConfigs["role2"].PolicyPath)
	assert.Equal(t, "conf2", acm.aclConfigs["role2"].ConfPath)

	assert.Equal(t, "role3", acm.aclConfigs["role3"].Role)
	assert.Equal(t, "policy3", acm.aclConfigs["role3"].PolicyPath)
	assert.Equal(t, "conf3", acm.aclConfigs["role3"].ConfPath)
}

func TestCheckAuthorizationForRegisteredRoleWithValidResourceAndMethod(t *testing.T) {
	// Arrange:
	rootDir := filesystem.GetRootDir("../../")
	policyPath := fmt.Sprintf("%s/test/data/policy.csv", rootDir)
	modelPath := fmt.Sprintf("%s/test/data/model.conf", rootDir)
	acm := NewAccessControlManager()
	acm.Register("admin", policyPath, modelPath)
	acm.Register("user", policyPath, modelPath)
	acm.Register("developer", policyPath, modelPath)

	scenarios := []struct {
		name       string
		role       string
		resource   string
		method     string
		authorized bool
	}{
		{name: "should authorize admin role on allowed resource", role: "admin", resource: "/admin/create", method: http.MethodGet, authorized: true},
		{name: "should not authorize user role on admin resource", role: "user", resource: "/admin/create", method: http.MethodGet, authorized: false},
		{name: "should authorize user role on allowed resource", role: "user", resource: "/api/register", method: http.MethodGet, authorized: true},
		{name: "should not authorize developer role on admin resource", role: "developer", resource: "/admin/create", method: http.MethodGet, authorized: false},
		{name: "should authorize developer role on allowed resource", role: "developer", resource: "/dev/logs", method: http.MethodGet, authorized: true},
	}

	// Act:
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			authorized, err := acm.IsAuthroized(scenario.role, scenario.resource, scenario.method)
			if err != nil {
				t.Errorf("AcmManager.IsAuthorized(), got error %v\n", err)
			}

			// Assert:
			assert.Equal(t, scenario.authorized, authorized)
		})
	}
}
