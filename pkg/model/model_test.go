package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse__Empty(t *testing.T) {
	// empty text - unparsable
	_, err := ParseMsg([]byte(``))
	require.Equal(t, `failed to parse message`, err.Error())
}

func TestParse(t *testing.T) {
	// error message
	res, err := ParseMsg([]byte(`{"type":"error","message":"Failed to subscribe","reason":"ETH-USD is not a valid product"}`))
	require.Nil(t, err)
	require.Equal(t, ErrorMsg{Type: "error", Message: "Failed to subscribe", Reason: "ETH-USD is not a valid product"}, res)

	// subscription message
	res, err = ParseMsg([]byte(`{"type":"subscriptions","channels":[{"name":"level2","product_ids":["ETH-BTC","BTC-USD"]}]}`))
	require.Nil(t, err)
	require.Equal(t, SubscriptionsMsg{Type: "subscriptions", Channels: []Channel{{Name: "level2", ProductIDs: []string{"ETH-BTC", "BTC-USD"}}}}, res)

	// update message
	res, err = ParseMsg([]byte(`{"type":"l2update","product_id":"BTC-USD","changes":[["sell","45632.17","522.05252813"]],"time":"2022-01-17T02:57:26.530613Z"}`))
	require.Nil(t, err)
	require.Equal(t, TypedUpdateMsg{ProductID: "BTC-USD", Changes: []UpdateChange{{Price: 45632.17, Qty: 522.05252813}}}, res)

}
