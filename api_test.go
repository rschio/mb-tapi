package tapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestMakeRequest(t *testing.T) {
	ok := func(w http.ResponseWriter, r *http.Request) {
		rsp := tReq(r)
		j := &Response{}
		if rsp == "" {
			j.StatusCode = 100
		} else {
			j.StatusCode = -1
			j.ErrorMessage = rsp
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(j)
	}
	notOk := func(w http.ResponseWriter, r *http.Request) {
		j := &Response{}
		j.StatusCode = 201
		j.ErrorMessage = "Valor de *TAPI-ID* inválido."
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(j)
	}

	s1 := httptest.NewServer(http.HandlerFunc(ok))
	c := NewClient(s1.URL, fakeID, fakeKey, nil)
	if _, err := c.MakeRequest(make(url.Values)); err != nil {
		t.Error(err)
	}
	s1.Close()

	s2 := httptest.NewServer(http.HandlerFunc(notOk))
	c = NewClient(s2.URL, fakeID, fakeKey, nil)
	_, err := c.MakeRequest(make(url.Values))
	if err == nil {
		t.Error("this function should return error")
	}
	if !errors.Is(err, &Error{Code: 201}) {
		t.Error("failed to match error")
	}
	s2.Close()
}

func tReq(r *http.Request) string {
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		return "invalid Content-Type"
	}
	if r.Header.Get("TAPI-ID") != fakeID {
		return "invalid TAPI-ID"
	}
	if r.Header.Get("TAPI-MAC") == "" {
		return "empty TAPI-MAC"
	}
	if r.FormValue("tapi_nonce") == "" {
		return "empty nonce"
	}
	return ""
}

func TestListSystemMessages(t *testing.T) {
	tests := []struct {
		str string
		lvl string
	}{
		{"Info", "INFO"},
		{"INFO", "INFO"},
		{"WaRniNg", "WARNING"},
		{"error", "ERROR"},
		{"invalid", ""},
		{"", ""},
	}
	for _, tt := range tests {
		lvl := tt.lvl
		srv := httptest.NewServer(handler(tLstSysMsgs, jsonLstSysMsgs, lvl))
		c := NewClient(srv.URL, fakeID, fakeKey, nil)
		msgs, err := c.ListSystemMessages(tt.str)
		srv.Close()
		if err != nil {
			t.Error(err)
		}
		if len(msgs) != 2 {
			t.Errorf("got %d msgs, expected %d", len(msgs), 2)
		}
	}

}

func TestGetAccountInfo(t *testing.T) {
	srv := httptest.NewServer(handler(tGetAccInfo, jsonGetAccInfo))
	c := NewClient(srv.URL, fakeID, fakeKey, nil)
	defer srv.Close()
	accinfo, err := c.GetAccountInfo()
	if err != nil {
		t.Error(err)
	}
	cmpjson, err := json.Marshal(accinfo)
	if err != nil {
		t.Error(err)
	}
	if string(cmpjson) != accInfoPart {
		t.Error("different json")
	}
}

func tGetAccInfo(r *http.Request, strs ...string) string {
	if r.FormValue("tapi_method") != "get_account_info" {
		return "invalid method"
	}
	return ""
}

func TestGetOrder(t *testing.T) {
	tests := []struct {
		c1   Coin
		c2   Coin
		id   int
		strs []string
	}{
		{BRL, BTC, 10, []string{"coin_pair", "BRLBTC", "order_id", "10"}},
		{BRL, BCH, 42, []string{"coin_pair", "BRLBCH", "order_id", "42"}},
		{BRL, ETH, 2, []string{"coin_pair", "BRLETH", "order_id", "2"}},
	}
	for _, tt := range tests {
		srv := httptest.NewServer(handler(tGetOrder, jsonGetOrder, tt.strs...))
		c := NewClient(srv.URL, fakeID, fakeKey, nil)
		o, err := c.GetOrder(tt.c1, tt.c2, tt.id)
		if err != nil {
			t.Errorf("get order failed: %v", err)
		}
		cmpjson, err := json.Marshal(o)
		if err != nil {
			t.Error(err)
		}
		if string(cmpjson) != getOrder {
			t.Error("different json")
		}
		srv.Close()
	}
}

