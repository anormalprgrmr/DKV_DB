package dal

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

type Options struct {
	PageSize       int
	MinFillPercent float32
	MaxFillPercent float32
}

var DefaultOptions = &Options{
	MinFillPercent: 0.5,
	MaxFillPercent: 0.95,
}

type DAL struct {
	minFillPercent float32
	maxFillPercent float32

	File     *os.File
	PageSize int

	*meta
	*freeList
}

var (
	dalInstance *DAL
	dalOnce     sync.Once
)

func GetDal(path string, option *Options) (*DAL, error) {
	var err error
	dalOnce.Do(func() {

		dalInstance = &DAL{
			meta:           GetMeta(),
			PageSize:       os.Getpagesize(),
			minFillPercent: option.MinFillPercent,
			maxFillPercent: option.MaxFillPercent,
		}

		if _, err = os.Stat(path); err == nil {
			fmt.Printf("we are in if")

			dalInstance.File, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
			if err != nil {
				_ = dalInstance.Close()
				return
			}

			dalInstance.meta, err = dalInstance.ReadMeta()
			if err != nil {
				return
			}
			fmt.Printf("%s ==> %d", "freePageNumber", dalInstance.meta.FreeListPage)

			freeList, err := dalInstance.ReadFreeList()
			if err != nil {
				return
			}

			dalInstance.freeList = freeList

		} else if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("we are in else if")
			dalInstance.File, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
			if err != nil {
				return
			}

			dalInstance.freeList = GetFreeList()
			dalInstance.FreeListPage = dalInstance.GetNextPage()
			_, err = dalInstance.WriteFreeList()
			if err != nil {
				return
			}

			_, err = dalInstance.WriteMeta(dalInstance.meta)
			if err != nil {
				return
			}

		} else {
			fmt.Printf("we are in else")
			return
		}

	})
	return dalInstance, err
}

func (d *DAL) Close() error {
	if d.File != nil {
		err := d.File.Close()
		if err != nil {
			return fmt.Errorf("couldn't close file : %s", err)
		}
	}
	return nil
}
func (d *DAL) WriteMeta(meta *meta) (*page, error) {
	p := d.AllocateEmptyPage()
	p.Num = 0
	meta.Serialize(p.Data)

	err := d.WritePage(p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (d *DAL) ReadMeta() (*meta, error) {
	p, err := d.ReadPage(0)
	if err != nil {
		return nil, err
	}

	meta := GetMeta()
	meta.Deserialize(p.Data)
	return meta, nil
}

func (d *DAL) WriteFreeList() (*page, error) {
	p := d.AllocateEmptyPage()
	p.Num = d.FreeListPage
	d.freeList.Serialize(p.Data)

	err := d.WritePage(p)
	if err != nil {
		return nil, err
	}
	d.FreeListPage = p.Num
	return p, nil
}

func (d *DAL) ReadFreeList() (*freeList, error) {
	p, err := d.ReadPage(d.meta.FreeListPage)
	if err != nil {
		return nil, err
	}

	freelist := GetFreeList()
	freelist.Deserialize(p.Data)
	return freelist, nil
}

func (d *DAL) maxThreshold() float32 {
	return d.maxFillPercent * float32(d.PageSize)
}

func (d *DAL) isOverPopulated(node *Node) bool {
	return float32(node.nodeSize()) > d.maxThreshold()
}

func (d *DAL) minThreshold() float32 {
	return d.minFillPercent * float32(d.PageSize)
}

func (d *DAL) isUnderPopulated(node *Node) bool {
	return float32(node.nodeSize()) < d.minThreshold()
}

func (d *DAL) getSplitIndex(node *Node) int {
	size := 0
	size += NodeHeaderSize

	for i := range node.Items {
		size += node.elementSize(i)

		// if we have a big enough page size (more than minimum), and didn't reach the last node, which means we can
		// spare an element
		if float32(size) > d.minThreshold() && i < len(node.Items)-1 {
			return i + 1
		}
	}

	return -1
}
