package channel

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/devind-src/syncnet-sdk/message"
)

// toInternal maps the channel's request to a switch request.
func toInternal(req *PaymentRequest, r route) (*message.Request, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, fmt.Errorf("channel: invalid amount %q: %w", req.Amount, err)
	}

	m := message.NewRequest()
	m.MsgType = r.MsgType
	m.TranType = r.TranType
	m.FromAccType = "00"
	m.ToAccType = "00"
	m.AmountTran = amount
	m.TraceNumber = req.TraceNumber
	m.DatetimeTran = req.DateTime
	m.MerchantId = req.MerchantID
	m.TerminalId = req.TerminalID
	m.Refnum = req.RefNum
	m.ReceivingInstId = req.ProductID
	m.ToAccNumber = req.CustomerID
	return m, nil
}

// correlationKey identifies a request and its asynchronous response from
// synccore; both sides carry these fields unchanged.
func correlationKey(tranType, datetime, merchantID, traceNumber string) string {
	return tranType + datetime + merchantID + traceNumber
}
