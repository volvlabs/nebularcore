package auth

import (
	"github.com/casbin/casbin/v2"
)

type AclConfig struct {
	Role       string
	PolicyPath string
	ConfPath   string
}

type AccessControlManager struct {
	aclConfigs map[string]AclConfig
}

func NewAccessControlManager() *AccessControlManager {
	return &AccessControlManager{
		aclConfigs: make(map[string]AclConfig),
	}
}

func (a *AccessControlManager) Register(role, policyPath, confPath string) {
	a.aclConfigs[role] = AclConfig{
		Role:       role,
		PolicyPath: policyPath,
		ConfPath:   confPath,
	}
}

func (a *AccessControlManager) RegisterAll(aclConfigs []AclConfig) {
	for _, aclConfig := range aclConfigs {
		a.aclConfigs[aclConfig.Role] = aclConfig
	}
}

func (a *AccessControlManager) IsAuthroized(role, resource, method string) (bool, error) {
	aclConfig := a.aclConfigs[role]
	authEnforcer, err := casbin.NewEnforcer(aclConfig.ConfPath, aclConfig.PolicyPath)
	if err != nil {
		return false, err
	}
	res, err := authEnforcer.Enforce(role, resource, method)
	if err != nil {
		return false, err
	}

	return res, nil
}
