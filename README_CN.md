# AUTODEPLOYMENT
## 项目介绍
本项目是一个简单的linux环境下自动部署cli命令行工具，使用Go编写。  
其主要功能是帮助运维及实施人员在部署过程中通过预先编写好的playbook.yml剧本文件来自动完成"文件远程拷贝" 和 "shell脚本远程执行" 的操作，从而完成快速部署，提高部署效率。  
## 用法
通过执行`adp init`命令，初始化配置文件，将在当前目录下生成一个playbook.yml剧本文件示例，可以按照自己需求修改剧本后执行```adp run```命令来执行任务。
```shell
# 标准初始化配置
adp init
# 全量初始化配置
adp init [-a]
# 修改配置文件
vim playbook.yml
# 执行默认剧本
adp run
# 执行指定剧本
adp run your_playbook.yml
```

## 安装

### 脚本安装（推荐）

#### Window

##### PowerShell

```shell
# Optional: Needed to run a remote script the first time
# 可选：第一次运行远程脚本时需要
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
irm https://github.com/ghostelement/auto-deployment/releases/download/install/install.ps1 | iex
# if you can't access github, you can use proxy
# 如果访问github很慢，可以使用镜像代理
irm https://github.com/ghostelement/auto-deployment/releases/download/install/install.ps1 -Proxy '<host>:<ip>' | iex
```

*国内访问*
```shell
# 可选：第一次运行远程脚本时需要
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
irm https://ghproxy.com/https://github.com/ghostelement/auto-deployment/releases/download/install/install_ZH-CN.ps1 | iex
```

#### Linux & Mac
```shell
curl -fsSL https://github.com/ghostelement/auto-deployment/releases/download/install/install.sh | sh
```

*国内访问*
```shell
curl -fsSL https://ghproxy.com/https://github.com/ghostelement/auto-deployment/releases/download/install/install.sh | sh -s https://ghproxy.com
```

### 手动安装(离线环境推荐)

#### Windows

##### PowerShell（推荐）
```shell
# 下载
wget https://github.com/ghostelement/auto-deployment/releases/download/{latest-version}/autodeployment_{version}_windows_amd64.tgz

# 解压
tar -xvzf autodeployment_{version}_windows_amd64.tgz -C /your/path
# 设置环境变量
[environment]::SetEnvironmentvariable("PATH", "$([environment]::GetEnvironmentvariable("Path", "User"));/your/path", "User")
### 或者直接将手动解压出的adp.exe拷贝到/windows/system32目录下
```

##### Cmd

```shell
# 下载
wget https://github.com/ghostelement/auto-deployment/releases/download/{version}/autodeployment_{version}_windows_amd64.tgz

# 解压
tar -xvzf autodeployment_{version}_windows_amd64.tgz -C /your/path
# 手动添加环境变量
```



#### Linux

```shell
# 下载
wget https://github.com/ghostelement/auto-deployment/releases/download/{version}/autodeployment_{version}_linux_amd64.tgz
# 解压
tar -zxvf autodeployment_{version}_linux_amd64.tgz -C /your/path/
chmod 755 /your/path/adp
# 设置环境变量（追加）
export PATH=$PATH:/your/path
### 或者直接放到/usr/bin目录下，则不用添加环境变量
sudo cp adp /usr/bin/
sudo chmod 755 /usr/bin/adp
```

### Mac
```shell
# 下载
wget https://github.com/ghostelement/auto-deployment/releases/download/{version}/autodeployment_{version}_darwin_amd64.tgz
# 解压
tar -zxvf autodeployment_{version}_darwin_amd64.tgz -C /your/path/
# 设置环境变量
export PATH=$PATH:/your/path
```

## 命令说明

#### init

> 初始化配置

`adp init`会在当前路径下创建一个名为`playbook.yml`的剧本文件及scripts内置脚本文件夹

```shell
# init playbook.yml
> adp init
# Or select Display all configurations
> adp init -a
```
`adp init /path/yourpath/`会在指定的目录下进行初始化操作  


playbook.yml示例：

```yaml
job:
    - name: jobname
      host:
        - 192.168.0.1:22
        - 192.168.0.2:22
      user: root
      password: yourpassword
      parallelNum: 5
      srcFile: filename or dirname
      destDir: /tmp
      cmd: ls /tmp
      shell: cd /tmp && pwd
```

**配置说明**

| 名称          | 必填 | 说明                                 | 样例         |
| ------------- | ---- | ------------------------------------ | ------------ |
| name       | Y    | 任务名        | initServer |
| host        | Y    | 远程主机列表                               | 192.168.0.1:22 |
| user     | Y    | 远程主机用户名                               |      root        |
| password      | Y   | 远程主机密码                                 |     123456     |
| parallelNum  | N    | 并发数量             |         5     |
| srcFile       | N  | 需要上传的文件路径或目目录路径 |    /yourpath/filename    |
| destDir     | N    | 远程服务器目标路径                       |  /tmp            |
| cmd        | N    | 需要执行的命令  |  cd /tmp && pwd         |
| shell      | N    | 需要执行的较复杂的命令(包含换行等)                 |   cd /tmp; pwd          |
#### run
> 执行任务

`adp run`执行当前目录下的playbook.yml剧本任务  
`adp run /path/your_playbook.yml`执行指定的剧本任务

#### version

> 查看版本

```shell
adp -v
autodeployment version v0.0.1 linux/amd64
```



#### help

> 帮助

```shell
adp --help
# or
adp -h

NAME:
   adp - this is a simple cli app that automates deploy

USAGE:
   adp run [/path/to/playbook.yml]

VERSION:
   v0.0.1

DESCRIPTION:
   This is a simple cli app that automates deploy.
   e.g. This is a common way to perform deploy, according to playbook.yml in the current path
     adp
   This is manually specifying the configuration file
     adp run /path/to/playbook.yml

COMMANDS:
   run
   init
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

