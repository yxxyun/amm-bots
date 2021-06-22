package algorithm

import (
	"amm-bots/utils"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/yxxyun/ripple/crypto"
	"github.com/yxxyun/ripple/data"
	"github.com/yxxyun/ripple/websockets"
)

func NewConstProductBot(
	client *websockets.Remote,
	baseToken *data.Amount,
	quoteToken *data.Amount,
	minPrice decimal.Decimal,
	maxPrice decimal.Decimal,
	priceGap decimal.Decimal,
	expandInventory decimal.Decimal,
	address *data.Account,
	privateKey string) *ConstProductBot {

	var lock sync.Mutex
	xfee, _ := data.NewNativeValue(int64(12))
	txflag := new(data.TransactionFlag)
	*txflag = *txflag | data.TxCanonicalSignature
	bot := ConstProductBot{
		client,
		baseToken,
		quoteToken,
		map[uint32]OrderLadder{},
		minPrice,
		maxPrice,
		priceGap,
		expandInventory,
		&lock,
		address,
		privateKey,
		xfee,
		txflag,
	}
	return &bot
}

type OrderLadder struct {
	Ladder ConstProductLadder
	Side   string
}
type ConstProductBot struct {
	client          *websockets.Remote
	baseToken       *data.Amount
	quoteToken      *data.Amount
	ladderMap       map[uint32]OrderLadder
	minPrice        decimal.Decimal
	maxPrice        decimal.Decimal
	priceGap        decimal.Decimal
	expandInventory decimal.Decimal
	updateLock      *sync.Mutex
	Address         *data.Account
	privateKey      string
	fee             *data.Value
	flag            *data.TransactionFlag
}

func (b *ConstProductBot) Run() {
	b.init()
	time.Sleep(30 * time.Second)
	for true {
		b.updateLock.Lock()
		b.maintainOrder()
		b.updateLock.Unlock()
		time.Sleep(30 * time.Second)
	}
}

func (b *ConstProductBot) init() {
	b.CancelAllPendingOrders()
	baseTokenAmount := decimal.NewFromFloat(b.baseToken.Float())
	quoteTokenAmount := decimal.NewFromFloat(b.quoteToken.Float())
	ladders, err := GenerateConstProductLadders(
		baseTokenAmount,
		quoteTokenAmount,
		b.minPrice,
		b.maxPrice,
		b.priceGap,
		b.expandInventory,
	)
	if err != nil {
		logrus.Error("ladders init failed ", err)
		panic(err)
	}
	centerPrice := quoteTokenAmount.Div(baseTokenAmount)
	for _, ladder := range ladders {
		if ladder.UpPrice.LessThanOrEqual(centerPrice) {
			b.createOrder(ladder, utils.BUY)
		} else {
			b.createOrder(ladder, utils.SELL)
		}
	}
}

func (b *ConstProductBot) createOrder(ladder ConstProductLadder, side string) {
	var price decimal.Decimal
	var takerpays, takergets *data.Amount
	var KeySequence *uint32
	seq := uint32(0)
	KeySequence = &seq
	if side == utils.SELL {
		price = ladder.UpPrice
		takergets, _ = data.NewAmount(ladder.Amount.StringFixed(6) + "/" + b.baseToken.Asset().String())
		takerpays, _ = data.NewAmount(ladder.Amount.Mul(price).StringFixed(6) + "/" + b.quoteToken.Asset().String())
	} else {
		price = ladder.DownPrice
		takerpays, _ = data.NewAmount(ladder.Amount.StringFixed(6) + "/" + b.baseToken.Asset().String())
		takergets, _ = data.NewAmount(ladder.Amount.Mul(price).StringFixed(6) + "/" + b.quoteToken.Asset().String())
	}

	airesult, err := b.client.AccountInfo(*b.Address)
	if err != nil {
		logrus.Warn("get acct info failed ", err)
	} else {
		AccountSequence := *airesult.AccountData.Sequence
		LedgerSequence := airesult.LedgerSequence + 4
		tx := &data.OfferCreate{
			TakerPays: *takerpays,
			TakerGets: *takergets,
			TxBase: data.TxBase{
				Account:            *b.Address,
				LastLedgerSequence: &LedgerSequence,
				Sequence:           AccountSequence,
				Fee:                *b.fee,
				TransactionType:    data.OFFER_CREATE,
				Flags:              b.flag,
			},
		}

		seed, _ := crypto.NewRippleHashCheck(b.privateKey, crypto.RIPPLE_FAMILY_SEED)
		key, _ := crypto.NewECDSAKey(seed.Payload())
		err = data.Sign(tx, key, KeySequence)
		if err != nil {
			logrus.Warn("sigh order failed ", err)
		}

		_, err := b.client.Submit(tx, false)
		time.Sleep(1 * time.Second)
		if err != nil {
			logrus.Warn("create order failed ", err)
		} else {
			go b.CheckTx(tx.Hash, OrderLadder{Ladder: ladder, Side: side}, AccountSequence)
		}

	}

}

