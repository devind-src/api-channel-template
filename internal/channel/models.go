package channel

// PaymentRequest is what the channel sends us.
// TODO: replace with the API contract agreed with the channel.
type PaymentRequest struct {
	MerchantID  string `json:"merchantId"`
	TerminalID  string `json:"terminalId"`
	ProductID   string `json:"productId"`
	CustomerID  string `json:"customerId"`
	Amount      string `json:"amount"`
	TraceNumber string `json:"traceNumber"`
	DateTime    string `json:"dateTime"` // MMddHHmmss
	RefNum      string `json:"referenceNumber"`
}

// PaymentResponse is what we answer.
// TODO: replace with the API contract agreed with the channel.
type PaymentResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	MerchantID      string `json:"merchantId,omitempty"`
	TraceNumber     string `json:"traceNumber,omitempty"`
	RefNum          string `json:"referenceNumber,omitempty"`
	CustomerID      string `json:"customerId,omitempty"`
	Amount          string `json:"amount,omitempty"`
}
