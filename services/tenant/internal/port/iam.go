package port

import "context"

// IAMClient bootstraps IAM state for a new tenant.
type IAMClient interface {
	BootstrapOwner(ctx context.Context, tenantID, userID, authorization string) error
}
