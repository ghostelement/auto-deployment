#!/usr/bin/bash
#
#***************************************************************************************************
#Author:        Leon
#Description:   reset for CentOS 6/7/8 & CentOS Stream 8/9 & Ubuntu 18.04/20.04/22.04 & Rocky 8/9 & Kylin 10
#Copyright (C): 2023 All rights reserved
#Readme:  使用bash运行脚本，尤其注意在ubuntu系统下的运行方式，bash init.sh或者./init.sh
#**************************************************************************************************
COLOR="echo -e \\033[01;31m"
END='\033[0m'

os(){
    if grep -Eqi "Centos"  /etc/issue && [ $(sed -rn 's#^.* ([0-9]+)\..*#\1#p' /etc/redhat-release) == 6 ] ;then
        OS_ID=`sed -rn 's#^([[:alpha:]]+) .*#\1#p' /etc/redhat-release`
        OS_RELEASE=`sed -rn 's#^.* ([0-9.]+).*#\1#p' /etc/redhat-release`
        OS_RELEASE_VERSION=`sed -rn 's#^.* ([0-9]+)\..*#\1#p' /etc/redhat-release`
    else
        OS_ID=`sed -rn '/^NAME=/s@.*="([[:alpha:]]+).*"$@\1@p' /etc/os-release`
        OS_NAME=`sed -rn '/^NAME=/s@.*="([[:alpha:]]+) (.*)"$@\2@p' /etc/os-release`
        #适配麒麟
        if [ ${OS_ID} == "Kylin" ]; then
            OS_NAME="Linux"
            OS_RELEASE="V10"
            OS_RELEASE_VERSION="8"
        else
            OS_RELEASE=`sed -rn '/^VERSION_ID=/s@.*="?([0-9.]+)"?@\1@p' /etc/os-release`
            OS_RELEASE_VERSION=`sed -rn '/^VERSION_ID=/s@.*="?([0-9]+)\.?.*"?@\1@p' /etc/os-release`
        fi
    fi
}

disable_selinux(){
    if [ ${OS_ID} == "CentOS" -o ${OS_ID} == "Kylin"   -o ${OS_ID} == "Rocky" ];then
        if [ `getenforce` == "Enforcing" ];then
            sed -ri.bak 's/^(SELINUX=).*/\1disabled/' /etc/selinux/config
            ${COLOR}"${OS_ID} ${OS_RELEASE} SELinux已禁用,请重新启动系统后才能生效!"${END}
        else
            ${COLOR}"${OS_ID} ${OS_RELEASE} SELinux已被禁用,不用设置!"${END}
        fi
    else
        ${COLOR}"${OS_ID} ${OS_RELEASE} SELinux默认没有安装,不用设置!"${END}
    fi
}

disable_firewall(){
    if [ ${OS_ID} == "CentOS" -o ${OS_ID} == "Kylin"   -o ${OS_ID} == "Rocky" ];then
        rpm -q firewalld &> /dev/null && { systemctl disable --now firewalld &> /dev/null; ${COLOR}"${OS_ID} ${OS_RELEASE} Firewall防火墙已关闭!"${END}; } || { service iptables stop ; chkconfig iptables off; ${COLOR}"${OS_ID} ${OS_RELEASE} iptables防火墙已关闭!"${END}; }
    else
        dpkg -s ufw &> /dev/null && { systemctl disable --now ufw &> /dev/null; ${COLOR}"${OS_ID} ${OS_RELEASE} ufw防火墙已关闭!"${END}; } || ${COLOR}"${OS_ID} ${OS_RELEASE}  没有ufw防火墙服务,不用关闭！"${END}
    fi
}

optimization_sshd(){
    if [ ${OS_ID} == "CentOS" -o ${OS_ID} == "Kylin"   -o ${OS_ID} == "Rocky" ];then
        sed -ri.bak -e 's/^#(UseDNS).*/\1 no/' -e 's/^(GSSAPIAuthentication).*/\1 no/' /etc/ssh/sshd_config
    else
        sed -ri.bak -e 's/^#(UseDNS).*/\1 no/' -e 's/^#(GSSAPIAuthentication).*/\1 no/' /etc/ssh/sshd_config
    fi
    if [ ${OS_RELEASE_VERSION} == "6" ];then
        service sshd restart
    else
        systemctl restart sshd
    fi
    ${COLOR}"${OS_ID} ${OS_RELEASE} SSH已优化完成!"${END}
}

