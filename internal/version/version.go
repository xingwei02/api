package version

// Version is the application version, injected at build time via ldflags.
// Example: go build -ldflags "-X github.com/dujiao-next/internal/version.Version=v1.2.3"
var Version = "latest-cliv5"
