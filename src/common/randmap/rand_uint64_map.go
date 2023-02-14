package randmap

// todo
// import (
// 	"math/rand"
// 	"sync"
// )

// type element struct {
// 	content     interface{}
// 	slice_index int
// }

// type Rand_uint64_map struct {
// 	m sync.RWMutex

// 	// Where the objects you care about are stored.
// 	container map[uint64]element

// 	// A slice of the map keys used in the map above. We put them in a slice
// 	// so that we can get a random key by choosing a random index.
// 	keys []uint64

// 	// We store the index of each key, so that when we remove an item, we can
// 	// quickly remove it from the slice above.
// 	sliceKeyIndex map[uint64]int
// }

// func NewRandUint64Map() *Rand_uint64_map {
// 	return &Rand_uint64_map{
// 		container:     make(map[uint64]element),
// 		sliceKeyIndex: make(map[uint64]int),
// 	}
// }

// func (s *Rand_uint64_map) Set(key uint64, item interface{}) {
// 	s.m.Lock()
// 	defer s.m.Unlock()

// 	if old_ele, ok := s.container[key]; ok {
// 		//old exist already
// 		s.container[key] = element{item, old_ele.slice_index}
// 	} else {
// 		// add map key to slice of map keys
// 		s.keys = append(s.keys, key)
// 		// store the index of the map key
// 		index := len(s.keys) - 1
// 		// store object in map
// 		s.container[key] = element{item, index}
// 		s.sliceKeyIndex[key] = index
// 	}
// }

// func (s *Rand_uint64_map) Get(key uint64) interface{} {
// 	s.m.RLock()
// 	defer s.m.RUnlock()

// 	if ele, ok := s.container[key]; ok {
// 		return ele.content
// 	} else {
// 		return nil
// 	}

// }

// func (s *Rand_uint64_map) Remove(key uint64) {
// 	s.m.Lock()
// 	defer s.m.Unlock()

// 	// get index in key slice for key
// 	index, exists := s.sliceKeyIndex[key]
// 	if !exists {
// 		// item does not exist
// 		return
// 	}

// 	delete(s.sliceKeyIndex, key)

// 	wasLastIndex := len(s.keys)-1 == index

// 	// remove key from slice of keys
// 	s.keys[index] = s.keys[len(s.keys)-1]
// 	s.keys = s.keys[:len(s.keys)-1]

// 	// we just swapped the last element to another position.
// 	// so we need to update it's index (if it was not in last position)
// 	if !wasLastIndex {
// 		otherKey := s.keys[index]
// 		s.sliceKeyIndex[otherKey] = index
// 	}

// 	// remove object from map
// 	delete(s.container, key)
// }

// func (s *Rand_uint64_map) Random() interface{} {

// 	if s.Len() == 0 {
// 		return nil
// 	}

// 	s.m.RLock()
// 	defer s.m.RUnlock()

// 	randomIndex := rand.Intn(len(s.keys))
// 	key := s.keys[randomIndex]

// 	if ele, ok := s.container[key]; ok {
// 		return ele.content
// 	} else {
// 		return nil
// 	}

// }

// func (s *Rand_uint64_map) PopRandom() interface{} {

// 	if s.Len() == 0 {
// 		return nil
// 	}

// 	s.m.RLock()
// 	randomIndex := rand.Intn(len(s.keys))
// 	key := s.keys[randomIndex]

// 	item := s.container[key]
// 	s.m.RUnlock()

// 	s.Remove(key)

// 	return item.content
// }

// func (s *Rand_uint64_map) Len() int {
// 	s.m.RLock()
// 	defer s.m.RUnlock()

// 	return len(s.container)
// }
