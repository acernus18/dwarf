package cacheutil

import (
	"fmt"
	"github.com/bluele/gcache"
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	cache := gcache.New(10).LRU().Build()
	supplier := func() (string, error) {
		fmt.Println("Use Supplier")
		return "111", nil
	}
	result, err := Load[string](cache, "Test", time.Second*10, supplier)
	logrus.Println(result, err)
	result2, err := Load[string](cache, "Test", time.Second*10, supplier)
	logrus.Println(result2, err)
	time.Sleep(time.Second * 10)
	result3, err := Load[string](cache, "Test", time.Second*10, supplier)
	logrus.Println(result3, err)
}
