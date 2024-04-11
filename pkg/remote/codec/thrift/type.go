package thrift

import (
	"github.com/apache/thrift/lib/go/thrift"
)

type Type struct {
	TypeID  thrift.TType
	KeyType *Type
	ValType *Type
	Fields  map[int16]*Type
}

func (t *Type) Equal(src *Type) bool {
	if t == nil && src == nil {
		return true
	}
	if t == nil || src == nil {
		return false
	}
	if t.TypeID != src.TypeID {
		return false
	}
	if !t.KeyType.Equal(src.KeyType) {
		return false
	}
	if !t.ValType.Equal(src.KeyType) {
		return false
	}
	if len(t.Fields) != len(src.Fields) {
		return false
	}
	for k, v := range t.Fields {
		if !v.Equal(src.Fields[k]) {
			return false
		}
	}
	return true
}

func (t *Type) Assignable(src *Type) bool {
	if t == nil || src == nil {
		return true
	}
	if t.TypeID != src.TypeID {
		return false
	}
	if !t.KeyType.Assignable(src.KeyType) {
		return false
	}
	if !t.ValType.Assignable(src.ValType) {
		return false
	}
	for k, v := range t.Fields {
		if srcField, ok := src.Fields[k]; ok {
			if !v.Assignable(srcField) {
				return false
			}
		}
	}
	return true
}

type Types []*Type

func (ts Types) Conflict() bool {
	if len(ts) >= 2 {
		for i := 1; i < len(ts); i++ {
			if !ts[0].Assignable(ts[i]) {
				return true
			}
		}
	}
	return false
}
