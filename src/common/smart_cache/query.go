package smart_cache

import (
	"context"
	"errors"
	"sync"

	"github.com/coreservice-io/cli-template/basic"
	"github.com/coreservice-io/cli-template/plugin/redis_plugin"
	"github.com/coreservice-io/cli-template/plugin/reference_plugin"
	"github.com/coreservice-io/cli-template/src/common/json"
	"github.com/coreservice-io/reference"
	"github.com/go-redis/redis/v8"
)

const UPDATE_REF_CHAN_TTL_SECS = 300

type smartCacheRefElement struct {
	Obj        interface{}
	Token_chan chan struct{}
}

var lockMap sync.Map

func refGet222(localRef *reference.Reference, keystr string) (result *smartCacheRefElement, to_update bool) {
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

func redisGet222(ctx context.Context, Redis *redis.ClusterClient, serialization bool, keystr string, result interface{}) error {

	scmd := Redis.Get(ctx, keystr) //trigger remote redis get
	r_bytes, err := scmd.Bytes()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}

	// switch string(r_bytes) {
	// case query_nil_err_str:
	// 	return QueryNilErr
	// case query_err_str:
	// 	return QueryErr
	// default:
	// }

	if serialization {
		return json.Unmarshal(r_bytes, result)
	} else {
		return scmd.Scan(result)
	}
}

func smartQueryCacheSlow_getRedis222(refCacheTTLSecs int64, resultHolder interface{}, serialization bool, key string, query_description string) error {
	err := redisGet222(context.Background(), redis_plugin.GetInstance().ClusterClient, serialization, key, resultHolder)
	if err == nil {
		basic.Logger.Debugln(query_description, " hit from redis")
		refSetTTL(reference_plugin.GetInstance(), key, resultHolder, refCacheTTLSecs)
		return nil //just return resultHolder, nil
	} else {
		//redis may broken, just return to make slow query safe
		return err //return with error
	}
}

func SmartQueryCacheSlow222(key string, resultHolderAlloc func() interface{},
	serialization bool, fromCache bool, updateCache bool,
	redisCacheTTLSecs int64, refCacheTTLSecs int64,
	slowQuery func(resultHolder interface{}) error,
	queryDescription string) (interface{}, error) {

	redisCacheTTLSecs = redisCacheTTLSecs + REDIS_TTL_DELAY_SEC
	refCacheTTLSecs = refCacheTTLSecs + REF_TTL_DELAY_SECS

	var resultHolder interface{}

	if fromCache {
		// try to get from reference
		refElement, to_update_ref := refGet222(reference_plugin.GetInstance(), key)

		if refElement != nil && !to_update_ref {
			basic.Logger.Debugln(queryDescription + " SmartQueryCacheSlow hit from reference")
			switch value := refElement.Obj.(type) {
			case *error: // if error
				return nil, *value
			default:
				return refElement.Obj, nil
			}
		} else if refElement != nil && to_update_ref {
			select {
			case <-refElement.Token_chan: //get update token
				go func() {

					// _, redis_err := smartQueryCacheSlow_getRedis(refCacheTTLSecs, resultHolder, serialization, key, queryDescription)
					// if redis_err == nil  {
					// 	return resultHolder, nil
					// } else if redis_err == QueryNilErr && !redis_update {
					// 	sq_s_r_c <- struct{}{}
					// 	return nil, QueryNilErr
					// } else if redis_update {
					// 	sq_s_r_c <- struct{}{}
					// 	//continue to break
					// } else {
					// 	sq_s_r_c <- struct{}{}
					// 	return nil, redis_err
					// }

					// get from redis
					// if exist
					// set to ref

					resultHolder = resultHolderAlloc()
					ele := &smartCacheRefElement{
						Obj:        resultHolder,
						Token_chan: refElement.Token_chan,
					}
					refSetTTL(reference_plugin.GetInstance(), key, ele, refCacheTTLSecs)

					// if not exist
					// query from db
					// update redis
					// update ref

					refElement.Token_chan <- struct{}{}

				}()
				return refElement.Obj, nil
			default:
				return refElement.Obj, nil
			}
		} else {

			lc, loaded := lockMap.LoadOrStore(key, make(chan struct{}, 1))
			if loaded {
				<-lc.(chan struct{}) //unblock when chan closed
				// get from ref
				refElement, _ := refGet222(reference_plugin.GetInstance(), key)
				if refElement != nil {
					switch value := refElement.Obj.(type) {
					case *error: // if error
						return nil, *value
					default:
						return refElement.Obj, nil
					}
				} else {
					return nil, errors.New("query error")
				}
			} else {

				// get from redis
				// if exist
				// set to ref

				resultHolder = resultHolderAlloc()
				tokenChan := make(chan struct{})
				tokenChan <- struct{}{}
				ele := &smartCacheRefElement{
					Obj:        resultHolder,
					Token_chan: tokenChan,
				}
				refSetTTL(reference_plugin.GetInstance(), key, ele, refCacheTTLSecs)

				// if not exist
				// query from db
				// update redis
				// update ref

				close(lc.(chan struct{})) //just close chan
				lockMap.Delete(key)
				return resultHolder, nil
			}
		}
	}

	/////////////////////////////////

	// after cache miss ,try from remote database
	if resultHolder == nil {
		resultHolder = resultHolderAlloc()
	}
	return SlowQuery(key, resultHolder, serialization, updateCache, redisCacheTTLSecs, refCacheTTLSecs, slowQuery, queryDescription)

	// query_err := slowQuery(resultHolder)

	// if query_err != nil {
	// 	if query_err == QueryNilErr {
	// 		rrSetQueryNilErrTTL(context.Background(), redis_plugin.GetInstance().ClusterClient, key, redisCacheTTLSecs)
	// 		return nil, QueryNilErr
	// 	} else {
	// 		basic.Logger.Errorln(queryDescription, " slowQuery err :", query_err)
	// 		rrSetQueryErr(context.Background(), redis_plugin.GetInstance().ClusterClient, key)
	// 		return nil, QueryErr
	// 	}
	// } else {
	// 	if updateCache {
	// 		tokenChan := make(chan struct{})
	// 		tokenChan <- struct{}{}
	// 		ele := &smartCacheRefElement{
	// 			Obj:        resultHolder,
	// 			Token_chan: tokenChan,
	// 		}
	// 		rrSet(context.Background(), redis_plugin.GetInstance().ClusterClient, reference_plugin.GetInstance(), serialization, key, ele, redisCacheTTLSecs, refCacheTTLSecs)
	// 	}
	// 	return resultHolder, nil
	// }
}

