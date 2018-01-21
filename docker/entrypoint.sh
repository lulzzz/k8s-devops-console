#!/bin/sh

CONF_PATH="/app/src/k8s-devops-console/conf/app.conf"

if [ -z "${APP_SECRET+x}" ]; then
    echo "[WARN] environment var APP_SECRET is empty, generating it"
    export APP_SECRET=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
fi

exec "$@"
