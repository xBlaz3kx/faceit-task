package notifier

import (
	"sync"

	"go.uber.org/zap"
)

type Notifier[T any] struct {
	receiverChannels map[string]chan T
	mu               sync.Mutex
	logger           *zap.Logger
}

func NewNotifier[T any]() *Notifier[T] {
	return &Notifier[T]{
		receiverChannels: map[string]chan T{},
		logger:           zap.L().Named("notifier"),
	}
}

func (d *Notifier[T]) AddSubscriber(id string) chan T {
	d.logger.Debug("Adding a new subscriber", zap.String("id", id))

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

	d.logger.Debug("Removing a subscriber", zap.String("id", id))

	// Close the channel
	close(d.receiverChannels[id])

	// Remove the channel from the map
	delete(d.receiverChannels, id)
}

func (d *Notifier[T]) Broadcast(msg T) {
	d.logger.Debug("Broadcasting a message to subscribers", zap.Any("message", msg))

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
	d.logger.Debug("Closing the notifier")

	d.mu.Lock()
	defer d.mu.Unlock()

	for _, ch := range d.receiverChannels {
		close(ch)
	}
}
