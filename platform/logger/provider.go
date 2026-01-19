// Package logger
package logger

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewWatermillLogger)
