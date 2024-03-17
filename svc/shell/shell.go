package shell

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"nightowl/logger"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	// "golang.org/x/text/encoding/simplifiedchinese"
)

const DefaultFileMode os.FileMode = 0777

type ShellRun struct {
	WorkDir string
}
type LogAsyncHandle func(name string, msg string)

func If(cond bool, a, b interface{}) interface{} {
	if cond {
		return a
	}
	return b
}
func CopyFile(s, d string) {
	s2, err := os.Open(s)
	if err != nil {
		fmt.Sprintf("copy file src error:%s", err)
		return
	}
	d2, err := os.OpenFile(d, os.O_WRONLY|os.O_CREATE, DefaultFileMode)
	if err != nil {
		fmt.Sprintf("copy file dst error:%s", err)
		return
	}
	defer s2.Close()
	defer d2.Close()
	io.Copy(d2, s2)
}

func Copy(src, dst string) error {
	var cp func(string, string)
	cp = func(s, d string) {
		fmt.Printf("cp %s -> %s\n", s, d)
		fi, err := os.Stat(s)
		if err != nil {
			fmt.Printf("err:\n", err)
			return
		}
		//如果是文件，复制文件
		if !fi.IsDir() {
			CopyFile(s, d)
			return
		}
		//如果是目录
		files, err := ioutil.ReadDir(s) //读取目录下文件
		if err != nil {
			fmt.Printf("err:\n", err)
			return
		}
		//创建目标目录
		err = os.MkdirAll(d, DefaultFileMode)
		if err != nil {
			fmt.Printf("err:\n", err)
			return
		}
		for _, file := range files {
			s2 := fmt.Sprintf("%s/%s", s, file.Name())
			d2 := fmt.Sprintf("%s/%s", d, file.Name())
			cp(s2, d2)
		}
	}
	cp(src, dst)
	return nil
}

