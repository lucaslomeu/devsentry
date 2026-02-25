//go:build !darwin

package status

import (
	"context"
	"fmt"
)

func GetStatus(ctx context.Context, _ Options) (Status, error) {
	_ = ctx
	return Status{}, fmt.Errorf("status: not supported on this system")
}
