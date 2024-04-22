package db

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Playbook struct {
	Database []Database `yaml:"db"`
}

type Database struct {
	DbName   string `yaml:"name"`
	Dbtype   string `yaml:"dbtype"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// 解析yaml文件
func DbConfig(p string) (*Playbook, error) {
	file, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}

	c := Playbook{}
	if err = yaml.Unmarshal(file, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

// 提取db连接信息
func (c *Playbook) GetDb(DbName string) (*Database, error) {
	for _, db := range c.Database {
		if db.DbName == DbName {
			return &db, nil
		}
	}
	return nil, fmt.Errorf("db named %s not found", DbName)
}
