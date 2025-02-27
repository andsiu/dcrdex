// This code is available on the terms of the project LICENSE.md file,
// also available online at https://blueoakcouncil.org/license/1.0.0.

package market

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	"decred.org/dcrdex/dex"
	"decred.org/dcrdex/dex/calc"
	"decred.org/dcrdex/dex/msgjson"
	"decred.org/dcrdex/dex/order"
	"decred.org/dcrdex/dex/wait"
	"decred.org/dcrdex/server/account"
	"decred.org/dcrdex/server/asset"
	"decred.org/dcrdex/server/comms"
	"decred.org/dcrdex/server/matcher"
)

// The AuthManager handles client-related actions, including authorization and
// communications.
type AuthManager interface {
	Route(route string, handler func(account.AccountID, *msgjson.Message) *msgjson.Error)
	Auth(user account.AccountID, msg, sig []byte) error
	Suspended(user account.AccountID) (found, suspended bool)
	Sign(...msgjson.Signable)
	Send(account.AccountID, *msgjson.Message) error
	Request(account.AccountID, *msgjson.Message, func(comms.Link, *msgjson.Message)) error
	RequestWithTimeout(account.AccountID, *msgjson.Message, func(comms.Link, *msgjson.Message), time.Duration, func()) error
	PreimageSuccess(user account.AccountID, refTime time.Time, oid order.OrderID)
	MissedPreimage(user account.AccountID, refTime time.Time, oid order.OrderID)
	RecordCancel(user account.AccountID, oid, target order.OrderID, t time.Time)
	RecordCompletedOrder(user account.AccountID, oid order.OrderID, t time.Time)
	UserSettlingLimit(user account.AccountID, mkt *dex.MarketInfo) int64
}

const (
	maxClockOffset = 600_000 // milliseconds => 600 sec => 10 minutes
	fundingTxWait  = time.Minute
	// ZeroConfFeeRateThreshold is multiplied by the last known fee rate for an
	// asset to attain a minimum fee rate acceptable for zero-conf funding
	// coins.
	ZeroConfFeeRateThreshold = 0.9
)

// MarketTunnel is a connection to a market.
type MarketTunnel interface {
	// SubmitOrder submits the order to the market for insertion into the epoch
	// queue.
	SubmitOrder(*orderRecord) error
	// MidGap returns the mid-gap market rate, which is ths rate halfway between
	// the best buy order and the best sell order in the order book.
	MidGap() uint64
	// MarketBuyBuffer is a coefficient that when multiplied by the market's lot
	// size specifies the minimum required amount for a market buy order.
	MarketBuyBuffer() float64
	// LotSize is the market's lot size in units of the base asset.
	LotSize() uint64
	// RateStep is the market's rate step in units of the quote asset.
	RateStep() uint64
	// CoinLocked should return true if the CoinID is currently a funding Coin
	// for an active DEX order. This is required for Coin validation to prevent
	// a user from submitting multiple orders spending the same Coin. This
	// method will likely need to check all orders currently in the epoch queue,
	// the order book, and the swap monitor, since UTXOs will still be unspent
	// according to the asset backends until the client broadcasts their
	// initialization transaction.
	//
	// DRAFT NOTE: This function could also potentially be handled by persistent
	// storage, since active orders and active matches are tracked there.
	CoinLocked(assetID uint32, coinID order.CoinID) bool
	// Cancelable determines whether an order is cancelable. A cancelable order
	// is a limit order with time-in-force standing either in the epoch queue or
	// in the order book.
	Cancelable(order.OrderID) bool

	// Suspend suspends the market as soon as a given time, returning the final
	// epoch index and and time at which that epoch closes.
	Suspend(asSoonAs time.Time, persistBook bool) (finalEpochIdx int64, finalEpochEnd time.Time)

	// Running indicates is the market is accepting new orders. This will return
	// false when suspended, but false does not necessarily mean Run has stopped
	// since a start epoch may be set.
	Running() bool

	// CheckUnfilled checks a user's unfilled book orders that are funded by
	// coins for a given asset to ensure that their funding coins are not spent.
	// If any of an unfilled order's funding coins are spent, the order is
	// unbooked (removed from the in-memory book, revoked in the DB, a
	// cancellation marked against the user, coins unlocked, and orderbook
	// subscribers notified). See Unbook for details.
	CheckUnfilled(assetID uint32, user account.AccountID) (unbooked []*order.LimitOrder)
}

// orderRecord contains the information necessary to respond to an order
// request.
type orderRecord struct {
	order order.Order
	req   msgjson.Stampable
	msgID uint64
}

// assetSet is pointers to two different assets, but with 4 ways of addressing
// them.
type assetSet struct {
	funding   *asset.BackedAsset
	receiving *asset.BackedAsset
	base      *asset.BackedAsset
	quote     *asset.BackedAsset
}

