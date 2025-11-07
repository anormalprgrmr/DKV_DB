package dal

import (
	"encoding/binary"
	"sync"
)

type meta struct {
	FreeListPage uint64
	Root         uint64
}

var (
	MetaInstance *meta
	MetaOnce     sync.Once
)

func GetMeta() *meta {
	MetaOnce.Do(func() {
		MetaInstance = &meta{}
	})

	return MetaInstance
}

func (m *meta) Serialize(buf []byte) {
	pos := 0

	binary.LittleEndian.PutUint64(buf[pos:], uint64(m.Root))
	pos += PageNumSize

	binary.LittleEndian.PutUint64(buf[pos:], uint64(m.FreeListPage))
	pos += PageNumSize
}

func (m *meta) Deserialize(buf []byte) {
	pos := 0

	m.Root = uint64(binary.LittleEndian.Uint64(buf[pos:]))
	pos += PageNumSize

	m.FreeListPage = uint64(binary.LittleEndian.Uint64(buf[pos:]))
	pos += PageNumSize
}
