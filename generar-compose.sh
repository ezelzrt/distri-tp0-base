#!/bin/bash

outfile="$1"
num_clients="$2"
echo "output file: $outfile"
echo "number of clients: $num_clients"

cat > $outfile <<'EOF'
name: tp0

networks:
  tp0_net:
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
      - LOGGING_LEVEL=DEBUG
    networks:
      - tp0_net
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
      - CLI_LOG_LEVEL=DEBUG
    volumes:
      - ./client/config.yaml:/config.yaml
    networks:
      - tp0_net
    depends_on:
      - server

EOF
done
