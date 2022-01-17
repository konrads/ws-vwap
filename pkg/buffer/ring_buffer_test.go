package buffer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRingBuffer__Errors(t *testing.T) {
	_, err := NewRingBuffer(0)
	require.Equal(t, err, ErrZeroCapacity)
	b, err := NewRingBuffer(1)
	require.NoError(t, err)

	_, err = b.Read()
	require.Equal(t, err, ErrArrayTooSmall)
}

func TestRingBuffer__Circular(t *testing.T) {
	b, err := NewRingBuffer(3)
	require.NoError(t, err)

	b.Push(PriceVol{1, 1})
	res, err := b.Read()
	require.NoError(t, err)
	require.Equal(t, []PriceVol{{1, 1}}, res)

	b.Push(PriceVol{2, 1})
	res, err = b.Read()
	require.NoError(t, err)
	require.Equal(t, []PriceVol{{1, 1}, {2, 1}}, res)

	b.Push(PriceVol{3, 2})
	res, err = b.Read()
	require.NoError(t, err)
	require.Equal(t, []PriceVol{{1, 1}, {2, 1}, {3, 2}}, res)

	b.Push(PriceVol{4, 3})
	res, err = b.Read()
	require.NoError(t, err)
	require.Equal(t, []PriceVol{{2, 1}, {3, 2}, {4, 3}}, res)

	b.Push(PriceVol{5, 4})
	res, err = b.Read()
	require.NoError(t, err)
	require.Equal(t, []PriceVol{{3, 2}, {4, 3}, {5, 4}}, res)
}
