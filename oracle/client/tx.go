package client

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
// It will return an error upon failure.
//
// Note, BroadcastTx is copied from the SDK except it removes a few unnecessary
// things like prompting for confirmation and printing the response. Instead,
// we return the TxResponse.