// newAssetSet is a constructor for an assetSet.
func newAssetSet(base, quote *asset.BackedAsset, sell bool) *assetSet {
	coins := &assetSet{
		quote:     quote,
		base:      base,
		funding:   quote,
		receiving: base,
	}
	if sell {
		coins.funding, coins.receiving = base, quote
	}
	return coins
}

// FeeSource is a source of the last reported tx fee rate estimate for an asset.
type FeeSource interface {
	LastRate(assetID uint32) (feeRate uint64)
}

// OrderRouter handles the 'limit', 'market', and 'cancel' DEX routes. These
// are authenticated routes used for placing and canceling orders.
type OrderRouter struct {
	auth        AuthManager
	assets      map[uint32]*asset.BackedAsset
	tunnels     map[string]MarketTunnel
	latencyQ    *wait.TickerQueue
	feeSource   FeeSource
	dexBalancer *DEXBalancer
}

// OrderRouterConfig is the configuration settings for an OrderRouter.
type OrderRouterConfig struct {
	AuthManager AuthManager
	Assets      map[uint32]*asset.BackedAsset
	Markets     map[string]MarketTunnel
	FeeSource   FeeSource
	DEXBalancer *DEXBalancer
}

// NewOrderRouter is a constructor for an OrderRouter.
func NewOrderRouter(cfg *OrderRouterConfig) *OrderRouter {
	router := &OrderRouter{
		auth:        cfg.AuthManager,
		assets:      cfg.Assets,
		tunnels:     cfg.Markets,
		latencyQ:    wait.NewTickerQueue(2 * time.Second),
		feeSource:   cfg.FeeSource,
		dexBalancer: cfg.DEXBalancer,
	}
	cfg.AuthManager.Route(msgjson.LimitRoute, router.handleLimit)
	cfg.AuthManager.Route(msgjson.MarketRoute, router.handleMarket)
	cfg.AuthManager.Route(msgjson.CancelRoute, router.handleCancel)
	return router
}

func (r *OrderRouter) Run(ctx context.Context) {
	r.latencyQ.Run(ctx)
}

func (r *OrderRouter) respondError(reqID uint64, user account.AccountID, msgErr *msgjson.Error) {
	log.Debugf("Error going to user %v: %s", user, msgErr)
	msg, err := msgjson.NewResponse(reqID, nil, msgErr)
	if err != nil {
		log.Errorf("Failed to create error response with message '%s': %v", msg, err)
		return // this should not be possible, but don't pass nil msg to Send
	}
	if err := r.auth.Send(user, msg); err != nil {
		log.Infof("Failed to send %s error response (msg = %s) to disconnected user %v: %q",
			msg.Route, msgErr, user, err)
	}
}

func fundingCoin(backend asset.Backend, coinID []byte, redeemScript []byte) (asset.FundingCoin, error) {
	outputTracker, is := backend.(asset.OutputTracker)
	if !is {
		return nil, fmt.Errorf("fundingCoin requested for incapable asset")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return outputTracker.FundingCoin(ctx, coinID, redeemScript)
}

func coinConfirmations(coin asset.Coin) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return coin.Confirmations(ctx)
}

