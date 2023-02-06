package component

import (
	"github.com/coreservice-io/cli-template/basic"
	"github.com/coreservice-io/cli-template/basic/config"
	"github.com/coreservice-io/cli-template/plugin/spr_plugin"
	"github.com/coreservice-io/redis_spr"
)

func InitSpr(toml_conf *config.TomlConfig) error {

	if toml_conf.Spr.Enable {

		spr_redis_conf := redis_spr.RedisConfig{
			Addr:     toml_conf.Redis.Host,
			UserName: toml_conf.Redis.Username,
			Password: toml_conf.Redis.Password,
			Port:     toml_conf.Redis.Port,
			Prefix:   toml_conf.Redis.Prefix,
			UseTLS:   toml_conf.Redis.Use_tls,
		}

		basic.Logger.Infoln("Init spr plugin with config:", spr_redis_conf)
		if err := spr_plugin.Init(&spr_redis_conf, basic.Logger); err == nil {
			basic.Logger.Infoln("### InitSpr success")
			return nil
		} else {
			return err
		}
	} else {
		return nil
	}

}
