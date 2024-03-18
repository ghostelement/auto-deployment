package deploy

import (
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Playbook struct {
	Jobs []Job `yaml:"job"`
}

type Job struct {
	JobName     string   `yaml:"name"`
	Hosts       []string `yaml:"host"`
	User        string   `yaml:"user"`
	Password    string   `yaml:"password"`
	ParallelNum int      `yaml:"parallelNum"`
	SrcFile     string   `yaml:"srcFile"`
	DestDir     string   `yaml:"destDir"`
	Cmd         string   `yaml:"cmd"`
	Shell       string   `yaml:"shell"`
}

// 解析yaml文件
func Config(p string) (*Playbook, error) {
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

// 提取任务Job
func (c *Playbook) GetJob(jobName string) (*Job, error) {
	for _, job := range c.Jobs {
		if err := job.Validate(); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// 验证配置文件
func (c *Job) Validate() error {
	if len(c.Hosts) == 0 {
		return errors.New("address can't be empty")
	}
	if c.User == "" {
		return errors.New("username can't be empty")
	}
	if c.Password == "" {
		return errors.New("password and publicKey can't be empty at the same time")
	}

	// check if the srcFiles exists
	if c.SrcFile != "" {
		for _, filepath := range strings.Split(c.SrcFile, ",") {
			if _, err := os.Stat(filepath); err != nil {
				if os.IsNotExist(err) {
					return errors.New(filepath + " not exists")
				}
				return err
			}
		}
	}

	return nil
}

// ExampleConfig Config
/*
func ExampleConfig() *Playbook {
	job := &Job{
		JobName:  "init",
		Hosts:    hosts{"host1", "host2"},
		User:     "root",
		Password: "password",
		SrcFile:  "/opt/exm.txt",
		DestDir:  "/opt/",
		Cmd:      "ls -l",
		Shell:    "pwd",
	}
	return &Playbook{
		Jobs: []*Job,
	}
}

// ExampleAllConfig Config
func ExampleAllConfig() *Playbook {
	return &Playbook{
		Addr:          "host1:port1,host2:port2,...",
		User:          "username",
		Pass:          "password",
		PublicKey:     "ssh public key",
		Timeout:       5,
		SrcFile:       "file1,file2,...",
		WorkDir:       "/path/to/remote/dir",
		ChangeWorkDir: true,
		PreCmd:        []string{"cmd1", "cmd2", "..."},
		PostCmd:       []string{"cmd1", "cmd2", "..."},
	}
}
*/
