package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestGetServiceNameFromResource(t *testing.T) {
	t.Run("With_Service_Name", func(t *testing.T) {
		resource := pcommon.NewResource()
		resource.Attributes().PutStr("service.name", "test-service")
		got := GetServiceNameFromResource(resource)

		assert.Equal(t, "test-service", got)
	})
	t.Run("Without_Service_Name", func(t *testing.T) {
		resource := pcommon.NewResource()
		got := GetServiceNameFromResource(resource)

		assert.Equal(t, "unknown", got)
	})
}
