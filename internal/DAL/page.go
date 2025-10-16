package dal

type page struct {
	Num  uint64
	Data []byte
}

func (d *dal) AllocateEmptyPage() *page {
	return &page{
		Data: make([]byte, d.PageSize),
	}
}

func (d *dal) ReadPage(pageNumber uint64) (*page, error) {

	p := d.AllocateEmptyPage()

	offset := pageNumber * uint64(d.PageSize)

	_, err := d.File.ReadAt(p.Data, int64(offset))
	if err != nil {
		return nil, err
	}

	return p, err
}

func (d *dal) WritePage(p *page) error {

	offset := p.Num * uint64(d.PageSize)
	_, err := d.File.WriteAt(p.Data, int64(offset))

	return err
}
