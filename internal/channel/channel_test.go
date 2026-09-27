package channel

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/devind-src/syncnet-sdk/config"
	"github.com/devind-src/syncnet-sdk/constants"
	"github.com/devind-src/syncnet-sdk/message"
	"github.com/devind-src/syncnet-sdk/networking"
)

var fixedNow = time.Date(2026, 1, 2, 3, 4, 5, 0, time.Local)

func testConfig() *Config {
	cfg := &Config{}
	cfg.Channel.ClientID = "client"
	cfg.Channel.SecretKey = "secret"
	return cfg
}

// newTestApp builds the App without a database: nothing here calls Start,
// so no sw_app/sw_nodes rows are touched.
func newTestApp(t *testing.T) *App {
	t.Helper()
	sdkCfg := &config.Config{AppName: "API Channel Template Test", LogDir: t.TempDir(), TraceDir: t.TempDir(), MaxRetryLog: 1}
	a := newApp(sdkCfg, testConfig(), nil)
	a.now = func() time.Time { return fixedNow }
	t.Cleanup(func() { close(a.done) })
	return a
}

func samplePayment() PaymentRequest {
	return PaymentRequest{
		MerchantID: "M1", TerminalID: "T1", ProductID: "PLN", CustomerID: "123",
		Amount: "150000", TraceNumber: "000001", DateTime: "0102030405", RefNum: "R1",
	}
}

// signedRequest builds a request context the way a well-behaved channel
// would send it.
func signedRequest(t *testing.T, path string, body []byte, ts time.Time) *networking.RequestContext {
	t.Helper()
	timestamp := ts.Format(timestampLayout)
	h := http.Header{}
	h.Set("X-client-id", "client")
	h.Set("X-timestamp", timestamp)
	h.Set("X-signature", sign("secret", path, "client", timestamp, body))
	return &networking.RequestContext{Path: path, Header: h, RemoteAddr: "10.0.0.1:5555", Method: http.MethodPost}
}

func TestLoadConfigRequiresCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appsettings.json")
	if err := os.WriteFile(path, []byte(`{"Channel": {"ClientId": "c"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(path); err == nil {
		t.Error("expected an error when Channel.SecretKey is missing")
	}
}

func TestAuthenticate(t *testing.T) {
	body := []byte(`{"a":1}`)
	cfg := testConfig()

	if !authenticate(cfg, signedRequest(t, "/api/payment", body, fixedNow), body, fixedNow) {
		t.Error("a correctly signed, fresh request must be accepted")
	}
	if authenticate(cfg, signedRequest(t, "/api/payment", body, fixedNow), []byte(`{"a":2}`), fixedNow) {
		t.Error("a tampered body must be rejected")
	}
	if authenticate(cfg, signedRequest(t, "/api/payment", body, fixedNow.Add(-10*time.Minute)), body, fixedNow) {
		t.Error("a stale timestamp must be rejected")
	}
	if authenticate(cfg, nil, body, fixedNow) {
		t.Error("a nil request context must be rejected")
	}
}

func TestToInternal(t *testing.T) {
	req := samplePayment()
	m, err := toInternal(&req, routes["/api/payment"])
	if err != nil {
		t.Fatal(err)
	}
	if m.TranType != constants.TranTypePayment || m.MsgType != "0200" || m.AmountTran.String() != "150000" ||
		m.ToAccNumber != "123" || m.ReceivingInstId != "PLN" {
		t.Errorf("toInternal = %+v", m)
	}

	req.Amount = "abc"
	if _, err := toInternal(&req, routes["/api/payment"]); err == nil {
		t.Error("expected an error for a non-numeric amount")
	}
}

func TestRequestIsCorrelatedWithResponse(t *testing.T) {
	a := newTestApp(t)
	ctx := context.Background()

	body, _ := json.Marshal(samplePayment())
	a.ProcessMsgFromRemoteWsServer(ctx, "CHANNEL", "CONN", body, signedRequest(t, "/api/payment", body, fixedNow))

	key := correlationKey(constants.TranTypePayment, "0102030405", "M1", "000001")
	if _, ok := a.pending.GetValue(key); !ok {
		t.Fatal("an accepted request must be kept until synccore answers")
	}

	// the same request again is a duplicate while the first is pending
	if a.pending.TryAdd(key, pending{}) {
		t.Error("duplicate request was not detected")
	}

	rsp := message.NewResponse()
	rsp.TranType, rsp.DatetimeTran, rsp.MerchantId, rsp.TraceNumber = constants.TranTypePayment, "0102030405", "M1", "000001"
	a.ProcessMsgFromSourceNode(ctx, "CHANNEL", "CONN", rsp)

	if _, ok := a.pending.GetValue(key); ok {
		t.Error("the pending request must be removed once answered")
	}
}

func TestRejectedRequestsAreNotForwarded(t *testing.T) {
	a := newTestApp(t)
	body, _ := json.Marshal(samplePayment())

	reqCtx := signedRequest(t, "/api/payment", body, fixedNow)
	reqCtx.Header.Set("X-signature", "forged")
	a.ProcessMsgFromRemoteWsServer(context.Background(), "CHANNEL", "CONN", body, reqCtx)

	a.ProcessMsgFromRemoteWsServer(context.Background(), "CHANNEL", "CONN", body, signedRequest(t, "/api/unknown", body, fixedNow))

	key := correlationKey(constants.TranTypePayment, "0102030405", "M1", "000001")
	if _, ok := a.pending.GetValue(key); ok {
		t.Error("an unauthenticated or unroutable request must not be forwarded")
	}
}
