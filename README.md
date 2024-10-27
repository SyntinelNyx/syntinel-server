# Syntinel Server

## Setup

### Development Setup

Generate locally signed SSL/TLS certificate & key with the below commands:

```
openssl ecparam -genkey -name secp384r1 -out server.key
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```

Generate ECDSA P-256 public and private key with the below commands:

```
openssl ecparam -genkey -name prime256v1 -noout -out ecdsa_private.pem
openssl ec -in ecdsa_private.pem -pubout -out ecdsa_public.pem
```
