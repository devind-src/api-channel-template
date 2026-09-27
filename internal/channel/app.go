package channel

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/devind-src/syncnet-sdk/app"
	"github.com/devind-src/syncnet-sdk/appprocessor"
	"github.com/devind-src/syncnet-sdk/cache"
	"github.com/devind-src/syncnet-sdk/config"
	"github.com/devind-src/syncnet-sdk/database"
	"github.com/devind-src/syncnet-sdk/message"
	"github.com/devind-src/syncnet-sdk/networking"
)

// Version is reported on the command port (VERSION).
const Version = "v0.1.0"

// pendingTTLMinutes is how long an inbound request waits for synccore's
// response before it is forgotten.
const pendingTTLMinutes = 5

// pending is an inbound request waiting for synccore's response.
type pending struct {
	Req    PaymentRequest
	ReqCtx *networking.RequestContext
}

// App implements appprocessor.Processor. As an inbound channel it handles
// ProcessMsgFromRemoteWsServer (HTTP from the channel) and
// ProcessMsgFromSourceNode (the answer from synccore); UnimplementedProcessor
// covers the rest.
type App struct {
	appprocessor.UnimplementedProcessor

	ap      *appprocessor.AppProcessor
	cfg     *Config
	pending *cache.Cache[pending]
	done    chan struct{}
	now     func() time.Time
}

// New is the app.Options.New constructor.
func New(env app.Env) (app.Service, error) {
	cfg, err := LoadConfig(env.SettingsPath)
	if err != nil {
		return nil, err
	}
	return newApp(env.Config, cfg, env.DB), nil
}

func newApp(sdkCfg *config.Config, cfg *Config, svc *database.Service) *App {
	a := &App{
		cfg:     cfg,
		pending: cache.New[pending](pendingTTLMinutes),
		done:    make(chan struct{}),
		now:     time.Now,
	}
	a.pending.StartSweeper(a.done, time.Minute)
	a.ap = appprocessor.New(sdkCfg.AppName, Version, sdkCfg, svc, a)
	return a
}

// Start connects to synccore, starts the HTTP server(s) configured in
// sw_connections and the command port.
func (a *App) Start(ctx context.Context) { a.ap.Start(ctx) }

// Stop disconnects gracefully.
func (a *App) Stop(ctx context.Context) {
	close(a.done)
	a.ap.Stop(ctx)
}

// ProcessMsgFromRemoteWsServer handles an HTTP request from the channel.
func (a *App) ProcessMsgFromRemoteWsServer(ctx context.Context, nodeName, connName string, body []byte, reqCtx *networking.RequestContext) {
	if a.ap.IsTraceOn() {
		a.ap.WriteTrace(ctx, nodeName, "REQ", body, appprocessor.FromRemote, remoteEndpoint(reqCtx))
	}

	if !authenticate(a.cfg, reqCtx, body, a.now()) {
		a.reply(ctx, nodeName, connName, reqCtx, respAuthFailed)
		return
	}

	r, ok := routes[strings.ToLower(reqCtx.Path)]
	if !ok {
		a.reply(ctx, nodeName, connName, reqCtx, respInvalidURL)
		return
	}

	var req PaymentRequest
	if err := json.Unmarshal(body, &req); err != nil {
		a.ap.LoggerNamed(ctx, err.Error(), string(body), nodeName)
		a.reply(ctx, nodeName, connName, reqCtx, respBadRequest)
		return
	}
	msg, err := toInternal(&req, r)
	if err != nil {
		a.ap.LoggerNamed(ctx, err.Error(), string(body), nodeName)
		a.reply(ctx, nodeName, connName, reqCtx, respBadRequest)
		return
	}

	key := correlationKey(msg.TranType, msg.DatetimeTran, msg.MerchantId, msg.TraceNumber)
	if !a.pending.TryAdd(key, pending{Req: req, ReqCtx: reqCtx}) {
		a.reply(ctx, nodeName, connName, reqCtx, respDuplicateTx)
		return
	}

	a.ap.SendToSourceHTTP(ctx, nodeName, connName, reqCtx, msg)
}

// ProcessMsgFromSourceNode handles synccore's response and answers the
// original HTTP request.
func (a *App) ProcessMsgFromSourceNode(ctx context.Context, nodeName, connName string, rsp *message.Response) {
	key := correlationKey(rsp.TranType, rsp.DatetimeTran, rsp.MerchantId, rsp.TraceNumber)
	p, ok := a.pending.GetValue(key)
	if !ok {
		a.ap.LoggerNamed(ctx, "original request not found (expired?): "+key, "", nodeName)
		return
	}
	a.pending.Remove(key)

	a.reply(ctx, nodeName, connName, p.ReqCtx, toRemote(rsp, &p.Req))
}

func (a *App) reply(ctx context.Context, nodeName, connName string, reqCtx *networking.RequestContext, res PaymentResponse) {
	body, err := json.Marshal(res)
	if err != nil {
		a.ap.LoggerNamed(ctx, err.Error(), "", nodeName)
		return
	}
	if a.ap.IsTraceOn() {
		a.ap.WriteTrace(ctx, nodeName, "RSP", body, appprocessor.ToRemote, remoteEndpoint(reqCtx))
	}
	a.ap.ReplyToHttpServer(ctx, nodeName, connName, reqCtx, body, map[string]string{"Content-Type": "application/json"})
}

func remoteEndpoint(reqCtx *networking.RequestContext) string {
	if reqCtx == nil {
		return ""
	}
	return reqCtx.RemoteAddr + reqCtx.Path
}
