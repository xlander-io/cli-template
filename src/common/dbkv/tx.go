package dbkv

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/coreservice-io/cli-template/plugin/redis_plugin"
	"github.com/coreservice-io/cli-template/src/common/smart_cache"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const DBKV_CACHE_TIME_SECS = 300

func SetDBKV_Str(tx *gorm.DB, keystr string, value string, description string) error {
	tx_result := tx.Table(TABLE_NAME_DBKV).Clauses(clause.OnConflict{UpdateAll: true}).Create(&DBKVModel{Key: keystr, Value: value, Description: description})
	if tx_result.Error != nil {
		return tx_result.Error
	}
	if tx_result.RowsAffected == 0 {
		return errors.New("0 row affected")
	}

	QueryDBKV(tx, nil, &[]string{keystr}, false, true)
	return nil
}

func SetDBKV_UInt64(tx *gorm.DB, keystr string, value uint64, description string) error {
	return SetDBKV_Str(tx, keystr, strconv.FormatUint(value, 10), description)
}

func SetDBKV_Int64(tx *gorm.DB, keystr string, value int64, description string) error {
	return SetDBKV_Str(tx, keystr, strconv.FormatInt(value, 10), description)
}

func SetDBKV_Int32(tx *gorm.DB, keystr string, value int32, description string) error {
	return SetDBKV_Str(tx, keystr, strconv.FormatInt(int64(value), 10), description)
}

func SetDBKV_Int(tx *gorm.DB, keystr string, value int, description string) error {
	return SetDBKV_Str(tx, keystr, strconv.Itoa(value), description)
}

func SetDBKV_Bool(tx *gorm.DB, keystr string, value bool, description string) error {
	if value {
		return SetDBKV_Str(tx, keystr, "true", description)
	} else {
		return SetDBKV_Str(tx, keystr, "false", description)
	}
}

func SetDBKV_Float32(tx *gorm.DB, keystr string, value float32, description string) error {
	return SetDBKV_Str(tx, keystr, fmt.Sprintf("%f", value), description)
}

func SetDBKV_Float64(tx *gorm.DB, keystr string, value float64, description string) error {
	return SetDBKV_Str(tx, keystr, fmt.Sprintf("%f", value), description)
}

func DeleteDBKV_Key(tx *gorm.DB, keystr string) error {
	if err := tx.Table(TABLE_NAME_DBKV).Where(" `key` = ?", keystr).Delete(&DBKVModel{}).Error; err != nil {
		return err
	}
	return nil
}

func DeleteDBKV_Id(tx *gorm.DB, id int64) error {
	if err := tx.Table(TABLE_NAME_DBKV).Where(" `id` = ?", id).Delete(&DBKVModel{}).Error; err != nil {
		return err
	}
	return nil
}

func GetDBKV_Bool(tx *gorm.DB, key string) (bool, error) {
	result, err := GetDBKV(tx, nil, &key, true, true)
	if err != nil {
		return false, err
	}

	bool_result, bool_err := result.ToBool()
	if bool_err != nil {
		return false, bool_err
	}
	return bool_result, nil
}

func GetDBKV_Str(tx *gorm.DB, key string) (string, error) {
	result, err := GetDBKV(tx, nil, &key, true, true)
	if err != nil {
		return "", err
	}
	return result.ToString(), nil
}

func GetDBKV_Int(tx *gorm.DB, key string) (int, error) {
	result, err := GetDBKV(tx, nil, &key, true, true)
	if err != nil {
		return 0, err
	}

	int_result, int_err := result.ToInt()
	if int_err != nil {
		return 0, int_err
	}
	return int_result, nil
}

func GetDBKV_Int32(tx *gorm.DB, key string) (int32, error) {
	result, err := GetDBKV(tx, nil, &key, true, true)
	if err != nil {
		return 0, err
	}

	int32_result, int32_err := result.ToInt32()
	if int32_err != nil {
		return 0, int32_err
	}

	return int32_result, nil
}

func GetDBKV_Int64(tx *gorm.DB, key string) (int64, error) {
	result, err := GetDBKV(tx, nil, &key, true, true)
	if err != nil {
		return 0, err
	}

	int64_result, int64_err := result.ToInt64()
	if int64_err != nil {
		return 0, int64_err
	}

	return int64_result, nil
}

func GetDBKV_UInt64(tx *gorm.DB, key string) (uint64, error) {
	result, err := GetDBKV(tx, nil, &key, true, true)
	if err != nil {
		return 0, err
	}

	uint64_result, uint64_err := result.ToUInt64()
	if uint64_err != nil {
		return 0, uint64_err
	}

	return uint64_result, nil
}

func GetDBKV_Float32(tx *gorm.DB, key string) (float32, error) {
	result, err := GetDBKV(tx, nil, &key, true, true)
	if err != nil {
		return 0, err
	}

	float32_result, float32_err := result.ToFloat32()
	if float32_err != nil {
		return 0, float32_err
	}

	return float32_result, nil
}

func GetDBKV_Float64(tx *gorm.DB, key string) (float64, error) {
	result, err := GetDBKV(tx, nil, &key, true, true)
	if err != nil {
		return 0, err
	}

	float64_result, float64_err := result.ToFloat64()
	if float64_err != nil {
		return 0, float64_err
	}

	return float64_result, nil
}

func GetDBKV(tx *gorm.DB, id *int64, key *string, fromCache bool, updateCache bool) (*DBKVModel, error) {

	var result *DBKVQueryResults
	var err error

	if key == nil {
		result, err = QueryDBKV(tx, id, nil, fromCache, updateCache)
	} else {
		result, err = QueryDBKV(tx, id, &[]string{*key}, fromCache, updateCache)
	}

	if err != nil {
		return nil, err
	}

	if result.TotalCount == 0 {
		return nil, smart_cache.ErrQueryNil
	}

	return result.Kv[0], nil
}

type DBKVQueryResults struct {
	Kv         []*DBKVModel
	TotalCount int64
}

func QueryDBKV(tx *gorm.DB, id *int64, keys *[]string, fromCache bool, updateCache bool) (*DBKVQueryResults, error) {

	//gen_key
	ck := smart_cache.NewConnectKey("dbkv")
	ck.C_Int64_Ptr("id", id).C_Str_Array_Ptr("keys", keys)

	key := redis_plugin.GetInstance().GenKey(ck.String())

	/////
	resultHolderAlloc := func() interface{} {
		return &DBKVQueryResults{
			Kv:         []*DBKVModel{},
			TotalCount: 0,
		}
	}

	/////
	// return (redisCacheSec int64, refCacheSec int64, err error), if err!=nil,ignore redisCacheSec and refCacheSec
	query := func(resultHolder interface{}) (int64,int64,error) {
		queryResults := resultHolder.(*DBKVQueryResults)

		query := tx.Table(TABLE_NAME_DBKV)
		if id != nil {
			query.Where("id = ?", *id)
		}
		if keys != nil {
			query.Where(TABLE_NAME_DBKV+".key IN ?", *keys)
		}

		query.Count(&queryResults.TotalCount)

		err := query.Find(&queryResults.Kv).Error
		if err != nil {
			return 0,0,err
		}

		if len(queryResults.Kv)==0{
			return 30,5,nil
		}else{
			return DBKV_CACHE_TIME_SECS,60,nil
		}
	}

	/////
	sq_result, sq_err := smart_cache.SmartQueryCacheSlow(key, resultHolderAlloc, true, fromCache, updateCache,
		 60, query, "DBKV Query")

	/////
	if sq_err != nil {
		return nil, sq_err
	} else {
		return sq_result.(*DBKVQueryResults), nil
	}

}
