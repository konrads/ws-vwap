package buffer

import "errors"

var (
	ErrZeroCapacity  = errors.New("assigned 0 capacity")
	ErrArrayTooSmall = errors.New("parameter array smaller than current size")
)

type PriceVol struct {
	Price float64
	Vol   float64
}

type RingBuffer interface {
	Push(x PriceVol)
	FoldL(acc interface{}, f func(PriceVol, interface{} /* accumulator*/) interface{} /*final accumulator*/) (interface{}, error)
	Read() ([]PriceVol, error)
}

type RingBufferImpl struct {
	w_ind    uint
	size     uint
	capacity uint
	buffer   []PriceVol
}

func NewRingBuffer(capacity uint) (*RingBufferImpl, error) {
	if capacity == 0 {
		return nil, ErrZeroCapacity
	}
	return &RingBufferImpl{
		capacity: capacity,
		buffer:   make([]PriceVol, capacity),
	}, nil
}

func (b *RingBufferImpl) Push(x PriceVol) {
	// Note: will overwrite even regardless of whether read prior or not
	b.buffer[b.w_ind] = x
	b.w_ind += 1
	if b.size != 0 && b.w_ind%b.capacity == 0 {
		b.w_ind = 0
	}
	if b.size < b.capacity {
		b.size += 1
	}
}

func (b *RingBufferImpl) FoldL(acc interface{}, f func(PriceVol, interface{} /* accumulator*/) interface{} /*final accumulator*/) (interface{}, error) {
	if b.size == 0 {
		return nil, ErrArrayTooSmall
	}
	var prefixArr []PriceVol
	suffixArr := []PriceVol{}
	if b.size < b.capacity {
		// not yet filled to capacity, read from the start
		prefixArr = b.buffer[:b.size]
	} else {
		// copy array suffix first
		prefixArr = b.buffer[b.w_ind:]
		// copy array prefix, if required
		if b.w_ind > 0 {
			suffixArr = b.buffer[:b.w_ind]
		}
	}
	for _, x := range prefixArr {
		acc = f(x, acc)
	}
	for _, x := range suffixArr {
		acc = f(x, acc)
	}
	return acc, nil
}

// Performs a read as a (slow) implementation of FoldL
func (b *RingBufferImpl) Read() ([]PriceVol, error) {
	initAcc := []PriceVol{}
	readF := func(x PriceVol, acc interface{}) interface{} {
		soFar := acc.([]PriceVol)
		soFar = append(soFar, x)
		return soFar
	}
	res, err := b.FoldL(initAcc, readF)
	if err != nil {
		return []PriceVol{}, err
	}
	return res.([]PriceVol), err
}