func tGetOrder(r *http.Request, strs ...string) string {
	if r.FormValue("tapi_method") != "get_order" {
		return "invalid method"
	}
	err := checkParams(r, strs...)
	if err != nil {
		return err.Error()
	}
	return ""
}

func TestListOrders(t *testing.T) {
	tnow := strconv.FormatInt(time.Now().Unix(), 10)
	tests := []struct {
		c1   Coin
		c2   Coin
		opts *ListOrdersOpts
		strs []string
	}{
		{BRL, BTC, nil, []string{"coin_pair", "BRLBTC"}},
		{BRL, BTC, &ListOrdersOpts{}, []string{
			"coin_pair", "BRLBTC", "order_type", "",
			"status_list", "", "has_fills", "",
			"from_id", "", "to_id", "", "from_timestamp", "",
			"to_timestamp", "",
		}},
		{BRL, ETH, &ListOrdersOpts{
			OrderType:     -1,
			StatusList:    [3]int{0, 0, 1},
			HasFills:      1,
			FromID:        500,
			ToID:          1000,
			FromTimestamp: tnow,
			ToTimestamp:   tnow,
		}, []string{
			"coin_pair", "BRLETH", "order_type", "2",
			"status_list", "[4]", "has_fills", "true",
			"from_id", "500", "to_id", "1000", "from_timestamp", tnow,
			"to_timestamp", tnow,
		}},
		{BRL, ETH, &ListOrdersOpts{
			OrderType:  1,
			StatusList: [3]int{5, -1, 0},
			HasFills:   -1,
		}, []string{
			"coin_pair", "BRLETH", "status_list", "[2,3]",
			"has_fills", "false",
		}},
	}
	for _, tt := range tests {
		srv := httptest.NewServer(handler(tListOrders, jsonListOrders, tt.strs...))
		c := NewClient(srv.URL, fakeID, fakeKey, nil)
		lo, err := c.ListOrders(tt.c1, tt.c2, tt.opts)
		if err != nil {
			t.Errorf("failed to list orders: %v", err)
		}
		cmpjson, err := json.Marshal(lo)
		if err != nil {
			t.Error(err)
		}
		if string(cmpjson) != listOrders {
			t.Error("different json")
		}
		srv.Close()
	}
}

func tListOrders(r *http.Request, strs ...string) string {
	if r.FormValue("tapi_method") != "list_orders" {
		return "invalid method"
	}
	err := checkParams(r, strs...)
	if err != nil {
		return err.Error()
	}
	return ""
}

func TestListOrderbook(t *testing.T) {
	tests := []struct {
		c1   Coin
		c2   Coin
		full bool
		strs []string
	}{
		{BRL, LTC, true, []string{"coin_pair", "BRLLTC", "full", "true"}},
		{BRL, XRP, false, []string{"coin_pair", "BRLXRP", "full", ""}},
	}
	for _, tt := range tests {
		srv := httptest.NewServer(handler(tListOrderbook, jsonListOrderbook, tt.strs...))
		c := NewClient(srv.URL, fakeID, fakeKey, nil)
		b, err := c.ListOrderbook(tt.c1, tt.c2, tt.full)
		if err != nil {
			t.Errorf("failed to list orders: %v", err)
		}
		cmpjson, err := json.Marshal(b)
		if err != nil {
			t.Error(err)
		}
		if string(cmpjson) != listOrderbook {
			t.Error("different json")
		}
		srv.Close()
	}
}

func tListOrderbook(r *http.Request, strs ...string) string {
	if r.FormValue("tapi_method") != "list_orderbook" {
		return "invalid method"
	}
	err := checkParams(r, strs...)
	if err != nil {
		return err.Error()
	}
	return ""
}

