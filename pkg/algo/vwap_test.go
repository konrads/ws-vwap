package algo

import (
	"testing"

	"github.com/konrads/ws-vwap/pkg/buffer"
	"github.com/stretchr/testify/require"
)

func TestVWAP__Empty(t *testing.T) {
	b, err := buffer.NewRingBuffer(1)
	require.NoError(t, err)

	_, err = VWAP(b)
	require.Equal(t, err, buffer.ErrArrayTooSmall)
}

func TestVWAP__Circular(t *testing.T) {
	b, err := buffer.NewRingBuffer(3)
	require.NoError(t, err)

	b.Push(buffer.PriceVol{1, 1})
	vwap, err := VWAP(b)
	require.NoError(t, err)
	require.Equal(t, 1.0, vwap)

	b.Push(buffer.PriceVol{2, 1})
	vwap, err = VWAP(b)
	require.NoError(t, err)
	require.Equal(t, 3.0/2, vwap)

	b.Push(buffer.PriceVol{3, 2})
	vwap, err = VWAP(b)
	require.NoError(t, err)
	require.Equal(t, 9.0/4, vwap)

	b.Push(buffer.PriceVol{4, 3})
	vwap, err = VWAP(b)
	require.NoError(t, err)
	require.Equal(t, 20.0/6, vwap)

	b.Push(buffer.PriceVol{5, 4})
	vwap, err = VWAP(b)
	require.NoError(t, err)
	require.Equal(t, 38.0/9, vwap)
}
