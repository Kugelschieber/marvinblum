#!/bin/bash

mkdir cert
cd cert
openssl req -new -text -passout pass:$1 -subj /CN=localhost -out server.req -keyout privkey.pem
openssl rsa -in privkey.pem -passin pass:$1 -out server.key
openssl req -x509 -in server.req -text -key server.key -out server.crt
chown 0:70 server.key
chmod 640 server.key
echo "done"