func TestPlaceOrder(t *testing.T) {
	tests := []struct {
		c1    Coin
		c2    Coin
		qt    string
		limit string
		f     string
		strs  []string
	}{
		{BRL, BTC, "0.05", "50", "buy", []string{
			"coin_pair", "BRLBTC", "quantity", "0.05", "limit_price", "50"}},
		{BRL, ETH, "0.9", "700", "buy", []string{
			"coin_pair", "BRLETH", "quantity", "0.9", "limit_price", "700"}},
		{BRL, BTC, "0.05", "50", "sell", []string{
			"coin_pair", "BRLBTC", "quantity", "0.05", "limit_price", "50"}},
		{BRL, ETH, "0.9", "700", "sell", []string{
			"coin_pair", "BRLETH", "quantity", "0.9", "limit_price", "700"}},
	}
	for _, tt := range tests {
		var fn1 fnParams
		var fn2 func(c1, c2 Coin, qt, limit string) (*Order, error)
		if tt.f == "buy" {
			fn1 = tPlaceBuyOrder
		} else {
			fn1 = tPlaceSellOrder
		}
		srv := httptest.NewServer(handler(fn1, jsonGetOrder, tt.strs...))
		c := NewClient(srv.URL, fakeID, fakeKey, nil)
		if tt.f == "buy" {
			fn2 = c.PlaceBuyOrder
		} else {
			fn2 = c.PlaceSellOrder
		}
		b, err := fn2(tt.c1, tt.c2, tt.qt, tt.limit)
		if err != nil {
			t.Errorf("failed to list orders: %v", err)
		}
		cmpjson, err := json.Marshal(b)
		if err != nil {
			t.Error(err)
		}
		if string(cmpjson) != getOrder {
			t.Error("different json")
		}
		srv.Close()
	}
}

func tPlaceBuyOrder(r *http.Request, strs ...string) string {
	if r.FormValue("tapi_method") != "place_buy_order" {
		return "invalid method"
	}
	err := checkParams(r, strs...)
	if err != nil {
		return err.Error()
	}
	return ""
}

func tPlaceSellOrder(r *http.Request, strs ...string) string {
	if r.FormValue("tapi_method") != "place_sell_order" {
		return "invalid method"
	}
	err := checkParams(r, strs...)
	if err != nil {
		return err.Error()
	}
	return ""
}

func TestPlaceMarketBuyOrder(t *testing.T) {
	tests := []struct {
		c1   Coin
		c2   Coin
		cost string
		strs []string
	}{
		{BRL, BTC, "10.08", []string{"coin_pair", "BRLBTC", "cost", "10.08"}},
		{BRL, ETH, "500.0", []string{"coin_pair", "BRLETH", "cost", "500.0"}},
	}
	for _, tt := range tests {
		srv := httptest.NewServer(handler(tPlaceMarketBuyOrder, jsonGetOrder, tt.strs...))
		c := NewClient(srv.URL, fakeID, fakeKey, nil)
		o, err := c.PlaceMarketBuyOrder(tt.c1, tt.c2, tt.cost)
		if err != nil {
			t.Errorf("failed to place market buy order: %v", err)
		}
		cmpjson, err := json.Marshal(o)
		if err != nil {
			t.Error(err)
		}
		if string(cmpjson) != getOrder {
			t.Error("different json")
		}
		srv.Close()
	}
}

func tPlaceMarketBuyOrder(r *http.Request, strs ...string) string {
	if r.FormValue("tapi_method") != "place_market_buy_order" {
		return "invalid method"
	}
	err := checkParams(r, strs...)
	if err != nil {
		return err.Error()
	}
	return ""
}

func TestPlaceMarketSellOrder(t *testing.T) {
	tests := []struct {
		c1   Coin
		c2   Coin
		qt   string
		strs []string
	}{
		{BRL, BTC, "0.001", []string{"coin_pair", "BRLBTC", "quantity", "0.001"}},
		{BRL, ETH, "0.01", []string{"coin_pair", "BRLETH", "quantity", "0.01"}},
	}
	for _, tt := range tests {
		srv := httptest.NewServer(handler(tPlaceMarketSellOrder, jsonGetOrder, tt.strs...))
		c := NewClient(srv.URL, fakeID, fakeKey, nil)
		o, err := c.PlaceMarketSellOrder(tt.c1, tt.c2, tt.qt)
		if err != nil {
			t.Errorf("failed to place market sell order: %v", err)
		}
		cmpjson, err := json.Marshal(o)
		if err != nil {
			t.Error(err)
		}
		if string(cmpjson) != getOrder {
			t.Error("different json")
		}
		srv.Close()
	}
}

func tPlaceMarketSellOrder(r *http.Request, strs ...string) string {
	if r.FormValue("tapi_method") != "place_market_sell_order" {
		return "invalid method"
	}
	err := checkParams(r, strs...)
	if err != nil {
		return err.Error()
	}
	return ""
}