func (b *ConstProductBot) maintainOrder() {
	orderInfo, err := b.client.AccountOffers(*b.Address, "validated")
	if err != nil {
		logrus.Warn("get order info failed ", err)
	} else {
		offermap := map[uint32]data.AccountOffer{}
		neworderladder := map[uint32]OrderLadder{}
		for _, offer := range orderInfo.Offers {
			offermap[offer.Sequence] = offer
		}
		for k, v := range b.ladderMap {

			if _, ok := offermap[k]; !ok {
				neworderladder[k] = v
				//b.createOrder(v.Ladder, utils.ToggleSide(v.Side))
				delete(b.ladderMap, k)
				logrus.Info("delete #", k)
			}

		}
		for _, l := range neworderladder {
			b.createOrder(l.Ladder, utils.ToggleSide(l.Side))
		}
	}
}

func (b *ConstProductBot) ElegantExit() {
	b.updateLock.Lock()
	b.CancelAllPendingOrders()
}
func (b *ConstProductBot) CheckTx(hash data.Hash256, ladder OrderLadder, seq uint32) {
	for {
		ret, err := b.client.Tx(hash)
		if err != nil {
			logrus.Warn("get tx info failed ", err)
		} else {
			if ret.Validated {
				b.ladderMap[seq] = ladder
				logrus.Info("OfferCreate #", seq)
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
}
func (b *ConstProductBot) CancelAllPendingOrders() {
	var KeySequence *uint32
	seq := uint32(0)
	KeySequence = &seq
	orderInfo, err := b.client.AccountOffers(*b.Address, "validated")
	if err != nil {
		logrus.Warn("get order info failed ", err)
	} else {
		airesult, err := b.client.AccountInfo(*b.Address)
		if err != nil {
			logrus.Warn("get acct info failed ", err)
		} else {
			AccountSequence := *airesult.AccountData.Sequence
			LedgerSequence := airesult.LedgerSequence + 4
			if orderInfo.Offers.Len() > 0 {
				for _, of := range orderInfo.Offers {
					tx := &data.OfferCancel{
						OfferSequence: of.Sequence,
						TxBase: data.TxBase{
							Account:            *b.Address,
							LastLedgerSequence: &LedgerSequence,
							Sequence:           AccountSequence,
							Fee:                *b.fee,
							TransactionType:    data.OFFER_CANCEL,
							Flags:              b.flag,
						},
					}
					seed, _ := crypto.NewRippleHashCheck(b.privateKey, crypto.RIPPLE_FAMILY_SEED)
					key, _ := crypto.NewECDSAKey(seed.Payload())
					data.Sign(tx, key, KeySequence)
					_, err := b.client.Submit(tx, false)
					if err != nil {
						logrus.Warn("OfferCancel failed ", err)
					} else {
						logrus.Info("OfferCancel #", of.Sequence)
						AccountSequence++
						LedgerSequence++
					}
					time.Sleep(1 * time.Second)
				}
			}
		}

	}
}
