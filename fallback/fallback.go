package fallback

import (
	"context"
)

type IFallbackHandler interface {
	Handle(ctx context.Context, step, id, payload string, err error) error
}
