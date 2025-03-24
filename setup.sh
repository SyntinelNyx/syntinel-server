#!/bin/sh

# Create the "data" directory if it doesn't exist
mkdir -p data

# Create the server CA certs using ECDSA.
openssl req -x509                                     \
  -newkey ec                                         \
  -pkeyopt ec_paramgen_curve:prime256v1              \
  -nodes                                             \
  -days 3650                                         \
  -keyout data/ca_key.pem                            \
  -out data/ca_cert.pem                              \
  -subj /C=US/ST=CA/L=SVL/O=gRPC/CN=test-server_ca/  \
  -config ./openssl.cnf                              \
  -extensions test_ca                                \
  -sha256

# Generate a server key using ECDSA.
openssl ecparam -name prime256v1 -genkey -noout -out data/server_key.pem

# Generate a certificate signing request (CSR) for the server.
openssl req -new                                    \
  -key data/server_key.pem                          \
  -out data/server_csr.pem                          \
  -subj /C=US/ST=CA/L=SVL/O=gRPC/CN=test-server1/   \
  -config ./openssl.cnf                             \
  -reqexts test_server

# Change permissions of the server key and csr to allow postgre to read them
chmod 600 data/server_key.pem 
chmod 600 data/server_csr.pem

# Sign the server certificate with the CA.
openssl x509 -req           \
  -in data/server_csr.pem   \
  -CAkey data/ca_key.pem    \
  -CA data/ca_cert.pem      \
  -days 3650                \
  -set_serial 1000          \
  -out data/server_cert.pem \
  -extfile ./openssl.cnf    \
  -extensions test_server   \
  -sha256

# Change permissions of the server key and csr to allow postgre to read them
chmod 600 data/server_cert.pem

# Verify the generated server certificate.
openssl verify -verbose -CAfile data/ca_cert.pem data/server_cert.pem

# Generate ECDSA P-256 private key
openssl ecparam -genkey -name prime256v1 -noout -out data/ecdsa_private.pem

# Generate the corresponding public key
openssl ec -in data/ecdsa_private.pem -pubout -out data/ecdsa_public.pem

# Clean up CSR file.
rm data/server_csr.pem
