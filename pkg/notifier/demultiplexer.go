package notifier

import "sync"

type Notifier[T any] struct {
	receiverChannels map[string]chan T
	mu               sync.Mutex
}

func NewNotifier[T any]() *Notifier[T] {
	return &Notifier[T]{
		receiverChannels: map[string]chan T{},
	}
}

func (d *Notifier[T]) AddSubscriber(id string) chan T {
	d.mu.Lock()
	defer d.mu.Unlock()

	ch := make(chan T, 10)
	d.receiverChannels[id] = ch

	return ch
}

func (d *Notifier[T]) RemoveSubscriber(id string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Verify if channel exists
	_, exists := d.receiverChannels[id]
	if !exists {
		return
	}

	// Close the channel
	close(d.receiverChannels[id])

	// Remove the channel from the map
	delete(d.receiverChannels, id)
}

func (d *Notifier[T]) Broadcast(msg T) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Broadcast the message to all subscribers in a separate goroutine
	go func() {
		for _, ch := range d.receiverChannels {
			ch <- msg
		}
	}()

}

func (d *Notifier[T]) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, ch := range d.receiverChannels {
		close(ch)
	}
}