func TestCancelOrder(t *testing.T) {
	tests := []struct {
		c1   Coin
		c2   Coin
		id   int
		strs []string
	}{
		{BRL, BTC, 987, []string{"coin_pair", "BRLBTC", "order_id", "987"}},
		{BRL, ETH, 1020, []string{"coin_pair", "BRLETH", "order_id", "1020"}},
	}
	for _, tt := range tests {
		srv := httptest.NewServer(handler(tCancelOrder, jsonGetOrder, tt.strs...))
		c := NewClient(srv.URL, fakeID, fakeKey, nil)
		o, err := c.CancelOrder(tt.c1, tt.c2, tt.id)
		if err != nil {
			t.Errorf("failed to cancel order: %v", err)
		}
		cmpjson, err := json.Marshal(o)
		if err != nil {
			t.Error(err)
		}
		if string(cmpjson) != getOrder {
			t.Error("different json")
		}
		srv.Close()
	}
}

func tCancelOrder(r *http.Request, strs ...string) string {
	if r.FormValue("tapi_method") != "cancel_order" {
		return "invalid method"
	}
	err := checkParams(r, strs...)
	if err != nil {
		return err.Error()
	}
	return ""
}

func TestGetWithdrawal(t *testing.T) {
	tests := []struct {
		coin Coin
		id   int
		strs []string
	}{
		{BRL, 42, []string{"coin", "BRL", "withdrawal_id", "42"}},
		{BTC, 10012, []string{"coin", "BTC", "withdrawal_id", "10012"}},
	}
	for _, tt := range tests {
		srv := httptest.NewServer(handler(tGetWithdrawal, jsonGetWithdrawal, tt.strs...))
		c := NewClient(srv.URL, fakeID, fakeKey, nil)
		w, err := c.GetWithdrawal(tt.coin, tt.id)
		if err != nil {
			t.Errorf("failed to get withdrawal: %v", err)
		}
		cmpjson, err := json.Marshal(w)
		if err != nil {
			t.Error(err)
		}
		if string(cmpjson) != getWithdrawal {
			t.Error("different json")
		}
		srv.Close()
	}
}

func tGetWithdrawal(r *http.Request, strs ...string) string {
	if r.FormValue("tapi_method") != "get_withdrawal" {
		return "invalid method"
	}
	err := checkParams(r, strs...)
	if err != nil {
		return err.Error()
	}
	return ""
}

func TestWithdrawlBRL(t *testing.T) {
	tests := []struct {
		desc   string
		qt     string
		accRef string
		strs   []string
	}{
		{"transfer it", "500.25", "001122", []string{
			"description", "transfer it",
			"quantity", "500.25",
			"account_ref", "001122",
			"coin", "BRL",
		}},
	}
	for _, tt := range tests {
		srv := httptest.NewServer(handler(tWithdrawCoin, jsonWithdrawCoin, tt.strs...))
		c := NewClient(srv.URL, fakeID, fakeKey, nil)
		w, err := c.WithdrawBRL(tt.desc, tt.qt, tt.accRef)
		if err != nil {
			t.Errorf("failed to withdraw coin: %v", err)
		}
		cmpjson, err := json.Marshal(w)
		if err != nil {
			t.Error(err)
		}
		if string(cmpjson) != withdrawCoin {
			t.Error("different json")
		}
		srv.Close()
	}
}

