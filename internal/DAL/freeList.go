package dal

import (
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
