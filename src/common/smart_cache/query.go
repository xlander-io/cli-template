package smart_cache

import (
	"context"
	"errors"
	"sync"

	"github.com/coreservice-io/cli-template/basic"
	"github.com/coreservice-io/cli-template/plugin/redis_plugin"
	"github.com/coreservice-io/cli-template/plugin/reference_plugin"
)

type QueryCacheTTL struct {
	Redis_ttl_secs int64
	Ref_ttl_secs   int64
}

type SlowQuery struct {
	//return default of : redis_ttl_secs, ref_ttl_secs
	CacheTTL *QueryCacheTTL
	//return redis_ttl_secs, ref_ttl_secs,error
	Query func(resultHolder interface{}) (*QueryCacheTTL, error)
}

var SlowQueryTTL_NOT_FOUND = &QueryCacheTTL{
	Redis_ttl_secs: 30,
	Ref_ttl_secs:   5,
}

var SlowQueryTTL_Default = &QueryCacheTTL{
	Redis_ttl_secs: 300,
	Ref_ttl_secs:   5,
}

/////////////////

type smartCacheRefElement struct {
	Obj        interface{}
	Token_chan chan struct{}
}

var lockMap sync.Map

func SmartQueryCacheSlow(key string, resultHolderAlloc func() interface{},
	serialization bool, fromCache bool, updateCache bool, slowQuery *SlowQuery,
	queryDescription string) (interface{}, error) {

	if fromCache {
		// try to get from reference
		refElement, to_update_ref := refGet(reference_plugin.GetInstance(), key)

		if refElement != nil && !to_update_ref { // 1. ref exist and no need to update
			basic.Logger.Debugln(queryDescription + " SmartQueryCacheSlow hit from reference")
			switch value := refElement.Obj.(type) {
			case error: // if error
				return nil, value
			default:
				return refElement.Obj, nil
			}
		} else if refElement != nil && to_update_ref { //2. ref exist and need update
			select {
			case <-refElement.Token_chan: //get run token
				go func() {
					defer func() {
						refElement.Token_chan <- struct{}{} // release run token
					}()

					resultHolder := resultHolderAlloc()
					// get from redis
					basic.Logger.Debugln(queryDescription, " SmartQueryCacheSlow try from redis")
					redis_err := redisGet(context.Background(), redis_plugin.GetInstance().ClusterClient, serialization, key, resultHolder)
					// redis_err => 1.nil,no err 2.ErrQueryNil 3.other err
					if redis_err == nil { //1.nil,no err
						// exist in redis
						// ref update
						ele := &smartCacheRefElement{
							Obj:        resultHolder,
							Token_chan: refElement.Token_chan, // use exist chan
						}
						refSetTTL(reference_plugin.GetInstance(), key, ele, slowQuery.CacheTTL.Ref_ttl_secs+REF_TTL_DELAY_SECS)
					} else if redis_err == ErrQueryNil { //2.ErrQueryNil
						// try from origin (example form db)
						// must update ref and redis
						slowQueryGetOrigin(key, resultHolder, serialization, true, slowQuery, queryDescription)
					} else { //3.other err
						// cache other error in ref for a short time
						refSetErr(context.Background(), reference_plugin.GetInstance(), key, redis_err)
					}
				}()
				return refElement.Obj, nil
			default:
				return refElement.Obj, nil
			}
		} else { //3. ref not exist
			lc, loaded := lockMap.LoadOrStore(key, make(chan struct{}, 1))

			if loaded { //most query enter this filed
				<-lc.(chan struct{}) //all processes unblock when chan closed
				refElement, _ := refGet(reference_plugin.GetInstance(), key)
				if refElement != nil {
					switch value := refElement.Obj.(type) {
					case error: // if error
						return nil, value
					default:
						return refElement.Obj, nil
					}
				} else {
					//ref nothing found which may caused by too short ref ttl which seems all most impossible
					basic.Logger.Errorln("SmartQueryCacheSlow ref nothing found which may caused by too short ref ttl which seems all most impossible")
					return nil, errors.New("SmartQueryCacheSlow query ref nil error")
				}

			} else { // only 1 query enter below
				resultHolder := resultHolderAlloc()
				// get from redis
				basic.Logger.Debugln(queryDescription, " SmartQueryCacheSlow try from redis")
				redis_err := redisGet(context.Background(), redis_plugin.GetInstance().ClusterClient, serialization, key, resultHolder)
				// redis_err => 1.nil,no err 2.QueryNilErr 3.other err
				if redis_err == nil { //1.nil,no err
					// exist in redis
					// ref update
					tokenChan := make(chan struct{}, 1)
					tokenChan <- struct{}{}
					ele := &smartCacheRefElement{
						Obj:        resultHolder,
						Token_chan: tokenChan, // a new chan
					}
					refSetTTL(reference_plugin.GetInstance(), key, ele, slowQuery.CacheTTL.Ref_ttl_secs+REF_TTL_DELAY_SECS)

					//close and delete chan after ref set
					close(lc.(chan struct{}))
					lockMap.Delete(key)
					return resultHolder, nil

				} else if redis_err == ErrQueryNil { //2.ErrQueryNil

					// try from origin (example form db)
					// must update ref and redis
					origin_q_err := slowQueryGetOrigin(key, resultHolder, serialization, true, slowQuery, queryDescription)

					//close and delete chan after ref set
					close(lc.(chan struct{}))
					lockMap.Delete(key)
					if origin_q_err != nil {
						return nil, origin_q_err
					} else {
						return resultHolder, nil
					}

				} else { //3.other err
					// cache other error in ref for a short time
					refSetErr(context.Background(), reference_plugin.GetInstance(), key, redis_err)

					//close and delete chan after ref set
					close(lc.(chan struct{}))
					lockMap.Delete(key)
					return nil, redis_err
				}
			}
		}
	} else {
		// after cache miss ,try from remote database
		resultHolder := resultHolderAlloc()
		origin_q_err := slowQueryGetOrigin(key, resultHolder, serialization, updateCache, slowQuery, queryDescription)
		if origin_q_err != nil {
			return nil, origin_q_err
		} else {
			return resultHolder, nil
		}
	}
}