func SlowQuery(key string, resultHolder interface{},
	serialization bool, updateCache bool,
	redisCacheTTLSecs int64, refCacheTTLSecs int64,
	slowQuery func(resultHolder interface{}) error,
	queryDescription string) (interface{}, error) {

	basic.Logger.Debugln(queryDescription, " try from db query")

	// if resultHolder == nil {
	// 	resultHolder = resultHolderAlloc()
	// }
	query_err := slowQuery(resultHolder)

	if query_err != nil {
		if query_err == QueryNilErr {
			rrSetQueryNilErrTTL(context.Background(), redis_plugin.GetInstance().ClusterClient, key, redisCacheTTLSecs)
			return nil, QueryNilErr
		} else {
			basic.Logger.Errorln(queryDescription, " slowQuery err :", query_err)
			rrSetQueryErr(context.Background(), redis_plugin.GetInstance().ClusterClient, key)
			return nil, QueryErr
		}
	} else {
		if updateCache {
			tokenChan := make(chan struct{})
			tokenChan <- struct{}{}
			ele := &smartCacheRefElement{
				Obj:        resultHolder,
				Token_chan: tokenChan,
			}
			rrSet(context.Background(), redis_plugin.GetInstance().ClusterClient, reference_plugin.GetInstance(), serialization, key, ele, redisCacheTTLSecs, refCacheTTLSecs)
		}
		return resultHolder, nil
	}
}

////////////

