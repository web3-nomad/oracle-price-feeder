package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"price-feeder/oracle/derivative"
	"price-feeder/oracle/provider"

	"github.com/BurntSushi/toml"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-playground/validator/v10"
)

const (
	DenomUSD = "USD"

	defaultListenAddr         = "0.0.0.0:7171"
	defaultSrvWriteTimeout    = 15 * time.Second
	defaultSrvReadTimeout     = 15 * time.Second
	defaultProviderTimeout    = 100 * time.Millisecond
	defaultHeightPollInterval = 1 * time.Second
	defaultHistoryDb          = "prices.db"
	defaultDerivativePeriod   = 30 * time.Minute
)

var (
	validate = validator.New()

	// ErrEmptyConfigPath defines a sentinel error for an empty config path.
	ErrEmptyConfigPath = errors.New("empty configuration file path")

	// SupportedProviders defines a lookup table of all the supported currency API
	// providers.
	SupportedProviders = map[provider.Name]struct{}{
		provider.ProviderBybit:              {},
		provider.ProviderBitfinex:           {},
		provider.ProviderBitforex:           {},
		provider.ProviderBkex:               {},
		provider.ProviderBitmart:            {},
		provider.ProviderBitstamp:           {},
		provider.ProviderCamelotV2:          {},
		provider.ProviderCamelotV3:          {},
		provider.ProviderFin:                {},
		provider.ProviderFinV2:              {},
		provider.ProviderPoloniex:           {},
		provider.ProviderPhemex:             {},
		provider.ProviderLbank:              {},
		provider.ProviderHitBtc:             {},
		provider.ProviderKraken:             {},
		provider.ProviderKucoin:             {},
		provider.ProviderBinance:            {},
		provider.ProviderBinanceUS:          {},
		provider.ProviderOsmosis:            {},
		provider.ProviderOsmosisV2:          {},
		provider.ProviderOkx:                {},
		provider.ProviderHuobi:              {},
		provider.ProviderGate:               {},
		provider.ProviderCoinbase:           {},
		provider.ProviderBitget:             {},
		provider.ProviderMexc:               {},
		provider.ProviderCrypto:             {},
		provider.ProviderCurve:              {},
		provider.ProviderMock:               {},
		provider.ProviderStride:             {},
		provider.ProviderXt:                 {},
		provider.ProviderIdxOsmosis:         {},
		provider.ProviderPyth:               {},
		provider.ProviderZero:               {},
		provider.ProviderUniswapV3:          {},
		provider.ProviderAstroportNeutron:   {},
		provider.ProviderAstroportTerra2:    {},
		provider.ProviderAstroportInjective: {},
	}

	SupportedDerivatives = map[string]struct{}{
		derivative.DerivativeTwap: {},
	}

	// maxDeviationThreshold is the maxmimum allowed amount of standard
	// deviations which validators are able to set for a given asset.
	maxDeviationThreshold = sdk.MustNewDecFromStr("3.0")
)

