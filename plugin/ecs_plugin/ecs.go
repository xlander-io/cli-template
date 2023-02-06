package ecs_plugin

import (
	"fmt"
	"time"

	elasticSearch "github.com/olivere/elastic/v7"
)

var instanceMap = map[string]*elasticSearch.Client{}

func GetInstance() *elasticSearch.Client {
	return instanceMap["default"]
}

func GetInstance_(name string) *elasticSearch.Client {
	return instanceMap[name]
}

/*
elasticSearchAddr
elasticSearchUserName
elasticSearchPassword
*/
type Config struct {
	Address  string
	UserName string
	Password string
}

func Init(esConfig *Config) error {
	return Init_("default", esConfig)
}

//  Init a new instance.
//  If only need one instance, use empty name "". Use GetDefaultInstance() to get.
//  If you need several instance, run Init() with different <name>. Use GetInstance(<name>) to get.
func Init_(name string, esConfig *Config) error {
	if name == "" {
		name = "default"
	}

	_, exist := instanceMap[name]
	if exist {
		return fmt.Errorf("ecs instance <%s> has already been initialized", name)
	}

	es, err := elasticSearch.NewClient(
		elasticSearch.SetURL(esConfig.Address),
		elasticSearch.SetBasicAuth(esConfig.UserName, esConfig.Password),
		elasticSearch.SetSniff(false),
		elasticSearch.SetHealthcheckInterval(30*time.Second),
		elasticSearch.SetGzip(true),
	)
	if err != nil {
		return err
	}
	instanceMap[name] = es
	return nil
}
