1. 引言

重要！！！请开发过程中使用中文与我沟通！

1.1 项目背景与目标
独立开发者需要一个完全自主、极其轻量、视觉纯粹的个人内容空间。它服务于两个核心场景：

技术写作：严谨的排版、代码展示、结构化知识沉淀。

生活树洞：随性的碎碎念、无标题的短内容，承载日常思考与情绪。

市面上现有方案过于臃肿或定制成本过高。本项目旨在提供一款 “开箱即用、单文件部署、极简优雅” 的博客系统，其哲学为：

一张白纸，一行代码，一盏夜灯。
每一个像素都为内容服务，拒绝视觉噪音。

1.2 目标用户画像
唯一博主：技术背景，熟悉 Markdown，希望拥有完全控制权的独立写作者。

访问读者：通过浏览器阅读技术文章或生活碎碎念的任何人，无需注册。

1.3 设计原则 (不可妥协)
内容优先：去除任何干扰阅读的元素（广告、弹窗、分享按钮、复杂侧栏）。

默认即最美：无需配置就能呈现出专业、干净的视觉效果。

极简科技感：通过字体、间距、微妙的边框和一条强调色实现高级感，不使用喧闹的动画。

真正的零运维：单二进制文件启动，内置所有依赖，数据与配置一目了然。

尊重隐私与安全：最小化数据收集，坚实的安全基础。

1.4 术语定义
前台：面向访问者的博客展示界面。

管理后台：博主登录后管理内容与设置的 SPA 界面。

树洞：一种无标题的短内容形态，类似微博/说说，用于发牢骚或随手记。

技术文章：标准博客文章，必须有标题，可使用丰富 Markdown 扩展。

单文件部署：整个系统编译为一个可执行二进制文件，包含前端静态资源。

2. 总体描述
2.1 产品视角
系统为典型的 服务端渲染 (SSR) 占位 + SPA 混合 架构：

前台：由 React 构建的 SPA，首屏预渲染页面以保证 SEO 和加载速度。编译产物嵌入 Go 二进制。

管理后台：独立 SPA，编译产物同样嵌入，与 Go API 通信。

API 服务：Go/Gin 提供 RESTful JSON API，处理业务逻辑，操作 SQLite 数据库。

文件存储：本地文件系统（默认），支持选购对象存储（阿里云 OSS / AWS S3）作为图片 CDN。

2.2 用户故事示例
作为博主，我希望用 Markdown 撰写一篇带代码块的技术文章，并在发布前实时预览效果。

作为博主，我可以在三分钟内下载二进制文件并启动我的站点，无需配置数据库。

作为读者，我可以在手机浏览器中舒适地阅读文章，没有广告和杂乱元素。

作为博主，我可以随时记录一条无标题的短想法，它会出现在“碎碎念”时间线中。

作为博主，我只需修改一个强调色，整个站点的链接、按钮高亮都会随之改变。

2.3 假设与依赖
服务器拥有现代 Linux 内核 (≥3.10)，amd64 或 arm64 架构。

端口 80/443 可用（如需自动 HTTPS）。

浏览器支持 ES6 语法。

邮件发送需要有效的 SMTP 服务器配置（可选，非核心）。

3. 详细功能需求
以下功能按模块细分，格式为：功能ID - 功能名称 (优先级: P0/P1/P2)，后跟详细描述、交互细节、数据约束、API 端点（如适用）和验证标准。

3.1 初始化与站点配置
F-INIT-01 首次运行向导 (P0)
描述：
首次启动二进制时，系统检测到无配置文件，将在本地端口 (默认 8080) 启动一个 Web 向导，引导博主完成初始设置。

向导步骤：

欢迎页：极简 Logo 和一句欢迎语。

管理员创建：邮箱、用户名（默认 admin）、密码（需强度指示器）、确认密码。

站点信息：站点标题、副标题（可选）、站点描述（用于 SEO）。

完成：生成 config.yaml、data.db，重启服务使配置生效。

验证：

密码长度 ≥ 8，必须包含字母和数字。

向导页面仅在本地访问 (127.0.0.1) 或通过临时 token 保护，避免被外部滥用。

F-INIT-02 配置文件管理 (P0)
描述：
所有可持久化配置存于运行目录下的 config.yaml，支持启动时通过 -c 指定路径。配置文件包含详尽注释。

