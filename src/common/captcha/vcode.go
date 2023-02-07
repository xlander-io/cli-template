package captcha

import (
	"context"
	"errors"
	"time"

	"github.com/coreservice-io/cli-template/basic"
	"github.com/coreservice-io/cli-template/plugin/redis_plugin"
	"github.com/coreservice-io/utils/rand_util"
	goredis "github.com/go-redis/redis/v8"
)

const redis_vcode_prefix = "vcode"

// send vcode to user
func GenVCode(vCodeKey string) (string, error) {
	key := redis_plugin.GetInstance().GenKey(redis_vcode_prefix, vCodeKey)
	code, _ := redis_plugin.GetInstance().Get(context.Background(), key).Result()
	if code == "" {
		code = rand_util.GenRandStr(4)
	}
	_, err := redis_plugin.GetInstance().Set(context.Background(), key, code, 4*time.Hour).Result()
	if err != nil {
		basic.Logger.Errorln("GenVCode set email vcode to redis error", "err", err)
		return "", errors.New("set email vcode error")
	}

	basic.Logger.Debugln("vcode", "code", code, "vCodeKey", vCodeKey)
	return code, nil
}

func ValidateVCode(vCodeKey string, code string) bool {
	key := redis_plugin.GetInstance().GenKey(redis_vcode_prefix, vCodeKey)
	value, err := redis_plugin.GetInstance().Get(context.Background(), key).Result()
	if err == goredis.Nil {
		return false
	} else if err != nil {
		basic.Logger.Debugln("ValidateVCode from redis err", "err", err, "vCodeKey", vCodeKey)
		return false
	}

	if value == code {
		return true
	}
	return false
}

func ClearVCode(vCodeKey string) {
	key := redis_plugin.GetInstance().GenKey(redis_vcode_prefix, vCodeKey)
	redis_plugin.GetInstance().Del(context.Background(), key)
}