func (ss ShellRun) ReadAsyncShellLog(reader io.Reader, handler func(string)) error {
	var cache string = ""
	buf := make([]byte, 1024, 1024)
	for {
		num, err := reader.Read(buf)
		// logger.Debug("ReadAsyncShellLog:",num,err)
		if err != nil {
			if len(cache) > 0 && handler != nil {

				handler(cache)
			}
			if err == io.EOF || strings.Contains(err.Error(), "closed") {
				err = nil
			}
			return err
		}
		if num > 0 {
			d := buf[:num]
			// d,_ = simplifiedchinese.GB18030.NewDecoder().Bytes(d)
			a := strings.Split(string(d), "\n")
			line := strings.Join(a[:len(a)-1], "\n")
			if handler != nil {
				handler(fmt.Sprintf("%s%s\n", cache, line))
			}
			// fmt.Printf("|%s%s\n", cache, line)
			cache = a[len(a)-1]
		}
	}
	return nil
}
func (ss ShellRun) RunAsyncShell(name string, args []string, handler LogAsyncHandle) error {
	cmd := exec.Command(name, args...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if ss.WorkDir != "" {
		cmd.Dir = ss.WorkDir
	}

	if err := cmd.Start(); err != nil {
		logger.Error("Error starting command: %s......", err.Error())
		return err
	}

	go ss.ReadAsyncShellLog(stdout, func(log string) {
		if handler != nil {
			handler(name, log)
		}
	})
	go ss.ReadAsyncShellLog(stderr, func(log string) {
		if handler != nil {
			handler(name, log)
		}
	})
	if err := cmd.Wait(); err != nil {
		log.Printf("Error waiting for command execution: %s......", err.Error())
		return err
	}
	return nil
}

type SshLoginHandle func(client ssh.Client)
type SshLoginArgs struct {
	Host string
	Port uint32
	User string
	Pwd  string
}

func (ss ShellRun) SshLogin(args SshLoginArgs, sshLoginHandle SshLoginHandle) {

	//创建sshp登陆配置
	config := &ssh.ClientConfig{
		// Timeout:         30, //ssh 连接time out 时间一秒钟, 如果ssh验证错误 会在一秒内返回
		User:            args.User,
		Auth:            []ssh.AuthMethod{ssh.Password(args.Pwd)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以, 但是不够安全
		//HostKeyCallback: hostKeyCallBackFunc(h.Host),
	}

	// config.Auth = []ssh.AuthMethod{publicKeyAuthFunc(sshKeyPath)}

	//dial 获取ssh client
	addr := fmt.Sprintf("%s:%d", args.Host, args.Port)
	logger.Debug(fmt.Sprintf("login %s@%s", args.User, addr))
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		logger.Error("create ssh client error:", err)
	}
	defer sshClient.Close()

	if sshLoginHandle != nil {
		sshLoginHandle(*sshClient)
	}

	// //执行远程命令
	// combo, err := session.CombinedOutput("whoami;uptime;df -h; free -m;top -bn 3 | grep 'Cpu'")
	// if err != nil {
	// 	log.Fatal("远程执行cmd 失败", err)
	// }
	// log.Println("命令输出:", string(combo))
}
func (ss ShellRun) RunWithSshSession(s ssh.Session, cmd string, args []string, handler LogAsyncHandle) error {
	stdout, _ := s.StdoutPipe()
	stderr, _ := s.StderrPipe()
	// if ss.WorkDir != ""{
	// 	cmd.Dir = ss.WorkDir
	// }
	cmds := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	// cmds = "ifconfig"
	if err := s.Start(cmds); err != nil {
		logger.Error("|run cmd:", cmds, "|error: ", err.Error())
		return err
	}

	go ss.ReadAsyncShellLog(stdout, func(log string) {
		if handler != nil {
			handler("", log)
		}
	})
	go ss.ReadAsyncShellLog(stderr, func(log string) {
		if handler != nil {
			handler("", log)
		}
	})
	if err := s.Wait(); err != nil {
		logger.Error("Error waiting for command:", cmds, "|error:", err.Error())
		return err
	}
	return nil
}
func (ss ShellRun) SshLoginAndRun(sshArgs SshLoginArgs, cmd string, args []string, handler LogAsyncHandle) {
	r := rand.Intn(100)
	ss.SshLogin(sshArgs, func(sshClient ssh.Client) {
		// logger.Debug("args:", sshArgs)
		//创建ssh-session
		session, err := sshClient.NewSession()
		if err != nil {
			logger.Error("create client session error:", err)
		}
		defer session.Close()
		ss.RunWithSshSession(*session, cmd, args, handler)
		logger.Debug("===RunWithSshSession-END:", r)
	})
	logger.Debug("===SshLoginAndRun-END:", r)
}

func (ss ShellRun) Scp(args SshLoginArgs, src string, dst string) {

	logger.Debug("scp", src, dst)
	ss.SshLogin(args, func(sshClient ssh.Client) {
		sftp, err := sftp.NewClient(&sshClient)
		if err != nil {
			logger.Error("sftp.NewClient error:", err)
			return
		}
		//
		scpFile := func(s, d string) {
			logger.Debug("scpFile", s, d)
			sf, err := os.Open(s)
			if nil != err {
				return
			}
			defer sf.Close()
			a := strings.Split(d, "/")
			err = sftp.MkdirAll(strings.Join(a[:len(a)-1], "/"))

			dfi, err := sftp.Stat(d)
			if nil == err && dfi.IsDir() {
				a = strings.Split(s, "/")
				d = fmt.Sprintf("%s/%s", d, a[len(a)-1])
			}
			df, err := sftp.Create(d)
			if nil != err {
				logger.Error("sftp.Create error:", err)
				return
			}
			defer df.Close()
			io.Copy(df, sf)
		}
		mkdir := func(d string) {
			logger.Debug("sftp.mkdir:", d)
			session, err := sshClient.NewSession()
			if err != nil {
				logger.Error("sshClient.NewSession error:", err)
				return
			}
			defer session.Close()
			err = ss.RunWithSshSession(*session, "mkdir", []string{"-p", d}, nil)
			if err != nil {
				logger.Error("RunWithSshSession error:", err)
				return

			}
		}
		var scp func(s, d string)
		scp = func(s, d string) {
			logger.Debug("scp(children) ", s, "->", d)
			fi, err := os.Stat(s)
			if err != nil {
				logger.Error("os.Stat error:", err)
				return
			}
			//如果是文件，复制文件
			if !fi.IsDir() {
				scpFile(s, d)
				return
			}
			//如果是目录
			files, err := ioutil.ReadDir(s) //读取目录下文件

			if err != nil {
				logger.Error("ioutil.ReadDir error:", err)
				return
			}
			//创建目标目录
			mkdir(d)
			for _, file := range files {
				s2 := fmt.Sprintf("%s/%s", s, file.Name())
				d2 := fmt.Sprintf("%s/%s", d, file.Name())
				scp(s2, d2)
			}
		}
		scp(src, dst)
	})
}