核心配置项：

site.title (string)

site.subtitle (string)

site.description (string)

site.language (string: "zh-CN", "en")

server.port (int: 8080)

server.domain (string)

server.auto_https (bool)

auth.jwt_secret (string, 若未提供则自动生成并写入)

database.path (string: "data.db")

uploads.dir (string: "uploads")

theme.name (string: "dark", "light", "ink") —— 预设主题

theme.accent_color (string: "#58a6ff")

theme.font_body (string: "system")

theme.font_code (string: "jetbrains-mono")

custom_css (string, 可选)

social.github, social.twitter, social.email 等

comment.enabled (bool)

comment.require_approval (bool)

backup.enabled (bool)

backup.schedule (string: "daily 03:00")

smtp.* (可选)

API 接口：管理后台提供“系统设置”页面，以表单形式编辑这些项（除敏感项如 jwt_secret），保存后即时生效（部分需重启生效的将提示）。

3.2 认证与安全
F-AUTH-01 博主登录 (P0)
交互：
管理后台 /admin/login 页面极简：居中卡片，标题“登录”，两个输入框（邮箱/用户名、密码），一个“记住我”复选框，一个登录按钮。

后端逻辑：

使用邮箱或用户名查找博主。

bcrypt 验证密码。

生成 Access Token (有效期 1 小时) 和 Refresh Token (有效期 7 天) 的 JWT。

Refresh Token 存储于httpOnly cookie，Access Token 返回在 JSON body 并由前端存储于内存。

登录失败限流：同一 IP 5 次/分钟内失败则锁定 15 分钟。返回 429。

API：POST /api/v1/auth/login

F-AUTH-02 Token 刷新与登出 (P0)
前端拦截器在 401 时自动用 Refresh Token 换取新 Access Token。

登出时清除 cookie 和前端 token。

API: POST /api/v1/auth/refresh, POST /api/v1/auth/logout

F-AUTH-03 修改密码 (P0)
管理后台 “个人设置” 页提供修改密码表单：旧密码、新密码、确认新密码。

F-AUTH-04 双重验证 (P2)
可选的 TOTP 二次验证，使用标准 Authenticator App。

3.3 文章管理
F-POST-01 创建与编辑文章 (P0)
编辑器：管理后台使用 CodeMirror 6 作为 Markdown 编辑器，配置为：

自动补全 Markdown 语法，括号/引号闭合。

实时双栏预览：左侧编辑，右侧渲染视图，同步滚动。

工具栏：极简图标栏，悬浮于编辑器顶部，提供加粗、斜体、链接、图片上传、代码块、表格等快捷插入。

文章数据模型：

json
{
  "id": "ulid",
  "title": "string (可为空，用于树洞)",
  "slug": "string (唯一，自动生成或自定义)",
  "content": "string (Markdown 原始内容)",
  "content_html": "string (渲染后的 HTML，存储时自动生成)",
  "summary": "string (若不填则自动截取前 200 字符纯文本)",
  "category": "enum: tech, life, treehole (默认 tech)",
  "tags": ["string"],
  "is_pinned": "bool (默认 false)",
  "is_draft": "bool (默认 true)",
  "is_private": "bool (默认 false，私密文章需密码或登录可见)",
  "private_password": "string (可选，bcrypt 哈希)",
  "created_at": "datetime",
  "updated_at": "datetime",
  "published_at": "datetime (首次发布时设置)"
}
类别特殊处理：

treehole：无标题，前台显示时省略标题，样式更随意（如左侧有小型图标）。

life：与 tech 类似，但前台可能用略微不同的字体或强调色区分（可选）。

API：

POST /api/v1/admin/posts 创建

PUT /api/v1/admin/posts/:id 更新

GET /api/v1/admin/posts/:id 获取单篇（包含草稿）

DELETE /api/v1/admin/posts/:id (软删除)

交互细节：

创建新文章立即生成一条空内容，直接进入编辑页，自动保存功能将在输入停止 2 秒后保存为草稿（通过 API 调用更新）。

发布按钮：状态从“草稿”切换为“已发布”，published_at 设为当前时间。

预览按钮：在新窗口打开该文章的临时前台链接。