// handleLimit is the handler for the 'limit' route. This route accepts a
// msgjson.Limit payload, validates the information, constructs an
// order.LimitOrder and submits it to the epoch queue.
func (r *OrderRouter) handleLimit(user account.AccountID, msg *msgjson.Message) *msgjson.Error {
	limit := new(msgjson.LimitOrder)
	err := msg.Unmarshal(&limit)
	if err != nil || limit == nil {
		return msgjson.NewError(msgjson.RPCParseError, "error decoding 'limit' payload")
	}

	rpcErr := r.verifyAccount(user, limit.AccountID, limit)
	if rpcErr != nil {
		return rpcErr
	}

	if _, suspended := r.auth.Suspended(user); suspended {
		return msgjson.NewError(msgjson.MarketNotRunningError, "suspended account %v may not submit trade orders", user)
	}

	tunnel, assets, sell, rpcErr := r.extractMarketDetails(&limit.Prefix, &limit.Trade)
	if rpcErr != nil {
		return rpcErr
	}

	// Spare some resources if the market is closed now. Any orders that make it
	// through to a closed market will receive a similar error from SubmitOrder.
	if !tunnel.Running() {
		return msgjson.NewError(msgjson.MarketNotRunningError, "market closed to new orders")
	}

	// Check that OrderType is set correctly
	if limit.OrderType != msgjson.LimitOrderNum {
		return msgjson.NewError(msgjson.OrderParameterError, "wrong order type set for limit order. wanted %d, got %d", msgjson.LimitOrderNum, limit.OrderType)
	}

	// Check that the rate is non-zero and obeys the rate step interval.
	if limit.Rate == 0 {
		return msgjson.NewError(msgjson.OrderParameterError, "rate = 0 not allowed")
	}
	if rateStep := tunnel.RateStep(); limit.Rate%rateStep != 0 {
		return msgjson.NewError(msgjson.OrderParameterError, "rate (%d) not a multiple of ratestep (%d)",
			limit.Rate, rateStep)
	}

	// Check time-in-force
	var force order.TimeInForce
	switch limit.TiF {
	case msgjson.StandingOrderNum:
		force = order.StandingTiF
	case msgjson.ImmediateOrderNum:
		force = order.ImmediateTiF
	default:
		return msgjson.NewError(msgjson.OrderParameterError, "unknown time-in-force")
	}

	lotSize := tunnel.LotSize()
	rpcErr = r.checkPrefixTrade(assets, lotSize, &limit.Prefix, &limit.Trade, true)
	if rpcErr != nil {
		return rpcErr
	}

	// Commitment
	if len(limit.Commit) != order.CommitmentSize {
		return msgjson.NewError(msgjson.OrderParameterError, "invalid commitment")
	}
	var commit order.Commitment
	copy(commit[:], limit.Commit)

	coinIDs := make([]order.CoinID, 0, len(limit.Trade.Coins))
	for _, coin := range limit.Trade.Coins {
		coinID := order.CoinID(coin.ID)
		coinIDs = append(coinIDs, coinID)
	}

	// Create the limit order.
	lo := &order.LimitOrder{
		P: order.Prefix{
			AccountID:  user,
			BaseAsset:  limit.Base,
			QuoteAsset: limit.Quote,
			OrderType:  order.LimitOrderType,
			ClientTime: time.UnixMilli(int64(limit.ClientTime)),
			//ServerTime set in epoch queue processing pipeline.
			Commit: commit,
		},
		T: order.Trade{
			Coins:    coinIDs,
			Sell:     sell,
			Quantity: limit.Quantity,
			Address:  limit.Address,
		},
		Rate:  limit.Rate,
		Force: force,
	}

	// NOTE: ServerTime is not yet set, so the order's ID, which is computed
	// from the serialized order, is not yet valid. The Market will stamp the
	// order on receipt, and the order ID will be valid.

	oRecord := &orderRecord{
		order: lo,
		req:   limit,
		msgID: msg.ID,
	}

	return r.processTrade(oRecord, tunnel, assets, limit.Coins, sell, limit.Rate, limit.RedeemSig, limit.Serialize())
}

// handleMarket is the handler for the 'market' route. This route accepts a
// msgjson.MarketOrder payload, validates the information, constructs an
// order.MarketOrder and submits it to the epoch queue.
func (r *OrderRouter) handleMarket(user account.AccountID, msg *msgjson.Message) *msgjson.Error {
	market := new(msgjson.MarketOrder)
	err := msg.Unmarshal(&market)
	if err != nil || market == nil {
		return msgjson.NewError(msgjson.RPCParseError, "error decoding 'market' payload")
	}

	rpcErr := r.verifyAccount(user, market.AccountID, market)
	if rpcErr != nil {
		return rpcErr
	}

	if _, suspended := r.auth.Suspended(user); suspended {
		return msgjson.NewError(msgjson.MarketNotRunningError, "suspended account %v may not submit trade orders", user)
	}

	tunnel, assets, sell, rpcErr := r.extractMarketDetails(&market.Prefix, &market.Trade)
	if rpcErr != nil {
		return rpcErr
	}

	if !tunnel.Running() {
		mktName, _ := dex.MarketName(market.Base, market.Quote)
		return msgjson.NewError(msgjson.MarketNotRunningError, "market %s closed to new orders", mktName)
	}

	// Check that OrderType is set correctly
	if market.OrderType != msgjson.MarketOrderNum {
		return msgjson.NewError(msgjson.OrderParameterError, "wrong order type set for market order")
	}

	// Passing sell as the checkLot parameter causes the lot size check to be
	// ignored for market buy orders.
	lotSize := tunnel.LotSize()
	rpcErr = r.checkPrefixTrade(assets, lotSize, &market.Prefix, &market.Trade, sell)
	if rpcErr != nil {
		return rpcErr
	}

	// Commitment.
	if len(market.Commit) != order.CommitmentSize {
		return msgjson.NewError(msgjson.OrderParameterError, "invalid commitment")
	}
	var commit order.Commitment
	copy(commit[:], market.Commit)

	coinIDs := make([]order.CoinID, 0, len(market.Trade.Coins))
	for _, coin := range market.Trade.Coins {
		coinID := order.CoinID(coin.ID)
		coinIDs = append(coinIDs, coinID)
	}

	// Create the market order
	mo := &order.MarketOrder{
		P: order.Prefix{
			AccountID:  user,
			BaseAsset:  market.Base,
			QuoteAsset: market.Quote,
			OrderType:  order.MarketOrderType,
			ClientTime: time.UnixMilli(int64(market.ClientTime)),
			//ServerTime set in epoch queue processing pipeline.
			Commit: commit,
		},
		T: order.Trade{
			Coins:    coinIDs,
			Sell:     sell,
			Quantity: market.Quantity,
			Address:  market.Address,
		},
	}

	// Send the order to the epoch queue.
	oRecord := &orderRecord{
		order: mo,
		req:   market,
		msgID: msg.ID,
	}

	return r.processTrade(oRecord, tunnel, assets, market.Coins, sell, 0, market.RedeemSig, market.Serialize())
}

