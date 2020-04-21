package tapi

import (
	"encoding/json"
)

type listSystemMessagesResponse struct {
	Messages []SystemMessage `json:"messages"`
}

type listOrdersResponse struct {
	Orders []Order `json:"orders"`
}

type listOrderbookResponse struct {
	Orderbook `json:"orderbook"`
}

type orderResponse struct {
	Order `json:"order"`
}

type withdrawalResponse struct {
	Withdrawal `json:"withdrawal"`
}

func unmarshalSysMsgs(data json.RawMessage) ([]SystemMessage, error) {
	msgsResp := listSystemMessagesResponse{}
	if err := json.Unmarshal(data, &msgsResp); err != nil {
		return nil, err
	}
	return msgsResp.Messages, nil
}

func unmarshalListOrders(data json.RawMessage) ([]Order, error) {
	listOrders := listOrdersResponse{}
	if err := json.Unmarshal(data, &listOrders); err != nil {
		return nil, err
	}
	return listOrders.Orders, nil
}

func unmarshalOrderbook(data json.RawMessage) (*Orderbook, error) {
	orderbook := listOrderbookResponse{}
	if err := json.Unmarshal(data, &orderbook); err != nil {
		return nil, err
	}
	return &orderbook.Orderbook, nil
}

func unmarshalOrder(data json.RawMessage) (*Order, error) {
	order := orderResponse{}
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order.Order, nil
}

func unmarshalWithdrawal(data json.RawMessage) (*Withdrawal, error) {
	w := withdrawalResponse{}
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, err
	}
	return &w.Withdrawal, nil
}
