package dal

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

type DAL struct {
	File     *os.File
	PageSize int

	*meta
	*freeList
}

var (
	dalInstance *DAL
	dalOnce     sync.Once
)

func GetDal(path string) (*DAL, error) {
	var err error
	dalOnce.Do(func() {

		dalInstance = &DAL{
			meta:     GetMeta(),
			PageSize: os.Getpagesize(),
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
	p, err := d.ReadPage(d.FreeListPage)
	if err != nil {
		return nil, err
	}

	freelist := GetFreeList()
	freelist.Deserialize(p.Data)
	return freelist, nil
}
