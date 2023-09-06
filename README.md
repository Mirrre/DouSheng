# **抖声--第六届字节跳动青训营项目**


### 本项目由gorm,gin,minIO,MySQL实现极简版抖音客户端
>汇报文档https://u92d1oq3n1.feishu.cn/docx/OXzAdMVEyoSVngxDTojcft0Knpf

# 文件结构


- [config](config)             *应用程序的配置文件 初始化db连接*
- [consts](consts)  *常量定义*
- [middleware](middleware) *中间件*
- [modules](modules)   *API功能实现*
  - [comment](modules/comment) 
   -  [favorite](modules/favorite)   
   -  [message](modules/message) 
   - [models](modules/models) *表单结构体模块*
   - [relation](modules/relation)
   - [user](modules/user)
   - [video](modules/video)
- [utils](utils) *工具包*
   - [responses.go](utils/responses.go) *http响应结构体*
   - [testutils.go](utils/testutils.go)  *单元测试工具* 
   - [token.go](utils/token.go) *Token鉴权函数*
- [docker-compose.yaml](docker-compose.yaml) *Docker容器本地化配置文件*
- [Dockerfile](Dockerfile)
- [go.mod](go.mod) *Go模块定义文件*   
- [go.sum](go.sum)  *模块的预期内容*
- [main.go](main.go) *开启服务主函数*
- [wait-go-it.sh](wait-go-it.sh) *数据库端口响应等待脚本*
- [LICENSE](LICENSE.txt)
 
# 技术亮点
  ### 使用单元测试便于纠错，实现自动化postman测试
  ### 使用oss对象存储存放视频文件
  ### 数据库使用索引，二分法等提高数据检索和更新速率
  ### 返回标准HTTP状态码
  ### 按照功能拆分文件结构



# 部署环境
### 两种部署方式
 1. [1024code](https://1024code.com/)申请线上GO1.20环境
 2.  Docker本地部署

 

# 开启服务

 **1024部署的项目**：运行  `go run main.go`   
 和`./main` **获取URL** 
 
 [抖声APP](https://bytedance.feishu.cn/docs/doccnM9KkBAdyDhg8qaeGlIz7S7#) 中复制URL到**高级设置**
 
 ---
  本地Docker部署运行命令    `docker compose up `

# 单元测试

运行命令 `GO_TESTING=true go test ./... -v`

# 排除问题

-  shell运行`go mod tidy`

# 贡献者
### 啊又寸又寸队go
- [claw16](https://github.com/claw16)
- [Mirrre](https://github.com/Mirrre)
- [moocx](https://github.com/moocx)

# 如何贡献代码
从PR到merge，期待您的贡献！

fork本仓库到你的仓库.

创建一个新的分支修改代码.

提交一个 Pull Request 到主分支.

等待 review 和合并.
# 开源协议
本项目采用[MIT LICENSE](LICENSE.txt)
