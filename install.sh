#!/bin/bash
################################################################
# @Author: Leon
# @Date: 2024-03-20
# @Description: install auto-deployment tool
################################################################

function warn(){
  echo -e "\033[33m$1\033[0m"
}

function info(){
  echo $1
}

function error(){
  echo -e "\033[31m$1\033[0m"
}

# 判断命令是否执行成功
function check() {
  if [ $? -ne 0 ]; then
    error $1
    exit 1
  fi
}

# 获取版本号
VERSION=$(curl -s https://api.github.com/repos/ghostelement/auto-deployment/releases/latest | grep tag_name | cut -d '"' -f 4)
URL="https://github.com/ghostelement/auto-deployment/releases/download/$VERSION"
Proxy=$1

if [ -n "$Proxy" ]; then
  URL="$Proxy/$URL"
fi

DEPLOY_DIR=""

# 获取当前系统
os=$(uname -s)
if [ $os == "Darwin" ]; then
  os="darwin"
  DEPLOY_DIR="$HOME/Applications/autodeployment"
elif [ $os == "Linux" ]; then
  os="linux"
  DEPLOY_DIR="/usr/bin"
else
  error "不支持的系统 $os"
  exit 1
fi

# 获取当前系统架构 x86_64 or arm64
arch=$(uname -m)
if [ $arch == "x86_64" ]; then
  arch="amd64"
elif [ $arch == "arm64" ]; then
  arch="arm64"
else
  error "不支持的架构 $arch"
  exit 1
fi

DownloadUrl="$URL/autodeployment_"$os"_"$arch".tar.gz"
info "download $DownloadUrl to $DEPLOY_DIR"

tarFileTmpDir=$DEPLOY_DIR/tmp

# 创建临时目录，并判断是否有权限
if [ $os == "Darwin" ]; then
  if [ -w $DEPLOY_DIR ]; then
    mkdir -p $tarFileTmpDir
  else
    warn "permission denied, auto change to sudo"
    sudo mkdir -p $tarFileTmpDir
  fi
fi

# 下载文件
info "download $DownloadUrl"
wget $DownloadUrl
check "download $DownloadUrl failed"

# 解压并复制文件到目标目录
if [ $os == "Darwin" ]; then
tar -zxvf autodeployment_"$os"_"$arch".tar.gz -C $tarFileTmpDir
cp $tarFileTmpDir/adp  $DEPLOY_DIR
chmod 755 $DEPLOY_DIR/adp
# 删除临时文件
rm -rf $tarFileTmpDir/adp
else
tar -zxvf autodeployment_"$os"_"$arch".tar.gz
cp adp $DEPLOY_DIR
fi

# 获取当前系统的环境变量Path，判断是否已经存在，不存在则添加
if [ $os == "Darwin" ]; then
  path=$(echo $PATH | grep $DEPLOY_DIR)
  if [ -z "$path" ]; then
    info "export PATH=$DEPLOY_DIR:$PATH" >>~/.bash_profile
    # tips
    warn "please run 'source ~/.bash_profile'"
    warn "if you are using zsh, please run 'echo export PATH=$DEPLOY_DIR:'\$PATH' >> ~/.zshrc && source ~/.zshrc'"
  fi
fi
