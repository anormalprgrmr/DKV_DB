package dal

import (
	"encoding/binary"
	"sync"
)

type freeList struct {
	maxPage   uint64
	freePages []uint64
}

var (
	freeListInstance *freeList
	freeListOnce     sync.Once
)

func GetFreeList() *freeList {

	freeListOnce.Do(func() {
		freeListInstance = &freeList{
			maxPage:   0,
			freePages: []uint64{},
		}
	})
	return freeListInstance
}

func (fr *freeList) GetNextPage() uint64 {
	if len(fr.freePages) > 0 {
		nextFreePageNum := fr.freePages[len(fr.freePages)-1]
		fr.freePages = fr.freePages[:len(fr.freePages)-1]
		return nextFreePageNum
	}
	fr.maxPage += 1
	return fr.maxPage
}

func (fr *freeList) ReleasePage(pageNumber uint64) {
	fr.freePages = append(fr.freePages, pageNumber)
}

func (fr *freeList) Serialize(buf []byte) []byte {
	pos := 0

	binary.LittleEndian.PutUint16(buf[pos:], uint16(fr.maxPage))
	pos += 2

	binary.LittleEndian.PutUint16(buf[pos:], uint16(len(fr.freePages)))
	pos += 2

	for _, page := range fr.freePages {
		binary.LittleEndian.PutUint64(buf[pos:], page)
		pos += PageNumSize
	}

	return buf
}

func (fr *freeList) Deserialize(buf []byte) {
	pos := 0

	fr.maxPage = uint64(binary.LittleEndian.Uint16(buf[pos:]))
	pos += 2

	freePagesLength := binary.LittleEndian.Uint16(buf[pos:])
	pos += 2

	for i := 0; i < int(freePagesLength); i++ {
		fr.freePages = append(fr.freePages, binary.LittleEndian.Uint64(buf[pos:]))
		pos += PageNumSize
	}
}
