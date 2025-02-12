package services

import (
	"math/big"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/tomochain/tomox-sdk/interfaces"
	"github.com/tomochain/tomox-sdk/types"
	"github.com/tomochain/tomox-sdk/utils"
	"github.com/tomochain/tomox-sdk/utils/math"
	"github.com/tomochain/tomox-sdk/ws"
)

// MarketsService struct with daos required, responsible for communicating with daos.
// MarketsService functions are responsible for interacting with daos and implements business logics.
type MarketsService struct {
	PairDao          interfaces.PairDao
	OrderDao         interfaces.OrderDao
	TradeDao         interfaces.TradeDao
	FiatPriceDao     interfaces.FiatPriceDao
	OHLCVService     interfaces.OHLCVService
	FiatPriceService interfaces.FiatPriceService
	PairService      interfaces.PairService
}

// NewMarketsService returns a new instance of TradeService
func NewMarketsService(
	pairDao interfaces.PairDao,
	orderdao interfaces.OrderDao,
	tradeDao interfaces.TradeDao,
	ohlcvService interfaces.OHLCVService,
	fiatPriceDao interfaces.FiatPriceDao,
	fiatPriceService interfaces.FiatPriceService,
	pairService interfaces.PairService,
) *MarketsService {
	return &MarketsService{
		PairDao:          pairDao,
		OrderDao:         orderdao,
		TradeDao:         tradeDao,
		FiatPriceDao:     fiatPriceDao,
		OHLCVService:     ohlcvService,
		FiatPriceService: fiatPriceService,
		PairService:      pairService,
	}
}

// Subscribe market
func (s *MarketsService) Subscribe(c *ws.Client) {
	socket := ws.GetMarketSocket()

	pairData, err := s.PairService.GetAllTokenPairData()

	if err != nil {
		logger.Error(err)
		socket.SendErrorMessage(c, err.Error())
		return
	}

	id := utils.GetMarketsChannelID(ws.MarketsChannel)
	err = socket.Subscribe(id, c)
	if err != nil {
		logger.Error(err)
		socket.SendErrorMessage(c, err.Error())
		return
	}

	smallChartsDataResult, err := s.FiatPriceService.GetFiatPriceChart()

	data := &types.MarketData{
		PairData:        pairData,
		SmallChartsData: smallChartsDataResult,
	}

	ws.RegisterConnectionUnsubscribeHandler(c, socket.UnsubscribeChannelHandler(id))
	socket.SendInitMessage(c, data)
}

// Unsubscribe
func (s *MarketsService) UnsubscribeChannel(c *ws.Client) {
	socket := ws.GetMarketSocket()

	id := utils.GetMarketsChannelID(ws.MarketsChannel)
	socket.UnsubscribeChannel(id, c)
}

// Unsubscribe
func (s *MarketsService) Unsubscribe(c *ws.Client) {
	socket := ws.GetMarketSocket()
	socket.Unsubscribe(c)
}

