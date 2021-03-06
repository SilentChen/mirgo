package common

import (
	"os"
	"testing"
)

func TestNewDropInfo(t *testing.T) {
	di1, err := NewDropInfo("1/10 Gold 500")
	if err != nil {
		t.Error(err)
	}
	t.Log(di1)
	di2, err := NewDropInfo("1/5 (MP)DrugLarge")
	if err != nil {
		t.Error(err)
	}
	t.Log(di2)
}

func TestGetDropInfosByMonsterName(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	monsterName := "Ancient_AxeSkeleton"
	dropDirPath := gopath + "/src/github.com/yenkeia/mirgo/dotnettools/database/Envir/Drops/"
	drops, err := GetDropInfosByMonsterName(dropDirPath, monsterName)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(drops)
	}
}
