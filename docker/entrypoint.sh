#!/bin/sh

CONF_PATH="/app/src/k8s-devops-console/conf/app.conf"

if [ -n "${OAUTH_PROVIDER+x}" ]; then
    echo "oauth.provider = ${OAUTH_PROVIDER}" >> "$CONF_PATH"
fi

if [ -n "${OAUTH_CLIENT_ID+x}" ]; then
    echo "oauth.client.id = ${OAUTH_CLIENT_ID}" >> "$CONF_PATH"
fi

if [ -n "${OAUTH_CLIENT_SECRET+x}" ]; then
    echo "oauth.client.secret = ${OAUTH_CLIENT_SECRET}" >> "$CONF_PATH"
fi

exec "$@"
