#!/bin/bash

outfile="$1"
num_clients="$2"
echo "output file: $outfile"
echo "number of clients: $num_clients"

cat > $outfile <<'EOF'
name: tp0

networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24

services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
    networks:
      - testing_net
    volumes:
      - ./server/config.ini:/config.ini

EOF

for i in $(seq 1 "$num_clients"); do
    cat >> $outfile <<EOF
  client${i}:
    container_name: client${i}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=${i}
    volumes:
      - ./client/config.yaml:/config.yaml
      - ./.data/agency-${i}.csv:/data/agency-${i}.csv
    networks:
      - testing_net
    depends_on:
      - server

EOF
done
