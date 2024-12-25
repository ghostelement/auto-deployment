package shell

import (
	"auto-deployment/logger"
	"errors"
	"fmt"
	"io"
	"io/fs"
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

// ReadAsyncShellLog 从给定的reader中异步读取shell日志，并通过handler处理每一行。
// reader: 用于读取日志的io.Reader接口。
// handler: 当读取到一行日志时调用的处理函数，每一行日志（包括最后一行，如果它不完整）都会被传递给这个函数。
// 返回值表示读取过程中遇到的任何错误。
func (ss ShellRun) ReadAsyncShellLog(reader io.Reader, handler func(string)) error {
	var cache string = ""
	buf := make([]byte, 8192)
	for {
		num, err := reader.Read(buf)
		// logger.Debug("ReadAsyncShellLog:",num,err)
		if err != nil {
			if len(cache) > 0 && handler != nil {

				handler(cache)
			}
			// 如果错误为EOF或包含"closed"，则将其重置为nil，表示读取完成
			if err == io.EOF || strings.Contains(err.Error(), "closed") {
				err = nil
			}
			return err
		}
		// 如果成功读取到数据
		if num > 0 {
			d := buf[:num]                      // 获取实际读取到的数据
			a := strings.Split(string(d), "\n") // 按行分割数据
			line := strings.Join(a[:len(a)-1], "\n")
			if handler != nil {
				// 调用handler处理每一行完整的日志（包括之前的缓存）
				handler(fmt.Sprintf("%s%s\n", cache, line))
			}
			// fmt.Printf("|%s%s\n", cache, line)
			cache = a[len(a)-1]
		}
	}
}

type SshLoginHandle func(client *ssh.Client)
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
		sshLoginHandle(sshClient)
	}
	return nil
}

// RunWithSshSession 通过ssh会话在指定主机上运行命令。
//
// 参数:
// host - 主机地址，用于日志记录。
// s - ssh.Session对象，用于执行远程命令。
// cmd - 要执行的命令。
// args - 命令的参数数组。
// handler - 异步日志处理函数，用于处理命令输出的日志。
//
// 返回值:
// 返回执行过程中可能出现的错误。
func (ss ShellRun) RunWithSshSession(host string, s ssh.Session, cmd string, args []string, handler LogAsyncHandle) error {
	stdout, _ := s.StdoutPipe()
	stderr, _ := s.StderrPipe()
	// if ss.WorkDir != ""{
	// 	cmd.Dir = ss.WorkDir
	// }
	// 构建要执行的完整命令字符串
	cmds := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))

	if err := s.Start(cmds); err != nil {
		logger.Error("|run cmd:", cmds, "|error: ", err.Error())
		return err
	}

	// 异步读取命令的标准输出，并通过handler处理
	go ss.ReadAsyncShellLog(stdout, func(log string) {
		if handler != nil {
			handler(host, log)
		}
	})
	// 异步读取命令的标准错误，并通过handler处理
	go ss.ReadAsyncShellLog(stderr, func(log string) {
		if handler != nil {
			handler(host, log)
		}
	})
	// 等待命令执行完成，如有错误则记录并返回
	if err := s.Wait(); err != nil {
		logger.Error("Error waiting for command:", cmds, "|error:", err.Error())
		return err
	}
	return nil
}

// SshLoginAndRun 通过SSH登录并执行命令
// sshArgs: 包含SSH登录所需参数，如用户名、密码、主机地址等
// cmd: 需要在远程主机上执行的命令
// args: 命令的参数数组
// handler: 日志异步处理函数，用于处理执行命令时的日志
// 返回值: 执行过程中可能出现的错误
func (ss ShellRun) SshLoginAndRun(sshArgs SshLoginArgs, cmd string, args []string, handler LogAsyncHandle) error {
	r := rand.Intn(100) //TODO:后期需改为uuid 随机生成一个数字，用于日志标识
	// 尝试通过sshArgs进行SSH登录，并在登录成功后执行相应的操作
	err := ss.SshLogin(sshArgs, func(sshClient *ssh.Client) {

		// logger.Debug("args:", sshArgs)
		//创建ssh-session
		session, err := sshClient.NewSession()
		if err != nil {
			logger.Error("create client session error:", err)
		}
		defer session.Close()
		// 使用创建的SSH会话执行命令
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

	err := ss.SshLogin(args, func(sshClient *ssh.Client) {

		sftp, err := sftp.NewClient(sshClient)
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
			// 打开源文件并准备读取
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
			//使用sftp创建目标文件
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
			if displayBar {
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

		//根据传入参数处理scp逻辑
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
				// 遍历目录下的所有元素
				files, err := os.ReadDir(srcPath)
				if err != nil {
					fmt.Printf("Error getting dir paht: %v\n", err)
					return
				}

				// 遍历目录下的所有元素
				//files, err := ioutil.ReadDir(srcPath)

				//if err != nil {
				//	logger.Error("ioutil.ReadDir error:", err)
				//	return
				//}
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
