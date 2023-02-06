package component

import (
	"github.com/coreservice-io/cli-template/basic"
	"github.com/coreservice-io/cli-template/plugin/reference_plugin"
)

func InitReference() error {
	if err := reference_plugin.Init(); err == nil {
		basic.Logger.Infoln("### InitReference success")
		return nil
	} else {
		return err
	}
}
