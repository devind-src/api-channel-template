package channel

// Responses this app returns on its own, without going to synccore.
// TODO: align codes/messages with the API agreed with the channel.
var (
	respAuthFailed  = PaymentResponse{ResponseCode: "X8", ResponseMessage: "OTENTIKASI TIDAK VALID"}
	respInvalidURL  = PaymentResponse{ResponseCode: "X6", ResponseMessage: "URL TIDAK VALID"}
	respBadRequest  = PaymentResponse{ResponseCode: "30", ResponseMessage: "FORMAT REQUEST TIDAK VALID"}
	respDuplicateTx = PaymentResponse{ResponseCode: "X2", ResponseMessage: "TRANSAKSI DUPLIKAT"}
)
