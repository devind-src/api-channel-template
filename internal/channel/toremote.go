package channel

import "github.com/devind-src/syncnet-sdk/message"

// toRemote maps synccore's response for original back to the channel's
// response shape.
func toRemote(rsp *message.Response, original *PaymentRequest) PaymentResponse {
	return PaymentResponse{
		ResponseCode:    rsp.RespCode,
		ResponseMessage: rsp.RespMessage,
		MerchantID:      original.MerchantID,
		TraceNumber:     original.TraceNumber,
		RefNum:          original.RefNum,
		CustomerID:      original.CustomerID,
		Amount:          original.Amount,
	}
}
