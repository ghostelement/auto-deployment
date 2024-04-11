package shell

import (
	"auto-deployment/logger"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"path"
	"path/filepath"
	"sync"

	"math/rand"
	"os"
	"strings"

	"github.com/pkg/sftp"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
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

type ProgressUpdateEvent struct {
	BytesCopied int64
	TotalBytes  int64
	Message     string
	Status      string
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
	buf := make([]byte, 8192, 8192)
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

type SshLoginHandle func(client ssh.Client)
type SshLoginArgs struct {
	Host     string
	Port     uint32
	User     string
	Password string
}

func (ss ShellRun) SshLogin(args SshLoginArgs, sshLoginHandle SshLoginHandle) error {

	//创建sshp登陆配置
	config := &ssh.ClientConfig{
		//Timeout:         30, //ssh 连接time out 时间一秒钟, 如果ssh验证错误 会在一秒内返回
		User:            args.User,
		Auth:            []ssh.AuthMethod{ssh.Password(args.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// config.Auth = []ssh.AuthMethod{publicKeyAuthFunc(sshKeyPath)}

	//dial 获取ssh client
	//addr := fmt.Sprintf("%s:%d", args.Host, args.Port)
	logger.Debug(fmt.Sprintf("login %s@%s", args.User, args.Host))
	sshClient, err := ssh.Dial("tcp", args.Host, config)
	if err != nil {
		logger.Error("create ssh client error:", err)
		return err
	}
	defer sshClient.Close()

	if sshLoginHandle != nil {
		sshLoginHandle(*sshClient)
	}
	return nil
}
func (ss ShellRun) RunWithSshSession(host string, s ssh.Session, cmd string, args []string, handler LogAsyncHandle) error {
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
			handler(host, log)
		}
	})
	go ss.ReadAsyncShellLog(stderr, func(log string) {
		if handler != nil {
			handler(host, log)
		}
	})
	if err := s.Wait(); err != nil {
		logger.Error("Error waiting for command:", cmds, "|error:", err.Error())
		return err
	}
	return nil
}
func (ss ShellRun) SshLoginAndRun(sshArgs SshLoginArgs, cmd string, args []string, handler LogAsyncHandle) error {
	r := rand.Intn(100)
	err := ss.SshLogin(sshArgs, func(sshClient ssh.Client) {

		// logger.Debug("args:", sshArgs)
		//创建ssh-session
		session, err := sshClient.NewSession()
		if err != nil {
			logger.Error("create client session error:", err)
		}
		defer session.Close()
		ss.RunWithSshSession(sshArgs.Host, *session, cmd, args, handler)
		logger.Debug("===RunWithSshSession-END:", r)
	})
	logger.Debug("===SshLoginAndRun-END:", r)
	return err
}

// test goroutine
type ScpProgressLock struct {
	mutex   sync.Mutex
	bar     *mpb.Bar
	proxied io.Reader
}

func (lock *ScpProgressLock) Read(p []byte) (n int, err error) {
	n, err = lock.proxied.Read(p)
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	lock.bar.IncrBy(n)
	return n, err
}

func (ss ShellRun) Scp(args SshLoginArgs, src string, dst string, p *mpb.Progress, displayBar bool) error {
	logger.Debug("scp", src, dst)

	err := ss.SshLogin(args, func(sshClient ssh.Client) {

		sftp, err := sftp.NewClient(&sshClient)
		if err != nil {
			logger.Error("sftp.NewClient error:", err)
			return
		}

		// sftp拷贝文件
		scpFile := func(srcPath, dstPath string) error {
			logger.Debug("scpFile", srcPath, dstPath)

			// 创建目标路径的上级目录
			dirName := path.Dir(dstPath)
			err = sftp.MkdirAll(dirName)
			if err != nil {
				logger.Error("sftp.MkdirAll error:", err)
				return err
			}

			sf, err := os.Open(srcPath)
			if nil != err {
				logger.Error("os.Open error:", err)
				return err
			}
			defer sf.Close()

			// 检查远程文件是否存在并决定是否拷贝
			if fileInfo, err := sftp.Stat(dstPath); err == nil {
				if fileInfo.IsDir() {
					logger.Error("Destination is a directory, skipping copy:", dstPath)
					return err
				} else {
					// 如果目标已经存在且为文件，则不再拷贝
					logger.Error("File already exists on the remote server, skipping copy:", dstPath)
					return err
				}
			}

			df, err := sftp.Create(dstPath)
			if nil != err {
				logger.Error("sftp.Create error:", err)
				return err
			}
			//defer df.Close()
			defer func() {
				if err := df.Close(); err != nil {
					logger.Error("df.Close error:", err)
				}
			}()

			// Create progress bar
			info, err := sf.Stat()
			if err != nil {
				logger.Error("sf.Stat error:", err)
				return err
			}

			//如果是拷贝文件则显示进度条，临时脚本则不显示
			if displayBar == true {
				totalBytes := info.Size()
				bar := p.AddBar(
					totalBytes,
					mpb.PrependDecorators(
						decor.Name(args.Host+" Copying "+srcPath+"> "),
						decor.CountersKibiByte("% .1f / % .1f"),
					),
					mpb.AppendDecorators(
						decor.Percentage(),
						decor.EwmaETA(decor.ET_STYLE_GO, float64(totalBytes)),
					),
				)
				// 使用互斥锁保护进度条更新
				progressLock := &ScpProgressLock{
					bar:     bar,
					proxied: bar.ProxyReader(sf),
				}
				// 使用缓冲区提高拷贝效率
				buf := make([]byte, 64*1024*1024)

				_, err = io.CopyBuffer(df, progressLock, buf)
				if err != nil && !errors.Is(err, io.EOF) {
					logger.Error("io.CopyBuffer error:", err)
					return err
				}
			} else {
				// 使用缓冲区提高拷贝效率
				buf := make([]byte, 64*1024*1024)
				_, err = io.CopyBuffer(df, sf, buf)
				if err != nil {
					logger.Error("io.CopyBuffer error:", err)
				}
			}

			//p.Wait()
			return nil
		}

		var scp func(srcPath, dstBase string)
		scp = func(srcPath, dstBase string) {
			logger.Debug("scp(children) ", srcPath, "->", dstBase)
			fi, err := os.Stat(srcPath)

			if err != nil {
				logger.Error("os.Stat error:", err)
				return
			}

			// 如果是目录
			if fi.IsDir() {
				if err != nil {
					fmt.Printf("Error getting dir size: %v\n", err)
					return
				}

				// 遍历目录下的所有元素
				files, err := ioutil.ReadDir(srcPath)

				if err != nil {
					logger.Error("ioutil.ReadDir error:", err)
					return
				}
				for _, file := range files {
					srcChild := fmt.Sprintf("%s/%s", srcPath, file.Name())
					dstChild := path.Join(dstBase, file.Name())

					// 对于子目录，先创建子目录再递归处理
					if file.IsDir() {
						err = sftp.MkdirAll(dstChild)
						if err != nil && !errors.Is(err, fs.ErrExist) {
							logger.Error("sftp.Mkdir error:", err)
							continue
						}
						scp(srcChild, dstChild)
					} else {
						// 处理文件时更新进度条

						// 对于文件，直接拷贝
						scpFile(srcChild, dstChild)
					}
				}
			} else {
				// 是文件，计算完整的目标路径并拷贝
				scpFile(srcPath, path.Join(dstBase, filepath.Base(srcPath)))
			}
		}
		// 开始递归拷贝，初始目标路径为指定的目标目录
		scp(src, dst)
	})

	return err
}
