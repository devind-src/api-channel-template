package channel

import (
	"testing"

	"github.com/devind-src/syncnet-sdk/depcheck"
)

// api-channel-xxx apps may depend on syncnet-sdk only.
func TestDependsOnlyOnSDK(t *testing.T) {
	depcheck.OnlySDK(t, "github.com/devind-src/api-channel-template")
}
