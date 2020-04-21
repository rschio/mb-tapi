package tapi

import "encoding/json"

type Coin uint8

const (
	BRL Coin = iota
	BTC
	LTC
	BCH
	XRP
	ETH
)

type Response struct {
	Data                json.RawMessage `json:"response_data"`
	StatusCode          int             `json:"status_code"`
	ErrorMessage        string          `json:"error_message"`
	ServerUnixTimestamp string          `json:"server_unix_timestamp"`
}

type SystemMessage struct {
	MsgDate    string `json:"msg_date"`
	Level      string `json:"level"`
	EventCode  int    `json:"event_code"`
	MsgContent string `json:"msg_content"`
}

type BalanceCrypto struct {
	Amount
	OpenOrders int `json:"amount_open_orders"`
}

type Amount struct {
	Available string `json:"available"`
	Total     string `json:"total"`
}

type AccountInfo struct {
	Balance struct {
		BCH BalanceCrypto `json:"bch"`
		BRL Amount        `json:"brl"`
		BTC BalanceCrypto `json:"btc"`
		ETH BalanceCrypto `json:"eth"`
		LTC BalanceCrypto `json:"ltc"`
		XRP BalanceCrypto `json:"xrp"`
	} `json:"balance"`
	WithdrawalLimits struct {
		BCH Amount `json:"bch"`
		BRL Amount `json:"brl"`
		BTC Amount `json:"btc"`
		ETH Amount `json:"eth"`
		LTC Amount `json:"ltc"`
		XRP Amount `json:"xrp"`
	} `json:"withdrawal_limits"`
}

type Operation struct {
	ID                int    `json:"operation_id"`
	Quantity          string `json:"quantity"`
	Price             string `json:"price"`
	FeeRate           string `json:"fee_rate"`
	ExecutedTimestamp string `json:"executed_timestamp"`
}

type Order struct {
	ID               int         `json:"order_id"`
	CoinPair         string      `json:"coin_pair"`
	Type             int         `json:"order_type"`
	Status           int         `json:"status"`
	HasFills         bool        `json:"has_fills"`
	Quantity         string      `json:"quantity"`
	LimitPrice       string      `json:"limit_price"`
	ExecutedQuantity string      `json:"executed_quantity"`
	ExecutedPriceAvg string      `json:"executed_price_avg"`
	Fee              string      `json:"fee"`
	CreatedTimestamp string      `json:"created_timestamp"`
	UpdatedTimestamp string      `json:"updated_timestamp"`
	Operations       []Operation `json:"operations"`
}

type OrderInfo struct {
	OrderID    int    `json:"order_id"`
	Quantity   string `json:"quantity"`
	LimitPrice string `json:"limit_price"`
	IsOwner    bool   `json:"is_owner"`
}

type Orderbook struct {
	Bids          []OrderInfo `json:"bids"`
	Asks          []OrderInfo `json:"asks"`
	LatestOrderID int         `json:"latest_order_id"`
}

type Withdrawal struct {
	ID               int    `json:"id"`
	Coin             string `json:"coin"`
	Quantity         string `json:"quantity"`
	NetQuantity      string `json:"net_quantity,omitempty"`
	Fee              string `json:"fee"`
	Account          string `json:"account,omitempty"`
	Address          string `json:"address,omitempty"`
	Status           int    `json:"status"`
	Tx               string `json:"tx,omitempty"`
	DestinationTag   int    `json:"destination_tag,omitempty"`
	CreatedTimestamp string `json:"created_timestamp"`
	UpdatedTimestamp string `json:"updated_timestamp"`
}
