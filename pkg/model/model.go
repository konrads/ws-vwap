package model

import (
	"encoding/json"
	"errors"
	"strconv"
)

type SubscribeMsg struct {
	Type       string   `json:"type"`
	ProductIDs []string `json:"product_ids"`
	Channels   []string `json:"channels"`
}

type Channel struct {
	Name       string   `json:"name"`
	ProductIDs []string `json:"product_ids"`
}

type SubscriptionsMsg struct {
	Type     string    `json:"type"`
	Channels []Channel `json:"channels"`
}

func (m *SubscriptionsMsg) IsValid() bool {
	return m.Type == "subscriptions" && len(m.Channels) > 0
}

type ErrorMsg struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
}

func (m *ErrorMsg) IsValid() bool {
	return m.Type == "error" && m.Message != ""
}

type SnapshotMsg struct {
	Type      string     `json:"type"`
	ProductID string     `json:"product_id"`
	Asks      [][]string `json:"asks"`
	Bids      [][]string `json:"bids"`
}

func (m *SnapshotMsg) IsValid() bool {
	return m.Type == "snapshot" && m.ProductID != ""
}

type UpdateMsg struct {
	Type      string     `json:"type"`
	ProductID string     `json:"product_id"`
	Changes   [][]string `json:"changes"`
}

func (m *UpdateMsg) IsValid() bool {
	return m.Type == "l2update" && m.ProductID != ""
}

type UpdateChange struct {
	Price float64
	Qty   float64
}

type TypedUpdateMsg struct {
	ProductID string
	Changes   []UpdateChange
}

var (
	ErrFailedToParse = errors.New("failed to parse message")
)

func ToTypedUpdateMsg(msg UpdateMsg) (*TypedUpdateMsg, error) {
	changes := []UpdateChange{}
	for _, change := range msg.Changes {
		price, err := strconv.ParseFloat(change[1], 64)
		if err != nil {
			return nil, err
		}
		qty, err := strconv.ParseFloat(change[2], 64)
		if err != nil {
			return nil, err
		}
		changes = append(changes, UpdateChange{Price: price, Qty: qty})
	}

	return &TypedUpdateMsg{ProductID: msg.ProductID, Changes: changes}, nil
}

func ParseMsg(bs []byte) (interface{}, error) {
	// processing from most expected to least
	var updateMsg UpdateMsg
	if err := json.Unmarshal(bs, &updateMsg); err == nil && updateMsg.IsValid() {
		typedUpdateMsg, err := ToTypedUpdateMsg(updateMsg)
		if err != nil {
			return nil, err
		}
		return *typedUpdateMsg, nil
	}

	var errorMsg ErrorMsg
	if err := json.Unmarshal(bs, &errorMsg); err == nil && errorMsg.IsValid() {
		return errorMsg, nil
	}

	var subscriptionsMsg SubscriptionsMsg
	if err := json.Unmarshal(bs, &subscriptionsMsg); err == nil && subscriptionsMsg.IsValid() {
		return subscriptionsMsg, nil
	}

	var snapshotMsg SnapshotMsg
	if err := json.Unmarshal(bs, &snapshotMsg); err == nil && snapshotMsg.IsValid() {
		return snapshotMsg, nil
	}

	return nil, ErrFailedToParse
}
