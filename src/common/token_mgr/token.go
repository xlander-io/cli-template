package token_mgr

import (
	"errors"
	"strings"

	"github.com/coreservice-io/cli-template/basic"
	"github.com/coreservice-io/cli-template/basic/config"
	"github.com/coreservice-io/utils/token_util"
)

var TokenMgr *token_util.TokenUtil

func InitTokenMgr(toml_conf *config.TomlConfig) error {
	if toml_conf.Token.Salt == "" || strings.Trim(toml_conf.Token.Salt, "") == "" {
		return errors.New("empty token salt now allowed")
	}
	TokenMgr = token_util.NewTokenUtil(toml_conf.Token.Salt)
	basic.Logger.Debugln("TokenMgr init success:", TokenMgr)
	return nil
}
