package contracts

import "context"

// DeploymentPolicyProvider defines interface for deployment policy operations
type DeploymentPolicyProvider interface {
	// Validate checks if policy name is valid
	Validate(ctx context.Context, name string) (bool, error)
	
	// Generate creates deployment configuration based on host model
	Generate(ctx context.Context, name string, host interface{}) ([]byte, error)
}