package template

import "bytes"

type bufferPool struct {
	c chan *bytes.Buffer
}

func newBufferPool(size int) (bp *bufferPool) {
	return &bufferPool{
		c: make(chan *bytes.Buffer, size),
	}
}

func (bp *bufferPool) get() (b *bytes.Buffer) {
	select {
	case b = <-bp.c:
		//
	default:
		b = bytes.NewBuffer([]byte{})
	}
	return
}

func (bp *bufferPool) put(b *bytes.Buffer) {
	b.Reset()
	select {
	case bp.c <- b:
	default:
	}
}
