package geo_ip_plugin

import (
	"fmt"

	"github.com/coreservice-io/geo_ip/lib"
)

type GeoIp struct {
	lib.GeoIpInterface
}

var instanceMap = map[string]*GeoIp{}

func GetInstance() *GeoIp {
	return instanceMap["default"]
}

func GetInstance_(name string) *GeoIp {
	return instanceMap[name]
}

func Init(version string, dataset_folder string, logger func(string), err_logger func(string)) error {
	return Init_("default", version, dataset_folder, logger, err_logger)
}

func Init_(name string, version string, dataset_folder string, logger func(string), err_logger func(string)) error {
	if name == "" {
		name = "default"
	}

	_, exist := instanceMap[name]
	if exist {
		return fmt.Errorf("ip_geo instance <%s> has already initialized", name)
	}

	ipClient := &GeoIp{}
	// new instance
	client, err := lib.NewClient(version, dataset_folder, logger, err_logger)
	if err != nil {
		return err
	}
	ipClient.GeoIpInterface = client

	instanceMap[name] = ipClient
	return nil
}
