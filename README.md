# 西电教师综合服务平台

面向高校教师服务场景的一体化应用原型，包含微信小程序教师端、Vue 管理后台和 Go API 服务。系统聚焦教师个人信息管理、教师树洞诉求办理、思政培训活动报名与台账管理，覆盖教师端、二级党委管理员端和校级管理员端的核心业务闭环。

建议 GitHub 仓库名：`xidian-teacher-service-platform`

建议仓库简介：`西电教师综合服务平台，提供教师信息管理、树洞诉求办理和思政培训报名台账的一体化小程序与后台服务。`

## 功能范围

| 使用端 | 当前功能 |
| --- | --- |
| 教师端 | 个人信息管理、头像上传、教师树洞提交与进度查看、思政培训报名与取消报名、培训台账查看 |
| 二级党委管理员端 | 教师信息管理、树洞诉求受理反馈、思政培训发布编辑删除、报名名单查看 |
| 校级管理员端 | 教师信息管理、树洞诉求管理、思政培训管理、系统工作台与容量统计 |

## 核心能力

- 教师信息：支持教师资料维护，后台可新增、编辑、删除教师账号。
- 教师树洞：支持标题、内容、匿名方式、类目、紧急程度和图片附件；后台可查看详情、反馈、标记处理和删除。
- 思政培训：支持后台发布、编辑、删除培训；教师端按工号报名、取消报名并避免重复报名；后台可查看报名名单。
- 工作台：展示树洞待处理、培训状态、教师数量、报名记录、附件占用和数据库占用。

## 技术栈

- 小程序：微信小程序原生页面
- 管理后台：Vue 3、Vite、Element Plus
- 后端服务：Go、Gin、MySQL
- 部署参考：Nginx、systemd、logrotate

## 目录结构

```text
.
├── miniprogram/          # 微信小程序教师端
├── admin/                # Vue 3 + Element Plus 管理后台
├── server/               # Gin API 服务与数据库脚本
├── deploy/               # 部署模板
└── docs/                 # 架构、功能与需求说明
```

## 本地启动

### API 服务

```bash
cd server
go mod tidy
go run ./cmd/api
```

默认监听 `http://127.0.0.1:8090`。环境变量示例见 `server/.env.example`。

### 管理后台

```bash
cd admin
npm install
npm run dev
```

默认访问 `http://127.0.0.1:5173`，并通过 Vite 代理访问 `/api/v1`。

### 微信小程序

用微信开发者工具打开 `miniprogram` 目录。公开仓库中的 `appid` 使用 `touristappid` 占位，真实小程序 AppID 请放在本地私有配置中。

## 文档

- [系统架构与功能结构](docs/architecture.md)
- [技术架构与部署设计](docs/technical-architecture.md)
- [业务流程设计](docs/business-flow.md)
- [功能模块落地清单](docs/function-modules.md)
- [系统建模用例说明](docs/use-cases.md)
- [系统需求规格说明](docs/requirements.md)
- [上线步骤清单](docs/deployment.md)
