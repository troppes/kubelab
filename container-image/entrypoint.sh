#!/bin/bash

echo "Entrypoint for Base Image Kubelab"
echo "Script:       0.0.1"
echo "User:         '$(whoami)'"
echo "Group:        '$(id -g -n)'"
echo "Working dir:  '$(pwd)'"



# Read the environment variables
username=$USERNAME
password=$PASSWORD

# Check if both variables are provided
if [ -z "$username" ] || [ -z "$password" ]; then
  echo "Both USERNAME and PASSWORD environment variables must be set."
  exit 1
fi

# Generate an encrypted password
encrypted_password=$(openssl passwd -6 "$password")

# Create the user
useradd -m -p "$encrypted_password" -s /bin/bash "$username"

# Check if the user creation was successful
if [ $? -eq 0 ]; then
  echo "User $username created successfully."
else
  echo "Failed to create user $username."
fi

service ssh start
rsyslogd # start rsyslog to fill /var/log/auth.log

touch /var/log/auth.log # create file directly for tail to work
chown syslog:adm /var/log/auth.log

exec tail -f /var/log/auth.log