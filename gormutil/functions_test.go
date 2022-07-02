package gormutil

import (
	"fmt"
	"github.com/duke-git/lancet/v2/strutil"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func setupDB(password, address, database string) *gorm.DB {
	dsn := fmt.Sprintf("root:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", password, address, database)
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(err)
	}
	return db
}

func TestCaseToCamel(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test1",
			args: args{name: ""},
			want: "",
		},
		{
			name: "Test2",
			args: args{name: "Item"},
			want: "item",
		},
		{
			name: "Test3",
			args: args{name: "ItemDetail"},
			want: "item_detail",
		},
		{
			name: "Test4",
			args: args{name: "ItemID"},
			want: "item_id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("%s", strutil.SnakeCase(tt.args.name))
			if got := caseToCamel(tt.args.name); got != tt.want {
				t.Errorf("caseToCamel() = %v, want %v", got, tt.want)
			}
		})
	}
}

type ItemDetail struct {
	gorm.Model
}

func TestSetAndGet(t *testing.T) {
	db := setupDB("", "", "")
	item, err := Take[ItemDetail](db, Filter{
		Key:    "ItemID",
		Action: EQ,
		Value:  "11251",
	})
	if err != nil {
		t.Error(err)
		return
	}
	if err = Set[ItemDetail](db, "TestKey", item); err != nil {
		t.Error(err)
		return
	}
	result, err := Get[ItemDetail](db, "TestKey")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(result)
}
