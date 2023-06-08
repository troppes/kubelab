#!/bin/bash

echo "Setup for Base Image of Kubelab"
echo "Script:       0.2.0"
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

# add group for user
groupadd -g 1000 "$username"

# Create the user
useradd -m -p "$password" -s /bin/bash "$username" -u 1000 -g 1000
if [ $? -eq 0 ]; then # Check if the user creation was successful
  echo "User $username created successfully."
else
  echo "Failed to create user $username."
fi

# Manually load skel
cp -r /etc/skel/. /home/"$username"/

# Set Folder rights, since the folder was created by the mount
chmod 755 /home/"$username"
chown -R "$username":"$username" /home/"$username"

if [ "$isSudo" = "true" ]; then
  usermod -aG sudo "$username"
  echo "User $username has been added to the sudoers group."
fi

# Change root pw
usermod --password "$rootPassword" root
if [ $? -eq 0 ]; then
  echo "Root pw changed successfully."
else
  echo "Failed to change root pw."
fi

# setup permanent host key
mkdir -p /home/"$username"/private/.kubelab
chmod 755 /home/"$username"/private/.kubelab 
chown root:root /home/"$username"/private/.kubelab
# Only copy if not already there with -n
cp -n /etc/ssh/ssh_host_rsa_key /home/"$username"/private/.kubelab/ssh_host_rsa_key

# setup private key if exists
if [ -f "/home/$username/private/.kubelab/kubelab_key" ]; then
  mkdir -p /home/"$username"/.ssh
  chmod 700 /home/"$username"/.ssh
  chown "$username":"$username" /home/"$username"/.ssh

  cp /home/"$username"/private/.kubelab/kubelab_key /home/"$username"/.ssh/kubelab_key
  chmod 600 /home/"$username"/.ssh/kubelab_key
  chown "$username":"$username" /home/"$username"/.ssh/kubelab_key
fi

# set hostkey in sshd_config
sed -i "s/#HostKey \/etc\/ssh\/ssh_host_rsa_key/HostKey \/home\/$username\/private\/.kubelab\/ssh_host_rsa_key/" /etc/ssh/sshd_config
# set new paths for authorized_keys
sed -i "s/.*AuthorizedKeysFile.*/AuthorizedKeysFile\t\.ssh\/authorized_keys .ssh\/kubelab_key /g" /etc/ssh/sshd_config
# allow key auth
sed -i 's/#\?PubkeyAuthentication\s\+.*$/PubkeyAuthentication yes/' /etc/ssh/sshd_config

service ssh start
rsyslogd # start rsyslog to fill /var/log/auth.log

touch /var/log/auth.log # create file directly for tail to work
chown syslog:adm /var/log/auth.log

exec tail -f /var/log/auth.log