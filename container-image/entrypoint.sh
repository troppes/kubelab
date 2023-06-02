#!/bin/bash

echo "Entrypoint for Base Image Kubelab"
echo "Script:       0.0.1"
echo "User:         '$(whoami)'"
echo "Group:        '$(id -g -n)'"
echo "Working dir:  '$(pwd)'"



# Read the environment variables
username=$USER_NAME
password=$USER_PASSWORD
isSudo=$(echo "$SUDO_ACCESS" | tr '[:upper:]' '[:lower:]')
rootPassword=$ROOT_PASSWORD


# Check if both variables are provided
if [ -z "$username" ] || [ -z "$password" ] || [ -z "$rootPassword" ]; then
  echo "USER_NAME and USER_PASSWORD and ROOT_PASSWORD environment variables must be set."
  exit 1
fi

# Create the user
useradd -m -p "$password" -s /bin/bash "$username"
if [ "$isSudo" = "true" ]; then
  usermod -aG sudo "$username"
  echo "User $username has been added to the sudoers group."
fi

# Check if the user creation was successful
if [ $? -eq 0 ]; then
  echo "User $username created successfully."
else
  echo "Failed to create user $username."
fi

# Change root pw
usermod --password "$rootPassword" root
if [ $? -eq 0 ]; then
  echo "Root pw changed successfully."
else
  echo "Failed to change root pw."
fi

# setup permanent host key, so that ssh does not complain
mkdir -p /home/"$username"/.kubelab
cp /etc/ssh/ssh_host_rsa_key /home/"$username"/.kubelab/ssh_host_rsa_key
sed -i "s/#HostKey \/etc\/ssh\/ssh_host_rsa_key/HostKey \/home\/$username\/.kubelab\/ssh_host_rsa_key/" /etc/ssh/sshd_config

service ssh start
rsyslogd # start rsyslog to fill /var/log/auth.log

touch /var/log/auth.log # create file directly for tail to work
chown syslog:adm /var/log/auth.log

exec tail -f /var/log/auth.log