// processTrade checks that the trade is valid and submits it to the market.
func (r *OrderRouter) processTrade(oRecord *orderRecord, tunnel MarketTunnel, assets *assetSet,
	coins []*msgjson.Coin, sell bool, rate uint64, redeemSig *msgjson.RedeemSig, sigMsg []byte) *msgjson.Error {

	fundingAsset := assets.funding
	user := oRecord.order.User()
	trade := oRecord.order.Trade()

	// If the receiving asset is account-based, we need to check that they can
	// cover fees for the redemption, since they can't be subtracted from the
	// received amount.
	receivingBalancer, isToAccount := assets.receiving.Backend.(asset.AccountBalancer)
	if isToAccount {
		if redeemSig == nil {
			log.Infof("user %s did not include a RedeemSig for received asset %s", user, assets.receiving.Symbol)
			return msgjson.NewError(msgjson.OrderParameterError, "no redeem address verification included for asset %s", assets.receiving.Symbol)
		}

		acctAddr := trade.ToAccount()
		if err := receivingBalancer.ValidateSignature(acctAddr, redeemSig.PubKey, sigMsg, redeemSig.Sig); err != nil {
			log.Infof("user %s failed redeem signature validation for order: %v",
				user, err)
			return msgjson.NewError(msgjson.SignatureError, "redeem signature validation failed")
		}

		if !r.sufficientAccountBalance(acctAddr, oRecord.order, &assets.receiving.Asset, assets.receiving.ID, tunnel) {
			return msgjson.NewError(msgjson.FundingError, "insufficient balance")
		}
	}

	// If the funding asset is account-based, we'll check balance and submit the
	// order immediately, since we don't need to find coins.
	fundingBalancer, isAccountFunded := assets.funding.Backend.(asset.AccountBalancer)
	if isAccountFunded {
		// Validate that the coins are correct for an account-based-asset-funded
		// order. There should be 1 coin, 1 sig, 1 pubkey, and no redeem script.
		if len(coins) != 1 {
			log.Infof("user %s submitted an %s-funded order with %d coin IDs", user, assets.funding.Symbol, len(coins))
			return msgjson.NewError(msgjson.OrderParameterError, "account-type asset funding requires exactly one coin ID")
		}
		acctProof := coins[0]
		if len(acctProof.PubKeys) != 1 || len(acctProof.Sigs) != 1 || len(acctProof.Redeem) > 0 {
			log.Infof("user %s submitted an %s-funded order with %d pubkeys, %d sigs, redeem script length %d",
				user, assets.funding.Symbol, len(acctProof.PubKeys), len(acctProof.Sigs), len(acctProof.Redeem))
			return msgjson.NewError(msgjson.OrderParameterError, "account-type asset funding requires exactly one coin ID")
		}

		acctAddr := trade.FromAccount()
		pubKey := acctProof.PubKeys[0]
		sig := acctProof.Sigs[0]
		if err := fundingBalancer.ValidateSignature(acctAddr, pubKey, sigMsg, sig); err != nil {
			log.Infof("user %s failed signature validation for order: %v",
				user, err)
			return msgjson.NewError(msgjson.SignatureError, "signature validation failed")
		}

		if !r.sufficientAccountBalance(acctAddr, oRecord.order, &assets.funding.Asset, assets.receiving.ID, tunnel) {
			return msgjson.NewError(msgjson.FundingError, "insufficient balance")
		}
		return r.submitOrderToMarket(tunnel, oRecord)
	}

	// Funding coins are from a utxo-based asset. Need to find them.

	// Validate coin IDs and prepare some strings for debug logging.
	coinStrs := make([]string, 0, len(coins))
	for _, coinID := range trade.Coins {
		coinStr, err := fundingAsset.Backend.ValidateCoinID(coinID)
		if err != nil {
			return msgjson.NewError(msgjson.FundingError, fmt.Sprintf("invalid coin ID %v: %v", coinID, err))
		}
		// TODO: Check all markets here?
		if tunnel.CoinLocked(assets.funding.ID, coinID) {
			return msgjson.NewError(msgjson.FundingError, fmt.Sprintf("coin %s is locked", fmtCoinID(assets.funding.ID, coinID)))
		}
		coinStrs = append(coinStrs, coinStr)
	}

	// Use this as a chance to check user's existing market orders.
	// TODO: check all markets?
	for mktName, tunnel := range r.tunnels {
		unbookedUnfunded := tunnel.CheckUnfilled(assets.funding.ID, oRecord.order.User())
		for _, badLo := range unbookedUnfunded {
			log.Infof("Unbooked unfunded order %v from market %s for user %v", badLo, mktName, oRecord.order.User())
		}
	}

	lotSize := tunnel.LotSize()

	midGap := tunnel.MidGap()
	if midGap == 0 {
		midGap = tunnel.RateStep()
	}

	lots := trade.Quantity / lotSize
	if !sell && rate == 0 {
		lots = matcher.QuoteToBase(midGap, trade.Quantity) / lotSize
	}

	var valSum uint64
	var spendSize uint32
	neededCoins := make(map[int]*msgjson.Coin, len(trade.Coins))
	for i, coin := range coins {
		neededCoins[i] = coin
	}

	checkCoins := func() (tryAgain bool, msgErr *msgjson.Error) {
		for key, coin := range neededCoins {
			// Get the coin from the backend and validate it.
			dexCoin, err := fundingCoin(fundingAsset.Backend, coin.ID, coin.Redeem)
			if err != nil {
				if errors.Is(err, asset.CoinNotFoundError) {
					return true, nil
				}
				if errors.Is(err, asset.ErrRequestTimeout) {
					log.Errorf("Deadline exceeded attempting to verify funding coin %v (%s). Will try again.",
						coin.ID, fundingAsset.Symbol)
					return true, nil
				}
				log.Errorf("Error retrieving limit order funding coin ID %s. user = %s: %v", coin.ID, user, err)
				return false, msgjson.NewError(msgjson.FundingError, fmt.Sprintf("error retrieving coin ID %v", coin.ID))
			}

			// Verify that the user controls the funding coins.
			err = dexCoin.Auth(msgBytesToBytes(coin.PubKeys), msgBytesToBytes(coin.Sigs), coin.ID)
			if err != nil {
				log.Debugf("Auth error for %s coin %s: %v", fundingAsset.Symbol, dexCoin, err)
				return false, msgjson.NewError(msgjson.CoinAuthError, fmt.Sprintf("failed to authorize coin %v", dexCoin))
			}

			msgErr := r.checkZeroConfs(dexCoin, fundingAsset)
			if msgErr != nil {
				return false, msgErr
			}

			delete(neededCoins, key) // don't check this coin again
			valSum += dexCoin.Value()
			// NOTE: Summing like this is actually not quite sufficient to
			// estimate the size associated with the input, because if it's a
			// BTC segwit output, we would also have to account for the marker
			// and flag weight, but only once per tx. The weight would add
			// either 0 or 1 byte to the tx virtual size, so we have a chance of
			// under-estimating by 1 byte to the advantage of the client. It
			// won't ever cause issues though, because we also require funding
			// for a change output in the final swap, which is actually not
			// needed, so there's some buffer.
			spendSize += dexCoin.SpendSize()
		}

		if valSum == 0 {
			return false, msgjson.NewError(msgjson.FundingError, "zero value funding coins not permitted")
		}

		// Calculate the fees and check that the utxo sum is enough.
		var reqVal uint64
		if sell {
			reqVal = calc.RequiredOrderFunds(trade.Quantity, uint64(spendSize), lots, &fundingAsset.Asset)
		} else {
			if rate > 0 { // limit buy
				quoteQty := calc.BaseToQuote(rate, trade.Quantity)
				reqVal = calc.RequiredOrderFunds(quoteQty, uint64(spendSize), lots, &assets.quote.Asset)
			} else {
				// This is a market buy order, so the quantity gets special handling.
				// 1. The quantity is in units of the quote asset.
				// 2. The quantity has to satisfy the market buy buffer.
				midGap := tunnel.MidGap()
				if midGap == 0 {
					midGap = tunnel.RateStep()
				}
				buyBuffer := tunnel.MarketBuyBuffer()
				lotWithBuffer := uint64(float64(lotSize) * buyBuffer)
				minReq := matcher.BaseToQuote(midGap, lotWithBuffer)
				if trade.Quantity < minReq {
					errStr := fmt.Sprintf("order quantity does not satisfy market buy buffer. %d < %d. midGap = %d", trade.Quantity, minReq, midGap)
					return false, msgjson.NewError(msgjson.FundingError, errStr)
				}
				reqVal = calc.RequiredOrderFunds(minReq, uint64(spendSize), 1, &assets.quote.Asset)
			}

		}
		if valSum < reqVal {
			return false, msgjson.NewError(msgjson.FundingError,
				fmt.Sprintf("not enough funds. need at least %d, got %d", reqVal, valSum))
		}

		return false, nil
	}

	log.Tracef("Searching for %s coins %v for new order", fundingAsset.Symbol, coinStrs)
	r.latencyQ.Wait(&wait.Waiter{
		Expiration: time.Now().Add(fundingTxWait),
		TryFunc: func() wait.TryDirective {
			tryAgain, msgErr := checkCoins()
			if tryAgain {
				return wait.TryAgain
			}
			if msgErr != nil {
				r.respondError(oRecord.msgID, user, msgErr)
				return wait.DontTryAgain
			}

			// Send the order to the epoch queue where it will be time stamped.
			log.Tracef("Found and validated %s coins %v for new order", fundingAsset.Symbol, coinStrs)
			if msgErr := r.submitOrderToMarket(tunnel, oRecord); msgErr != nil {
				r.respondError(oRecord.msgID, user, msgErr)
			}
			return wait.DontTryAgain
		},
		ExpireFunc: func() {
			// Tell them to broadcast again or check their node before broadcast
			// timeout is reached and the match is revoked.
			r.respondError(oRecord.msgID, user, msgjson.NewError(msgjson.TransactionUndiscovered,
				fmt.Sprintf("failed to find funding coins %v", coinStrs)))
		},
	})

	return nil
}

