#!/bin/bash

echo "Entrypoint for Kubelab Backend"
echo "Script:       0.0.1"
echo "User:         '$(whoami)'"
echo "Group:        '$(id -g -n)'"
echo "Working dir:  '$(pwd)'"

# Create folder for db
[ -d database ] || mkdir database

# create DB file
if [[ ! -f "database/database.db" ]]; then
    npm run setup
fi

echo "Pre-flight checks sucessful, starting now."
exec node server.js