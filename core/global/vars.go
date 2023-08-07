package global

import (
	"context"
)

var (
	Ctx    context.Context
	Cancel context.CancelFunc
)

func init() {
	Ctx, Cancel = context.WithCancel(context.Background())
}
