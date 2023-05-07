#!/bin/bash

# 启动Nginx
nginx -s stop

nginx
echo "nginx已在后台启动"

# 检查nats-server是否已经在运行
if pgrep nats-server > /dev/null
then
    # 如果已经在运行，先暂停它
    pkill nats-server
fi

# 启动NATS
nats-server &

# 输出提示信息
echo "NATS server已在后台启动"