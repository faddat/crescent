package amm

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ OrderSource = (*OrderBook)(nil)

type OrderBook struct {
	buys, sells orderBookTicks
}

func NewOrderBook(tickPrec int, orders ...Order) *OrderBook {
	ob := &OrderBook{}
	for _, order := range orders {
		ob.Add(order)
	}
	return ob
}

func (ob *OrderBook) Add(orders ...Order) {
	for _, order := range orders {
		switch order.GetDirection() {
		case Buy:
			ob.buys.add(order)
		case Sell:
			ob.sells.add(order)
		}
	}
}

func (ob *OrderBook) HighestBuyPrice() (sdk.Dec, bool) {
	return ob.buys.highestPrice()
}

func (ob *OrderBook) LowestSellPrice() (sdk.Dec, bool) {
	return ob.sells.lowestPrice()
}

func (ob *OrderBook) BuyAmountOver(price sdk.Dec) sdk.Int {
	return ob.buys.amountOver(price)
}

func (ob *OrderBook) BuyOrdersOver(price sdk.Dec) []Order {
	return ob.buys.ordersOver(price)
}

func (ob *OrderBook) SellAmountUnder(price sdk.Dec) sdk.Int {
	return ob.sells.amountUnder(price)
}

func (ob *OrderBook) SellOrdersUnder(price sdk.Dec) []Order {
	return ob.sells.ordersUnder(price)
}

type orderBookTicks []*orderBookTick

func (ticks *orderBookTicks) findPrice(price sdk.Dec) (i int, exact bool) {
	i = sort.Search(len(*ticks), func(i int) bool {
		return (*ticks)[i].price.LTE(price)
	})
	if i < len(*ticks) && (*ticks)[i].price.Equal(price) {
		exact = true
	}
	return
}

func (ticks *orderBookTicks) add(order Order) {
	i, exact := ticks.findPrice(order.GetPrice())
	if exact {
		(*ticks)[i].add(order)
	} else {
		if i < len(*ticks) {
			// Insert a new order book tick at index i.
			*ticks = append((*ticks)[:i], append([]*orderBookTick{newOrderBookTick(order)}, (*ticks)[i:]...)...)
		} else {
			// Append a new order book tick at the end.
			*ticks = append(*ticks, newOrderBookTick(order))
		}
	}
}

func (ticks *orderBookTicks) highestPrice() (sdk.Dec, bool) {
	if len(*ticks) == 0 {
		return sdk.Dec{}, false
	}
	for _, tick := range *ticks {
		if tick.totalOpenAmount().IsPositive() {
			return tick.price, true
		}
	}
	return sdk.Dec{}, false
}

func (ticks *orderBookTicks) lowestPrice() (sdk.Dec, bool) {
	if len(*ticks) == 0 {
		return sdk.Dec{}, false
	}
	for i := len(*ticks) - 1; i >= 0; i-- {
		if (*ticks)[i].totalOpenAmount().IsPositive() {
			return (*ticks)[i].price, true
		}
	}
	return sdk.Dec{}, false
}

func (ticks *orderBookTicks) amountOver(price sdk.Dec) sdk.Int {
	panic("not implemented")
}

func (ticks *orderBookTicks) amountUnder(price sdk.Dec) sdk.Int {
	panic("not implemented")
}

func (ticks *orderBookTicks) ordersOver(price sdk.Dec) []Order {
	panic("not implemented")
}

func (ticks *orderBookTicks) ordersUnder(price sdk.Dec) []Order {
	panic("not implemented")
}

type orderBookTick struct {
	price  sdk.Dec
	orders []Order
}

func newOrderBookTick(order Order) *orderBookTick {
	return &orderBookTick{
		price:  order.GetPrice(),
		orders: []Order{order},
	}
}

func (tick *orderBookTick) add(order Order) {
	if order.GetPrice().Equal(tick.price) {
		panic(fmt.Sprintf("order price %q != tick price %q", order.GetPrice(), tick.price))
	}
	if first := tick.orders[0]; first.GetDirection() != order.GetDirection() {
		panic(fmt.Sprintf("order direction %q != tick direction %q", order.GetDirection(), first.GetDirection()))
	}
	tick.orders = append(tick.orders, order)
}

func (tick *orderBookTick) totalOpenAmount() sdk.Int {
	amt := sdk.ZeroInt()
	for _, order := range tick.orders {
		amt = amt.Add(order.GetOpenAmount())
	}
	return amt
}
