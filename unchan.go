package unchan

import (
	"math/rand"
	"runtime"
)

// channel is the general functionality of a go channel.
type channel interface {
	Send(interface{})
	Receive() interface{}
	Len() int
	Cap() int
}

// Unchan is an unordered channel.
// Values that are sent are received in any order.
type Unchan struct {
	ch          chan interface{}
	length      int // Must store as shuffling removes elements temporarily
	killshuffle chan struct{}
}

// New generates a new unordered channel of given length/capacity.
func New(l int) *Unchan {
	ch := make(chan interface{}, l)
	unchan := &Unchan{
		ch:          ch,
		killshuffle: make(chan struct{}),
	}
	go unchan.shuffle()
	runtime.SetFinalizer(unchan, func(c *Unchan) {
		finalize(c)
	})
	return unchan
}

// Send pushes a value down the channel.
func (c *Unchan) Send(v interface{}) {
	c.ch <- v
	c.length++
	return
}

// Receive pulls a value out of the channel.
func (c *Unchan) Receive() interface{} {
	v := <-c.ch
	c.length--
	return v
}

// Len is the current number of values in the channel.
func (c *Unchan) Len() int {
	return c.length
}

// Cap is the maximum number of values the channel can hold.
func (c *Unchan) Cap() int {
	return cap(c.ch)
}

// finalize is any task that must be run when an Unchan is garbage collected.
func finalize(c *Unchan) {
	c.killshuffle <- struct{}{}
}

// shuffle runs in the background to constantly reorder values in the channel.
func (c *Unchan) shuffle() {
	var v1, v2 interface{} // Preallocate all necessary space
	select {
	case <-c.killshuffle:
		return
	case v1 = <-c.ch: // Able to get first value
		select {
		case v2 = <-c.ch: // Able to get second value
			if rand.Intn(2) == 1 { // Shuffle order half the time
				c.Send(v1)
				c.Send(v2)
			} else {
				c.Send(v2)
				c.Send(v1)
			}
		default: // Unable to get second value, resend first value
			c.ch <- v1
		}
	}
}
