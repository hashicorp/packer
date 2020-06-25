package external

import (
	"hash/crc32"
	"strconv"
)

func hashcode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	return -v
}

func hashString(v interface{}) int {
	return hashcode(v.(string))
}

func hashInt(v interface{}) int {
	return hashcode(strconv.Itoa(v.(int)))
}

// setIDFunc is the function to identify the unique id for the item of set
type setIDFunc func(interface{}) int

// set is a structure to distinct instance
type set struct {
	idFunc setIDFunc
	idMap  map[int]interface{}
}

// newSet will expected a list, reserving only one item with same id and return a set-collection
func newSet(idFunc setIDFunc, vL []interface{}) *set {
	s := &set{
		idMap:  make(map[int]interface{}, len(vL)),
		idFunc: idFunc,
	}

	for _, v := range vL {
		s.Add(v)
	}

	return s
}

func (s *set) Add(v interface{}) {
	id := s.idFunc(v)
	if _, ok := s.idMap[id]; !ok {
		s.idMap[id] = v
	}
}

func (s *set) Remove(v interface{}) {
	delete(s.idMap, s.idFunc(v))
}

func (s *set) List() []interface{} {
	vL := make([]interface{}, s.Len())

	var i int
	for _, v := range s.idMap {
		vL[i] = v
		i++
	}

	return vL
}

func (s *set) Len() int {
	return len(s.idMap)
}
