package smart_cache

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/coreservice-io/cli-template/src/common/json"
	"github.com/coreservice-io/reference"
	"github.com/go-redis/redis/v8"
)

const query_err_str = "|query_err|"
const query_nil_err_str = "|query_nil_err|"

var ErrQuery = errors.New(query_err_str)
var ErrQueryNil = errors.New(query_nil_err_str)

const REF_TTL_DELAY_SECS = 600  //add REF_TTL_DELAY_SECS to local ref when set
const REDIS_TTL_DELAY_SEC = 300 // add REDIS_TTL_DELAY_SEC to redis when set

const QUERY_ERR_SECS = 5 //if query failed (not nil err) , set a temporary mark in redis

func refGet(localRef *reference.Reference, keystr string) (result *smartCacheRefElement, to_update bool) {
	refElement, ttl := localRef.Get(keystr)
	if refElement == nil {
		return nil, true
	}

	if ttl <= REF_TTL_DELAY_SECS {
		return refElement.(*smartCacheRefElement), true
	} else {
		return refElement.(*smartCacheRefElement), false
	}
}

func refSetTTL(localRef *reference.Reference, keystr string, element *smartCacheRefElement, ref_ttl_second int64) error {
	return localRef.Set(keystr, element, ref_ttl_second)
}

// //first try from localRef if not found then try from remote redis
func redisGet(ctx context.Context, Redis *redis.ClusterClient, serialization bool, keystr string, result interface{}) error {

	scmd := Redis.Get(ctx, keystr) //trigger remote redis get
	r_bytes, err := scmd.Bytes()
	if err == redis.Nil {
		return ErrQueryNil
	}
	if err != nil {
		return err
	}

	if serialization {
		return json.Unmarshal(r_bytes, result)
	} else {
		return scmd.Scan(result)
	}
}

func rrSet(ctx context.Context, Redis *redis.ClusterClient, localRef *reference.Reference, serialization bool, keystr string, element *smartCacheRefElement, redis_ttl_second int64, ref_ttl_second int64) error {
	return rrSetTTL(ctx, Redis, localRef, serialization, keystr, element, redis_ttl_second, ref_ttl_second)
}

// reference set && redis set
// set both value to both local reference & remote redis
func rrSetTTL(ctx context.Context, Redis *redis.ClusterClient, localRef *reference.Reference, serialization bool, keystr string, element *smartCacheRefElement, redis_ttl_second int64, ref_ttl_second int64) error {
	if element == nil {
		return errors.New("value nil not allowed")
	}
	if serialization {
		err := localRef.Set(keystr, element, ref_ttl_second)
		if err != nil {
			return err
		}
		v_json, err := json.Marshal(element.Obj)
		if err != nil {
			return err
		}
		return Redis.Set(ctx, keystr, v_json, time.Duration(redis_ttl_second)*time.Second).Err()
	} else {
		err := localRef.Set(keystr, element, ref_ttl_second)
		if err != nil {
			return err
		}
		tp := reflect.TypeOf(element.Obj).Kind()
		if tp == reflect.Ptr {
			pointer_v_type := reflect.TypeOf(element.Obj).Elem().Kind()
			if pointer_v_type == reflect.Slice || pointer_v_type == reflect.Struct {
				return errors.New("pointer to slice/struct type must set with serialization=true")
			}
			return Redis.Set(ctx, keystr, reflect.ValueOf(element.Obj).Elem().Interface(), time.Duration(redis_ttl_second)*time.Second).Err()
		} else {
			if tp == reflect.Slice || tp == reflect.Struct {
				return errors.New("slice/struct type must set with serialization=true")
			}
			return Redis.Set(ctx, keystr, element.Obj, time.Duration(redis_ttl_second)*time.Second).Err()
		}
	}
}

func refSetErr(ctx context.Context, localRef *reference.Reference, keystr string, err error) error {
	tokenChan := make(chan struct{})
	tokenChan <- struct{}{}
	ele := &smartCacheRefElement{
		Obj:        err,
		Token_chan: tokenChan,
	}
	return refSetTTL(localRef, keystr, ele, QUERY_ERR_SECS+REF_TTL_DELAY_SECS)

}
