package main

import (
	"reflect"
	"sync"
)

type Map[K comparable, V any] struct {
	mu     sync.RWMutex
	_map   map[K]V
	length int
}

func (m *Map[K, V]) copy(from *Map[K, V]) *Map[K, V] {
	from.forEach(func(key K, value V) {
		m.set(key, value)
	})
	return m
}

func (m *Map[K, V]) delete(key K) {
	delete(m._map, key)
}

func (m *Map[K, V]) has(key K) bool {
	for k := range m._map {
		if reflect.DeepEqual(k, key) {
			return true
		}
	}
	return false
}

func (m *Map[K, V]) get(key K) V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k := range m._map {
		if reflect.DeepEqual(k, key) {
			return m._map[k]
		}
	}
	return m._map[key]
}

func (m *Map[K, V]) set(key K, value V) *Map[K, V] {
	m.mu.Lock()
	defer m.mu.Unlock()
	m._map[key] = value
	m.length = len(m._map)
	return m
}

func MapEntries[K, V comparable](_map *Map[K, V]) [][]any {
	var slice [][]any
	for k, v := range _map._map {
		slice = append(slice, []any{k, v})
	}
	return slice
}

// type callback[K, V comparable, Obj comparable] func(key K, value V, obj Obj)
// func (m *Map[K, V]) forEach(callback callback[K, V, *Map[K, V]]) {
// 	for key, value := range m._map {
// 		callback(key, value, m)
// 	}
// }

type callback[K comparable, V any] func(key K, value V)

func (m *Map[K, V]) forEach(callback callback[K, V]) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for key, value := range m._map {
		callback(key, value)
	}
}

func NewMap[K comparable, V comparable]() *Map[K, V] {
	return &Map[K, V]{_map: map[K]V{}, length: 0}
}
