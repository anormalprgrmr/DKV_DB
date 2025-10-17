package main

import (
	"os/exec"

	dal "github.com/anormalprgrmr/DKV_DB/internal/DAL"
)

func hexDump() {
	cmd := exec.Command("hexdump", "-C", "db.db")
	output, _ := cmd.CombinedOutput()
	println(string(output))
}

func main() {
	// initialize db
	dalIns, _ := dal.GetDal("db.db")

	// // create a new page
	p := dalIns.AllocateEmptyPage()
	p.Num = dalIns.GetNextPage()
	println("next pageee iss in start : ", p.Num)
	// copy(p.Data[:], "eyy babaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	// // commit it
	// _ = dalIns.WritePage(p)
	// _, _ = dalIns.WriteFreeList()
	// println("next pageee iss phase 1 finish: ", dalIns.GetNextPage())

	// // Close the db
	// _ = dalIns.Close()

	// We expect the freelist state was saved, so we write to
	// page number 3 and not overwrite the one at number 2
	println("going to teset2")
	println("next pageee iss phase 2 before start: ", dalIns.GetNextPage())
	// dalIns, _ = dal.GetDal("db.db")

	println("next pageee iss phase 2 start: ", dalIns.GetNextPage())
	p = dalIns.AllocateEmptyPage()
	p.Num = dalIns.GetNextPage()
	copy(p.Data[:], "wewew")
	_ = dalIns.WritePage(p)

	// // Create a page and free it so the released pages will be updated
	pageNum := dalIns.GetNextPage()
	println("next pageee iss : ", pageNum)
	dalIns.ReleasePage(pageNum)
	println("next pageee iss after release : ", pageNum)

	// // commit it
	_, _ = dalIns.WriteFreeList()
	_ = dalIns.Close()

	println("finiished")
}
