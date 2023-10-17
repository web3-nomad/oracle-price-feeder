package provider

import (
	"context"
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"

	"price-feeder/oracle/types"

	"github.com/rs/zerolog"
)

var (
	_                     Provider = (*FinV2Provider)(nil)
	finV2DefaultEndpoints          = Endpoint{
		Name: ProviderFinV2,
		Urls: []string{
			"https://cosmos.directory/kujira",
			"https://lcd.kaiyo.kujira.setten.io",
			"https://lcd-kujira.mintthemoon.xyz",
		},
		PollInterval: 3 * time.Second,
		ContractAddresses: map[string]string{
			"USDCUSK": "kujira1rwx6w02alc4kaz7xpyg3rlxpjl4g63x5jq292mkxgg65zqpn5llq202vh5",
		},
	}
)

type (
	// FinV2 defines an oracle provider that uses the API of an Kujira node
	// to directly retrieve the price from the fin contract
	FinV2Provider struct {
		provider
		contracts map[string]string
	}

	FinV2BookResponse struct {
		Data FinV2BookData `json:"data"`
	}

	FinV2BookData struct {
		Base  []FinV2Order `json:"base"`
		Quote []FinV2Order `json:"quote"`
	}

	FinV2Order struct {
		Price string `json:"quote_price"`
	}
)

func NewFinV2Provider(
	ctx context.Context,
	logger zerolog.Logger,
	endpoints Endpoint,
	pairs ...types.CurrencyPair,
) (*FinV2Provider, error) {
	provider := &FinV2Provider{}
	provider.Init(
		ctx,
		endpoints,
		logger,
		pairs,
		nil,
		nil,
	)

	provider.contracts = provider.endpoints.ContractAddresses

	availablePairs, _ := provider.GetAvailablePairs()
	provider.setPairs(pairs, availablePairs, nil)

	go startPolling(provider, provider.endpoints.PollInterval, logger)
	return provider, nil
}

func (p *FinV2Provider) Poll() error {
	timestamp := time.Now()

	p.mtx.Lock()
	defer p.mtx.Unlock()

	for symbol, pair := range p.getAllPairs() {

		contract, err := p.getContractAddress(pair)
		if err != nil {
			p.logger.Warn().
				Str("symbol", symbol).
				Msg("no contract address found")
			continue
		}

		path := fmt.Sprintf(
			"/cosmwasm/wasm/v1/contract/%s/smart/eyJib29rIjp7ImxpbWl0IjoxfX0K",
			contract,
		)

		content, err := p.httpGet(path)
		if err != nil {
			return err
		}

		var bookResponse FinV2BookResponse
		err = json.Unmarshal(content, &bookResponse)
		if err != nil {
			return err
		}

		if len(bookResponse.Data.Base) < 1 || len(bookResponse.Data.Quote) < 1 {
			return fmt.Errorf("no order found")
		}

		base := strToDec(bookResponse.Data.Base[0].Price)
		quote := strToDec(bookResponse.Data.Quote[0].Price)

		price := base.Add(quote).QuoInt64(2)

		p.setTickerPrice(
			symbol,
			price,
			sdk.ZeroDec(),
			timestamp,
		)
	}

	return nil
}

func (p *FinV2Provider) GetAvailablePairs() (map[string]struct{}, error) {
	return p.getAvailablePairsFromContracts()
}
