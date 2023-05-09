#!/bin/bash

nginx -s stop
if pgrep nats-server > /dev/null
then
    pkill nats-server
fi

if pgrep redis-server > /dev/null
then
    # 如果已经在运行，先暂停它
    pkill redis-server
fi

