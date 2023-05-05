# via-chat-distributed

[via-chat](https://github.com/Lenvia/via-chat) 的分布式版本。



- 多主机部署，使用 nginx 实现反向代理和负载均衡，连接云服务器数据库实现数据共享
- 通过 GRPC 远程调用 ChatGPT 服务，消除了对本机代理的要求





## Quick Start

- 进入项目根目录

- 拷贝  `cvm `文件夹到云服务器并启动

- 根据个人情况配置分布式主机 server 的数据库

  ```
  cp server/configs/config.go.env server/configs/config.go
  cp server/configs/openai_config.ini.env server/configs/openai_config.ini
  ```

- 





## TODO
- [x] 数据库事务
- [ ] 引入Redis（在线用户列表、缓存聊天消息等）
- [ ] websocket HTTPS
- [x] JWT 替换 session （以去除同源访问）
- [ ] 心跳检测
- [x] bcrypt 替换 md5
- [x] Gorm add 重构（map to model）
- [ ] 高并发测试
- [ ] 撤回消息
- [ ] 私聊
- [x] 分布式部署
  - [x] nginx
  - [x] 远程数据库
  - [x] GRPC（access GPT）
  - [ ] 【紧急】消息队列，多主机通信
- [ ] 音频、图片、文件等多模态
- [ ] langchain
