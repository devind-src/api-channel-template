// Package channel is an inbound channel interface (api-channel-xxx): an
// external channel (merchant, aggregator, mobile backend, ...) calls this
// app over HTTP, the app authenticates and maps the request to a
// message.Request and sends it into synccore; when synccore answers, the
// message.Response is mapped back and the still-open HTTP request is
// answered.
//
// Files, and what to change for a new channel:
//
//	config.go     the channel's section of appsettings.json
//	models.go     the channel's request/response shapes (our API contract to them)
//	auth.go       how the channel authenticates its requests
//	routes.go     URL path -> transaction type
//	tointernal.go channel request  -> message.Request
//	toremote.go   message.Response -> channel response
//	respcode.go   error responses returned directly by this app
//	app.go        wiring: authenticate, route, correlate, reply (rarely changes)
//
// Everything else (HTTP server, connection to synccore, command port,
// logging, tracing, lifecycle) is provided by syncnet-sdk.
package channel