aliyun(){
    URL=mirrors.aliyun.com
}

huawei(){
    URL=repo.huaweicloud.com
}

tencent(){
    URL=mirrors.tencent.com
}

tuna(){
    URL=mirrors.tuna.tsinghua.edu.cn
}

netease(){
    URL=mirrors.163.com
}

sohu(){
    URL=mirrors.sohu.com
}

nju(){
    URL=mirrors.nju.edu.cn
}

ustc(){
    URL=mirrors.ustc.edu.cn
}

sjtu(){
    URL=mirrors.sjtug.sjtu.edu.cn
}

fedora(){
    URL=archives.fedoraproject.org
}

set_yum_rocky8_9(){
    sed -i.bak -e 's|^mirrorlist=|#mirrorlist=|g' -e 's|^#baseurl=http://dl.rockylinux.org/$contentdir|baseurl=https://'${URL}'/rocky|g' /etc/yum.repos.d/[Rr]ocky*.repo
    ${COLOR}"更新镜像源中,请稍等..."${END}
    dnf clean all > /dev/null
    dnf makecache > /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} YUM源设置完成!"${END}
}

set_yum_rocky8_9_2(){
    sed -i.bak -e 's|^mirrorlist=|#mirrorlist=|g' -e 's|^#baseurl=http://dl.rockylinux.org/$contentdir|baseurl=https://'${URL}'/rockylinux|g' /etc/yum.repos.d/[Rr]ocky*.repo
    ${COLOR}"更新镜像源中,请稍等..."${END}
    dnf clean all > /dev/null
    dnf makecache > /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} YUM源设置完成!"${END}
}

