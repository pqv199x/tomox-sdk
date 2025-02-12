package services

import (
	"math/big"
	"testing"

	"github.com/tomochain/tomox-sdk/rabbitmq"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tomochain/tomox-sdk/types"
	"github.com/tomochain/tomox-sdk/utils/testutils"
	"github.com/tomochain/tomox-sdk/utils/testutils/mocks"
)

func TestCancelTrades(t *testing.T) {
	orderDao := new(mocks.OrderDao)
	pairDao := new(mocks.PairDao)
	accountDao := new(mocks.AccountDao)
	tradeDao := new(mocks.TradeDao)
	engine := new(mocks.Engine)
	ethereum := new(mocks.EthereumProvider)

	amqp := rabbitmq.InitConnection("amqp://guest:guest@localhost:5672/")
	orderService := NewOrderService(
		orderDao,
		pairDao,
		accountDao,
		tradeDao,
		engine,
		ethereum,
		amqp,
	)

	t1 := testutils.GetTestTrade1()
	t2 := testutils.GetTestTrade2()
	o1 := testutils.GetTestOrder1()
	o2 := testutils.GetTestOrder2()

	trades := []*types.Trade{&t1, &t2}
	hashes := []common.Hash{t1.OrderHash, t2.OrderHash}
	amounts := []*big.Int{t1.Amount, t2.Amount}
	orders := []*types.Order{&o1, &o2}

	orderDao.On("GetByHashes", hashes).Return(orders, nil)
	engine.On("CancelTrades", orders, amounts).Return(nil)

	err := orderService.CancelTrades(trades)
	if err != nil {
		t.Error("Could not cancel trades", err)
	}

	orderDao.AssertCalled(t, "GetByHashes", hashes)
	engine.AssertCalled(t, "CancelTrades", orders, amounts)
}
