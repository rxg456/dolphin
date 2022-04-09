package config

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/rxg456/dolphin/api/models"
	"gopkg.in/yaml.v2"
)

type Config struct {
	RpcServerAddr string                `yaml:"rpc_server_addr"`
	HttpAddr      string                `yaml:"http_addr"`
	Job           JobSection            `yaml:"job"`
	LogStrategies []*models.LogStrategy `yaml:"log_strategies"`
	EnableLogJob  bool                  `yaml:"enable_log_job"`
}

type JobSection struct {
	MetaDir  string `yaml:"metadir"`
	Interval int    `yaml:"interval"`
}

// 根据io read读取配置文件后的字符串解析yaml
func Load(s []byte) (*Config, error) {
	cfg := &Config{}

	err := yaml.Unmarshal(s, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// 根据conf路径读取内容
func LoadFile(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg, err := Load(content)
	if err != nil {
		fmt.Printf("[parsing Yaml file err...][err:%v]\n", err)
		return nil, err
	}
	return cfg, nil
}

// 解析用户配置的日志策略正则
func SetLogRegs(input []*models.LogStrategy) []*models.LogStrategy {
	res := []*models.LogStrategy{}
	for _, st := range input {
		st := st
		st.TagRegs = make(map[string]*regexp.Regexp)
		// 处理主正则
		if len(st.Pattern) != 0 {
			reg, err := regexp.Compile(st.Pattern)
			if err != nil {
				fmt.Printf("compile pattern regexp failed:[name:%v][pat:%v][err:%v]\n",
					st.MetricName,
					st.Pattern,
					err,
				)
				continue
			}
			st.PatternReg = reg
		}
		// 处理标签的正则
		for tagK, tagv := range st.Tags {
			reg, err := regexp.Compile(tagv)
			if err != nil {
				fmt.Printf("compile pattern regexp failed:[name:%v][pat:%v][err:%v]\n",
					st.MetricName,
					tagv,
					err,
				)
				continue
			}
			st.TagRegs[tagK] = reg
		}
		res = append(res, st)
	}
	return res
}
