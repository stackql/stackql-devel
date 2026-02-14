package providerinvoker

import "context"

// Request carries an invoker-specific payload. Step 1 keeps this opaque so we can
// stabilise call-shapes in StackQL before lift/shift into any-sdk.
type Request struct {
	Payload any
}

// Result is a protocol-agnostic execution result from StackQL's point of view.
type Result struct {
	Body     any
	Messages []string
}

type Invoker interface {
	Invoke(ctx context.Context, req Request) (Result, error)
}
