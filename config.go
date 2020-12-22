package main

import (
	"fmt"
	"regexp"
	"strconv"
	"os"

	"github.com/tkanos/gonfig"
)

var (
	config        zookeeperExporterConfig
	defaultConfig = zookeeperExporterConfig{
		ZkHost:            "127.0.0.1:2181",
		Timeout:            30,
		PublishPort:        "9419",
		PublishAddr:        "",
		OutputFormat:       "TTY", //JSON
		EnabledExporters:   []string{"node", "overview"},
		SubSystemName:      "",
		SubSystemID:        "",
		ClusterName: 		"",
		//ExtraLabels:		[]map[string]string{},
	}
)

type zookeeperExporterConfig struct {
	ZkHost                	 string              `json:"zk_host"`
	Timeout                  int                 `json:"timeout"`
	PublishPort              string              `json:"publish_port"`
	PublishAddr              string              `json:"publish_addr"`
	OutputFormat             string              `json:"output_format"`
	EnabledExporters		 []string            `json:"enabled_exporters"`
	SubSystemName            string              `json:"sub_system_name"`
	SubSystemID              string              `json:"sub_system_id"`
	ClusterName				 string				 `json:"cluster_name"`
	//ExtraLabels            []map[string]string `json:"extra_labels"`
}

func initConfigFromFile(config_file string) error {
	config = zookeeperExporterConfig{}
	err := gonfig.GetConf(config_file, &config)
	if err != nil {
		return err
	}

	return nil
}

func initConfig() {
	config = defaultConfig
	if host := os.Getenv("ZK_HOST"); host != "" {
		if valid, _ := regexp.MatchString("[:.0-9]+", host); valid {
			config.ZkHost = host
		} else {
			panic(fmt.Errorf("Rabbit URL must start with http:// or https://"))
		}
	}

	if port := os.Getenv("PUBLISH_PORT"); port != "" {
		if _, err := strconv.Atoi(port); err == nil {
			config.PublishPort = port
		} else {
			panic(fmt.Errorf("The configured port is not a valid number: %v", port))
		}

	}

	if addr := os.Getenv("PUBLISH_ADDR"); addr != "" {
		config.PublishAddr = addr
	}

	if output := os.Getenv("OUTPUT_FORMAT"); output != "" {
		config.OutputFormat = output
	}

	if timeout := os.Getenv("TIMEOUT"); timeout != "" {
		t, err := strconv.Atoi(timeout)
		if err != nil {
			panic(fmt.Errorf("timeout is not a number: %v", err))
		}
		config.Timeout = t
	}


	if subSystemName := os.Getenv("SUB_SYSTEM_NAME"); subSystemName != "" {
		config.SubSystemName = subSystemName
	}

	if subSystemID := os.Getenv("SUB_SYSTEM_ID"); subSystemID != "" {
		config.SubSystemID = subSystemID
	}

	if ClusterName := os.Getenv("CLUSTER_NAME"); ClusterName != "" {
		config.ClusterName = ClusterName
	}
}