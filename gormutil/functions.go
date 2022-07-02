package gormutil

import (
	"fmt"
	"github.com/acernus18/dwarf/serializeutil"
	"gorm.io/gorm"
	"strings"
	"unicode"
)

const (
	LIKE = iota
	EQ
	NEQ
	GT
	LT
	GTE
	LTE
	IN
	NIN
)

func caseToCamel(name string) string {
	characters := []rune(name)
	builder := strings.Builder{}
	continueFlag := false
	for i, r := range characters {
		if unicode.IsUpper(r) && i != 0 {
			if !continueFlag {
				builder.WriteRune('_')
				continueFlag = true
			}
		} else {
			continueFlag = false
		}
		builder.WriteRune(unicode.ToLower(r))
	}
	return builder.String()
}

func (s ScopesParams) Where() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		tx := db
		validActions := map[int]string{
			LIKE: "LIKE",
			EQ:   "=",
			NEQ:  "!=",
			GT:   ">",
			LT:   "<",
			GTE:  ">=",
			LTE:  "<=",
			IN:   "IN",
			NIN:  "NOT IN",
		}
		if len(s.Filters) > 0 {
			for i := range s.Filters {
				action, valid := validActions[s.Filters[i].Action]
				if !valid {
					continue
				}
				filter := fmt.Sprintf("`%s` %s ?", caseToCamel(s.Filters[i].Key), action)
				tx = tx.Where(filter, s.Filters[i].Value)
			}
		}
		return tx
	}
}

func (s ScopesParams) Order() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		tx := db
		if len(s.Orders) > 0 {
			orders := make([]string, len(s.Orders))
			for i := range s.Orders {
				if s.Orders[i].Desc {
					orders = append(orders, fmt.Sprintf("`%s` DESC", caseToCamel(s.Orders[i].Order)))
				} else {
					orders = append(orders, fmt.Sprintf("`%s`", caseToCamel(s.Orders[i].Order)))
				}
			}
			tx = tx.Order(orders)
		}
		return tx
	}
}

func (s ScopesParams) Page() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (s.Pagination.PageIndex - 1) * s.Pagination.PageSize
		return db.Limit(s.Pagination.PageSize).Offset(offset)
	}
}

func Query[T any](tx *gorm.DB, params ScopesParams) ([]T, error) {
	result := make([]T, 0)
	if err := tx.Scopes(params.Where(), params.Order()).Find(&result).Error; err != nil {
		return result, err
	}
	return result, nil
}

func QueryWithPage[T any](tx *gorm.DB, params ScopesParams) ([]T, int64, error) {
	var model T
	var count int64 = 0
	result := make([]T, 0)
	scopes := tx.Model(&model).Scopes(params.Where(), params.Order())
	if err := scopes.Count(&count).Error; err != nil {
		return result, count, err
	}
	if err := scopes.Scopes(params.Page()).Find(&result).Error; err != nil {
		return result, count, err
	}
	return result, count, nil
}

func Take[T any](tx *gorm.DB, filter Filter) (T, error) {
	var defaultResult T
	result, err := Query[T](tx, ScopesParams{
		Filters: []Filter{filter},
		Orders:  nil,
	})
	if err != nil {
		return defaultResult, err
	}
	if len(result) < 1 {
		return defaultResult, gorm.ErrRecordNotFound
	}
	return result[0], nil
}

func Update[T any](tx *gorm.DB, filter Filter, operator func(T) (T, error)) error {
	record, err := Take[T](tx, filter)
	if err != nil {
		return err
	}
	destination, err := operator(record)
	if err != nil {
		return err
	}
	if err = tx.Model(&record).Updates(destination).Error; err != nil {
		return err
	}
	return nil
}

func Set[T any](tx *gorm.DB, key string, value T) error {

	var count int64
	if err := tx.Model(&KeyValue{}).Where("`key` = ?", key).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		return tx.Create(&KeyValue{
			Key:   key,
			Value: serializeutil.Serialize(value),
		}).Error
	}

	return tx.Model(&KeyValue{}).Where("`key` = ?", key).Update("`value`", serializeutil.Serialize(value)).Error
}

func Get[T any](tx *gorm.DB, key string) (T, error) {
	var kv KeyValue
	if err := tx.Where("`key` = ?", key).Take(&kv).Error; err != nil {
		var defaultValue T
		return defaultValue, err
	}
	return serializeutil.Deserialize[T](kv.Value)
}
