package tapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ListSystemMessages return the system messages. Use lvl = "" to get all
// messages or set lvl to desired message level. An invalid lvl is set as "".
func (c *Client) ListSystemMessages(lvl string) ([]SystemMessage, error) {
	params := make(url.Values)
	params.Set("tapi_method", "list_system_messages")
	switch strings.ToLower(lvl) {
	case "info":
		params.Set("level", "INFO")
	case "warning":
		params.Set("level", "WARNING")
	case "error":
		params.Set("level", "ERROR")
	}
	resp, err := c.MakeRequest(params)
	if err != nil {
		return nil, err
	}
	return unmarshalSysMsgs(resp.Data)
}

// GetAccountInfo get account data such currency balances
// and withdrawal limits.
func (c *Client) GetAccountInfo() (*AccountInfo, error) {
	params := make(url.Values)
	params.Set("tapi_method", "get_account_info")
	resp, err := c.MakeRequest(params)
	if err != nil {
		return nil, err
	}
	accInfo := new(AccountInfo)
	if err := json.Unmarshal(resp.Data, accInfo); err != nil {
		return nil, err
	}
	return accInfo, nil
}

// GetOrder returns the order data according to the given id.
func (c *Client) GetOrder(c1, c2 Coin, id int) (*Order, error) {
	params := make(url.Values)
	params.Set("tapi_method", "get_order")
	params.Set("coin_pair", c1.String()+c2.String())
	params.Set("order_id", strconv.Itoa(id))
	resp, err := c.MakeRequest(params)
	if err != nil {
		return nil, err
	}
	return unmarshalOrder(resp.Data)
}

// ListOrdersOpts contains the optional values of ListOrder.
// Any field with zero value means not set.
type ListOrdersOpts struct {
	// OrderType filter the order by type (buy or sell).
	// OrderType = 0 means all (not set).
	// Ordertype < 0 means sell.
	// OrderType > 0 means buy.
	OrderType int

	// StatusList filter by order status. More than one
	// option can be set.
	// StatusList[0] != 0 means filter by open order.
	// StatusList[1] != 0 means filter by cancelled order.
	// StatusList[2] != 0 means filter by filled order.
	StatusList [3]int

	// HasFills filter orders with or without execution.
	// HasFills = 0 means not set.
	// HasFills < 0 means without execution.
	// HasFills > 0 means with execution.
	HasFills int

	// FromID filter orders from ID (inclusive).
	FromID int

	// ToID filter orders until ID (inclusive).
	ToID int

	// FromTimestamp filter orders since timestamp.
	FromTimestamp string

	// ToTimestamp filter orders created until timestamp (inclusive).
	ToTimestamp string
}

func parseOpts(params url.Values, opts *ListOrdersOpts) {
	switch t := opts.OrderType; {
	case t < 0:
		params.Set("order_type", strconv.Itoa(2))
	case t > 0:
		params.Set("order_type", strconv.Itoa(1))
	}
	// Range over StatusList and add the values
	// +2 because the count starts on 2.
	sl := ""
	for i, v := range opts.StatusList {
		if v != 0 {
			sl += strconv.Itoa(i+2) + ","
		}
	}
	// If at least one of status was setted, create
	// a list and remove the last comma.
	if sl != "" {
		sl = "[" + sl[:len(sl)-1] + "]"
		params.Set("status_list", sl)
	}
	switch h := opts.HasFills; {
	case h < 0:
		params.Set("has_fills", "false")
	case h > 0:
		params.Set("has_fills", "true")
	}
	if opts.FromID != 0 {
		params.Set("from_id", strconv.Itoa(opts.FromID))
	}
	if opts.ToID != 0 {
		params.Set("to_id", strconv.Itoa(opts.ToID))
	}
	if opts.FromTimestamp != "" {
		params.Set("from_timestamp", opts.FromTimestamp)
	}
	if opts.ToTimestamp != "" {
		params.Set("to_timestamp", opts.ToTimestamp)
	}
}

