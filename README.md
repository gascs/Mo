<p align="center">
  <h1 align="center">Mo Blog</h1>
  <p align="center">一张白纸，一行代码，一盏夜灯。</p>
</p>

---

极简、零运维、单文件部署的个人博客系统。技术写作 + 生活树洞，为内容而生的纯粹空间。

## 快速开始

```bash
# 方式一：一键启动（自动安装 Go/Node，构建，运行）
./bootstrap.sh

# 方式二：Docker（只需装 Docker）
make docker-up

# 方式三：传统构建
make build-linux && ./myblog
```

Windows 用户运行 `.\bootstrap.ps1`。

打开 `http://localhost:8080`，首次访问会进入初始化向导。

## 一键部署到服务器

```bash
# 首次部署（自动安装 Go/Node/systemd/防火墙）
make setup HOST=12.34.56.78

# 后续更新
make deploy HOST=12.34.56.78

# 回滚
make rollback HOST=12.34.56.78
```

## 功能

- **Markdown 编辑器** — CodeMirror 6，双栏实时预览，代码高亮 (Shiki)，数学公式 (KaTeX)，图表 (Mermaid)
- **三种内容形态** — 技术文章、生活随笔、树洞碎碎念
- **评论系统** — 嵌套回复、审核机制、反垃圾
- **全站搜索** — SQLite FTS5 全文索引
- **RSS / 归档 / 友链** — 开箱即用
- **三种主题** — 暗夜、素白、墨绿，支持自定义配色
- **单文件部署** — Go 二进制内嵌前端，拷贝即运行
- **自动 HTTPS** — Let's Encrypt 证书自动获取
- **自动备份** — SQLite + 文件定时备份，支持 S3

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go + Gin |
| 数据库 | SQLite (WAL 模式，纯 Go 驱动) |
| 前端 | React + TypeScript + Vite |
| CSS | Tailwind CSS |
| 部署 | 单二进制，Docker，systemd |

## 目录结构

```
Mo/
  main.go              # 入口
  embed.go             # 内嵌前端
  Makefile             # 构建/部署入口
  bootstrap.sh         # 一键启动 (Linux/macOS)
  bootstrap.ps1        # 一键启动 (Windows)
  Dockerfile           # Docker 构建
  web/                 # 前端 (React)
  internal/            # 后端
    handler/           # API 处理器
    service/           # 业务逻辑
    database/          # 数据库 & 迁移
    model/             # 数据模型
    router/            # 路由
    auth/              # 认证 & 安全
    config/            # 配置加载
  scripts/             # 部署脚本
  deploy/              # systemd 配置
```

## 命令行

```bash
./myblog -h              # 帮助
./myblog -v              # 版本信息
./myblog -p 9090         # 指定端口
./myblog -c config.yaml  # 指定配置文件
./myblog -m              # 仅执行数据库迁移
```

## License

MIT
