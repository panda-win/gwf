package gwf

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/go-yaml/yaml"
)

type Config struct {
	// app名称
	AppName string `yaml:"appName"`
	// listen addr
	Listen string `yaml:"listen"`
	// app 版本
	AppVersion string `yaml:"appVersion"`
}

var config *Config
var initOnce sync.Once

var rootPath string

func initConfig() {
	initOnce.Do(func() {
		rPath, err := getDeployRootPath(true)
		if err != nil {
			panic(fmt.Errorf("获取项目跟目录失败：%s", err))
		}
		rootPath = rPath
		configFilename := fmt.Sprintf("%s/config/app.yaml", rootPath)
		if !fileExists(configFilename) {
			panic("app配置文件不存在 filename:" + configFilename)
		}

		c := &Config{}
		b, err := ioutil.ReadFile(configFilename)
		if err != nil {
			panic(fmt.Sprintf("读取app配置文件失败 filename:%s err:%s", configFilename, err))
		}
		err = yaml.Unmarshal(b, c)
		if err != nil {
			panic(fmt.Sprintf("解析app配置文件失败 filename:%s err:%s content:%s", configFilename, err, string(b)))
		}
		config = c
	})
}
