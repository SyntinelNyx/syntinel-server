#!/bin/sh

mkdir -p /app/data
mkdir -p /app/data/certs

FILES = "config.yaml ecdsa_private.pem ecdsa_public.pem"
CERTS="ca_cert.pem ca_key.pem server_cert.pem server_key.pem"

for file in $FILES; do
    if [ ! -f "/app/data/$file" ]; then
        cp "/app/defaults/$file" "/app/data/$file"
    fi
done

for cert in $CERTS; do
    if [ ! -f "/app/data/$cert" ]; then
        cp "/app/defaults/$cert" "/app/data/certs/$cert"
    fi
done

exec ./syntinel-server -e production -p 8080