// //////////////////
// usually used for db cases which has slow query
func SmartQueryCacheSlow(key string, resultHolderAlloc func() interface{},
	serialization bool, fromCache bool, updateCache bool,
	redisCacheTTLSecs int64, refCacheTTLSecs int64,
	slowQuery func(resultHolder interface{}) error,
	queryDescription string) (interface{}, error) {

	redisCacheTTLSecs = redisCacheTTLSecs + REDIS_TTL_DELAY_SEC
	refCacheTTLSecs = refCacheTTLSecs + REF_TTL_DELAY_SECS

	var resultHolder interface{}

	if fromCache {

		// try to get from reference
		result, to_update_ref := refGet(reference_plugin.GetInstance(), key)

		if !to_update_ref {
			basic.Logger.Debugln(queryDescription + " SmartQueryCacheSlow hit from reference")
			return result, nil
		} else {

			update_ref_key := "sq_s_r_c_" + key
			//slow_query_ref_chan
			sq_s_r_c_i, sq_s_r_c_ttl_left_secs := reference_plugin.GetInstance().Get(update_ref_key)
			var sq_s_r_c chan struct{}
			if sq_s_r_c_i == nil {
				sq_s_r_c = make(chan struct{}, 1)
				sq_s_r_c <- struct{}{}
			} else {
				sq_s_r_c = sq_s_r_c_i.(chan struct{})
			}

			if sq_s_r_c_ttl_left_secs < UPDATE_REF_CHAN_TTL_SECS/2 {
				reference_plugin.GetInstance().Set(update_ref_key, sq_s_r_c, UPDATE_REF_CHAN_TTL_SECS)
			}

			if result != nil {
				select {
				case <-sq_s_r_c:
					//don't need check again
					//do the slow query
					resultHolder = resultHolderAlloc()
					redis_update, redis_err := smartQueryCacheSlow_getRedis(refCacheTTLSecs, resultHolder, serialization, key, queryDescription)
					if redis_err == nil && !redis_update {
						sq_s_r_c <- struct{}{}
						return resultHolder, nil
					} else if redis_err == QueryNilErr && !redis_update {
						sq_s_r_c <- struct{}{}
						return nil, QueryNilErr
					} else if redis_update {
						sq_s_r_c <- struct{}{}
						//continue to break
					} else {
						sq_s_r_c <- struct{}{}
						return nil, redis_err
					}
					break //will go to outsider code
				default:
					return result, nil
				}
			} else {
				<-sq_s_r_c
				//check again
				result, to_update_ref_check := refGet(reference_plugin.GetInstance(), key)
				if !to_update_ref_check {
					sq_s_r_c <- struct{}{}
					return result, nil
				}
				//
				resultHolder = resultHolderAlloc()
				redis_update, redis_err := smartQueryCacheSlow_getRedis(refCacheTTLSecs, resultHolder, serialization, key, queryDescription)
				if redis_err == nil && !redis_update {
					sq_s_r_c <- struct{}{}
					return resultHolder, nil
				} else if redis_err == QueryNilErr && !redis_update {
					sq_s_r_c <- struct{}{}
					return nil, QueryNilErr
				} else if redis_update {
					sq_s_r_c <- struct{}{}
					//continue
				} else {
					sq_s_r_c <- struct{}{}
					return nil, redis_err
				}
			}
		}

	}

	//after cache miss ,try from remote database
	basic.Logger.Debugln(queryDescription, " try from query")

	//slow_query_db_chan
	update_db_key := "sq_s_db_c_" + key

	sq_s_r_db_i, sq_s_db_c_ttl_left_secs := reference_plugin.GetInstance().Get(update_db_key)
	var sq_s_db_c chan struct{}
	if sq_s_r_db_i == nil {
		sq_s_db_c = make(chan struct{}, 1)
		sq_s_db_c <- struct{}{}
	} else {
		sq_s_db_c = sq_s_r_db_i.(chan struct{})
	}

	if sq_s_db_c_ttl_left_secs < UPDATE_REF_CHAN_TTL_SECS/2 {
		reference_plugin.GetInstance().Set(update_db_key, sq_s_db_c, UPDATE_REF_CHAN_TTL_SECS)
	}

	/////////////

	if resultHolder != nil {
		select {
		case <-sq_s_db_c:

			query_err := slowQuery(resultHolder)

			if query_err != nil {
				if query_err == QueryNilErr {
					rrSetQueryNilErrTTL(context.Background(), redis_plugin.GetInstance().ClusterClient, key, redisCacheTTLSecs)
					sq_s_db_c <- struct{}{}
					return nil, QueryNilErr
				} else {
					basic.Logger.Errorln(queryDescription, " slowQuery err :", query_err)
					rrSetQueryErr(context.Background(), redis_plugin.GetInstance().ClusterClient, key)
					sq_s_db_c <- struct{}{}
					return nil, QueryErr
				}
			} else {
				if updateCache {
					rrSet(context.Background(), redis_plugin.GetInstance().ClusterClient, reference_plugin.GetInstance(), serialization, key, resultHolder, redisCacheTTLSecs, refCacheTTLSecs)
				}
				sq_s_db_c <- struct{}{}
				return resultHolder, nil
			}

		default:
			return resultHolder, nil
		}

	} else {

		resultHolder = resultHolderAlloc()

		<-sq_s_db_c
		//check again
		result, to_update_ref_check := refGet(reference_plugin.GetInstance(), key)
		if !to_update_ref_check {
			sq_s_db_c <- struct{}{}
			return result, nil
		}

		query_err := slowQuery(resultHolder)

		if query_err != nil {
			if query_err == QueryNilErr {
				rrSetQueryNilErrTTL(context.Background(), redis_plugin.GetInstance().ClusterClient, key, redisCacheTTLSecs)
				sq_s_db_c <- struct{}{}
				return nil, QueryNilErr
			} else {
				basic.Logger.Errorln(queryDescription, " slowQuery err :", query_err)
				rrSetQueryErr(context.Background(), redis_plugin.GetInstance().ClusterClient, key)
				sq_s_db_c <- struct{}{}
				return nil, QueryErr
			}
		} else {
			if updateCache {
				rrSet(context.Background(), redis_plugin.GetInstance().ClusterClient, reference_plugin.GetInstance(), serialization, key, resultHolder, redisCacheTTLSecs, refCacheTTLSecs)
			}
			sq_s_db_c <- struct{}{}
			return resultHolder, nil
		}

	}

}

