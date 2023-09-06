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
- [LICENSE](LICENSE)

# 技术亮点
  - 编写了尽可能覆盖率高的单元测试和集成测试用例，保证代码的可靠性，以及提高了在业务逻辑更新的过程中的容错率。即使出错，也可以快速定位到有问题的逻辑代码
  - 使用对象存储存放视频文件，测试过的方案有 Amazon S3 和 minIO
  - 设计优化了数据库索引，最大化提高了查询效率
  - 利用GORM框架
    - 与数据库的交互都是用结构体的实例对象，而不是直接用SQL语句，避免了SQL恶意注入
    - 利用了GORM的hook，在一些对象被创建和删除的时候更新相关数据，利用了MySQL的事务保证了数据库一致性

  - 按照功能拆分文件结构
  - 利用Docker配置了本地测试和开发的环境，提高开发效率的同时，保证了全平台的设置统一



# 部署环境
 1. [1024code](https://1024code.com/)申请线上GO1.20环境
 2.  Docker本地部署

 

# 开启服务

### 1024code

在1024code的代码空间中拉取仓库的`master` 分支，运行：

```bash
go run .
```

###  本地环境

```bash
docker compose up
```



# 单元测试

```bash
GO_TESTING=true go test ./...
```

**常用参数**：

```bash
-v 		# 详细输出
-p 1 	# 强制非并行模式
-cover	# 提供覆盖率报告
```



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
本项目采用[MIT LICENSE](LICENSE)