func TestWithdrawlCrypto(t *testing.T) {
	tests := []struct {
		coin Coin
		desc string
		i    *WithdrawInfo
		strs []string
	}{
		{BTC, "", &WithdrawInfo{
			Address:       "18d2ogsrMXsspcxzz3DgecePNdxcZUpaUX",
			Quantity:      "0.678",
			TxFee:         "0.0005",
			ViaBlockchain: true,
		}, []string{
			"coin", "BTC",
			"address", "18d2ogsrMXsspcxzz3DgecePNdxcZUpaUX",
			"description", "",
			"quantity", "0.678",
			"account_ref", "",
			"tx_fee", "0.0005",
			"tx_aggregate", "",
			"via_blockchain", "true",
			"destination_tag", "",
		}},
		{XRP, "hello", &WithdrawInfo{
			Address:        "18d2ogsrMXsspcxzz3DgecePNdxcZUpaUY",
			Quantity:       "0.9",
			TxFee:          "0.08",
			TxNotAggregate: true,
			DestinationTag: 20,
		}, []string{
			"coin", "XRP",
			"address", "18d2ogsrMXsspcxzz3DgecePNdxcZUpaUY",
			"description", "hello",
			"quantity", "0.9",
			"account_ref", "",
			"tx_fee", "0.08",
			"tx_aggregate", "false",
			"via_blockchain", "",
			"destination_tag", "20",
		}},
	}
	for _, tt := range tests {
		srv := httptest.NewServer(handler(tWithdrawCoin, jsonWithdrawCoin, tt.strs...))
		c := NewClient(srv.URL, fakeID, fakeKey, nil)
		w, err := c.WithdrawCrypto(tt.coin, tt.desc, tt.i)
		if err != nil {
			t.Errorf("failed to withdraw coin: %v", err)
		}
		cmpjson, err := json.Marshal(w)
		if err != nil {
			t.Error(err)
		}
		if string(cmpjson) != withdrawCoin {
			t.Error("different json")
		}
		srv.Close()
	}
}

func tWithdrawCoin(r *http.Request, strs ...string) string {
	if r.FormValue("tapi_method") != "withdraw_coin" {
		return "invalid method"
	}
	err := checkParams(r, strs...)
	if err != nil {
		return err.Error()
	}
	return ""
}

type fnParams func(r *http.Request, strs ...string) string

func handler(fn fnParams, payload []byte, strs ...string) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		j := &Response{StatusCode: 100}
		errstr := fn(r, strs...)
		if errstr != "" {
			j.StatusCode = -1
			j.ErrorMessage = errstr
			json.NewEncoder(w).Encode(j)
			return
		}
		w.Write(payload)
	}
	return http.HandlerFunc(f)
}

func tLstSysMsgs(r *http.Request, strs ...string) string {
	if r.FormValue("tapi_method") != "list_system_messages" {
		return "invalid method"
	}
	err := checkParams(r, "level", strs[0])
	if err != nil {
		return err.Error()
	}
	return ""
}

func checkParams(r *http.Request, pairs ...string) error {
	if len(pairs)%2 != 0 {
		return fmt.Errorf("invalid pair of name vals")
	}
	for i := 0; i < len(pairs); i += 2 {
		name := pairs[i]
		val := pairs[i+1]
		if got := r.FormValue(name); got != val {
			return fmt.Errorf("expected %s, got %s", val, got)
		}
	}
	return nil
}

var jsonLstSysMsgs = []byte(`
{
    "response_data": {
        "messages": [
            {
                "msg_date": "1453827748",
                "level": "INFO",
                "event_code": 7000,
                "msg_content": "Manutenção programada para 2015-DEZ-25, janela de até 2 horas, a partir das 14hs. O sistema estará indisponível durante esse período."
            },
            {
                "msg_date": "1453827748",
                "level": "INFO",
                "event_code": 7002,
                "msg_content": "Novo filtro de datas disponível para o método *list_orders*. Veja mais detalhes em https://www.mercadobitcoin.com.br/trade-api/."
            }
        ]
    },
    "status_code": 100,
    "server_unix_timestamp": "1453827748"
}`)

