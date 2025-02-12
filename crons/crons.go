package crons

import (
	"github.com/robfig/cron"
	"github.com/tomochain/tomox-sdk/engine"
	"github.com/tomochain/tomox-sdk/services"
)

// CronService contains the services required to initialize crons
type CronService struct {
	OHLCVService      *services.OHLCVService
	PriceBoardService *services.PriceBoardService
	PairService       *services.PairService
	FiatPriceService  *services.FiatPriceService
	RelayService      *services.RelayerService
	Engine            *engine.Engine
}

// NewCronService returns a new instance of CronService
func NewCronService(
	ohlcvService *services.OHLCVService,
	priceBoardService *services.PriceBoardService,
	pairService *services.PairService,
	fiatPriceService *services.FiatPriceService,
	relayService *services.RelayerService,
	engine *engine.Engine,
) *CronService {
	return &CronService{
		OHLCVService:      ohlcvService,
		PriceBoardService: priceBoardService,
		PairService:       pairService,
		FiatPriceService:  fiatPriceService,
		RelayService:      relayService,
		Engine:            engine,
	}
}

// InitCrons is responsible for initializing all the crons in the system
func (s *CronService) InitCrons() {

	s.RelayService.UpdateRelayer()
	s.FiatPriceService.SyncFiatPrice()
	s.FiatPriceService.UpdateFiatPrice()

	c := cron.New()
	s.startRelayerUpdate(c)
	s.tickStreamingCron(c)   // Cron to fetch OHLCV data
	s.getFiatPriceCron(c)    // Cron to query USD price from coinmarketcap.com and update "tokens" collection
	s.startPriceBoardCron(c) // Cron to fetch data for top price board
	s.startMarketsCron(c)    // Cron to fetch markets data
	c.Start()
}
