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

	// TODO: add oracle sqlserver连接器
	//根据数据库类型连接db，如果是sql类可以复用mysql函数，只需引入不同的驱动
	switch dbInfo.Dbtype {
	case "mysql":
		datasource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbInfo.User, dbInfo.Password, dbInfo.Host, dbInfo.Port, dbInfo.Database)
		return ConnMysqlAndRun("mysql", datasource)
	case "postgres":
		datasource := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbInfo.User, dbInfo.Password, dbInfo.Host, dbInfo.Port, dbInfo.Database)
		return ConnMysqlAndRun("postgres", datasource)
	//case "oracle":
	//	datasource := fmt.Sprintf("oracle://%s:%s@%s:%s/%s", dbInfo.User, dbInfo.Password, dbInfo.Host, dbInfo.Port, dbInfo.Database)
	//	return ConnMysqlAndRun("oracle", datasource)
	case "redis":
		c, err := dbInfo.OpenRedisConnection()
		if err != nil {
			fmt.Printf("connection failed:%v\n", err)
			os.Exit(1)
		}
		dbInfo.InputReader(c)
	default:
		return fmt.Errorf("db type %s not supported", dbInfo.Dbtype)
	}
	return nil
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
	//if c.Password == "" {
	//	return errors.New("password and publicKey can't be empty at the same time")
	//}

	return nil
}
