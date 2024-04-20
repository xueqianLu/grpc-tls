#!/bin/bash

# 1. gen ca.key
openssl genrsa -out ca.key 4096
# 2. gen ca.csr
openssl req -new -sha256 -out ca.csr -key ca.key -config ca.conf
# 3. gen ca.crt
openssl x509 -req -days 3650 -in ca.csr -signkey ca.key -out ca.crt


# server side
# 1. gen server.key
openssl genrsa -out server.key 2048

# 2. gen server.csr
openssl req -new -sha256 -out server.csr -key server.key -config server.conf

# 3. gen server.pem with ca.crt and ca.key
openssl x509 -req -sha256 -CA ca.crt -CAkey ca.key -CAcreateserial -days 365 -in server.csr -out server.crt -extensions req_ext -extfile server.conf



# client side
# 1. gen client.key
openssl genrsa -out client.key 2048

# 2. gen client.csr
openssl req -new -key client.key -out client.csr 


# 3. gen client.crt with ca and ca.key
openssl x509 -req -sha256 -CA ca.crt -CAkey ca.key -CAcreateserial -days 365  -in client.csr -out client.crt
