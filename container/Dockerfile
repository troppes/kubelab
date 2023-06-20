FROM ubuntu:22.04

WORKDIR /setup

COPY . .

# unminimize ubuntu to make commands available to users need
RUN yes | unminimize

RUN apt update && apt upgrade -y

# install *.UTF-8 locales otherwise some apps get trouble
RUN apt -y install locales && locale-gen en_US.UTF-8 ja_JP.UTF-8 zh_CN.UTF-8 && update-locale LANG=en_US.UTF-8 

# install other utilities
RUN  apt -y install \
    apt-transport-https \
    bash-completion vim less man jq bc \
    lsof tree psmisc htop lshw sysstat dstat \
    iproute2 iputils-ping iptables dnsutils traceroute \
    netcat curl wget nmap socat netcat-openbsd rsync \
    p7zip-full \
    git tig \
    binutils acl pv \
    strace tcpdump \
    sudo \
    coreutils \
    openssh-server \ 
    rsyslog

# enable bash-completeion for root user (other users works by default)
RUN  (echo && echo '[ -f /etc/bash_completion ] && ! shopt -oq posix && . /etc/bash_completion') >> /etc/skel/.bashrc

# remove kernel log need
RUN sed -i '/imklog/s/^/#/' /etc/rsyslog.conf

# allow root login
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config

EXPOSE 22

CMD [ "sh", "entrypoint.sh" ]