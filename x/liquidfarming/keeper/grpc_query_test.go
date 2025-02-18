package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v3/types"
	farmtypes "github.com/crescent-network/crescent/v3/x/farm/types"
	"github.com/crescent-network/crescent/v3/x/liquidfarming/types"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestGRPCParams() {
	resp, err := s.querier.Params(sdk.WrapSDKContext(s.ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.keeper.GetParams(s.ctx), resp.Params)
}

func (s *KeeperTestSuite) TestGRPCLiquidFarms() {
	pair1 := s.createPair(s.addr(0), "denom1", "denom2")
	pool1 := s.createPool(s.addr(0), pair1.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	minFarmAmt1, minBidAmt1 := sdk.NewInt(10_000_000), sdk.NewInt(10_000_000)
	s.createLiquidFarm(pool1.Id, minFarmAmt1, minBidAmt1, sdk.ZeroDec())

	pair2 := s.createPair(s.addr(1), "denom3", "denom4")
	pool2 := s.createPool(s.addr(1), pair2.Id, utils.ParseCoins("100_000_000denom3, 100_000_000denom4"))
	minFarmAmt2, minBidAmt2 := sdk.NewInt(30_000_000), sdk.NewInt(30_000_000)
	s.createLiquidFarm(pool2.Id, minFarmAmt2, minBidAmt2, sdk.ZeroDec())

	for _, tc := range []struct {
		name      string
		req       *types.QueryLiquidFarmsRequest
		expectErr bool
		postRun   func(*types.QueryLiquidFarmsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"happy case",
			&types.QueryLiquidFarmsRequest{},
			false,
			func(resp *types.QueryLiquidFarmsResponse) {
				s.Require().Len(resp.LiquidFarms, 2)

				for _, liquidFarm := range resp.LiquidFarms {
					switch liquidFarm.PoolId {
					case 1:
						s.Require().Equal(minFarmAmt1, liquidFarm.MinFarmAmount)
						s.Require().Equal(minBidAmt1, liquidFarm.MinBidAmount)
					case 2:
						s.Require().Equal(minFarmAmt2, liquidFarm.MinFarmAmount)
						s.Require().Equal(minBidAmt2, liquidFarm.MinBidAmount)
					}
					poolCoinDenom := liquiditytypes.PoolCoinDenom(liquidFarm.PoolId)
					farm, found := s.app.FarmKeeper.GetFarm(s.ctx, poolCoinDenom)
					if !found {
						farm.TotalFarmingAmount = sdk.ZeroInt()
					}
					s.Require().Equal(farm.TotalFarmingAmount, liquidFarm.TotalFarmingAmount)
				}
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.LiquidFarms(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCLiquidFarm() {
	pair := s.createPair(s.addr(0), "denom1", "denom2")
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	minFarmAmt, minBidAmt := sdk.NewInt(10_000_000), sdk.NewInt(10_000_000)
	s.createLiquidFarm(pool.Id, minFarmAmt, minBidAmt, sdk.ZeroDec())

	for _, tc := range []struct {
		name      string
		req       *types.QueryLiquidFarmRequest
		expectErr bool
		postRun   func(*types.QueryLiquidFarmResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"happy case",
			&types.QueryLiquidFarmRequest{
				PoolId: pool.Id,
			},
			false,
			func(resp *types.QueryLiquidFarmResponse) {
				poolCoinDenom := liquiditytypes.PoolCoinDenom(resp.LiquidFarm.PoolId)
				farm, found := s.app.FarmKeeper.GetFarm(s.ctx, poolCoinDenom)
				if !found {
					farm.TotalFarmingAmount = sdk.ZeroInt()
				}
				s.Require().Equal(farm.TotalFarmingAmount, resp.LiquidFarm.TotalFarmingAmount)
				s.Require().Equal(types.LiquidFarmCoinDenom(pool.Id), resp.LiquidFarm.LFCoinDenom)
				s.Require().Equal(types.LiquidFarmReserveAddress(pool.Id).String(), resp.LiquidFarm.LiquidFarmReserveAddress)
				s.Require().Equal(minFarmAmt, resp.LiquidFarm.MinFarmAmount)
				s.Require().Equal(minBidAmt, resp.LiquidFarm.MinBidAmount)
			},
		},
		{
			"query by invalid pool id",
			&types.QueryLiquidFarmRequest{
				PoolId: 5,
			},
			true,
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.LiquidFarm(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCRewardsAuctions() {
	pair := s.createPair(s.addr(0), "denom1", "denom2")
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	s.deposit(s.addr(1), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(2), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(3), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.nextBlock()

	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(1), utils.ParseCoin("100_000pool1"), true)
	s.liquidFarm(pool.Id, s.addr(2), utils.ParseCoin("300_000pool1"), true)
	s.liquidFarm(pool.Id, s.addr(3), utils.ParseCoin("500_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	s.placeBid(pool.Id, s.addr(5), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(pool.Id, s.addr(6), utils.ParseCoin("110_000pool1"), true)
	s.placeBid(pool.Id, s.addr(7), utils.ParseCoin("150_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	s.placeBid(pool.Id, s.addr(5), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(pool.Id, s.addr(6), utils.ParseCoin("110_000pool1"), true)
	s.placeBid(pool.Id, s.addr(7), utils.ParseCoin("150_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	for _, tc := range []struct {
		name      string
		req       *types.QueryRewardsAuctionsRequest
		expectErr bool
		postRun   func(*types.QueryRewardsAuctionsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by invalid pool id",
			&types.QueryRewardsAuctionsRequest{
				PoolId: 0,
			},
			true,
			func(resp *types.QueryRewardsAuctionsResponse) {},
		},
		{
			"query by invalid pool id",
			&types.QueryRewardsAuctionsRequest{
				PoolId: 10,
			},
			false,
			func(resp *types.QueryRewardsAuctionsResponse) {
				s.Require().Len(resp.RewardAuctions, 0)
			},
		},
		{
			"query by pool id",
			&types.QueryRewardsAuctionsRequest{
				PoolId: pool.Id,
			},
			false,
			func(resp *types.QueryRewardsAuctionsResponse) {
				s.Require().Len(resp.RewardAuctions, 3)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.RewardsAuctions(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCRewardsAuction() {
	pair := s.createPair(s.addr(0), "denom1", "denom2")
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	s.deposit(s.addr(1), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(2), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(3), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(1), utils.ParseCoin("100_000pool1"), true)
	s.liquidFarm(pool.Id, s.addr(2), utils.ParseCoin("300_000pool1"), true)
	s.liquidFarm(pool.Id, s.addr(3), utils.ParseCoin("500_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	s.placeBid(pool.Id, s.addr(5), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(pool.Id, s.addr(6), utils.ParseCoin("110_000pool1"), true)
	s.placeBid(pool.Id, s.addr(7), utils.ParseCoin("150_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	for _, tc := range []struct {
		name      string
		req       *types.QueryRewardsAuctionRequest
		expectErr bool
		postRun   func(*types.QueryRewardsAuctionResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by invalid pool id",
			&types.QueryRewardsAuctionRequest{
				PoolId:    10,
				AuctionId: 1,
			},
			true,
			func(resp *types.QueryRewardsAuctionResponse) {},
		},
		{
			"query by invalid auction id",
			&types.QueryRewardsAuctionRequest{
				PoolId:    1,
				AuctionId: 10,
			},
			true,
			func(resp *types.QueryRewardsAuctionResponse) {},
		},
		{
			"query finished auction",
			&types.QueryRewardsAuctionRequest{
				PoolId:    pool.Id,
				AuctionId: 1,
			},
			false,
			func(resp *types.QueryRewardsAuctionResponse) {
				s.Require().Equal(pool.PoolCoinDenom, resp.RewardAuction.BiddingCoinDenom)
				s.Require().Equal(types.PayingReserveAddress(pool.Id), resp.RewardAuction.GetPayingReserveAddress())
				s.Require().Equal(types.AuctionStatusFinished, resp.RewardAuction.Status)
			},
		},
		{
			"query started auction",
			&types.QueryRewardsAuctionRequest{
				PoolId:    pool.Id,
				AuctionId: 2,
			},
			false,
			func(resp *types.QueryRewardsAuctionResponse) {
				s.Require().Equal(pool.PoolCoinDenom, resp.RewardAuction.BiddingCoinDenom)
				s.Require().Equal(types.PayingReserveAddress(pool.Id), resp.RewardAuction.GetPayingReserveAddress())
				s.Require().Equal(types.AuctionStatusStarted, resp.RewardAuction.Status)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.RewardsAuction(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCBids() {
	pair := s.createPair(s.addr(0), "denom1", "denom2")
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	s.deposit(s.addr(1), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(2), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(3), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(1), utils.ParseCoin("100_000pool1"), true)
	s.liquidFarm(pool.Id, s.addr(2), utils.ParseCoin("300_000pool1"), true)
	s.liquidFarm(pool.Id, s.addr(3), utils.ParseCoin("500_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	s.placeBid(pool.Id, s.addr(5), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(pool.Id, s.addr(6), utils.ParseCoin("110_000pool1"), true)
	s.placeBid(pool.Id, s.addr(7), utils.ParseCoin("150_000pool1"), true)
	s.nextBlock()

	for _, tc := range []struct {
		name      string
		req       *types.QueryBidsRequest
		expectErr bool
		postRun   func(*types.QueryBidsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by invalid pool id",
			&types.QueryBidsRequest{
				PoolId: 0,
			},
			true,
			func(resp *types.QueryBidsResponse) {},
		},
		{
			"query by pool id",
			&types.QueryBidsRequest{
				PoolId: pool.Id,
			},
			false,
			func(resp *types.QueryBidsResponse) {
				s.Require().Len(resp.Bids, 3)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Bids(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCMintRate() {
	pair := s.createPairWithLastPrice(s.addr(0), "denom1", "denom2", sdk.NewDec(1))
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	plan := s.createPrivatePlan(s.addr(0), []farmtypes.RewardAllocation{
		{
			PairId:        pool.PairId,
			RewardsPerDay: utils.ParseCoins("100_000_000denom3"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("500_000_000denom3"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(1), utils.ParseCoin("200_000pool1"), true)
	s.liquidFarm(pool.Id, s.addr(2), utils.ParseCoin("300_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	s.liquidFarm(pool.Id, s.addr(3), utils.ParseCoin("500_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	s.placeBid(pool.Id, s.addr(10), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(pool.Id, s.addr(11), utils.ParseCoin("200_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	s.liquidFarm(pool.Id, s.addr(4), utils.ParseCoin("500_000pool1"), true) // Mint rate is changed
	s.nextBlock()

	for _, tc := range []struct {
		name      string
		req       *types.QueryMintRateRequest
		expectErr bool
		postRun   func(*types.QueryMintRateResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by invalid pool id",
			&types.QueryMintRateRequest{
				PoolId: 0,
			},
			true,
			func(resp *types.QueryMintRateResponse) {},
		},
		{
			"query by not found pool id",
			&types.QueryMintRateRequest{
				PoolId: 10,
			},
			false,
			func(resp *types.QueryMintRateResponse) {
				s.Require().True(resp.MintRate.IsZero())
			},
		},
		{
			"query by valid pool id",
			&types.QueryMintRateRequest{
				PoolId: 1,
			},
			false,
			func(resp *types.QueryMintRateResponse) {
				s.Require().True(resp.MintRate.LT(sdk.OneDec()))
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.MintRate(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCBurnRate() {
	pair := s.createPair(s.addr(0), "denom1", "denom2")
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	plan := s.createPrivatePlan(s.addr(0), []farmtypes.RewardAllocation{
		{
			PairId:        pool.PairId,
			RewardsPerDay: utils.ParseCoins("100_000_000denom3"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("500_000_000denom3"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(1), utils.ParseCoin("200_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	s.placeBid(pool.Id, s.addr(10), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(pool.Id, s.addr(11), utils.ParseCoin("200_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	s.liquidFarm(pool.Id, s.addr(2), utils.ParseCoin("200_000pool1"), true)
	s.nextBlock()

	s.liquidUnfarm(pool.Id, s.addr(2), utils.ParseCoin("100_000lf1"), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(3), utils.ParseCoin("200_000pool1"), true) // Burn rate is changed
	s.nextBlock()

	for _, tc := range []struct {
		name      string
		req       *types.QueryBurnRateRequest
		expectErr bool
		postRun   func(*types.QueryBurnRateResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by invalid pool id",
			&types.QueryBurnRateRequest{
				PoolId: 0,
			},
			true,
			func(resp *types.QueryBurnRateResponse) {},
		},
		{
			"query by not found pool id",
			&types.QueryBurnRateRequest{
				PoolId: 10,
			},
			false,
			func(resp *types.QueryBurnRateResponse) {
				s.Require().True(resp.BurnRate.IsZero())
			},
		},
		{
			"query by valid pool id",
			&types.QueryBurnRateRequest{
				PoolId: 1,
			},
			false,
			func(resp *types.QueryBurnRateResponse) {
				s.Require().True(resp.BurnRate.GT(sdk.OneDec()))
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.BurnRate(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}
