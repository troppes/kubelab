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

# install Node
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && apt-get install -y nodejs

# enable x11 support
RUN apt -y install xauth x11-apps software-properties-common libgconf-2-4 xorg libnotify4 libnss3 libxss1 libatk1.0-0 libatk-bridge2.0-0 libgdk-pixbuf2.0-0 libgtk-3-0 libnss3-dev libxss-dev libasound2 dbus-x11 xvfb
RUN sed -i "s/^.*X11Forwarding.*$/X11Forwarding yes/" /etc/ssh/sshd_config 
RUN sed -i "s/^.*X11UseLocalhost.*$/X11UseLocalhost no/" /etc/ssh/sshd_config 
RUN npm install -g electron@latest --unsafe-perm=true --allow-root

# enable bash-completeion for root user (other users works by default)
RUN  (echo && echo '[ -f /etc/bash_completion ] && ! shopt -oq posix && . /etc/bash_completion') >> /etc/skel/.bashrc

# remove kernel log need
RUN sed -i '/imklog/s/^/#/' /etc/rsyslog.conf

# allow root login
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config

EXPOSE 22

CMD [ "sh", "entrypoint.sh" ]