type (
	// Config defines all necessary price-feeder configuration parameters.
	Config struct {
		Server               Server                       `toml:"server"`
		CurrencyPairs        []CurrencyPair               `toml:"currency_pairs" validate:"required,gt=0,dive,required"`
		Deviations           []Deviation                  `toml:"deviation_thresholds"`
		ProviderMinOverrides []ProviderMinOverrides       `toml:"provider_min_overrides"`
		Account              Account                      `toml:"account" validate:"required,gt=0,dive,required"`
		Telemetry            Telemetry                    `toml:"telemetry"`
		VotePeriod           string                       `toml:"vote_period" validate:"required"`
		ProviderTimeout      string                       `toml:"provider_timeout"`
		ProviderEndpoints    []ProviderEndpoints          `toml:"provider_endpoints" validate:"dive"`
		EnableServer         bool                         `toml:"enable_server"`
		EnableVoter          bool                         `toml:"enable_voter"`
		Healthchecks         []Healthchecks               `toml:"healthchecks" validate:"dive"`
		HeightPollInterval   string                       `toml:"height_poll_interval"`
		HistoryDb            string                       `toml:"history_db"`
		ContractAdresses     map[string]map[string]string `toml:"contract_addresses"`
	}

	// Server defines the API server configuration.
	Server struct {
		ListenAddr     string   `toml:"listen_addr"`
		WriteTimeout   string   `toml:"write_timeout"`
		ReadTimeout    string   `toml:"read_timeout"`
		VerboseCORS    bool     `toml:"verbose_cors"`
		AllowedOrigins []string `toml:"allowed_origins"`
	}

	// CurrencyPair defines a price quote of the exchange rate for two different
	// currencies and the supported providers for getting the exchange rate.
	CurrencyPair struct {
		Base             string          `toml:"base" validate:"required"`
		Quote            string          `toml:"quote" validate:"required"`
		Providers        []provider.Name `toml:"providers" validate:"required,gt=0,dive,required"`
		Derivative       string          `toml:"derivative"`
		DerivativePeriod string          `toml:"derivative_period"`
	}

	// Deviation defines a maximum amount of standard deviations that a given asset can
	// be from the median without being filtered out before voting.
	Deviation struct {
		Base      string `toml:"base" validate:"required"`
		Threshold string `toml:"threshold" validate:"required"`
	}

	// ProviderMinOverrides defines the minimum amount of sources that need
	// to *sucessfully* provide price data for a certain asset
	ProviderMinOverrides struct {
		Denoms    []string `toml:"denoms" validate:"required"`
		Providers uint     `toml:"providers" validate:"required"`
	}

	// Account defines account related configuration that is related to the
	// network and transaction signing functionality.
	Account struct {
		NetworkName  string `toml:"network_name" validate:"required"`
		OperatorID   string `toml:"operator_id" validate:"required"`
		OperatorSeed string `toml:"operator_seed" validate:"required"`
		TopicID      string `toml:"topic_id" validate:"required"`
	}

	// Telemetry defines the configuration options for application telemetry.
	Telemetry struct {
		// Prefixed with keys to separate services
		ServiceName string `toml:"service_name" mapstructure:"service-name"`

		// Enabled enables the application telemetry functionality. When enabled,
		// an in-memory sink is also enabled by default. Operators may also enabled
		// other sinks such as Prometheus.
		Enabled bool `toml:"enabled" mapstructure:"enabled"`

		// Enable prefixing gauge values with hostname
		EnableHostname bool `toml:"enable_hostname" mapstructure:"enable-hostname"`

		// Enable adding hostname to labels
		EnableHostnameLabel bool `toml:"enable_hostname_label" mapstructure:"enable-hostname-label"`

		// Enable adding service to labels
		EnableServiceLabel bool `toml:"enable_service_label" mapstructure:"enable-service-label"`

		// GlobalLabels defines a global set of name/value label tuples applied to all
		// metrics emitted using the wrapper functions defined in telemetry package.
		//
		// Example:
		// [["network_name", "testnet"]]
		GlobalLabels [][]string `toml:"global_labels" mapstructure:"global-labels"`

		// PrometheusRetentionTime, when positive, enables a Prometheus metrics sink.
		// It defines the retention duration in seconds.
		PrometheusRetentionTime int64 `toml:"prometheus_retention" mapstructure:"prometheus-retention-time"`
	}

	Healthchecks struct {
		URL     string `toml:"url" validate:"required"`
		Timeout string `toml:"timeout" validate:"required"`
	}

	ProviderEndpoints struct {
		Name          provider.Name `toml:"name" validate:"required"`
		Urls          []string      `toml:"urls"`
		Websocket     string        `toml:"websocket"`
		WebsocketPath string        `toml:"websocket_path"`
		PollInterval  string        `toml:"poll_interval"`
		Contracts     []string      `toml:"contracts"`
	}
)

// telemetryValidation is custom validation for the Telemetry struct.
func telemetryValidation(sl validator.StructLevel) {
	tel := sl.Current().Interface().(Telemetry)

	if tel.Enabled && (len(tel.GlobalLabels) == 0 || len(tel.ServiceName) == 0) {
		sl.ReportError(tel.Enabled, "enabled", "Enabled", "enabledNoOptions", "")
	}
}

// endpointValidation is custom validation for the ProviderEndpoint struct.
func endpointValidation(sl validator.StructLevel) {
	endpoint := sl.Current().Interface().(ProviderEndpoints)

	if len(endpoint.Name) < 1 || (len(endpoint.Urls) < 1 && len(endpoint.Websocket) < 1) {
		sl.ReportError(endpoint, "endpoint", "Endpoint", "unsupportedEndpointType", "")
	}
	if _, ok := SupportedProviders[endpoint.Name]; !ok {
		sl.ReportError(endpoint.Name, "name", "Name", "unsupportedEndpointProvider", "")
	}
}

