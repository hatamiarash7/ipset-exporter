// Package buildinfo holds metadata injected with go build -ldflags -X.
package buildinfo

var (
	Version   = "dev"
	Revision  = "unknown"
	BuildDate = "unknown"
)
