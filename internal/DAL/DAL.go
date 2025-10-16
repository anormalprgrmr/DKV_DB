package dal

import (
	"fmt"
	"os"
	"sync"
)

type dal struct {
	File     *os.File
	PageSize int
	*freeList
}

var (
	dalInstance *dal
	dalOnce     sync.Once
)

func GetDal(path string, pageSize int) (*dal, error) {
	var err error
	dalOnce.Do(func() {
		var file *os.File
		file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return
		}
		dalInstance = &dal{
			File:     file,
			PageSize: pageSize,
			freeList: GetFreeList(),
		}
	})
	return dalInstance, err
}

func (d *dal) Close() error {
	if d.File != nil {
		err := d.File.Close()
		if err != nil {
			return fmt.Errorf("couldn't close file : %s \n", err)
		}
	}
	return nil
}