var (
	jsonGetAccInfo = []byte(`{"response_data":` + accInfoPart + `,"status_code":100,"server_unix_timestamp":"1453831028"}`)
	accInfoPart    = `{"balance":{"bch":{"available":"5.00000000","total":"6.00000000","amount_open_orders":1},"brl":{"available":"3000.00000","total":"4900.00000"},"btc":{"available":"10.00000000","total":"11.00000000","amount_open_orders":3},"eth":{"available":"490.00000000","total":"500.00000000","amount_open_orders":1},"ltc":{"available":"500.00000000","total":"500.00000000","amount_open_orders":0},"xrp":{"available":"105.00000000","total":"106.00000000","amount_open_orders":0}},"withdrawal_limits":{"bch":{"available":"2.00000000","total":"2.00000000"},"brl":{"available":"988.00","total":"20000.00"},"btc":{"available":"3.76600000","total":"5.00000000"},"eth":{"available":"210.00000000","total":"300.00000000"},"ltc":{"available":"500.00000000","total":"500.00000000"},"xrp":{"available":"100.00000000","total":"200.00000000"}}}`

	jsonGetOrder = []byte(`{"response_data":{"order":` + getOrder + `},"status_code":100,"server_unix_timestamp":"1453835329"}`)
	getOrder     = `{"order_id":3,"coin_pair":"BRLBTC","order_type":2,"status":4,"has_fills":true,"quantity":"1.00000000","limit_price":"900.00000","executed_quantity":"1.00000000","executed_price_avg":"900.00000","fee":"6.30000000","created_timestamp":"1453835329","updated_timestamp":"1453835329","operations":[{"operation_id":1,"quantity":"1.00000000","price":"900.00000","fee_rate":"0.70","executed_timestamp":"1453835329"}]}`

	jsonListOrders = []byte(`{"response_data":{"orders":` + listOrders + `},"status_code":100,"server_unix_timestamp":"1453838494"}`)
	listOrders     = `[{"order_id":1,"coin_pair":"BRLBTC","order_type":1,"status":2,"has_fills":false,"quantity":"1.00000000","limit_price":"1000.00000","executed_quantity":"0.00000000","executed_price_avg":"0.00000","fee":"0.00000000","created_timestamp":"1453838494","updated_timestamp":"1453838494","operations":[]},{"order_id":2,"coin_pair":"BRLBTC","order_type":2,"status":2,"has_fills":false,"quantity":"1.00000000","limit_price":"1100.00000","executed_quantity":"0.00000000","executed_price_avg":"0.00000","fee":"0.00000000","created_timestamp":"1453838494","updated_timestamp":"1453838494","operations":[]},{"order_id":3,"coin_pair":"BRLBTC","order_type":2,"status":4,"has_fills":true,"quantity":"1.00000000","limit_price":"900.00000","executed_quantity":"1.00000000","executed_price_avg":"900.00000","fee":"6.30000000","created_timestamp":"1453838494","updated_timestamp":"1453838494","operations":[{"operation_id":1,"quantity":"1.00000000","price":"900.00000","fee_rate":"0.70","executed_timestamp":"1453838494"}]},{"order_id":4,"coin_pair":"BRLBTC","order_type":1,"status":2,"has_fills":true,"quantity":"2.00000000","limit_price":"900.00000","executed_quantity":"1.00000000","executed_price_avg":"900.00000","fee":"0.00300000","created_timestamp":"1453838494","updated_timestamp":"1453838494","operations":[{"operation_id":1,"quantity":"1.00000000","price":"900.00000","fee_rate":"0.30","executed_timestamp":"1453838494"}]}]`

	jsonListOrderbook = []byte(`{"response_data":{"orderbook":` + listOrderbook + `},"status_code":100,"server_unix_timestamp":"1453833861"}`)
	listOrderbook     = `{"bids":[{"order_id":1,"quantity":"1.00000000","limit_price":"1000.00000","is_owner":true},{"order_id":4,"quantity":"1.00000000","limit_price":"900.00000","is_owner":false}],"asks":[{"order_id":2,"quantity":"1.00000000","limit_price":"1100.00000","is_owner":true}],"latest_order_id":4}`

	jsonGetWithdrawal = []byte(`{"response_data":{"withdrawal":` + getWithdrawal + `},"status_code":100,"server_unix_timestamp":"1453912131"}`)
	getWithdrawal     = `{"id":1,"coin":"BTC","quantity":"1.50000000","fee":"0.00050000","address":"1G38ybvfUyn96aJbKnzkifX2eEMH9N87ww","status":2,"tx":"d9893c57880953f044bcf0c9f31b923459a2fc54e82e8c8544645b96da37726f","created_timestamp":"1453912131","updated_timestamp":"1453912131"}`

	jsonWithdrawCoin = []byte(`{"response_data":{"withdrawal":` + withdrawCoin + `},"status_code":100,"server_unix_timestamp":"1453912088"}`)
	withdrawCoin     = `{"id":1,"coin":"BRL","quantity":"300.56","net_quantity":"291.68","fee":"8.88","account":"bco: 341, ag: 1111, cta: 23456-X","status":1,"created_timestamp":"1453912088","updated_timestamp":"1453912088"}`
)
