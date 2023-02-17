package smart_cache

import (
	"context"
	"errors"
	"math/rand"
	"reflect"
	"time"

	"github.com/coreservice-io/cli-template/src/common/json"
	"github.com/coreservice-io/reference"
	"github.com/go-redis/redis/v8"
)

type query_err string

func (e query_err) Error() string { return string(e) }

type query_nil_err string

func (e query_nil_err) Error() string { return string(e) }

const query_err_str = "|query_err|"
const query_nil_err_str = "|query_nil_err|"

const CacheNilErr = redis.Nil //won't be used outside module
const QueryErr = query_err(query_err_str)
const QueryNilErr = query_nil_err(query_nil_err_str)

const QUERY_ERR_SECS = 5 //if query failed (not nil err) , set a temporary mark in redis

// check weather we need do refresh
// the probobility becomes lager when left seconds close to 0
// this goal of this function is to avoid big traffic glitch
func check_ref_ttl_refresh(secleft int64) bool {
	if secleft == 0 {
		return true
	}

	if secleft > 0 && secleft <= 3 {
		if rand.Intn(int(secleft)*5) == 0 {
			return true
		}
	}
	return false
}

func check_redis_ttl_refresh(secleft int64) bool {
	if secleft == 0 {
		return true
	}
	if secleft > 0 && secleft <= 3 {
		if rand.Intn(int(secleft)*2) == 0 {
			return true
		}
	}
	return false
}

func Ref_Get(localRef *reference.Reference, keystr string) (result interface{}) {
	localvalue, ttl := localRef.Get(keystr)
	if !check_ref_ttl_refresh(ttl) && localvalue != nil {
		return localvalue
	}
	return nil
}

// func Ref_Set(localRef *reference.Reference, keystr string, value interface{}) error {
// 	return Ref_Set_TTL(localRef, keystr, value, local_reference_secs)
// }

func Ref_Set_TTL(localRef *reference.Reference, keystr string, value interface{}, ref_ttl_second int64) error {
	return localRef.Set(keystr, value, ref_ttl_second)
}

// //first try from localRef if not found then try from remote redis
func Redis_Get(ctx context.Context, Redis *redis.ClusterClient, serialization bool, keystr string, result interface{}) error {
	// 1/5 check ttl
	if rand.Intn(5) == 0 {
		ttl, err := Redis.TTL(context.Background(), keystr).Result()
		// if ttl==-1 means no expire

		// if has expire time
		if err == nil && ttl != -1 && check_redis_ttl_refresh(int64(ttl.Seconds())) {
			//need refresh
			return CacheNilErr
		}
	}

	scmd := Redis.Get(ctx, keystr) //trigger remote redis get
	r_bytes, err := scmd.Bytes()
	if err == redis.Nil {
		return CacheNilErr
	}
	if err != nil {
		return err
	}

	switch string(r_bytes) {
	case query_nil_err_str:
		return QueryNilErr
	case query_err_str:
		return QueryErr
	default:
	}

	if serialization {
		return json.Unmarshal(r_bytes, result)
	} else {
		return scmd.Scan(result)
	}
}

func RR_Set(ctx context.Context, Redis *redis.ClusterClient, localRef *reference.Reference, serialization bool, keystr string, value interface{}, redis_ttl_second int64, ref_ttl_second int64) error {
	return RR_Set_TTL(ctx, Redis, localRef, serialization, keystr, value, redis_ttl_second, ref_ttl_second)
}

func RR_SetQueryErr(ctx context.Context, Redis *redis.ClusterClient, keystr string) error {
	return Redis.Set(ctx, keystr, query_err_str, time.Duration(QUERY_ERR_SECS)*time.Second).Err()
}

func RR_SetQueryErr_TTL(ctx context.Context, Redis *redis.ClusterClient, keystr string, ttl_second int64) error {
	return Redis.Set(ctx, keystr, query_err_str, time.Duration(ttl_second)*time.Second).Err()
}

func RR_SetQueryNilErr_TTL(ctx context.Context, Redis *redis.ClusterClient, keystr string, ttl_second int64) error {
	return Redis.Set(ctx, keystr, query_nil_err_str, time.Duration(ttl_second)*time.Second).Err()
}

// reference set && redis set
// set both value to both local reference & remote redis
func RR_Set_TTL(ctx context.Context, Redis *redis.ClusterClient, localRef *reference.Reference, serialization bool, keystr string, value interface{}, redis_ttl_second int64, ref_ttl_second int64) error {
	if value == nil {
		return errors.New("value nil not allowed")
	}
	if serialization {
		err := localRef.Set(keystr, value, ref_ttl_second)
		if err != nil {
			return err
		}
		v_json, err := json.Marshal(value)
		if err != nil {
			return err
		}
		return Redis.Set(ctx, keystr, v_json, time.Duration(redis_ttl_second)*time.Second).Err()
	} else {
		err := localRef.Set(keystr, value, ref_ttl_second)
		if err != nil {
			return err
		}
		tp := reflect.TypeOf(value).Kind()
		if tp == reflect.Ptr {
			pointer_v_type := reflect.TypeOf(value).Elem().Kind()
			if pointer_v_type == reflect.Slice || pointer_v_type == reflect.Struct {
				return errors.New("pointer to slice/struct type must set with serialization=true")
			}
			return Redis.Set(ctx, keystr, reflect.ValueOf(value).Elem().Interface(), time.Duration(redis_ttl_second)*time.Second).Err()
		} else {
			if tp == reflect.Slice || tp == reflect.Struct {
				return errors.New("slice/struct type must set with serialization=true")
			}
			return Redis.Set(ctx, keystr, value, time.Duration(redis_ttl_second)*time.Second).Err()
		}
	}
}

func RR_Del(ctx context.Context, Redis *redis.ClusterClient, localRef *reference.Reference, keystr string) {
	localRef.Delete(keystr)
	Redis.Del(ctx, keystr)
}
