# AUTODEPLOYMENT
## Application Introduction
This application is a simple command-line tool for automatic deployment in a Linux environment, written in `Go 1.20.4`.
Its primary function is to assist operations and implementation personnel in automating the process of "remote file copying" and "remote execution of shell scripts" through pre-written playbook.yml files during deployment. This facilitates rapid deployment and enhances the efficiency of the deployment process.  
## Usage
By executing the `adp init` command, the configuration file is initialized, and an example playbook.yml file will be generated in the current directory. You can modify the playbook according to your needs and then execute the `adp run` command to perform the task.
```shell
# Standard Initialization Configuration
adp init
# Full Initialization Configuration
adp init [-a]
# Edit the playbook Configuration
vim playbook.yml
# Executing the Default Playbook
adp run
# Executing a Specific Playbook
adp run your_playbook.yml
```
## Install

#### Linux && MacOS

```shell
curl -fsSL https://github.com/ghostelement/auto-deployment/releases/download/install/install.sh | sh
```

*China mainland users can use the following command to speed up the download*

```shell
curl -fsSL https://ghproxy.com/https://github.com/ghostelement/auto-deployment/releases/download/install/install.sh | sh -s https://ghproxy.com
```

#### Windows (PowerShell)

```powershell
# Optional: Needed to run a remote script the first time
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
irm https://github.com/ghostelement/auto-deployment/releases/download/install/install.ps1 | iex
# if you can't access github, you can use proxy
irm https://github.com/ghostelement/auto-deployment/releases/download/install/install.ps1 -Proxy '<host>:<ip>' | iex
```

*China mainland users can use the following command to speed up the download*

```powershell
# Optional: Needed to run a remote script the first time
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
irm https://ghproxy.com/https://github.com/ghostelement/auto-deployment/releases/download/install/install_ZH-CN.ps1 | iex
```

## Commands

### init
> Initialize the configuration

Run `adp init` command to initialize the configuration file.
The configuration file is located in the current directory. The default name is `playbook.yml`.
If you want to all the configuration file to be named `playbook.yml`, you can add `-a` parameter.

**Example:**

```shell
adp init
# Or select Display all configurations
adp init -a
```
`adp init /path/yourpath/`when followed by a path, is designed to initialize the configuration files for the CLI tool within the specified directory. This means that instead of creating the default playbook.yml and other related configuration files in the current working directory, it will create them in the directory you provide.  


playbook.yml example：

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
**Configuration Explanation**
| Name          | Required | Description                                | Example         |
| ------------- | ---- | ------------------------------------ | ------------ |
| name       | Y    | The name of the task. This is used to identify the task within the playbook.        | initServer |
| host        | Y    | A list of remote hosts to which the deployment will be executed. This can be a single IP address or a range of addresses.      | 192.168.0.1:22 |
| user     | Y    | The username used to authenticate to the remote host.            |      root        |
| password      | Y   | The password associated with the remote user account.           |     123456     |
| parallelNum  | N    | The number of concurrent connections to be established during the deployment process. This can speed up the process if multiple hosts are being deployed to simultaneously.      |         5     |
| srcFile       | N  | The path to the file or directory that needs to be uploaded to the remote host. If this is not specified, no file transfer will be performed. |     /yourpath/filename         |
| destDir     | N   | The destination directory on the remote server where files will be uploaded or where commands will be executed.     |     /tmp         |
| cmd        | N    | A simple command that needs to be executed on the remote host. This is useful for quick, single-line commands.  |  cd /tmp && pwd         |
| shell      | N    | A more complex command that may include newlines or multiple instructions. This is used when the command is too long or involves multiple steps.         |   cd /tmp; pwd          |

### run
> Executing playbook

Run `adp run` command to deploy program to server.
The configuration file is located in the current directory. The default name is `playbook.yml`.

**Example:**

```shell
adp run
# or
adp run /path/yourpath/playbook.yml
```
#### db
> Database Connection  
`adp db dbname`Connects to the database using the playbook.yml file in the current directory.  
`adp db dbname -f /path/your_playbook.yml`Connects to the database using a specified playbook file.  
- Supported Database Types: mysql, postgresql, redis

Database YAML Configuration Example:
```yaml
db:
  - name: mysql
    host: 192.168.223.5
    port: 3306
    dbtype: mysql
    username: root
    password: 'admin12345'
    database: test
  - name: postgresql
    host: 192.168.223.5
    port: 5432
    dbtype: postgresql
    username: postgres
    password: 'admin12345'
    database: test
  - name: redis
    host: 192.168.223.5
    port: 6379
    dbtype: redis
    username: 'root'
    password: 'admin12345'
    database: 0
```
### version
> Check the version of the CLI tool

```shell
adp -v
```
### help
> help

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