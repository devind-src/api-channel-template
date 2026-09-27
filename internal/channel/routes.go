package channel

import "github.com/devind-src/syncnet-sdk/constants"

// route is how an inbound URL path maps onto a switch transaction.
type route struct {
	MsgType  string
	TranType string
}

// routes lists the URL paths this channel accepts (matched lower-case).
// TODO: adjust to the API agreed with the channel.
var routes = map[string]route{
	"/api/inquiry":  {MsgType: "0100", TranType: constants.TranTypeInquiry},
	"/api/payment":  {MsgType: "0200", TranType: constants.TranTypePayment},
	"/api/advice":   {MsgType: "0220", TranType: constants.TranTypeAdvice},
	"/api/reversal": {MsgType: "0400", TranType: constants.TranTypeReversal},
}
