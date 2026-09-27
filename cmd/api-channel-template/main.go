// Command api-channel-template is the starting point for a new inbound
// channel interface (api-channel-xxx). Rename it together with the module
// path — see README.md.
package main

import (
	"github.com/devind-src/api-channel-template/internal/channel"
	"github.com/devind-src/syncnet-sdk/app"
)

func main() {
	app.Run(app.Options{Version: channel.Version, New: channel.New})
}