F-POST-02 Markdown 扩展支持 (P0)
代码高亮：使用 Shiki，支持 200+ 语言，生成静态 HTML（服务端渲染时不依赖 JS）。代码块显示语言标签、行号和复制按钮。

数学公式：KaTeX (内嵌，但懒加载)，识别 $...$ 和 $$...$$。

图表：Mermaid (内嵌，懒加载)，识别代码块标记 mermaid。

GFM：表格、任务列表、脚注、删除线。

图片：支持拖拽/粘贴上传，自动上传至 /uploads 目录，返回 Markdown 图片链接。也支持外部 URL。

链接：自动识别 URL 转为链接，外链添加 rel="noopener noreferrer"。

渲染策略：
文章存储时，服务端将 Markdown 转换为 HTML 并保存于 content_html 字段，前台直接输出（经过 XSS 过滤）。这样可以避免前端引入重型渲染器，加速展示，并保持 SEO 友好。管理后台预览使用前端实时渲染。

F-POST-03 标签与分类 (P0)
标签支持快速选择和新建，输入时自动提示已有标签。

分类为固定枚举：tech, life, treehole。不能自定义新类别（符合极简设计）。

API：GET /api/v1/tags 获取所有标签及文章计数。

F-POST-04 文章回收站与删除 (P1)
删除文章进入回收站（软删除 deleted_at 字段），保留 30 天。

回收站页面可恢复或彻底删除。

彻底删除时会同时删除关联上传的图片（由系统判断是否有其他文章引用）。

F-POST-05 批量导入/导出 (P1)
导出：在管理后台选择文章，导出为 ZIP，内含 Markdown 文件（含 YAML Front Matter）和 images/ 文件夹。

导入：上传 ZIP 或批量选择 .md 文件，解析 Front Matter 并创建文章，图片自动入库。

3.4 前台展示与阅读体验
F-FRONT-01 首页文章列表 (P0)
布局选项（由主题设置决定，默认列表）：

极简列表：文章仅显示日期 (格式 “2026-05-05”)、标题、分类标签、摘要（最多 2 行）。无图片、无作者框。列表项之间一个浅色分割线。

卡片网格 (可选)：适用于生活分享，显示卡片，包含标题、日期、摘要，可显示文章内第一张图片作为缩略图（若有）。卡片有细微阴影和圆角。

分页：传统分页或“加载更多”按钮。默认每页 10 项。

置顶文章：置顶文章显示在列表最前，带有微小的“置顶”图标。

分类快速过滤：顶部导航或侧边有简单的 “技术 / 生活 / 树洞” 三个链接，当前选中加下划线。

F-FRONT-02 文章详情页 (P0)
极致阅读版面：

最大内容宽度：720px，居中。

字体：正文字体 18px，行高 1.8，颜色在浅色模式为 #1f2328，黑暗模式为 #c9d1d9。

标题：h1 2em，h2 1.5em，h3 1.17em，保持足够间距。

