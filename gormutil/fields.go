package gormutil

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Filter struct {
	Key    string
	Action int
	Value  any
}

type Order struct {
	Order string
	Desc  bool
}

type Pagination struct {
	PageSize  int
	PageIndex int
}

type ScopesParams struct {
	Filters    []Filter
	Orders     []Order
	Pagination Pagination
}

// KeyValue Schema
//create table key_values
//(
//	`id`          int unsigned auto_increment,
//	`created_at`  datetime     not null,
//	`updated_at`  datetime     not null,
//	`deleted_at`  datetime,
//	`key`   	  varchar(32)  not null,
//	`value` 	  json         not null,
//	primary key (`id`),
//	key `idx_deleted_id` (`deleted_at`) using btree,
//	unique key `u_idx_key` (`key`) using btree
//) engine innodb
//default charset utf8mb4;
type KeyValue struct {
	gorm.Model
	Key   string
	Value datatypes.JSON
}
