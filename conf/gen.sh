#!/bin/bash

NodeList=("node1" "node2" "node3" "node4")

rm -rf ca > /dev/null 2>&1

cadir="./ca"
mkdir $cadir

# 1. gen ca.key
openssl genrsa -out ${cadir}/ca.key 4096
# 2. gen ca.csr
openssl req -new -sha256 -out ${cadir}/ca.csr -key ${cadir}/ca.key -config ca.conf
# 3. gen ca.crt
openssl x509 -req -days 3650 -in ${cadir}/ca.csr -signkey ${cadir}/ca.key -out ${cadir}/ca.crt

for element in ${NodeList[*]}
do
  rm -rf $element
  mkdir $element
  # server side
  # 1. gen server.key
  openssl genrsa -out ${element}/server.key 2048
  sed "s/{{servername}}/${element}/g" server.conf > ${element}/server.conf

  # 2. gen server.csr
  openssl req -new -sha256 -out ${element}/server.csr -key ${element}/server.key -config ${element}/server.conf

  # 3. gen server.pem with ca.crt and ca.key
  openssl x509 -req -sha256 -CA ${cadir}/ca.crt -CAkey ${cadir}/ca.key -CAcreateserial -days 365 -in ${element}/server.csr -out ${element}/server.crt -extensions req_ext -extfile ${element}/server.conf

  # client side
  # 1. gen client.key
  openssl genrsa -out ${element}/client.key 2048

  # 2. gen client.csr
  openssl req -new -key ${element}/client.key -out ${element}/client.csr

  # 3. gen client.crt with ca and ca.key
  openssl x509 -req -sha256 -CA ${cadir}/ca.crt -CAkey ${cadir}/ca.key -CAcreateserial -days 365  -in ${element}/client.csr -out ${element}/client.crt

done