package captcha

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/coreservice-io/cli-template/basic"
	"github.com/coreservice-io/cli-template/plugin/redis_plugin"
	"github.com/coreservice-io/utils/rand_util"
	goredis "github.com/go-redis/redis/v8"
)

const vcode_len = 16 //16 is safe ,don't make this shorter then 16
const vcode_expire_secs = 600
const redis_vcode_prefix = "vcode"

// send vcode to user
func GenVCode(vCodeKey string) (string, error) {
	key := redis_plugin.GetInstance().GenKey(redis_vcode_prefix, vCodeKey)
	code, _ := redis_plugin.GetInstance().Get(context.Background(), key).Result()
	if code == "" {
		code = rand_util.GenRandStr(vcode_len)
	}
	_, err := redis_plugin.GetInstance().Set(context.Background(), key, code, vcode_expire_secs*time.Second).Result()
	if err != nil {
		basic.Logger.Errorln("GenVCode set email vcode to redis error", "err", err)
		return "", errors.New("set email vcode error")
	}

	basic.Logger.Debugln("vcode", "code", code, "vCodeKey", vCodeKey)
	return code, nil
}

func ValidateVCode(vCodeKey string, code string) bool {
	//incase user have Cap or whitespace
	code = strings.ToLower(code)
	code = strings.TrimSpace(code)

	key := redis_plugin.GetInstance().GenKey(redis_vcode_prefix, vCodeKey)
	value, err := redis_plugin.GetInstance().Get(context.Background(), key).Result()
	if err == goredis.Nil {
		return false
	} else if err != nil {
		basic.Logger.Debugln("ValidateVCode from redis err", "err", err, "vCodeKey", vCodeKey)
		return false
	}
	redis_plugin.GetInstance().Del(context.Background(), key)
	return value == code
}
