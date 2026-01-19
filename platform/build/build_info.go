// Package build provides build info.
package build

type ServiceBuildInfo struct {
	ServiceName string
	Version     string

	// we can add more info here later.
}

func NewBuildInfo(serviceName, version string) *ServiceBuildInfo {
	return &ServiceBuildInfo{serviceName, version}
}
