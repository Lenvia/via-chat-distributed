#!/bin/bash

nginx -s stop
if pgrep nats-server > /dev/null
then
    pkill nats-server
fi