func slowQueryGetOrigin(key string, resultHolder interface{},
	serialization bool, updateCache bool, slowQuery *SlowQuery, queryDescription string) error {
	basic.Logger.Debugln(queryDescription, " SmartQueryCacheSlow try from db query")

	query_ttl, query_err := slowQuery.Query(resultHolder)

	if query_err != nil {
		refSetErr(context.Background(), reference_plugin.GetInstance(), key, query_err)
		return query_err
	} else {
		if updateCache {
			// ref and redis update
			tokenChan := make(chan struct{}, 1)
			tokenChan <- struct{}{}
			ele := &smartCacheRefElement{
				Obj:        resultHolder,
				Token_chan: tokenChan,
			}
			rrSet(context.Background(), redis_plugin.GetInstance().ClusterClient, reference_plugin.GetInstance(), serialization, key,
				ele, query_ttl.Redis_ttl_secs+REDIS_TTL_DELAY_SEC, query_ttl.Ref_ttl_secs+REF_TTL_DELAY_SECS)
		}
		return nil
	}
}

// fastQuery usually is a redis query
func SmartQueryCacheFast(
	key string,
	resultHolderAlloc func() interface{},
	fromRefCache bool,
	updateRefCache bool,
	fastQuery func(resultHolder interface{}) (int64, error), //return ref_ttl_secs ,err
	queryDescription string) (interface{}, error) {

	if fromRefCache {
		// try to get from reference
		refElement, to_update_ref := refGet(reference_plugin.GetInstance(), key)

		if refElement != nil && !to_update_ref { // 1. ref exist and no need to update
			basic.Logger.Debugln(queryDescription + " SmartQueryCacheFast hit from reference")
			switch value := refElement.Obj.(type) {
			case error: // if error
				return nil, value
			default:
				return refElement.Obj, nil
			}
		} else if refElement != nil && to_update_ref { //2. ref exist and need update
			select {
			case <-refElement.Token_chan: //get run token
				go func() {
					defer func() {
						refElement.Token_chan <- struct{}{} // release run token
					}()

					//try from origin
					resultHolder := resultHolderAlloc()
					refCacheTTLSecs, query_err := fastQuery(resultHolder)
					if query_err != nil {
						// cache other error in ref for a short time
						refSetErr(context.Background(), reference_plugin.GetInstance(), key, query_err)
					} else {
						// ref must update
						ele := &smartCacheRefElement{
							Obj:        resultHolder,
							Token_chan: refElement.Token_chan, // use exist chan
						}
						refSetTTL(reference_plugin.GetInstance(), key, ele, refCacheTTLSecs+REF_TTL_DELAY_SECS)
					}

				}()
				return refElement.Obj, nil
			default:
				return refElement.Obj, nil
			}
		} else { //3. ref not exist
			lc, loaded := lockMap.LoadOrStore(key, make(chan struct{}, 1))
			if loaded { //most query enter this filed
				<-lc.(chan struct{}) //unblock when chan closed
				// get from ref
				refElement, _ := refGet(reference_plugin.GetInstance(), key)
				if refElement != nil {
					switch value := refElement.Obj.(type) {
					case error: // if error
						return nil, value
					default:
						return refElement.Obj, nil
					}
				} else {
					//ref nothing found which may caused by too short ref ttl which seems all most impossible
					basic.Logger.Errorln("SmartQueryCacheFast ref nothing found which may caused by too short ref ttl which seems all most impossible")
					return nil, errors.New("SmartQueryCacheFast query error")
				}
			} else { // only 1 query enter below
				//try from origin
				resultHolder := resultHolderAlloc()
				refCacheTTLSecs, query_err := fastQuery(resultHolder)
				if query_err != nil {
					// cache other error in ref for a short time
					refSetErr(context.Background(), reference_plugin.GetInstance(), key, query_err)

					//close and delete chan after ref set
					close(lc.(chan struct{})) //just close to unlock chan
					lockMap.Delete(key)
					return nil, query_err

				} else {
					// ref must update
					tokenChan := make(chan struct{}, 1)
					tokenChan <- struct{}{}
					ele := &smartCacheRefElement{
						Obj:        resultHolder,
						Token_chan: tokenChan, // a new chan
					}
					refSetTTL(reference_plugin.GetInstance(), key, ele, refCacheTTLSecs+REF_TTL_DELAY_SECS)

					//close and delete chan after ref set
					close(lc.(chan struct{})) //just close to unlock chan
					lockMap.Delete(key)
					return resultHolder, nil
				}
			}
		}
	} else {

		//after cache miss ,try from remote database
		basic.Logger.Debugln(queryDescription, " SmartQueryCacheFast try from fast query")
		///
		resultHolder := resultHolderAlloc()
		refCacheTTLSecs, query_err := fastQuery(resultHolder)
		if query_err != nil {
			refSetErr(context.Background(), reference_plugin.GetInstance(), key, query_err)
			return nil, query_err
		} else {
			if updateRefCache {
				// ref and redis update
				tokenChan := make(chan struct{}, 1)
				tokenChan <- struct{}{}
				ele := &smartCacheRefElement{
					Obj:        resultHolder,
					Token_chan: tokenChan,
				}
				refSetTTL(reference_plugin.GetInstance(), key, ele, refCacheTTLSecs+REF_TTL_DELAY_SECS)
			}
			return resultHolder, nil
		}

	}

}