// ListOrders returns a list of at max 200 orders filtered by opts options.
// Use opts = nil or opts = &ListOrderOps{} to set no options.
func (c *Client) ListOrders(c1, c2 Coin, opts *ListOrdersOpts) ([]Order, error) {
	params := make(url.Values)
	params.Set("tapi_method", "list_orders")
	params.Set("coin_pair", c1.String()+c2.String())
	if opts != nil {
		parseOpts(params, opts)
	}
	resp, err := c.MakeRequest(params)
	if err != nil {
		return nil, err
	}
	return unmarshalListOrders(resp.Data)
}

// ListOrderbook returns the orderbook to the informed coins,
// if full is true returns at max 500 asks and 500 bids,
// if full is false returns at max 20 asks and 20 bids.
func (c *Client) ListOrderbook(c1, c2 Coin, full bool) (*Orderbook, error) {
	params := make(url.Values)
	params.Set("tapi_method", "list_orderbook")
	params.Set("coin_pair", c1.String()+c2.String())
	if full {
		params.Set("full", "true")
	}
	resp, err := c.MakeRequest(params)
	if err != nil {
		return nil, err
	}
	return unmarshalOrderbook(resp.Data)
}

// PlaceBuyOrder opens a buy order of coin pair c1 and c2 with quantity qt of
// digital coin and unit limit price limit.
func (c *Client) PlaceBuyOrder(c1, c2 Coin, qt, limit string) (*Order, error) {
	return c.placeOrder(c1, c2, "place_buy_order", qt, limit)
}

// PlaceSellOrder opens a sell order of coin pair c1 and c2 with quantity qt of
// digital coin and unit limit price limit.
func (c *Client) PlaceSellOrder(c1, c2 Coin, qt, limit string) (*Order, error) {
	return c.placeOrder(c1, c2, "place_sell_order", qt, limit)
}

func (c *Client) placeOrder(c1, c2 Coin, method, qt, limit string) (*Order, error) {
	params := make(url.Values)
	params.Set("tapi_method", method)
	params.Set("coin_pair", c1.String()+c2.String())
	params.Set("quantity", qt)
	params.Set("limit_price", limit)
	resp, err := c.MakeRequest(params)
	if err != nil {
		return nil, err
	}
	return unmarshalOrder(resp.Data)
}

// PlaceMarketBuyOrder opens a buy order of coin pair c1 and c2 with limit
// volume cost in BRL.
func (c *Client) PlaceMarketBuyOrder(c1, c2 Coin, cost string) (*Order, error) {
	params := make(url.Values)
	params.Set("tapi_method", "place_market_buy_order")
	params.Set("coin_pair", c1.String()+c2.String())
	params.Set("cost", cost)
	resp, err := c.MakeRequest(params)
	if err != nil {
		return nil, err
	}
	return unmarshalOrder(resp.Data)
}

// PlaceMarketSellOrder opens a sell order of coin pair c1 and c2 with qt
// quantity of digital coin.
func (c *Client) PlaceMarketSellOrder(c1, c2 Coin, qt string) (*Order, error) {
	params := make(url.Values)
	params.Set("tapi_method", "place_market_sell_order")
	params.Set("coin_pair", c1.String()+c2.String())
	params.Set("quantity", qt)
	resp, err := c.MakeRequest(params)
	if err != nil {
		return nil, err
	}
	return unmarshalOrder(resp.Data)
}

// CancelOrder cancels a buy or sell order by coin pair and id of order.
func (c *Client) CancelOrder(c1, c2 Coin, id int) (*Order, error) {
	params := make(url.Values)
	params.Set("tapi_method", "cancel_order")
	params.Set("coin_pair", c1.String()+c2.String())
	params.Set("order_id", strconv.Itoa(id))
	resp, err := c.MakeRequest(params)
	if err != nil {
		return nil, err
	}
	return unmarshalOrder(resp.Data)
}

