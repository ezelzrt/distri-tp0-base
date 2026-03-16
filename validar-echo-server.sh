#!/bin/bash

NETWORK="tp0_testing_net"
SERVER_HOST="server"
SERVER_PORT=12345
TEST_MSG="es igual"

docker run --rm --network "$NETWORK" nicolaka/netshoot \
  sh -c "printf '%s' '$TEST_MSG' | nc -w 3 -q 1 $SERVER_HOST $SERVER_PORT | grep -q '^$TEST_MSG$'"

if [ $? -eq 0 ]; then
  echo "action: test_echo_server | result: success"
else
  echo "action: test_echo_server | result: fail"
fi