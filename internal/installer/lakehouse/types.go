package lakehouse

import "context"

// RunCommand executes a command with the provided cancellation deadline.
type RunCommand func(context.Context, string, ...string) ([]byte, error)

// Progress receives credential-free installation progress messages.
type Progress func(string)