// Validate returns an error if the Config object is invalid.
func (c Config) Validate() error {
	validate.RegisterStructValidation(telemetryValidation, Telemetry{})
	validate.RegisterStructValidation(endpointValidation, ProviderEndpoints{})
	return validate.Struct(c)
}

func (p ProviderEndpoints) ToEndpoint() (provider.Endpoint, error) {
	var pollInterval time.Duration
	if p.PollInterval != "" {
		interval, err := time.ParseDuration(p.PollInterval)
		if err != nil {
			return provider.Endpoint{}, fmt.Errorf("failed to parse poll interval: %v", err)
		}
		pollInterval = interval
	}

	e := provider.Endpoint{
		Name:          p.Name,
		Urls:          p.Urls,
		Websocket:     p.Websocket,
		WebsocketPath: p.WebsocketPath,
		PollInterval:  pollInterval,
	}
	return e, nil
}

// ParseConfig attempts to read and parse configuration from the given file path.
// An error is returned if reading or parsing the config fails.
func ParseConfig(configPath string) (Config, error) {
	var cfg Config

	if configPath == "" {
		return cfg, ErrEmptyConfigPath
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config: %w", err)
	}

	if _, err := toml.Decode(string(configData), &cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode config: %w", err)
	}

	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = defaultListenAddr
	}
	if len(cfg.Server.WriteTimeout) == 0 {
		cfg.Server.WriteTimeout = defaultSrvWriteTimeout.String()
	}
	if len(cfg.Server.ReadTimeout) == 0 {
		cfg.Server.ReadTimeout = defaultSrvReadTimeout.String()
	}
	if len(cfg.ProviderTimeout) == 0 {
		cfg.ProviderTimeout = defaultProviderTimeout.String()
	}
	if cfg.HeightPollInterval == "" {
		cfg.HeightPollInterval = defaultHeightPollInterval.String()
	}
	if cfg.HistoryDb == "" {
		cfg.HistoryDb = defaultHistoryDb
	}

	derivativeDenoms := map[string]struct{}{}
	derivativeBases := map[string]struct{}{}
	pairs := make(map[string]map[provider.Name]struct{})
	coinQuotes := make(map[string]struct{})
	for i, cp := range cfg.CurrencyPairs {
		if _, ok := pairs[cp.Base]; !ok {
			pairs[cp.Base] = make(map[provider.Name]struct{})
		}
		if strings.ToUpper(cp.Quote) != DenomUSD {
			coinQuotes[cp.Quote] = struct{}{}
		}
		if cp.Derivative != "" {
			derivativeBases[cp.Base] = struct{}{}
			derivativeDenoms[cp.Base+cp.Quote] = struct{}{}
			_, ok := SupportedDerivatives[cp.Derivative]
			if !ok {
				return cfg, fmt.Errorf("unsupported derivative: %s", cp.Derivative)
			}
			if cp.DerivativePeriod != "" {
				_, err := time.ParseDuration(cp.DerivativePeriod)
				if err != nil {
					return cfg, err
				}
			} else {
				cfg.CurrencyPairs[i].DerivativePeriod = defaultDerivativePeriod.String()
			}
		} else {
			_, ok := derivativeDenoms[cp.Base]
			if ok {
				return cfg, fmt.Errorf("cannot combine derivative and nonderivative pairs for %s", cp.Base)
			}
		}
		for _, provider := range cp.Providers {
			if _, ok := SupportedProviders[provider]; !ok {
				return cfg, fmt.Errorf("unsupported provider: %s", provider)
			}
			pairs[cp.Base][provider] = struct{}{}
		}
	}

	for _, deviation := range cfg.Deviations {
		threshold, err := sdk.NewDecFromStr(deviation.Threshold)
		if err != nil {
			return cfg, fmt.Errorf("deviation thresholds must be numeric: %w", err)
		}

		if threshold.GT(maxDeviationThreshold) {
			return cfg, fmt.Errorf("deviation thresholds must not exceed 3.0")
		}
	}

	for _, override := range cfg.ProviderMinOverrides {
		if override.Providers < 1 {
			return cfg, fmt.Errorf("minimum providers must be greater than 0")
		}
	}

	return cfg, cfg.Validate()
}
