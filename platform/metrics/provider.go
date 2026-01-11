package metrics

import "github.com/google/wire"

// ProviderSet is metrics providers for Wire.
var ProviderSet = wire.NewSet(
	NewRegistry,
)
