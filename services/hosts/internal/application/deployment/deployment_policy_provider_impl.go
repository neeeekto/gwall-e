package host_deployment

import (
	"context"
	"github.com/gwall-e/hosts/internal/domain/projects/contracts"
)

// DeploymentPolicyProviderImpl implements DeploymentPolicyProvider interface
type DeploymentPolicyProviderImpl struct{}

// NewDeploymentPolicyProvider creates new DeploymentPolicyProvider instance
func NewDeploymentPolicyProvider() contracts.DeploymentPolicyProvider {
	return &DeploymentPolicyProviderImpl{}
}

// Validate checks if policy name is valid
func (p *DeploymentPolicyProviderImpl) Validate(ctx context.Context, name string) (bool, error) {
	// TODO: implement actual validation logic
	return true, nil
}

// Generate creates deployment configuration based on host model
func (p *DeploymentPolicyProviderImpl) Generate(ctx context.Context, name string, host interface{}) ([]byte, error) {
	// TODO: implement actual generation logic
	return []byte{}, nil
}