package db

import (
	"auto-deployment/logger"
	"errors"
	"fmt"
	"os"
	"strings"

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

// 连接db,根据dbname数据库类型连接db
func ConnetDb(file string, dbname string) error {
	//解析yaml
	text, err := DbConfig(file)
	if err != nil {
		return err
	}
	// 获取db信息
	dbInfo, err := text.GetDb(dbname)
	if err != nil {
		return err
	}
	// 校验db信息
	err = dbInfo.Validate()
	if err != nil {
		logger.Error("|Error playbook:", file, "|Error: ", err.Error())
		return err
	}
	switch dbInfo.Dbtype {
	case "mysql":
		return ConnMysqlAndRun(dbInfo)
	default:
		return fmt.Errorf("db type %s not supported", dbInfo.Dbtype)
	}
}

// 检查sql脚本，过滤掉注释
func checkSql(sql string) bool {
	check := true
	if strings.HasPrefix(sql, "/*") && strings.HasSuffix(sql, "*/") {
		check = false
	}
	if strings.HasPrefix(sql, "--") {
		check = false
	}

	return check
}

// 验证配置文件
func (c *Database) Validate() error {
	if c.DbName == "" {
		return errors.New("db name can't be empty")
	}
	if c.Host == "" {
		return errors.New("host can't be empty")
	}
	if c.Port == "" {
		return errors.New("port can't be empty")
	}
	if c.Dbtype == "" {
		return errors.New("dbtype can't be empty")
	}
	if c.User == "" {
		return errors.New("username can't be empty")
	}
	if c.Password == "" {
		return errors.New("password and publicKey can't be empty at the same time")
	}

	return nil
}
