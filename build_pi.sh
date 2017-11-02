#!/bin/bash
GOOS=linux
echo Cleanup deployment directory
rm -rf deployment
mkdir -p deployment

echo Build consumer
mkdir -p deployment/consumer/
cd consumer && env GOOS=linux GOARCH=arm GOARM=5 go build && mv consumer ../deployment/consumer/consumer && cd ../
cp consumer/wpwconfig.json deployment/consumer/wpwconfig.json
cp consumer/wpwconfig.json deployment/consumer/wpwconfig.json
mkdir -p deployment/consumer/logs

echo build producer
mkdir -p deployment/producer/
cd producer && env GOOS=linux GOARCH=arm GOARM=5 go build && mv producer ../deployment/producer/producer && cd ../
cp producer/wpwconfig.json deployment/producer/wpwconfig.json
cp producer/wpwconfig.json deployment/producer/wpwconfig.json
mkdir -p deployment/producer/logs
