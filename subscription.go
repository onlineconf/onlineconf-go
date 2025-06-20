package onlineconf

import (
	"bytes"
	"fmt"
	"hash"
	"log"

	"golang.org/x/crypto/blake2b"
)

// longer values are hashed, shorter values are stored directly
const maxCurrValLen = 128

// it's possible to subscribe to the path value itself and to its subtree as two different subscriptions
type subscriptionKey struct {
	path        string
	isRecursive bool
}

type subscription struct {
	channels map[chan<- struct{}]struct{}
	current  []byte // value including the type byte. nil/empty: value doesn't exist
	isHashed bool   // determined using maxCurrValLen. subtree subscriptions are hashed always
}

var rootKey = subscriptionKey{
	path:        "/",
	isRecursive: true,
}

func (m *Module) subscribeChan(path string, isRecursive bool, ch chan<- struct{}) error {
	path = cleanPath(path)
	key := subscriptionKey{
		path:        path,
		isRecursive: isRecursive,
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	sub, ok := m.subscriptions[key]
	if !ok {
		sub = subscription{
			channels: map[chan<- struct{}]struct{}{ch: struct{}{}},
		}

		if key != rootKey {
			current, isHashed, err := m.getSubscr(key)
			if err != nil {
				return err
			}

			sub.current = current
			sub.isHashed = isHashed
		}

		if m.subscriptions == nil {
			m.subscriptions = map[subscriptionKey]subscription{key: sub}
		} else {
			m.subscriptions[key] = sub
		}

		return nil
	}

	sub.channels[ch] = struct{}{}

	return nil
}

// SubscribeChan creates a subscription for the specified path.
//
// If the path's value is changed/deleted, a notification (struct{}{} value) is sent to the specified channel.
// If the channel is already closed before a notification is sent, the subscription for this channel
// is deleted. If the channel is busy (over the capacity) during a notification, no blocking occurs.
//
// It's possible to make several subscriptions to the same path.
// If the channel is already subscribed to the path, nothing happens.
//
// It's possible to subscribe to a non-existing path to be notified of its creation.
// When the path's value is deleted, no unsubscription occurs.
func (m *Module) SubscribeChan(path string, ch chan<- struct{}) error {
	return m.subscribeChan(path, false, ch)
}

// Subscribe makes a channel with a capacity of 1 and calls [Module.SubscribeChan].
func (m *Module) Subscribe(path string) (chan struct{}, error) {
	ch := make(chan struct{}, 1)
	return ch, m.subscribeChan(path, false, ch)
}

// SubscribeChanSubtree creates a subscription for the specified path itself and all descending paths.
// Creation or deletion of any empty value is considered a change too.
//
// Subscribe* and Subscribe*Subtree are different subscriptions and may use different channels.
//
// A subscription to the root path "/" (or "") is a special case - a notification will be sent
// always when the underlying database file is updated, even if there's no value is really changed/created/deleted.
//
// Path separators other than a slash aren't supported by this method.
// `child_lists` OnlineConf feature is required for subtree notifications.
//
// See [Module.SubscribeChan] for other details.
func (m *Module) SubscribeChanSubtree(path string, ch chan<- struct{}) error {
	return m.subscribeChan(path, true, ch)
}

// SubscribeSubtree makes a channel with a capacity of 1 and calls [Module.SubscribeChanSubtree].
func (m *Module) SubscribeSubtree(path string) (chan struct{}, error) {
	ch := make(chan struct{}, 1)
	return ch, m.subscribeChan(path, true, ch)
}

// UnsubscribeChan removes the subscription made by [Module.SubscribeChan] or [Module.Subscribe]
// for the path and the channel specified.
// The channel is closed after removing the subscription.
// If the channel is already closed, no panic occurs.
// If the channel isn't subscribed to the specified path, the channel is closed too.
func (m *Module) UnsubscribeChan(path string, ch chan<- struct{}) {
	m.unsubscribeChan(path, false, ch)
	safeClose(ch)
}

// UnsubscribeChanSubtree removes the subscription made by [Module.SubscribeChanSubtree] or [Module.SubscribeSubtree]
// for the path and the channel specified.
//
// See [Module.UnsubscribeChan] for a description of closing the channel.
func (m *Module) UnsubscribeChanSubtree(path string, ch chan<- struct{}) {
	m.unsubscribeChan(path, true, ch)
	safeClose(ch)
}

func (m *Module) unsubscribeChan(path string, isRecursive bool, ch chan<- struct{}) {
	path = cleanPath(path)
	key := subscriptionKey{
		path:        path,
		isRecursive: isRecursive,
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	sub, ok := m.subscriptions[key]
	if !ok {
		return
	}

	delete(sub.channels, ch)

	if len(sub.channels) == 0 {
		delete(m.subscriptions, key)
	}
}

// Unsubscribe removes the subscription made by [Module.SubscribeChan] or [Module.Subscribe]
// for the specified path. All subscribed channels are closed.
// If any of these channels are already closed, no panic occurs.
func (m *Module) Unsubscribe(path string) {
	for ch := range m.unsubscribe(path, false) {
		safeClose(ch)
	}
}

// UnsubscribeSubtree removes the subscription made by [Module.SubscribeChanSubtree] or [Module.SubscribeSubtree]
// for the specified path. All subscribed channels are closed.
// If any of these channels are already closed, no panic occurs.
func (m *Module) UnsubscribeSubtree(path string) {
	for ch := range m.unsubscribe(path, true) {
		safeClose(ch)
	}
}

func (m *Module) unsubscribe(path string, isRecursive bool) map[chan<- struct{}]struct{} {
	path = cleanPath(path)
	key := subscriptionKey{
		path:        path,
		isRecursive: isRecursive,
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	sub, ok := m.subscriptions[key]
	if !ok {
		return nil
	}

	delete(m.subscriptions, key)

	return sub.channels
}

func (m *Module) processSubscriptions() {
	// m.mutex is already write-locked since we are in reopen()
	for key, sub := range m.subscriptions {
		var (
			current  []byte
			isHashed bool
			err      error
		)

		isRootKey := key == rootKey

		if !isRootKey { // a special hack - the root path is treated as changed when CDB is changed
			current, isHashed, err = m.getSubscr(key)
			if err != nil {
				// may not happen during the initial module open because m.subscriptions is empty yet.
				// just log it and continue processing.
				// TODO CDB may be seriously broken, is it worth continuing?
				log.Print(err)
				continue
			}

			if sub.isHashed == isHashed && bytes.Equal(sub.current, current) {
				continue
			}
		}

		for ch := range sub.channels {
			if !notify(ch) {
				delete(sub.channels, ch)
			}
		}

		if len(sub.channels) == 0 {
			delete(m.subscriptions, key)
		} else if !isRootKey {
			sub.current = current
			sub.isHashed = isHashed
			m.subscriptions[key] = sub
		}
	}
}

// getSubscr returns raw value bytes (including the type byte) or it's blake2b-256 hash
// if it's longer than [maxCurrValLen] bytes.
//
// If isRecursive is true, the value itself _and_ all descending subtree values are hashed recursively,
// including any empty values (since an empty value is represented as a single-byte string "s").
// In recursive mode, hashing is always used without performing a length check.
func (m *Module) getSubscr(key subscriptionKey) ([]byte, bool, error) {
	if key.isRecursive {
		h, err := blake2b.New256(nil)
		if err != nil {
			return nil, false, fmt.Errorf("blake2b.New256: %w", err)
		}

		err = m.getRecursive(key.path, h)
		if err != nil {
			return nil, false, err
		}

		return h.Sum(nil), true, nil
	}

	data, err := m.getRaw(key.path)
	if len(data) == 0 {
		return nil, false, err
	}

	if len(data) <= maxCurrValLen {
		return data, false, nil
	}

	sum := blake2b.Sum256(data)

	return sum[:], true, nil
}

func (m *Module) getRecursive(path string, h hash.Hash) error {
	data, err := m.getRaw(path)
	if err != nil {
		return err
	}

	if len(data) != 0 { // empty values have a length of 1
		_, err := h.Write(data)
		if err != nil {
			return fmt.Errorf("getRecursive: error hashing path %s: %w", path, err)
		}
	}

	subtree := path + "/"

	children, err := m.getStringsRaw(subtree)
	if err != nil {
		return err
	}

	for _, child := range children {
		if err := m.getRecursive(subtree+child, h); err != nil {
			return err
		}
	}

	return nil
}

// safeClose closes the channel or does nothing if it's already closed.
func safeClose(ch chan<- struct{}) {
	defer func() {
		_ = recover()
	}()

	close(ch)
}

// notify returns false if the channel is closed,
// or true if the notification is sent or the channel is busy.
func notify(ch chan<- struct{}) (isNotified bool) {
	defer func() {
		isNotified = recover() == nil
	}()

	select {
	case ch <- struct{}{}:
	default: // the channel is busy - there are pending notification(s) so it's surely "isNotified"
	}

	return true
}