// sufficientAccountBalance checks that the user's account-based asset balance
// is sufficient to support the order, considering the user's other orders and
// active matches across all DEX markets.
func (r *OrderRouter) sufficientAccountBalance(accountAddr string, ord order.Order, assetInfo *dex.Asset, redeemAssetID uint32, tunnel MarketTunnel) bool {
	assetID := assetInfo.ID
	trade := ord.Trade()

	var fundingQty, fundingLots uint64
	var redeems int
	if ord.Base() == assetID {
		if trade.Sell {
			fundingQty = trade.Quantity
			fundingLots = trade.Quantity / tunnel.LotSize()
		} else {
			if lo, ok := ord.(*order.LimitOrder); ok {
				redeems = int(calc.QuoteToBase(lo.Rate, trade.Quantity) / tunnel.LotSize())
			} else {
				redeems = int(calc.QuoteToBase(safeMidGap(tunnel), trade.Quantity) / tunnel.LotSize())
			}
		}
	} else {
		if trade.Sell {
			redeems = int(trade.Quantity / tunnel.LotSize())
		} else {
			if lo, ok := ord.(*order.LimitOrder); ok {
				fundingQty = calc.BaseToQuote(lo.Rate, trade.Quantity)
				fundingLots = trade.Quantity / tunnel.LotSize()
			} else { // market buy
				fundingQty = trade.Quantity
				fundingLots = fundingQty / tunnel.LotSize()
			}
		}
	}

	return r.dexBalancer.CheckBalance(accountAddr, assetID, redeemAssetID, fundingQty, fundingLots, redeems)
}

