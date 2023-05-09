# via-chat-distributed

[via-chat](https://github.com/Lenvia/via-chat) 的分布式版本。



- 多主机部署，使用 nginx 实现反向代理和负载均衡，连接云服务器数据库实现数据共享
- 使用 NATS 消息队列，实现多主机消息广播
- 通过 GRPC 远程调用 ChatGPT 服务，消除了对本机代理的要求
- 使用 Redis 缓存历史消息，减少对数据库访问次数，新消息异步插入数据库





## Quick Start

- 配置数据库和 GPT 服务所在的服务器

  ```
  cd cvm
  go run main.go
  ```

- （请自行安装 Nginx 和 NATS 服务）在调度主机上启动 Nginx 和 NATS

  ```
  cd dispath
  cp nginx.conf <本地主机的 nginx 配置文件路径>
  sh scripts/start-services
  ```

- 分布式主机 server 配置文件，拷贝后根据实际情况修改

  ```
  cd server
  cp configs/config.go.env server/configs/config.go
  cp configs/openai_config.ini.env server/configs/openai_config.ini
  ```

- 分布式主机启动服务

  ```
  cd server
  go run main.go
  ```

  






## TODO
- [x] 数据库事务
- [ ] websocket HTTPS
- [x] JWT 替换 session （以去除同源访问）
- [ ] 心跳检测
- [x] bcrypt 替换 md5
- [x] Gorm add 重构（map to model）
- [ ] 高并发测试
- [ ] 撤回消息
- [ ] 私聊
- [x] 分布式部署
  - [x] nginx（ip_hash）
  - [x] 远程数据库
  - [x] GRPC（access GPT）
  - [x] NATS 消息队列，多主机通信
  - [x] Redis 缓存历史消息和新消息
- [ ] 音频、图片、文件等多模态
- [ ] langchain
