package main

import dal "github.com/anormalprgrmr/DKV_DB/internal/DAL"

func main() {
	// initialize db
	dalIns, _ := dal.GetDal("db.db")

	// create a new page
	p := dalIns.AllocateEmptyPage()
	p.Num = dalIns.GetNextPage()
	println("next pageee iss in start : ", p.Num)
	copy(p.Data[:], "data")

	// commit it
	_ = dalIns.WritePage(p)
	_, _ = dalIns.WriteFreeList()
	println("next pageee iss phase 1 finish: ", dalIns.GetNextPage())

	// Close the db
	_ = dalIns.Close()

	// We expect the freelist state was saved, so we write to
	// page number 3 and not overwrite the one at number 2
	println("going to teset2")
	dalIns, _ = dal.GetDal("db.db")
	println("next pageee iss phase 2 start: ", dalIns.GetNextPage())
	p = dalIns.AllocateEmptyPage()
	p.Num = dalIns.GetNextPage()
	copy(p.Data[:], "data2")
	_ = dalIns.WritePage(p)

	// // Create a page and free it so the released pages will be updated
	pageNum := dalIns.GetNextPage()
	println("next pageee iss : ", pageNum)
	dalIns.ReleasePage(pageNum)
	println("next pageee iss after release : ", pageNum)

	// // commit it
	_, _ = dalIns.WriteFreeList()
	println("finiished")
}