func (r *OrderRouter) submitOrderToMarket(tunnel MarketTunnel, oRecord *orderRecord) *msgjson.Error {
	if err := tunnel.SubmitOrder(oRecord); err != nil {
		code := msgjson.UnknownMarketError
		switch {
		case errors.Is(err, ErrInternalServer):
			log.Errorf("Market failed to SubmitOrder: %v", err)
		case errors.Is(err, ErrQuantityTooHigh):
			code = msgjson.OrderQuantityTooHigh
			fallthrough
		default:
			log.Debugf("Market failed to SubmitOrder: %v", err)
		}
		return msgjson.NewError(code, err.Error())
	}
	return nil
}

// Check the FundingCoin confirmations, and if zero, ensure the tx fee rate
// is sufficient, > 90% of our last recorded estimate for the asset.
func (r *OrderRouter) checkZeroConfs(dexCoin asset.FundingCoin, fundingAsset *asset.BackedAsset) *msgjson.Error {
	// Verify that zero-conf coins are within 10% of the last known fee
	// rate.
	confs, err := coinConfirmations(dexCoin)
	if err != nil {
		log.Debugf("Confirmations error for %s coin %s: %v", fundingAsset.Symbol, dexCoin, err)
		return msgjson.NewError(msgjson.FundingError, fmt.Sprintf("failed to verify coin %v", dexCoin))
	}
	if confs > 0 {
		return nil
	}
	lastKnownFeeRate := r.feeSource.LastRate(fundingAsset.ID) // MaxFeeRate applied inside feeSource
	feeMinimum := uint64(math.Round(float64(lastKnownFeeRate) * ZeroConfFeeRateThreshold))
	feeRate := dexCoin.FeeRate()
	if lastKnownFeeRate > 0 && feeRate < feeMinimum {
		log.Debugf("Fee rate too low %s coin %s: %d < %d", fundingAsset.Symbol, dexCoin, feeRate, feeMinimum)
		return msgjson.NewError(msgjson.FundingError,
			fmt.Sprintf("fee rate for %s is too low. %d < %d", dexCoin, feeRate, feeMinimum))
	}
	return nil
}