func (s *MarketsService) GetPairData() ([]*types.PairData, error) {
	now := time.Now()
	end := time.Unix(now.Unix(), 0)
	start := time.Unix(now.AddDate(0, 0, -1).Unix(), 0)
	one, _ := bson.ParseDecimal128("1")

	pairs, err := s.PairDao.GetActivePairs()
	if err != nil {
		return nil, err
	}

	tradeDataQuery := []bson.M{
		{
			"$match": bson.M{
				"createdAt": bson.M{
					"$gte": start,
					"$lt":  end,
				},
				"status": bson.M{"$in": []string{types.SUCCESS}},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"pairName":   "$pairName",
					"baseToken":  "$baseToken",
					"quoteToken": "$quoteToken",
				},
				"count":     bson.M{"$sum": one},
				"open":      bson.M{"$first": "$pricepoint"},
				"openTime":  bson.M{"$first": "$createdAt"},
				"high":      bson.M{"$max": "$pricepoint"},
				"low":       bson.M{"$min": "$pricepoint"},
				"close":     bson.M{"$last": "$pricepoint"},
				"closeTime": bson.M{"$last": "$createdAt"},
				"volume":    bson.M{"$sum": bson.M{"$toDecimal": "$amount"}},
			},
		},
	}

	bidsQuery := []bson.M{
		{
			"$match": bson.M{
				"status": bson.M{"$in": []string{types.OrderStatusNew, types.OrderStatusOpen, types.OrderStatusPartialFilled}},
				"side":   "BUY",
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"pairName":   "$pairName",
					"baseToken":  "$baseToken",
					"quoteToken": "$quoteToken",
				},
				"orderCount": bson.M{"$sum": one},
				"orderVolume": bson.M{
					"$sum": bson.M{
						"$subtract": []bson.M{{"$toDecimal": "$amount"}, {"$toDecimal": "$filledAmount"}},
					},
				},
				"bestPrice": bson.M{"$max": "$pricepoint"},
			},
		},
	}

	asksQuery := []bson.M{
		{
			"$match": bson.M{
				"status": bson.M{"$in": []string{types.OrderStatusNew, types.OrderStatusOpen, types.OrderStatusPartialFilled}},
				"side":   "SELL",
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"pairName":   "$pairName",
					"baseToken":  "$baseToken",
					"quoteToken": "$quoteToken",
				},
				"orderCount": bson.M{"$sum": one},
				"orderVolume": bson.M{
					"$sum": bson.M{
						"$subtract": []bson.M{{"$toDecimal": "$amount"}, {"$toDecimal": "$filledAmount"}},
					},
				},
				"bestPrice": bson.M{"$min": "$pricepoint"},
			},
		},
	}

	tradeData, err := s.TradeDao.Aggregate(tradeDataQuery)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	bidsData, err := s.OrderDao.Aggregate(bidsQuery)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	asksData, err := s.OrderDao.Aggregate(asksQuery)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	pairsData := make([]*types.PairData, 0)
	for _, p := range pairs {
		pairData := &types.PairData{
			Pair:         types.PairID{PairName: p.Name(), BaseToken: p.BaseTokenAddress, QuoteToken: p.QuoteTokenAddress},
			Open:         big.NewInt(0),
			High:         big.NewInt(0),
			Low:          big.NewInt(0),
			Volume:       big.NewInt(0),
			Close:        big.NewInt(0),
			CloseBaseUsd: big.NewFloat(0),
			Count:        big.NewInt(0),
			OrderVolume:  big.NewInt(0),
			OrderCount:   big.NewInt(0),
			BidPrice:     big.NewInt(0),
			AskPrice:     big.NewInt(0),
			Price:        big.NewInt(0),
		}

		for _, t := range tradeData {
			if t.AddressCode() == p.AddressCode() {
				pairData.Open = t.Open
				pairData.High = t.High
				pairData.Low = t.Low
				pairData.Volume = t.Volume
				pairData.Close = t.Close
				pairData.Count = t.Count
				fiatItem, err := s.FiatPriceDao.GetLastPriceCurrentByTime(p.BaseTokenSymbol, t.CloseTime)
				if err == nil {
					pairData.CloseBaseUsd, _ = pairData.CloseBaseUsd.SetString(fiatItem.Price)
				}

			}
		}

		for _, o := range bidsData {
			if o.AddressCode() == p.AddressCode() {
				pairData.OrderVolume = o.OrderVolume
				pairData.OrderCount = o.OrderCount
				pairData.BidPrice = o.BestPrice
			}
		}

		for _, o := range asksData {
			if o.AddressCode() == p.AddressCode() {
				pairData.OrderVolume = math.Add(pairData.OrderVolume, o.OrderVolume)
				pairData.OrderCount = math.Add(pairData.OrderCount, o.OrderCount)
				pairData.AskPrice = o.BestPrice

				if math.IsNotEqual(pairData.BidPrice, big.NewInt(0)) && math.IsNotEqual(pairData.AskPrice, big.NewInt(0)) {
					pairData.Price = math.Avg(pairData.BidPrice, pairData.AskPrice)
				} else {
					pairData.Price = big.NewInt(0)
				}
			}
		}

		pairsData = append(pairsData, pairData)
	}

	return pairsData, nil
}