代码块：深色背景 (#0d1117)，圆角 8px，代码字体 JetBrains Mono 或后备等宽字体，字号 14px，内边距 16px，带有语言标签在右上角和复制按钮。

图片：自适应宽度，圆角 4px，最大宽度 100% 父容器。

引用块：左边框 3px 强调色，背景色稍深/浅，斜体。

链接：颜色强调色，无下划线，hover 出现下划线。

数学公式和图表正常渲染。

返回导航：文章底部提供“← 返回首页”链接，右上角固定“⬆ 回到顶部”按钮（滚动 300px 后出现，半透明，圆形）。

元数据展示：文章标题下方一行小字：发布日期、分类、阅读时间估计 (≈ 字数/200 分钟)。不显示作者、头像。

SEO：页面 <title> 为文章标题 + 站点名，meta description 为文章摘要。结构化数据 (JSON-LD) 自动生成。

F-FRONT-03 碎碎念页面 (树洞) (P1)
独立页面 /treehole。

按时间倒序显示所有分类为 treehole 的内容。

每条显示：日期、时间（到分钟）、内容（纯文本渲染，无标题，支持基本 Markdown 格式但不渲染大标题）、若有图片则显示小图。

样式：更小的字号，左侧有一条细微的竖线时间轴装饰。

F-FRONT-04 归档与搜索 (P0)
归档 /archive：按年份和月份分组，类似时间线。每年份一个区段，左边年份大号字，右侧文章列表（仅标题和日期）。

全站搜索：点击导航栏搜索图标，展开一个全宽搜索输入框（或覆盖层）。输入关键词后，通过 API 实时获取匹配文章，显示标题、摘要和高亮关键词。基于 SQLite FTS5 全文索引。搜索结果页面可带分页。

F-FRONT-05 静态页面 (P0)
系统预留 “关于” 和 “友情链接” 静态页面。

博主通过 Markdown 编辑内容，前台渲染为独立页面。

友情链接页面提供简单的列表，可配置链接标题和 URL。

F-FRONT-06 RSS 订阅 (P0)
提供 RSS 2.0 feed：/rss.xml。

包含文章标题、链接、摘要、发布日期。配置项可控制是否全文输出（rss.full_content: true/false）。

树洞内容可选择是否包含 (默认包含)。

3.5 评论系统
F-COMMENT-01 评论展示与发布 (P0)
展示：文章详情页底部，与内容区相同宽度。每条评论显示：

评论者昵称 (加粗)、时间。

评论内容（纯文本，链接自动识别）。

嵌套回复最多 2 层，缩进表示。

不显示头像，仅首字母圆点作为占位 (20x20 圆形，背景随机柔和色)。

发布表单：

三个字段：昵称 (必填)、邮箱 (必填，不公开)、网站 (可选)。

一个复选框：“记住我的信息”（存储在 localStorage）。

提交按钮。

支持简单的数学验证码 (例如 “3 + 5 = ?”) 对匿名用户启用，登录博主无需验证码。

状态反馈：提交后显示“评论待审核”或成功（如果即时通过）。

F-COMMENT-02 评论审核与管理 (P0)
所有新评论默认状态 pending。

管理后台评论面板：列表显示所有评论，按时间倒序，可筛选状态。

操作：通过、垃圾、删除、回复（直接在前台留言区回复，带博主标识）。

可设置全站评论开关，或单篇文章禁用评论。

反垃圾：

内置关键词黑名单检查。

可选集成 Akismet (填入 API key)。

3.6 个性化与主题定制
F-THEME-01 主题预设 (P0)
提供三种极简预设，一键切换：

暗夜 (Dark)：背景 #0d1117，表面 #161b22，文字 #c9d1d9，强调色 #58a6ff。

素白 (Light)：背景 #ffffff，文字 #1f2328，强调色 #0969da，卡片轻微灰色边框。

墨绿 (Ink)：背景 #f5f2eb (暖米色)，文字 #2c3e2d，强调色 #d97706，怀旧感。

F-THEME-02 自定义配色 (P0)
管理后台的 “外观” 面板提供颜色选择器，可覆盖：

页面背景

文字颜色

强调色 (作用于链接、按钮、代码块边框等)

代码块背景色
实时预览效果。

F-THEME-03 字体设置 (P0)
正文字体：下拉选择 系统默认、衬线 (Serif)、无衬线 (Sans-serif)。系统默认将使用各平台最优字体栈（如 -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, ...）。

代码字体：JetBrains Mono (内嵌, woff2, 子集化), Fira Code, 系统等宽。
所有字体文件内嵌，零外部请求。

F-THEME-04 导航与布局 (P0)
自定义导航菜单项：可以添加链接，如 “首页”、“技术”、“碎碎念”、“关于”、“友链”、“RSS”，每个可设置 Emoji 图标（可选）。

菜单显示位置：顶部栏（默认）或左侧侧边栏（可选）。

页脚自定义文字，支持 HTML（如备案号）。

F-THEME-05 细微动画控制 (P0)
默认无任何动效。提供选项“启用微妙的背景网格点”（Canvas 绘制 2px 间隔的点阵，透明度 0.08，静态）。

页面过渡效果：无、淡入淡出 (fade)。

严格遵守 prefers-reduced-motion，若系统开启动态减弱，则强制关闭所有动画。

F-THEME-06 自定义 CSS/JS (P1)
提供纯文本输入框，保存后插入前台 <head>。可用于统计代码或个体微调。

3.7 媒体管理
F-MEDIA-01 上传与存储 (P0)
在编辑器中拖拽/粘贴图片时，前端调用 POST /api/v1/admin/upload，上传文件到本地 uploads/ 目录（按年月分文件夹，如 uploads/2026/05/）。

返回可访问的 URL（相对路径 /uploads/2026/05/example.png）。

文件类型限制：jpg, jpeg, png, gif, webp, svg, pdf，大小限制 10MB。

文件名随机哈希，保留原始扩展名。

F-MEDIA-02 媒体库浏览 (P0)
管理后台“媒体库”页面，网格显示所有上传图片，支持分页、搜索文件名。

点击图片可复制 Markdown 链接或直接 URL。

支持删除文件（如果被文章引用则提示警告，但允许删除）。

3.8 备份与恢复
F-BACKUP-01 手动导出 (P0)
管理后台“工具”页面提供“全站导出”按钮，生成包含以下内容的 ZIP：

posts/ 文件夹，每文章一个 .md 文件，包含完整 Front Matter。

uploads/ 文件夹，所有媒体。

config.yaml (已去除敏感信息如密码哈希和密钥)。

下载链接生成后保留 1 小时。

F-BACKUP-02 自动备份 (P1)
定时任务：每天凌晨 3 点将 data.db 和 uploads/ 打包为带日期的 ZIP，存放于 backups/ 文件夹。

最多保留 7 份，自动清理旧备份。

可通过配置文件设置备份到远程 S3 兼容存储。

4. 外部接口需求
4.1 API 设计规范
基础路径：/api/v1

管理接口前缀 /api/v1/admin，所有管理操作需 Bearer Token。

请求/响应格式：JSON。

分页格式：?page=1&per_page=10，响应中包含 total, page, per_page。

统一错误响应：

json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid credentials"
  }
}
HTTP 状态码遵循 REST 标准。