// GetWithdrawal returns the data of a transfer of digital coin or
// a withdrawal of BRL.
func (c *Client) GetWithdrawal(coin Coin, id int) (*Withdrawal, error) {
	params := make(url.Values)
	params.Set("tapi_method", "get_withdrawal")
	params.Set("coin", coin.String())
	params.Set("withdrawal_id", strconv.Itoa(id))
	resp, err := c.MakeRequest(params)
	if err != nil {
		return nil, err
	}
	return unmarshalWithdrawal(resp.Data)
}

// WithdrawInfo contains the information to complete a withdrawal.
type WithdrawInfo struct {
	// Address is the Bitcoin/Ethereum/Litecoin/BCash/XRP address.
	Address string

	// Quantity is the liquid transfer value.
	Quantity string

	// TxFee is the transaction fee paid to miners to process the
	// transaction.
	TxFee string

	// TxNotAggregate is setted to not aggregate the transfer with
	// others transfers in one Blockchain transaction. Default is to
	// aggregate.
	TxNotAggregate bool

	// ViaBlockchain set transfer via blockchain to generate
	// a transaction on the Bitcoin/BCash network. Defalut is not
	// via Blockchain.
	ViaBlockchain bool

	// DestinationTag should be set only for XRP.
	DestinationTag int
}

// WithdrawBRL requests a withdrawal BRL with desc decription, qt quantity,
// and accRef ID of a previous registered bank account.
func (c *Client) WithdrawBRL(desc, qt, accRef string) (*Withdrawal, error) {
	params := make(url.Values)
	params.Set("quantity", qt)
	params.Set("account_ref", accRef)
	return c.withdrawCoin(BRL, params, desc)
}

// WithdrawCrypto requests a digital coin transfer order with coin,
// description and withdraw info.
func (c *Client) WithdrawCrypto(coin Coin, desc string, i *WithdrawInfo) (*Withdrawal, error) {
	if i == nil {
		return nil, errors.New("nil WithdrawInfo")
	}
	if coin == BRL {
		return nil, errors.New("use WithdrawBRL for BRL")
	}
	params := make(url.Values)
	params.Set("address", i.Address)
	params.Set("quantity", i.Quantity)
	params.Set("tx_fee", i.TxFee)
	if i.TxNotAggregate {
		params.Set("tx_aggregate", "false")
	}
	if i.ViaBlockchain {
		params.Set("via_blockchain", "true")
	}
	if coin == XRP {
		params.Set("destination_tag", strconv.Itoa(i.DestinationTag))
	}
	return c.withdrawCoin(coin, params, desc)
}

func (c *Client) withdrawCoin(coin Coin, p url.Values, desc string) (*Withdrawal, error) {
	p.Set("tapi_method", "withdraw_coin")
	p.Set("coin", coin.String())
	if desc != "" {
		p.Set("description", desc)
	}
	resp, err := c.MakeRequest(p)
	if err != nil {
		return nil, err
	}
	return unmarshalWithdrawal(resp.Data)
}

// Error contains a code to error that can be
// compared to API error codes.
type Error struct {
	Code int
	Err  string
}

func (e *Error) Error() string { return e.Err }

// Is checks whether e is the same type as the target.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return (e.Code == t.Code || t.Code == 0)
}

// MakeRequest create and make a request with nonce, ID, MAC and params
// to c.service.
func (c *Client) MakeRequest(params url.Values) (*Response, error) {
	params.Add("tapi_nonce", c.Nonce())
	e := params.Encode()

	r, err := http.NewRequest("POST", c.service, strings.NewReader(e))
	if err != nil {
		return nil, err
	}

	mac := c.Hmac(r.URL.Path + "?" + e)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("TAPI-ID", c.apiID)
	r.Header.Set("TAPI-MAC", mac)

	resp, err := c.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := &Response{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 100 {
		err := &Error{Code: response.StatusCode, Err: response.ErrorMessage}
		return nil, err
	}
	return response, nil
}
