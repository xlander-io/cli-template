package redis_pipeline

import (
	"context"
	"math/rand"
	"time"

	"github.com/coreservice-io/cli-template/basic"
	"github.com/coreservice-io/cli-template/plugin/redis_plugin"
	"github.com/coreservice-io/job"

	"github.com/go-redis/redis/v8"
)

const thread_count = 5
const channel_limit = 6000

const cmd_count_limit = 60
const interval_limit_millisec = 2500

type PipelineCmd struct {
	Ctx       context.Context
	Operation string
	Key       string
	Args      []interface{}
}

type PipelineExecutor struct {
	cmdListChannel           chan *PipelineCmd
	last_exec_time_unixmilli int64

	exec_count_limit             int
	exec_interval_limit_millisec int64
	cmd_channel_limit            int
}

var pipelineExecutors []*PipelineExecutor = []*PipelineExecutor{}

func InitPipelineExecutors() {
	for {
		if len(pipelineExecutors) >= thread_count {
			break
		}

		pe := &PipelineExecutor{
			cmdListChannel:           make(chan *PipelineCmd, channel_limit),
			last_exec_time_unixmilli: time.Now().UTC().UnixMilli(),

			exec_count_limit:             cmd_count_limit,
			exec_interval_limit_millisec: interval_limit_millisec,
			cmd_channel_limit:            channel_limit,
		}
		if err := pe.startPipelineExec(); err == nil {
			pipelineExecutors = append(pipelineExecutors, pe)
		}
	}
}

func GetPipeline() *PipelineExecutor {
	return pipelineExecutors[rand.Intn(thread_count)]
}

func (pe *PipelineExecutor) startPipelineExec() error {
	const jobName = "ScheduleRedisPipelineExec"
	err := job.Start(
		context.Background(),
		job.JobConfig{
			Name:                jobName,
			Job_type:            job.TYPE_PANIC_REDO,
			Interval_secs:       1,
			Chk_before_start_fn: nil,
			Process_fn: func(j *job.Job) {
				for {
					if len(pe.cmdListChannel) < pe.exec_count_limit && time.Now().UTC().UnixMilli()-pe.last_exec_time_unixmilli < pe.exec_interval_limit_millisec {
						time.Sleep(250 * time.Millisecond)
						continue
					}
					pe.exec()
				}
			},
			On_panic: func(j *job.Job, panic_err interface{}) {
				basic.Logger.Errorln(jobName, panic_err)
			},
			Panic_sleep_secs: 1,
			Final_fn:         nil,
		},
		nil,
	)
	if err != nil {
		basic.Logger.Errorln("ScheduleRedisPipelineExec start err:", err)
		return err
	}
	return nil
}

func (pe *PipelineExecutor) exec() {

	pe.last_exec_time_unixmilli = time.Now().UTC().UnixMilli()

	pl := redis_plugin.GetInstance().Pipeline()

outLoop:
	for i := 0; i < pe.exec_count_limit; i++ {
		select {
		case cmd := <-pe.cmdListChannel:
			switch cmd.Operation {
			case operation_Set:
				pl.Set(cmd.Ctx, cmd.Key, cmd.Args[0], cmd.Args[1].(time.Duration))

			case operation_ZRem:
				z := []interface{}{}
				z = append(z, cmd.Args...)
				pl.ZRem(cmd.Ctx, cmd.Key, z...)

			case operation_ZAdd:
				z := []*redis.Z{}
				for _, v := range cmd.Args {
					z = append(z, v.(*redis.Z))
				}
				pl.ZAdd(cmd.Ctx, cmd.Key, z...)

			case operation_ZAddNX:
				z := []*redis.Z{}
				for _, v := range cmd.Args {
					z = append(z, v.(*redis.Z))
				}
				pl.ZAddNX(cmd.Ctx, cmd.Key, z...)

			case operation_HSet:
				pl.HSet(cmd.Ctx, cmd.Key, cmd.Args...)

			case operation_Expire:
				pl.Expire(cmd.Ctx, cmd.Key, cmd.Args[0].(time.Duration))

			case operation_ZRemRangeByScore:
				pl.ZRemRangeByScore(cmd.Ctx, cmd.Key, cmd.Args[0].(string), cmd.Args[1].(string))

			case operation_HIncrBy:
				pl.HIncrBy(cmd.Ctx, cmd.Key, cmd.Args[0].(string), cmd.Args[1].(int64))

			case operation_HIncrByFloat:
				pl.HIncrByFloat(cmd.Ctx, cmd.Key, cmd.Args[0].(string), cmd.Args[1].(float64))

			case operation_IncrByFloat:
				pl.IncrByFloat(cmd.Ctx, cmd.Key, cmd.Args[0].(float64))

			case operation_ZIncrBy:
				pl.ZIncrBy(cmd.Ctx, cmd.Key, cmd.Args[0].(float64), cmd.Args[1].(string))

			case operation_SAdd:
				pl.SAdd(cmd.Ctx, cmd.Key, cmd.Args...)

			default:
				basic.Logger.Errorln("unsupported cmd:", cmd.Operation)
			}

		default:
			break outLoop
		}
	}

	if pl.Len() == 0 {
		return
	}

	_, err := pl.Exec(context.Background())
	if err != nil {
		basic.Logger.Errorln("exec pipeline error:", err)
		time.Sleep(5 * time.Second) // sleep a while for exe err
		return
	}
}
