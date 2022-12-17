package komap

import "github.com/kamalshkeir/kutils/kchans"

type Map[K, V any] struct {
	root    *node[K, V]
	compare func(K, K) int
}

type node[K, V any] struct {
	k           K
	v           V
	left, right *node[K, V]
}

func New[K, V any](compare func(K, K) int) *Map[K, V] {
	return &Map[K, V]{compare: compare}
}

func (m *Map[K, V]) find(k K) **node[K, V] {
	pn := &m.root
	for *pn != nil {
		switch cmp := m.compare(k, (*pn).k); {
		case cmp < 0:
			pn = &(*pn).left
		case cmp > 0:
			pn = &(*pn).right
		default:
			return pn
		}
	}
	return pn
}

func (m *Map[K, V]) Insert(k K, v V) bool {
	pn := m.find(k)
	if *pn != nil {
		(*pn).v = v
		return false
	}
	*pn = &node[K, V]{k: k, v: v}
	return true
}

func (m *Map[K, V]) Find(k K) (V, bool) {
	pn := m.find(k)
	if *pn == nil {
		var zero V
		return zero, false
	}
	return (*pn).v, true
}

type keyValue[K, V any] struct {
	k K
	v V
}

func (m *Map[K, V]) InOrder() *Iterator[K, V] {
	type kv = keyValue[K, V]
	sender, receiver := kchans.Ranger[kv]()
	var f func(*node[K, V]) bool
	f = func(n *node[K, V]) bool {
		if n == nil {
			return true
		}
		return f(n.left) &&
			sender.Send(kv{n.k, n.v}) &&
			f(n.right)
	}
	go func() {
		f(m.root)
		sender.Close()
	}()
	return &Iterator[K, V]{receiver}
}

type Iterator[K, V any] struct {
	r *kchans.Receiver[keyValue[K, V]]
}

func (it *Iterator[K, V]) Next() (K, V, bool) {
	kv, ok := it.r.Next()
	return kv.k, kv.v, ok
}