// handleCancel is the handler for the 'cancel' route. This route accepts a
// msgjson.Cancel payload, validates the information, constructs an
// order.CancelOrder and submits it to the epoch queue.
func (r *OrderRouter) handleCancel(user account.AccountID, msg *msgjson.Message) *msgjson.Error {
	cancel := new(msgjson.CancelOrder)
	err := msg.Unmarshal(&cancel)
	if err != nil || cancel == nil {
		return msgjson.NewError(msgjson.RPCParseError, "error decoding 'cancel' payload")
	}

	rpcErr := r.verifyAccount(user, cancel.AccountID, cancel)
	if rpcErr != nil {
		return rpcErr
	}

	// NOTE: Allow suspended accounts to submit cancel orders.

	tunnel, rpcErr := r.extractMarket(&cancel.Prefix)
	if rpcErr != nil {
		return rpcErr
	}

	if len(cancel.TargetID) != order.OrderIDSize {
		return msgjson.NewError(msgjson.OrderParameterError, "invalid target ID format")
	}
	var targetID order.OrderID
	copy(targetID[:], cancel.TargetID)

	if !tunnel.Cancelable(targetID) {
		return msgjson.NewError(msgjson.OrderParameterError, "target order not known: %v", targetID)
	}

	// Check that OrderType is set correctly
	if cancel.OrderType != msgjson.CancelOrderNum {
		return msgjson.NewError(msgjson.OrderParameterError, "wrong order type set for cancel order")
	}

	rpcErr = checkTimes(&cancel.Prefix)
	if rpcErr != nil {
		return rpcErr
	}

	// Commitment.
	if len(cancel.Commit) != order.CommitmentSize {
		return msgjson.NewError(msgjson.OrderParameterError, "invalid commitment")
	}
	var commit order.Commitment
	copy(commit[:], cancel.Commit)

	// Create the cancel order
	co := &order.CancelOrder{
		P: order.Prefix{
			AccountID:  user,
			BaseAsset:  cancel.Base,
			QuoteAsset: cancel.Quote,
			OrderType:  order.CancelOrderType,
			ClientTime: time.UnixMilli(int64(cancel.ClientTime)),
			//ServerTime set in epoch queue processing pipeline.
			Commit: commit,
		},
		TargetOrderID: targetID,
	}

	// Send the order to the epoch queue.
	oRecord := &orderRecord{
		order: co,
		req:   cancel,
		msgID: msg.ID,
	}
	if err := tunnel.SubmitOrder(oRecord); err != nil {
		if errors.Is(err, ErrInternalServer) {
			log.Errorf("Market failed to SubmitOrder: %v", err)
		}
		return msgjson.NewError(msgjson.UnknownMarketError, err.Error())
	}
	return nil
}

// verifyAccount checks that the submitted order squares with the submitting user.
func (r *OrderRouter) verifyAccount(user account.AccountID, msgAcct msgjson.Bytes, signable msgjson.Signable) *msgjson.Error {
	// Verify account ID matches.
	if !bytes.Equal(user[:], msgAcct) {
		return msgjson.NewError(msgjson.OrderParameterError, "account ID mismatch")
	}
	// Check the clients signature of the order.
	sigMsg := signable.Serialize()
	err := r.auth.Auth(user, sigMsg, signable.SigBytes())
	if err != nil {
		return msgjson.NewError(msgjson.SignatureError, "signature error: "+err.Error())
	}
	return nil
}

// extractMarket finds the MarketTunnel for the provided prefix.
func (r *OrderRouter) extractMarket(prefix *msgjson.Prefix) (MarketTunnel, *msgjson.Error) {
	mktName, err := dex.MarketName(prefix.Base, prefix.Quote)
	if err != nil {
		return nil, msgjson.NewError(msgjson.UnknownMarketError, "asset lookup error: "+err.Error())
	}
	tunnel, found := r.tunnels[mktName]
	if !found {
		return nil, msgjson.NewError(msgjson.UnknownMarketError, "unknown market "+mktName)
	}
	return tunnel, nil
}

// SuspendEpoch holds the index and end time of final epoch marking the
// suspension of a market.
type SuspendEpoch struct {
	Idx int64
	End time.Time
}

// SuspendMarket schedules a suspension of a given market, with the option to
// persist the orders on the book (or purge the book automatically on market
// shutdown). The scheduled final epoch and suspend time are returned. Note that
// OrderRouter is a proxy for this request to the ultimate Market. This is done
// because OrderRouter is the entry point for new orders into the market. TODO:
// track running, suspended, and scheduled-suspended markets, appropriately
// blocking order submission according to the schedule rather than just checking
// Market.Running prior to submitting incoming orders to the Market.
func (r *OrderRouter) SuspendMarket(mktName string, asSoonAs time.Time, persistBooks bool) *SuspendEpoch {
	mkt, found := r.tunnels[mktName]
	if !found {
		return nil
	}

	idx, t := mkt.Suspend(asSoonAs, persistBooks)
	return &SuspendEpoch{
		Idx: idx,
		End: t,
	}
}

// Suspend is like SuspendMarket, but for all known markets.
func (r *OrderRouter) Suspend(asSoonAs time.Time, persistBooks bool) map[string]*SuspendEpoch {

	suspendTimes := make(map[string]*SuspendEpoch, len(r.tunnels))
	for name, mkt := range r.tunnels {
		idx, ts := mkt.Suspend(asSoonAs, persistBooks)
		suspendTimes[name] = &SuspendEpoch{Idx: idx, End: ts}
	}

	// MarketTunnel.Running will return false when the market closes, and true
	// when and if it opens again. Locking/blocking of the incoming order
	// handlers is not necessary since any orders that sneak in to a Market will
	// be rejected if there is no active epoch.

	return suspendTimes
}