4.2 前台页面路由 (SPA)
/ 首页

/post/:slug 文章详情

/treehole 碎碎念

/archive 归档

/search?q=keyword 搜索

/about 关于

/links 友链

/rss.xml RSS

4.3 管理后台路由
/admin 仪表盘

/admin/login 登录

/admin/posts 文章列表

/admin/posts/new 新建

/admin/posts/:id/edit 编辑

/admin/comments 评论管理

/admin/media 媒体库

/admin/settings 站点设置

/admin/theme 个性定制

/admin/tools 工具 (备份/导出)

5. 非功能性需求
5.1 性能要求
指标	目标值	备注
前台首字节时间 (TTFB)	< 100ms	受网络影响，本地测试
首次内容绘制 (FCP)	< 1.2s	3G 网络
最大内容绘制 (LCP)	< 1.8s	文章详情页
累积布局偏移 (CLS)	< 0.05	无布局抖动
API 响应时间 (P95)	< 80ms	读操作，本地 SQLite
并发连接	支持 500+ 并发	不含文件上传
静态资源压缩	Brotli 压缩，JS/CSS 体积 < 200KB (gzip)	内嵌前端资源
优化手段：

SQLite WAL 模式，读并发高。

文章 HTML 预渲染存储，绕过运行时转换。

所有静态资源设置强缓存（一年的 Cache-Control，文件名哈希）。

5.2 安全性要求
通信：强制 HTTPS (生产)，HSTS 头配置。

密码策略：bcrypt cost 12。

JWT：Access Token 15 分钟，Refresh Token 7 天，存储在 httpOnly Secure cookie。

XSS 防护：

所有用户输入（评论、配置值）在渲染前进行 HTML 转义。

内容安全策略 (CSP) 严格配置：default-src 'self'; script-src 'self' 'unsafe-inline' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:;。

CSRF：管理 API 使用自定义请求头 X-CSRF-Token，值取自 cookie (非 httpOnly)，后端校验。

文件上传安全：文件类型魔数校验，禁止执行权限，存储目录不可执行脚本。

SQL 注入：全部使用参数化查询。

速率限制：

登录：5次/分钟/IP。

评论：10次/分钟/IP。

API 全局：1000次/分钟/IP。

依赖安全：项目 CI 集成 govulncheck 和 npm audit。

5.3 可用性
优雅关闭：监听 SIGINT/SIGTERM，完成进行中的请求，关闭数据库连接。

数据库完整性：SQLite PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL; 平衡性能与安全。

