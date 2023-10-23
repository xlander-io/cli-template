package redis_pipeline

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	operation_Set              = "Set"
	operation_ZAdd             = "ZAdd"
	operation_ZRem             = "ZRem"
	operation_ZAddNX           = "ZAddNX"
	operation_HSet             = "HSet"
	operation_Expire           = "Expire"
	operation_ZRemRangeByScore = "ZRemRangeByScore"
	operation_HIncrByFloat     = "HIncrByFloat"
	operation_HIncrBy          = "HIncrBy"
	operation_IncrByFloat      = "IncrByFloat"
	operation_ZIncrBy          = "ZIncrBy"
	operation_SAdd             = "SAdd"
)

// don't use pipeline for high-safety scenario
// as redis pipeline may have data lost problem
func (pe *PipelineExecutor) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) {
	redisCmd := &PipelineCmd{
		Ctx:       ctx,
		Operation: operation_Set,
		Key:       key,
		Args:      []interface{}{value, expiration},
	}
	pe.cmdListChannel <- redisCmd
}

// don't use pipeline for high-safety scenario
// as redis pipeline may have data lost problem
func (pe *PipelineExecutor) ZRem(ctx context.Context, key string, members ...interface{}) {
	redisCmd := &PipelineCmd{
		Ctx:       ctx,
		Operation: operation_ZRem,
		Key:       key,
		Args:      []interface{}{},
	}
	redisCmd.Args = append(redisCmd.Args, members...)
	pe.cmdListChannel <- redisCmd
}

// don't use pipeline for high-safety scenario
// as redis pipeline may have data lost problem
func (pe *PipelineExecutor) ZAdd(ctx context.Context, key string, members ...*redis.Z) {
	redisCmd := &PipelineCmd{
		Ctx:       ctx,
		Operation: operation_ZAdd,
		Key:       key,
		Args:      []interface{}{},
	}
	for _, v := range members {
		redisCmd.Args = append(redisCmd.Args, v)
	}
	pe.cmdListChannel <- redisCmd
}

// don't use pipeline for high-safety scenario
// as redis pipeline may have data lost problem
func (pe *PipelineExecutor) ZAddNX(ctx context.Context, key string, members ...*redis.Z) {
	redisCmd := &PipelineCmd{
		Ctx:       ctx,
		Operation: operation_ZAddNX,
		Key:       key,
		Args:      []interface{}{},
	}
	for _, v := range members {
		redisCmd.Args = append(redisCmd.Args, v)
	}
	pe.cmdListChannel <- redisCmd
}

// don't use pipeline for high-safety scenario
// as redis pipeline may have data lost problem
func (pe *PipelineExecutor) HSet(ctx context.Context, key string, values ...interface{}) {
	redisCmd := &PipelineCmd{
		Ctx:       ctx,
		Operation: operation_HSet,
		Key:       key,
		Args:      []interface{}{},
	}
	redisCmd.Args = append(redisCmd.Args, values...)
	pe.cmdListChannel <- redisCmd
}

// don't use pipeline for high-safety scenario
// as redis pipeline may have data lost problem
func (pe *PipelineExecutor) Expire(ctx context.Context, key string, expiration time.Duration) {
	redisCmd := &PipelineCmd{
		Ctx:       ctx,
		Operation: operation_Expire,
		Key:       key,
		Args:      []interface{}{expiration},
	}
	pe.cmdListChannel <- redisCmd
}

// don't use pipeline for high-safety scenario
// as redis pipeline may have data lost problem
func (pe *PipelineExecutor) ZRemRangeByScore(ctx context.Context, key, min, max string) {
	redisCmd := &PipelineCmd{
		Ctx:       ctx,
		Operation: operation_ZRemRangeByScore,
		Key:       key,
		Args:      []interface{}{min, max},
	}
	pe.cmdListChannel <- redisCmd
}

// don't use pipeline for high-safety scenario
// as redis pipeline may have data lost problem
func (pe *PipelineExecutor) HIncrBy(ctx context.Context, key, field string, incr int64) {
	redisCmd := &PipelineCmd{
		Ctx:       ctx,
		Operation: operation_HIncrBy,
		Key:       key,
		Args:      []interface{}{field, incr},
	}
	pe.cmdListChannel <- redisCmd
}

// don't use pipeline for high-safety scenario
// as redis pipeline may have data lost problem
func (pe *PipelineExecutor) HIncrByFloat(ctx context.Context, key, field string, incr float64) {
	redisCmd := &PipelineCmd{
		Ctx:       ctx,
		Operation: operation_HIncrByFloat,
		Key:       key,
		Args:      []interface{}{field, incr},
	}
	pe.cmdListChannel <- redisCmd
}

// don't use pipeline for high-safety scenario
// as redis pipeline may have data lost problem
func (pe *PipelineExecutor) IncrByFloat(ctx context.Context, key string, incr float64) {
	redisCmd := &PipelineCmd{
		Ctx:       ctx,
		Operation: operation_IncrByFloat,
		Key:       key,
		Args:      []interface{}{incr},
	}
	pe.cmdListChannel <- redisCmd
}

// don't use pipeline for high-safety scenario
// as redis pipeline may have data lost problem
func (pe *PipelineExecutor) ZIncrBy(ctx context.Context, key string, increment float64, member string) {
	redisCmd := &PipelineCmd{
		Ctx:       ctx,
		Operation: operation_ZIncrBy,
		Key:       key,
		Args:      []interface{}{increment, member},
	}
	pe.cmdListChannel <- redisCmd
}

// don't use pipeline for high-safety scenario
// as redis pipeline may have data lost problem
func (pe *PipelineExecutor) SAdd(ctx context.Context, key string, members ...interface{}) {
	redisCmd := &PipelineCmd{
		Ctx:       ctx,
		Operation: operation_SAdd,
		Key:       key,
		Args:      members,
	}
	pe.cmdListChannel <- redisCmd
}
