#!/bin/bash
# run like this : bash ./makecert.sh
mkdir certs
rm certs/*
echo "make server cert"
openssl req -newkey rsa:2048 -x509 -nodes -keyout certs/server.key -new -out certs/server.crt -subj /CN=test.dollop.com
echo "make client cert"
openssl req -newkey rsa:2048 -x509 -nodes -keyout certs/client.key -new -out certs/client.crt -subj /CN=test.dollop.com