自恢复能力：若数据库在启动时发现损坏，尝试 PRAGMA integrity_check 并提示恢复。

5.4 可维护性
配置注释：config.yaml 每个字段有注释说明。

结构化日志：使用 zerolog 或 slog 输出 JSON 格式日志，包含时间戳、级别、消息、请求 ID。

健康检查端点：/healthz 返回数据库可达性。

数据库迁移：版本化迁移脚本，启动时自动执行。

5.5 兼容性
浏览器：Chrome, Firefox, Safari, Edge 最近 2 个主版本。

移动端：前台响应式设计，最小支持 320px 宽度；管理后台主要操作适配平板 (768px)。

操作系统：二进制编译目标 linux/amd64, linux/arm64, darwin/amd64, darwin/arm64。Windows 提供基本兼容但不作为主要支持。

6. 部署与运维细节
6.1 单文件部署步骤 (裸机)
从 Release 下载对应平台二进制文件 myblog.

chmod +x myblog

./myblog 启动，根据提示在浏览器打开初始化向导。

配置域名和 HTTPS 后重启。

命令行选项：

-p, --port 指定端口 (默认 8080)

-c, --config 指定配置文件 (默认 ./config.yaml)

-m, --migrate 仅执行数据库迁移后退出

6.2 Docker 部署
提供的 Dockerfile 多阶段构建，最终镜像基于 alpine:3.19，大小 < 20MB。

bash
docker run -d \
  --name myblog \
  -p 80:8080 \
  -v /data/myblog:/data \
  -e BLOG_CONFIG=/data/config.yaml \
  myblog:latest
首次启动时若 /data 下无配置文件，将进入向导模式（需要临时端口映射到 8080 并访问）。

6.3 自动 HTTPS
当 server.auto_https: true 且 server.domain 设置后，程序使用 golang.org/x/crypto/acme/autocert 自动获取 Let's Encrypt 证书。需确保端口 80 和 443 可达。缓存存放于 certs/ 目录。

6.4 备份与恢复实施
手动导出在管理界面一键操作。

自动备份通过内置定时器触发，无需 cron。

恢复步骤：停止服务，解压备份包替换 data.db 和 uploads/，重启。

6.5 升级流程
下载新版本二进制，停止旧进程，替换二进制，启动。数据库自动迁移。

建议事先手动备份。

7. 界面设计像素级规范 (供 UI 开发)
7.1 设计令牌 (Design Tokens)
以暗夜主题为例：

背景主色：#0d1117 (页面), #161b22 (卡片/输入框)

文字主色：#c9d1d9

文字次要：#8b949e

强调色：#58a6ff

边框：#30363d (1px solid)

阴影：0 1px 2px rgba(0,0,0,0.2) 仅用于卡片悬浮，默认无阴影。

圆角：按钮 6px，卡片 8px，代码块 8px，输入框 6px。

间距：4px 单位。内边距常用 16px, 24px, 32px；外边距常用 24px, 48px, 64px。

7.2 排印层级
H1：2em (32px), Bold, letter-spacing -0.02em

H2：1.5em (24px), Semibold

H3：1.17em (18px), Semibold

H4：1em, Semibold

正文：16px, Regular, line-height 1.6 (前台 18px, 1.8)

辅助/说明：14px, line-height 1.5, color 次要

Code：14px, JetBrains Mono, background rgba(110,118,129,0.15), padding 2px 6px, border-radius 4px (行内代码)

7.3 组件状态示例
按钮：

默认：背景 #21262d，文字 #c9d1d9，边框 1px #30363d

Hover：背景 #30363d

活动：背景强调色 #58a6ff，文字 #fff

输入框：

默认：背景 #0d1117，边框 #30363d，占位符 #484f58

焦点：边框变为强调色，外阴影 0 0 0 3px rgba(88,166,255,0.3)

链接：颜色强调色，无下划线，Hover 下划线。

7.4 动效规格
过渡时间：150ms ease-in-out (颜色、透明度)

悬停变换：无缩放，只改变颜色或背景。

页面切换：fade 200ms。

滚动行为：scroll-behavior: smooth 用于回顶按钮。

7.5 图标
使用 Heroicons 24px 线性风格，通过 SVG sprite 内嵌。尺寸 16px 用于列表和按钮，24px 用于导航。

