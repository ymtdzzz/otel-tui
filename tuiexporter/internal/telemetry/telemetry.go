package telemetry

import "go.opentelemetry.io/collector/pdata/pcommon"

// GetServiceNameFromResource returns service name from given resource attributes.
// If the key `service.name` is contained, it returns `unknown`.
func GetServiceNameFromResource(resource pcommon.Resource) string {
	sname, ok := resource.Attributes().Get("service.name")
	if ok {
		return sname.AsString()
	}
	return "unknown"
}
