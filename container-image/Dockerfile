FROM ubuntu:22.04

WORKDIR /setup

COPY . .

# unminimize ubuntu to make commands available to users need
RUN yes | unminimize

RUN apt update && apt upgrade -y

# install openssh-server
RUN apt install -y openssh-server

# install syslog for logging
RUN apt install rsyslog -y

# remove kernel log need
RUN sed -i '/imklog/s/^/#/' /etc/rsyslog.conf

# allow root login
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config

EXPOSE 22

CMD [ "sh", "entrypoint.sh" ]