package smart_cache

import (
	"context"

	"github.com/coreservice-io/cli-template/basic"
	"github.com/coreservice-io/cli-template/plugin/redis_plugin"
	"github.com/coreservice-io/cli-template/plugin/reference_plugin"
)

// for Query ,return QueryNilErr if Query result is nil  -> as set nil to cache is not supported
func SmartQuery(key string, resultHolderAlloc func() interface{}, serialization bool, fromCache bool, updateCache bool, cacheTTLSecs int64, Query func(resultHolder interface{}) error, queryDescription string) (interface{}, error) {

	var resultHolder interface{}

	if fromCache {
		// try to get from reference
		result := Ref_Get(reference_plugin.GetInstance(), key)
		if result != nil {
			basic.Logger.Debugln(queryDescription + " hit from reference")
			return result, nil
		}

		resultHolder = resultHolderAlloc()

		err := Redis_Get(context.Background(), redis_plugin.GetInstance().ClusterClient, serialization, key, resultHolder)
		if err == nil {
			basic.Logger.Debugln(queryDescription, " hit from redis")
			Ref_Set(reference_plugin.GetInstance(), key, resultHolder)
			return resultHolder, nil
		} else if err == CacheNilErr {
			//continue to get from db part
		} else if err == QueryNilErr {
			return nil, QueryNilErr
		} else if err == QueryErr {
			//this happens when query failed
			basic.Logger.Errorln(queryDescription, " QueryErr")
			return nil, QueryErr
		} else {
			//redis may broken, just return to make slow query safe
			return nil, err
		}
	}

	//after cache miss ,try from remote database
	basic.Logger.Debugln(queryDescription, " try from query")

	if resultHolder == nil {
		resultHolder = resultHolderAlloc()
	}
	err := Query(resultHolder)

	if err != nil {
		if err == QueryNilErr {
			RR_SetQueryNilErr_TTL(context.Background(), redis_plugin.GetInstance().ClusterClient, key, cacheTTLSecs)
			return nil, QueryNilErr
		} else {
			basic.Logger.Errorln(queryDescription, " Query err :", err)
			RR_SetQueryErr(context.Background(), redis_plugin.GetInstance().ClusterClient, key)
			return nil, QueryErr
		}
	} else {
		if updateCache {
			RR_Set(context.Background(), redis_plugin.GetInstance().ClusterClient, reference_plugin.GetInstance(), serialization, key, resultHolder, cacheTTLSecs)
		}
		return resultHolder, nil
	}
}
