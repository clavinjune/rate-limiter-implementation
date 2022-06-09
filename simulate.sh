#! /bin/bash


sleep 1
for i in $(seq 1 6); do
  curl -i "http://localhost:8000?id=$1"
  echo ""
done
