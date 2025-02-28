#!/bin/sh

mkdir -p /app/data

# List of required files
FILES="ca_cert.pem ca_key.pem config.yaml ecdsa_private.pem ecdsa_public.pem server_cert.pem server_key.pem"

for file in $FILES; do
    if [ ! -f "/app/data/$file" ]; then
        cp "/app/defaults/$file" "/app/data/$file"
    fi
done

exec ./syntinel-server -e production -p 8080