// extractMarketDetails finds the MarketTunnel, an assetSet, and market side for
// the provided prefix.
func (r *OrderRouter) extractMarketDetails(prefix *msgjson.Prefix, trade *msgjson.Trade) (MarketTunnel, *assetSet, bool, *msgjson.Error) {
	// Check that assets are for a valid market.
	tunnel, rpcErr := r.extractMarket(prefix)
	if rpcErr != nil {
		return nil, nil, false, rpcErr
	}
	// Side must be one of buy or sell
	var sell bool
	switch trade.Side {
	case msgjson.BuyOrderNum:
	case msgjson.SellOrderNum:
		sell = true
	default:
		return nil, nil, false, msgjson.NewError(msgjson.OrderParameterError,
			fmt.Sprintf("invalid side value %d", trade.Side))
	}
	quote, found := r.assets[prefix.Quote]
	if !found {
		panic("missing quote asset for known market should be impossible")
	}
	base, found := r.assets[prefix.Base]
	if !found {
		panic("missing base asset for known market should be impossible")
	}
	return tunnel, newAssetSet(base, quote, sell), sell, nil
}

// checkTimes validates the timestamps in an order prefix.
func checkTimes(prefix *msgjson.Prefix) *msgjson.Error {
	offset := time.Now().UnixMilli() - int64(prefix.ClientTime)
	if offset < 0 {
		offset *= -1
	}
	if offset >= maxClockOffset {
		return msgjson.NewError(msgjson.ClockRangeError, fmt.Sprintf(
			"clock offset of %d ms is larger than maximum allowed, %d ms",
			offset, maxClockOffset,
		))
	}
	// Server time should be unset.
	if prefix.ServerTime != 0 {
		return msgjson.NewError(msgjson.OrderParameterError, "non-zero server time not allowed")
	}
	return nil
}

// checkPrefixTrade validates the information in the prefix and trade portions
// of an order.
func (r *OrderRouter) checkPrefixTrade(assets *assetSet, lotSize uint64, prefix *msgjson.Prefix,
	trade *msgjson.Trade, checkLot bool) *msgjson.Error {
	// Check that the client's timestamp is still valid.
	rpcErr := checkTimes(prefix)
	if rpcErr != nil {
		return rpcErr
	}
	// Check that the address is valid.
	if !assets.receiving.Backend.CheckAddress(trade.Address) {
		return msgjson.NewError(msgjson.OrderParameterError, "address doesn't check")
	}
	// Quantity cannot be zero, and must be an integral multiple of the lot size.
	if trade.Quantity == 0 {
		return msgjson.NewError(msgjson.OrderParameterError, "zero quantity not allowed")
	}
	if checkLot && trade.Quantity%lotSize != 0 {
		return msgjson.NewError(msgjson.OrderParameterError, "order quantity not a multiple of lot size")
	}
	// Validate UTXOs
	// Check that all required arrays are of equal length.
	if len(trade.Coins) == 0 {
		return msgjson.NewError(msgjson.FundingError, "order must specify utxos")
	}

	for i, coin := range trade.Coins {
		sigCount := len(coin.Sigs)
		if sigCount == 0 {
			return msgjson.NewError(msgjson.SignatureError, fmt.Sprintf("no signature for coin %d", i))
		}
		if len(coin.PubKeys) != sigCount {
			return msgjson.NewError(msgjson.OrderParameterError, fmt.Sprintf(
				"pubkey count %d not equal to signature count %d for coin %d",
				len(coin.PubKeys), sigCount, i,
			))
		}
	}

	return nil
}

// msgBytesToBytes converts a []msgjson.Byte to a [][]byte.
func msgBytesToBytes(msgBs []msgjson.Bytes) [][]byte {
	b := make([][]byte, 0, len(msgBs))
	for _, msgB := range msgBs {
		b = append(b, msgB)
	}
	return b
}

// fmtCoinID formats the coin ID by asset. If an error is encountered, the
// coinID string returned hex-encoded and prepended with "unparsed:".
func fmtCoinID(assetID uint32, coinID []byte) string {
	strID, err := asset.DecodeCoinID(assetID, coinID)
	if err != nil {
		return "unparsed:" + hex.EncodeToString(coinID)
	}
	return strID
}

// fmtCoinIDs is like fmtCoinID but for a slice of CoinIDs, printing with
// default Go slice formatting like "[coin1 coin2 ...]".
func fmtCoinIDs(assetID uint32, coinIDs []order.CoinID) string {
	out := make([]string, len(coinIDs))
	for i := range coinIDs {
		out[i] = fmtCoinID(assetID, coinIDs[i])
	}
	return fmt.Sprint(out)
}

func safeMidGap(tunnel MarketTunnel) uint64 {
	midGap := tunnel.MidGap()
	if midGap == 0 {
		return tunnel.RateStep()
	}
	return midGap
}