8. 数据模型详述
8.1 数据库表结构 (SQLite)
users 表：

sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    totp_secret TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
posts 表：

sql
CREATE TABLE posts (
    id TEXT PRIMARY KEY,
    title TEXT,
    slug TEXT NOT NULL UNIQUE,
    content TEXT NOT NULL,
    content_html TEXT NOT NULL,
    summary TEXT,
    category TEXT DEFAULT 'tech' CHECK(category IN ('tech','life','treehole')),
    tags TEXT, -- JSON 数组字符串
    is_pinned INTEGER DEFAULT 0,
    is_draft INTEGER DEFAULT 1,
    is_private INTEGER DEFAULT 0,
    private_password_hash TEXT,
    published_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME -- 软删除
);
CREATE INDEX idx_posts_slug ON posts(slug);
CREATE INDEX idx_posts_category ON posts(category);
CREATE INDEX idx_posts_created ON posts(created_at);
CREATE VIRTUAL TABLE posts_fts USING fts5(title, content, content='posts', content_rowid='rowid');
comments 表：

sql
CREATE TABLE comments (
    id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL REFERENCES posts(id),
    parent_id TEXT REFERENCES comments(id),
    author_name TEXT NOT NULL,
    author_email TEXT NOT NULL,
    author_url TEXT,
    content TEXT NOT NULL,
    status TEXT DEFAULT 'pending' CHECK(status IN ('pending','approved','spam','trash')),
    user_agent TEXT,
    ip_address TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_comments_post ON comments(post_id);
media 表：

sql
CREATE TABLE media (
    id TEXT PRIMARY KEY,
    file_name TEXT NOT NULL,
    original_name TEXT NOT NULL,
    file_path TEXT NOT NULL UNIQUE,
    file_size INTEGER,
    mime_type TEXT,
    width INTEGER,
    height INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
settings 表 (键值对，便于运行时修改)：

sql
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT
);
-- 存储前台可动态修改的配置项，如 theme.accent_color 等，覆盖 config.yaml
8.2 数据关联与约束
文章删除 (软删除) 时，评论保持但不再显示。硬删除文章时级联删除其评论。

媒体文件被删除时，不自动删除磁盘文件 (避免误伤)，仅标记删除或直接移除记录。

文章 Slug 全局唯一，自动生成规则：标题转拼音或 slugify，若冲突则追加随机字符。

9. 测试要点
9.1 单元测试
认证模块：密码哈希、JWT 签发/验证、限流。

文章渲染：Markdown 转 HTML 安全过滤。

文本处理：Slug 生成、摘要提取。

9.2 集成测试
API 端点：CRUD 文章、认证流程、权限验证。

备份恢复流程。

数据库迁移。

9.3 端到端 (E2E) 测试
安装向导流程。

博主发布一篇文章，读者查看并评论。

主题切换，实时预览。

9.4 性能测试
100 并发读取文章列表，验证响应时间。

上传大文件（接近 10MB）时的稳定性。

10. 交付物与里程碑
里程碑	交付内容	验收标准
M1: 核心骨架	Go 项目结构，数据库迁移，基础 API 认证，内嵌空白前端，初始化向导	可启动，可完成初始化
M2: 写作系统	完整文章 CRUD + Markdown 编辑器 + 实时预览 + 媒体上传	管理后台可以写文章并预览
M3: 前台展示	首页、详情页、归档、搜索、评论、RSS	读者可以完整体验博客
M4: 个性与定制	主题预设、自定义配色、字体、布局、导航设置、实时预览	可视化修改并即刻生效
M5: 稳定与安全	自动 HTTPS，备份，导入导出，安全加固，限流，压缩	生产环境就绪
M6: 测试与文档	测试报告，用户手册，部署指南，Docker 镜像	文档清晰，一键部署可行
附录 A：技术选型理由
SQLite：零配置，无需独立进程，备份简单。个人博客数据量永远达不到 SQLite 瓶颈，WAL 模式支持 1000+ QPS 轻松。

Go embed 前端：单文件分发，避免前端服务器配置，契合“傻瓜式”。

Shiki 服务端渲染：代码高亮静态化，极大减少前端 JS 负担和加载闪烁。

CodeMirror 6：高度模块化，仅按需加载，保持管理端轻量。