set_yum_centos_stream9(){
    [ -d /etc/yum.repos.d/backup ] || mkdir /etc/yum.repos.d/backup
    (ls -l /etc/yum.repos.d/ |grep -q repo) && mv /etc/yum.repos.d/*.repo /etc/yum.repos.d/backup || ${COLOR}"没有repo文件!"${END}
    cat > /etc/yum.repos.d/base.repo <<-EOF
[BaseOS]
name=BaseOS
baseurl=https://${URL}/centos-stream/\$stream/BaseOS/\$basearch/os/
gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-centosofficial

[AppStream]
name=AppStream
baseurl=https://${URL}/centos-stream/\$stream/AppStream/\$basearch/os/
gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-centosofficial

[extras-common]
name=extras-common
baseurl=https://${URL}/centos-stream/SIGs/\$stream/extras/\$basearch/extras-common/
gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-centosofficial
EOF
    ${COLOR}"更新镜像源中,请稍等..."${END}
    dnf clean all > /dev/null
    dnf makecache > /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} YUM源设置完成!"${END}
}

set_yum_centos_stream8(){
    sed -i.bak -e 's|^mirrorlist=|#mirrorlist=|g' -e 's|^#baseurl=http://mirror.centos.org/$contentdir|baseurl=https://'${URL}'/centos|g' /etc/yum.repos.d/CentOS-*.repo
    ${COLOR}"更新镜像源中,请稍等..."${END}
    dnf clean all > /dev/null
    dnf makecache > /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} YUM源设置完成!"${END}
}

set_yum_centos8(){
    sed -i.bak -e 's|^mirrorlist=|#mirrorlist=|g' -e 's|^#baseurl=http://mirror.centos.org/$contentdir|baseurl=https://'${URL}'/centos-vault/centos|g' /etc/yum.repos.d/CentOS-*.repo
    ${COLOR}"更新镜像源中,请稍等..."${END}
    dnf clean all > /dev/null
    dnf makecache > /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} YUM源设置完成!"${END}
}

set_epel_centos8_9(){
    cat > /etc/yum.repos.d/epel.repo <<-EOF
[epel]
name=epel
baseurl=https://${URL}/epel/\$releasever/Everything/\$basearch/
gpgcheck=1
gpgkey=https://${URL}/epel/RPM-GPG-KEY-EPEL-\$releasever
EOF
    ${COLOR}"更新镜像源中,请稍等..."${END}
    dnf clean all > /dev/null
    dnf makecache > /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} EPEL源设置完成!"${END}
}

set_epel_2_centos8_9(){
    cat > /etc/yum.repos.d/epel.repo <<-EOF
[epel]
name=epel
baseurl=https://${URL}/fedora-epel/\$releasever/Everything/\$basearch/
gpgcheck=1
gpgkey=https://${URL}/fedora-epel/RPM-GPG-KEY-EPEL-\$releasever
EOF
    ${COLOR}"更新镜像源中,请稍等..."${END}
    dnf clean all > /dev/null
    dnf makecache > /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} EPEL源设置完成!"${END}
}

set_epel_3_centos8_9(){
    cat > /etc/yum.repos.d/epel.repo <<-EOF
[epel]
name=epel
baseurl=https://${URL}/fedora/epel/\$releasever/Everything/\$basearch/
gpgcheck=1
gpgkey=https://${URL}/fedora/epel/RPM-GPG-KEY-EPEL-\$releasever
EOF
    ${COLOR}"更新镜像源中,请稍等..."${END}
    dnf clean all > /dev/null
    dnf makecache > /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} EPEL源设置完成!"${END}
}

set_yum_centos7(){
    sed -i.bak -e 's|^mirrorlist=|#mirrorlist=|g' -e 's|^#baseurl=http://mirror.centos.org|baseurl=https://'${URL}'|g' /etc/yum.repos.d/CentOS-*.repo
    ${COLOR}"更新镜像源中,请稍等..."${END}
    yum clean all > /dev/null
    yum makecache > /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} YUM源设置完成!"${END}
}

set_epel_centos7(){
    cat > /etc/yum.repos.d/epel.repo <<-EOF
[epel]
name=epel
baseurl=https://${URL}/epel/\$releasever/\$basearch/
gpgcheck=1
gpgkey=https://${URL}/epel/RPM-GPG-KEY-EPEL-\$releasever
EOF
    ${COLOR}"更新镜像源中,请稍等..."${END}
    yum clean all > /dev/null
    yum makecache > /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} EPEL源设置完成!"${END}
}

set_epel_2_centos7(){
    cat > /etc/yum.repos.d/epel.repo <<-EOF
[epel]
name=epel
baseurl=https://${URL}/fedora-epel/\$releasever/\$basearch/
gpgcheck=1
gpgkey=https://${URL}/fedora-epel/RPM-GPG-KEY-EPEL-\$releasever
EOF
    ${COLOR}"更新镜像源中,请稍等..."${END}
    yum clean all > /dev/null
    yum makecache > /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} EPEL源设置完成!"${END}
}

set_epel_3_centos7(){
    cat > /etc/yum.repos.d/epel.repo <<-EOF
[epel]
name=epel
baseurl=https://${URL}/fedora/epel/\$releasever/\$basearch/
gpgcheck=1
gpgkey=https://${URL}/fedora/epel/RPM-GPG-KEY-EPEL-\$releasever
EOF
    ${COLOR}"更新镜像源中,请稍等..."${END}
    yum clean all > /dev/null
    yum makecache > /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} EPEL源设置完成!"${END}
}

set_yum_centos6(){
    sed -i.bak -e 's|^mirrorlist=|#mirrorlist=|g' -e 's|^#baseurl=http://mirror.centos.org|baseurl=https://'${URL}'/centos-vault|g' /etc/yum.repos.d/CentOS-*.repo
    ${COLOR}"更新镜像源中,请稍等..."${END}
    yum clean all > /dev/null
    yum makecache  &> /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} YUM源设置完成!"${END}
}

set_epel_centos6(){
    cat > /etc/yum.repos.d/epel.repo <<-EOF
[epel]
name=epel
baseurl=https://${URL}/epel/\$releasever/\$basearch/
gpgcheck=1
gpgkey=https://${URL}/epel/RPM-GPG-KEY-EPEL-\$releasever
EOF
    ${COLOR}"更新镜像源中,请稍等..."${END}
    yum clean all > /dev/null
    yum makecache > /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} EPEL源设置完成!"${END}
}

set_epel_2_centos6(){
    cat > /etc/yum.repos.d/epel.repo <<-EOF
[epel]
name=epel
baseurl=https://${URL}/pub/archive/epel/\$releasever/\$basearch/
gpgcheck=1
gpgkey=https://$(tencent)/epel/RPM-GPG-KEY-EPEL-\$releasever
EOF
    ${COLOR}"更新镜像源中,请稍等..."${END}
    yum clean all > /dev/null
    yum makecache > /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} EPEL源设置完成!"${END}
}

rocky8_9_base_menu(){
    
    aliyun
    set_yum_rocky8_9_2
    
}

centos_stream9_base_menu(){
    
    aliyun
    set_yum_centos_stream9
    
}

centos_stream8_base_menu(){
    
    aliyun
    set_yum_centos_stream8
    
}

centos8_base_menu(){
    
    aliyun
    set_yum_centos8
    
}

centos7_base_menu(){
    
    aliyun
    set_yum_centos7
    
}

centos6_base_menu(){
    
    aliyun
    set_yum_centos6
   
}

centos8_9_epel_menu(){
    
    aliyun
    set_epel_centos8_9
    
}

centos7_epel_menu(){

    aliyun
    set_epel_centos7

}

centos6_epel_menu(){

    tencent
    set_epel_centos6

}

rocky_menu(){

    rocky8_9_base_menu

}

centos_menu(){
    
    if [ ${OS_RELEASE_VERSION} == "6" ];then
        centos6_base_menu
    elif [ ${OS_NAME} == "Stream" ];then
        if [ ${OS_RELEASE_VERSION} == "8" ];then
            centos_stream8_base_menu
        else
            centos_stream9_base_menu
        fi
    elif [ ${OS_RELEASE_VERSION} == "8" -a ${OS_NAME} == "Linux" ];then
        centos8_base_menu
    else
        centos7_base_menu
    fi
    
}

set_apt(){
    OLD_URL=`sed -rn "s@^deb http://(.*)/ubuntu/? $(lsb_release -cs) main.*@\1@p" /etc/apt/sources.list`
    sed -i.bak 's/'${OLD_URL}'/'${URL}'/g' /etc/apt/sources.list
    if [ ${OS_RELEASE_VERSION} == "18" ];then
	    SEC_URL=`sed -rn "s@^deb http://(.*)/ubuntu $(lsb_release -cs)-security main.*@\1@p" /etc/apt/sources.list`
        sed -i.bak 's/'${SEC_URL}'/'${URL}'/g' /etc/apt/sources.list
    fi
    ${COLOR}"更新镜像源中,请稍等..."${END}
    apt-get update &> /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} APT源设置完成!"${END}
}

apt_menu(){
        aliyun
        set_apt
}

set_mirror_repository(){
    if [ ${OS_ID} == "CentOS" -o ${OS_ID} == "Kylin"   ];then
        centos_menu
    elif [ ${OS_ID} == "Rocky" ];then
        rocky_menu
    else
        apt_menu
    fi
}

centos_minimal_install(){
    ${COLOR}'开始安装“Minimal安装建议安装软件包”,请稍等......'${END}
    yum -y install gcc make autoconf gcc-c++ glibc glibc-devel pcre pcre-devel openssl openssl-devel systemd-devel zlib-devel vim lrzsz tree tmux lsof tcpdump wget net-tools iotop bc bzip2 zip unzip nfs-utils man-pages device-mapper-persistent-data lvm2 bash-completion chrony &> /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} Minimal安装建议安装软件包已安装完成!"${END}
}

ubuntu_minimal_install(){
    ${COLOR}'开始安装“Minimal安装建议安装软件包”,请稍等......'${END}
    apt-get -y install iproute2 ntpdate tcpdump telnet traceroute nfs-kernel-server nfs-common lrzsz tree openssl libssl-dev libpcre3 libpcre3-dev zlib1g-dev gcc openssh-server iotop unzip zip nfs-common  lvm2 bash-completion chrony &> /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} Minimal安装建议安装软件包已安装完成!"${END}
}

minimal_install(){
    if [ ${OS_ID} == "CentOS" -o ${OS_ID} == "Kylin"   -o ${OS_ID} == "Rocky" ] &> /dev/null;then
        centos_minimal_install
    else
        ubuntu_minimal_install
    fi
}



set_sshd_port(){
    disable_selinux
    disable_firewall
    read -p "请输入端口号: " PORT
    sed -i 's/#Port 22/Port '${PORT}'/' /etc/ssh/sshd_config
    ${COLOR}"${OS_ID} ${OS_RELEASE} 更改SSH端口号已完成,请重启系统后生效!"${END}
}

check_ip(){
    local IP=$1
    VALID_CHECK=$(echo ${IP}|awk -F. '$1<=255&&$2<=255&&$3<=255&&$4<=255{print "yes"}')
    if echo ${IP}|grep -E "^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$" >/dev/null; then
        if [ ${VALID_CHECK} == "yes" ]; then
            echo "IP ${IP}  available!"
            return 0
        else
            echo "IP ${IP} not available!"
            return 1
        fi
    else
        echo "IP format error!"
        return 1
    fi
}


red(){
    P_COLOR=31
}

green(){
    P_COLOR=32
}

yellow(){
    P_COLOR=33
}

blue(){
    P_COLOR=34
}

violet(){
    P_COLOR=35
}

cyan_blue(){
    P_COLOR=36
}

random_color(){
    P_COLOR="$[RANDOM%7+31]"
}

centos_ps1(){
    C_PS1=$(echo "PS1='\[\e[1;${P_COLOR}m\][\u@\h \W]\\$ \[\e[0m\]'" >> ~/.bashrc)
}

ubuntu_ps1(){
    U_PS1=$(echo 'PS1="\[\e[1;'''${P_COLOR}'''m\]${debian_chroot:+($debian_chroot)}\u@\h:\w\\$ \[\e[0m\]"' >> ~/.bashrc)
}


set_vim(){
    echo "export EDITOR=vim" >> ~/.bashrc
}

set_vim_env(){
    if grep -Eqi ".*EDITOR" ~/.bashrc;then
        sed -i '/.*EDITOR/d' ~/.bashrc
        set_vim
    else
        set_vim
    fi
    ${COLOR}"${OS_ID} ${OS_RELEASE} 默认文本编辑器设置成功,请重新登录生效!"${END}
}

set_history(){
if ! grep HISTTIMEFORMAT /etc/profile; then
cat >> /etc/profile <<EOF
#获取用户及IP
USER=\$(who -u am i 2>/dev/null| awk '{print \$1 \$NF}') 
if [ -z \$USER ]
then
  USER=\$(hostname)
fi
#新的history格式
HISTTIMEFORMAT="%F %T \$USER "
export HISTTIMEFORMAT
export HISTSIZE=10000
EOF
fi 
}

set_kernel(){
    cat > /etc/sysctl.conf <<-EOF
# Controls source route verification
net.ipv4.conf.default.rp_filter = 1
net.ipv4.ip_nonlocal_bind = 1
net.ipv4.ip_forward = 1

# Do not accept source routing
net.ipv4.conf.default.accept_source_route = 0

# Controls the System Request debugging functionality of the kernel
kernel.sysrq = 0

# Controls whether core dumps will append the PID to the core filename.
# Useful for debugging multi-threaded applications.
kernel.core_uses_pid = 1

# Controls the use of TCP syncookies
net.ipv4.tcp_syncookies = 1

# Disable netfilter on bridges.
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-arptables = 0

# Controls the default maxmimum size of a mesage queue
kernel.msgmnb = 65536

# Controls the maximum size of a message, in bytes
kernel.msgmax = 65536

# Controls the maximum shared segment size, in bytes
kernel.shmmax = 68719476736

# Controls the maximum number of shared memory segments, in pages
kernel.shmall = 4294967296

# TCP kernel paramater
net.ipv4.tcp_mem = 94500000 915000000 927000000
net.ipv4.tcp_rmem = 4096        87380   4194304
net.ipv4.tcp_wmem = 4096        16384   4194304
net.ipv4.tcp_window_scaling = 1
net.ipv4.tcp_sack = 1
net.ipv4.icmp_echo_ignore_broadcasts = 1
net.ipv4.icmp_ignore_bogus_error_responses = 1
net.ipv4.conf.all.rp_filter = 1
net.netfilter.nf_conntrack_buckets = 262144
net.netfilter.nf_conntrack_max = 1048576


# socket buffer
net.core.wmem_default = 8388608
net.core.rmem_default = 8388608
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.core.netdev_max_backlog = 262144
net.core.somaxconn = 65535
net.core.optmem_max = 81920

# TCP conn
net.ipv4.tcp_max_syn_backlog = 262144
net.ipv4.tcp_syn_retries = 3
net.ipv4.tcp_retries1 = 3
net.ipv4.tcp_retries2 = 15

# tcp conn reuse
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_tw_recycle = 0
net.ipv4.tcp_fin_timeout = 10
net.ipv4.tcp_timestamps = 0

net.ipv4.tcp_max_tw_buckets = 20000
net.ipv4.tcp_max_orphans = 3276800
net.ipv4.tcp_synack_retries = 1

# keepalive conn
net.ipv4.tcp_keepalive_time = 300
net.ipv4.tcp_keepalive_intvl = 30
net.ipv4.tcp_keepalive_probes = 3
net.ipv4.ip_local_port_range = 1024    65000

# swap
vm.overcommit_memory = 0
vm.swappiness = 10

#net.ipv4.conf.eth1.rp_filter = 0
#net.ipv4.conf.lo.arp_ignore = 1
#net.ipv4.conf.lo.arp_announce = 2
#net.ipv4.conf.all.arp_ignore = 1
#net.ipv4.conf.all.arp_announce = 2
EOF
    sysctl -p &> /dev/null
    ${COLOR}"${OS_ID} ${OS_RELEASE} 优化内核参数成功!"${END}
}

set_limits(){
    cat >> /etc/security/limits.conf <<-EOF
root     soft   core     unlimited
root     hard   core     unlimited
root     soft   nproc    1000000
root     hard   nproc    1000000
root     soft   nofile   1000000
root     hard   nofile   1000000
root     soft   memlock  32000
root     hard   memlock  32000
root     soft   msgqueue 8192000
root     hard   msgqueue 8192000
* soft nproc 65535
* hard nproc 65535
* soft nofile 65535
* hard nofile 65535
EOF
    ${COLOR}"${OS_ID} ${OS_RELEASE} 优化资源限制参数成功!"${END}
}

set_localtime(){
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
    echo 'Asia/Shanghai' >/etc/timezone
    if [ ${OS_ID} == "Ubuntu" ];then
        cat >> /etc/default/locale <<-EOF
LC_TIME=en_DK.UTF-8
EOF
    fi
    ${COLOR}"${OS_ID} ${OS_RELEASE} 系统时区已设置成功,请重启系统后生效!"${END}
}

network()
{
    #超时时间
    local timeout=1

    #目标网站
    local target=www.baidu.com

    #获取响应状态码
    local ret_code=`curl -I -s --connect-timeout ${timeout} ${target} -w %{http_code} | tail -n1`

    if [ "x$ret_code" = "x200" ]; then
        #网络畅通
        return 1
    else
        #网络不畅通
        return 0
    fi

    return 0
}

menu(){
##关闭防火墙
disable_selinux
disable_firewall
##配置ssh
optimization_sshd
#配置镜像源
echo "测试网络情况"
network
if [ $? -eq 0 ];then
    echo "服务器无Internet网络,跳过更新!!!"
else
    set_mirror_repository
    ##基础工具安装
    minimal_install
fi
##设置默认编辑器为vim
set_vim_env
##设置history格式
set_history
##优化内核
set_kernel
set_limits
##设置时区
set_localtime
}

main(){
    os
    menu
}

main