// return to_update bool , error
func smartQueryCacheSlow_getRedis(refCacheTTLSecs int64, resultHolder interface{}, serialization bool, key string, query_description string) (bool, error) {
	to_update, err := redisGet(context.Background(), redis_plugin.GetInstance().ClusterClient, serialization, key, resultHolder)
	if err == nil {
		basic.Logger.Debugln(query_description, " hit from redis")
		refSetTTL(reference_plugin.GetInstance(), key, resultHolder, refCacheTTLSecs)
		return to_update, nil //just return resultHolder, nil
	} else if err == QueryNilErr {
		return to_update, QueryNilErr //return with error
	} else if err == QueryErr {
		basic.Logger.Errorln(query_description, " QueryErr") //this happens when query failed
		return to_update, QueryErr                           //return with error
	} else {
		//redis may broken, just return to make slow query safe
		return to_update, err //return with error
	}
}

// fastQuery usually is a redis query
// for Query ,return QueryNilErr if Query result is nil  -> as set nil to cache is not supported
func SmartQueryCacheFast(
	key string,
	resultHolderAlloc func() interface{},
	fromRefCache bool,
	updateRefCache bool,
	refCacheTTLSecs int64,
	fastQuery func(resultHolder interface{}) error,
	queryDescription string) (interface{}, error) {

	refCacheTTLSecs = refCacheTTLSecs + REF_TTL_DELAY_SECS

	if fromRefCache {
		// try to get from reference
		result, to_update_ref := refGet(reference_plugin.GetInstance(), key)
		if !to_update_ref {
			basic.Logger.Debugln(queryDescription + " SmartQueryCacheFast hit from reference")
			return result, nil
		}
	}

	//after cache miss ,try from remote database
	basic.Logger.Debugln(queryDescription, " SmartQueryCacheFast try from fast query")

	//fast_query_db_chan
	update_f_key := "sq_f_c_" + key

	sq_f_i, sq_f_ttl_left_secs := reference_plugin.GetInstance().Get(update_f_key)
	var sq_f_c chan struct{}
	if sq_f_i == nil {
		sq_f_c = make(chan struct{}, 1)
		sq_f_c <- struct{}{}
	} else {
		sq_f_c = sq_f_i.(chan struct{})
	}

	if sq_f_ttl_left_secs < UPDATE_REF_CHAN_TTL_SECS/2 {
		reference_plugin.GetInstance().Set(update_f_key, sq_f_c, UPDATE_REF_CHAN_TTL_SECS)
	}

	////

	sq_f_c <- struct{}{}

	//check again
	result, to_update_ref_check := refGet(reference_plugin.GetInstance(), key)
	if !to_update_ref_check {
		sq_f_c <- struct{}{}
		return result, nil
	}

	///
	resultHolder := resultHolderAlloc()
	query_err := fastQuery(resultHolder)
	if query_err != nil {
		if query_err == QueryNilErr {
			return nil, QueryNilErr
		} else {
			return nil, QueryErr
		}
	} else {
		if updateRefCache {
			reference_plugin.GetInstance().Set(key, resultHolder, refCacheTTLSecs)
		}
		return resultHolder, nil
	}

}
