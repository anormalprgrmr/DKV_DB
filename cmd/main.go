package main

import (
	"os"

	dal "github.com/anormalprgrmr/DKV_DB/internal/DAL"
)

func main() {
	dal, _ := dal.GetDal("db.db", os.Getpagesize())
	p := dal.AllocateEmptyPage()
	p.Num = dal.GetNextPage()
	copy(p.Data[:], "hell")
	_ = dal.WritePage(p)

	// api.RunServer(8080)
}
