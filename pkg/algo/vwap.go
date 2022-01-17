package algo

import "github.com/konrads/ws-vwap/pkg/buffer"

func VWAP(b buffer.RingBuffer) (float64, error) {
	type PriceVolVol struct {
		priceVol float64
		vol      float64
	}

	initAcc := PriceVolVol{}
	f := func(x buffer.PriceVol, acc interface{}) interface{} {
		accTyped := acc.(PriceVolVol)
		accTyped.priceVol += x.Price * x.Vol
		accTyped.vol += x.Vol
		return accTyped
	}
	res, err := b.FoldL(initAcc, f)
	if err != nil {
		return 0, err
	}
	resTyped := res.(PriceVolVol)
	return resTyped.priceVol / resTyped.vol, nil
}
