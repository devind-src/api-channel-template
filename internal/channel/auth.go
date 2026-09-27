package channel

import (
	"crypto/hmac"
	"encoding/base64"
	"time"

	"github.com/devind-src/syncnet-sdk/crypto"
	"github.com/devind-src/syncnet-sdk/networking"
)

// The example scheme: the channel signs "path:clientId:timestamp:body" with
// HMAC-SHA512 and sends X-client-id, X-timestamp and X-signature headers;
// timestamps outside timestampWindow are rejected (replay protection).
// TODO: replace with the scheme agreed with the channel.
const (
	timestampLayout = "20060102150405"
	timestampWindow = 5 * time.Minute
)

func authenticate(cfg *Config, reqCtx *networking.RequestContext, body []byte, now time.Time) bool {
	if reqCtx == nil {
		return false
	}

	clientID := reqCtx.Header.Get("X-client-id")
	timestamp := reqCtx.Header.Get("X-timestamp")
	signature := reqCtx.Header.Get("X-signature")

	if clientID != cfg.Channel.ClientID || !timestampFresh(timestamp, now) {
		return false
	}

	expected := sign(cfg.Channel.SecretKey, reqCtx.Path, clientID, timestamp, body)
	return hmac.Equal([]byte(signature), []byte(expected))
}

func sign(secretKey, path, clientID, timestamp string, body []byte) string {
	toSign := path + ":" + clientID + ":" + timestamp + ":" + string(body)
	return base64.StdEncoding.EncodeToString(crypto.ComputeHMACSHA512Hash(toSign, secretKey))
}

func timestampFresh(timestamp string, now time.Time) bool {
	t, err := time.ParseInLocation(timestampLayout, timestamp, now.Location())
	if err != nil {
		return false
	}
	return now.Sub(t).Abs() <= timestampWindow
}
