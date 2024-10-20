# Syntinel Server

### Setup

Generate ECDSA P-256 public and private key with the below commands:

```
openssl ecparam -genkey -name prime256v1 -noout -out ecdsa_private.pem
openssl ec -in ecdsa_private.pem -pubout -out ecdsa_public.pem
```
