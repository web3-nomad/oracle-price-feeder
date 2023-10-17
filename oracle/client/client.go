package client

import (
	"context"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// OracleClient defines a structure that interfaces with the Umee node.
type (
	OracleClient struct {
		Logger          zerolog.Logger
		NetworkName     string
		HederaClient    *hedera.Client
		OperatorAccount hedera.AccountID
		VotePeriod      time.Duration
		//ChainHeight     *ChainHeight
		topicID hedera.TopicID
	}
)

func NewOracleClient(
	ctx context.Context,
	logger zerolog.Logger,
	networkName string,
	operatorID string,
	operatorSeed string,
	topicID string,
	votePeriod time.Duration,
	heightPollInterval time.Duration,
) (OracleClient, error) {
	hederaClient, err := hedera.ClientForName(networkName)
	if err != nil {
		return OracleClient{}, err
	}
	operatorAccountID, err := hedera.AccountIDFromString(operatorID)
	if err != nil {
		return OracleClient{}, err
	}
	mnemonic, err := hedera.NewMnemonic(strings.Split(operatorSeed, " "))
	if err != nil {
		return OracleClient{}, err
	}
	operatorKey, err := mnemonic.ToStandardEd25519PrivateKey("", 0)
	if err != nil {
		return OracleClient{}, err
	}
	hederaClient.SetOperator(operatorAccountID, operatorKey)
	topicIDParsed, err := hedera.TopicIDFromString(topicID)
	if err != nil {
		return OracleClient{}, err
	}

	oracleClient := OracleClient{
		Logger:          logger.With().Str("module", "oracle_client").Logger(),
		NetworkName:     networkName,
		HederaClient:    hederaClient,
		OperatorAccount: operatorAccountID,
		VotePeriod:      votePeriod,
		topicID:         topicIDParsed,
	}

	//clientCtx, err := oracleClient.CreateClientContext()
	if err != nil {
		return OracleClient{}, err
	}

	//	oracleClient.ChainHeight = chainHeight

	return oracleClient, nil
}
func (oc *OracleClient) PutTx(content []byte) error {
	submitTxn, err := hedera.NewTopicMessageSubmitTransaction().
		// The message we are submitting
		SetMessage(content).
		// To which topic ID
		SetTopicID(oc.topicID).
		Execute(oc.HederaClient)

	if err != nil {
		panic(fmt.Sprintf("%v : error submitting topic", err))
		return err

	}
	oc.Logger.Debug().Bytes("hash", submitTxn.Hash).
		Str("TopicID", oc.topicID.String()).
		Msg("Submitted TXN to topic")

	return nil
}
