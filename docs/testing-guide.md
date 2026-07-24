# 小程序测试指南

本项目由三部分组成：

- `miniprogram/`：微信原生小程序，教师端。
- `server/`：Go + Gin API 服务，默认端口 `8090`。
- `admin/`：Vue 管理后台，默认端口 `5173`。

真实使用时，小程序不是单独运行的。用户在微信里打开小程序，小程序通过 HTTPS 请求后端 API；管理人员在浏览器里打开后台，后台也请求同一个后端 API；后端连接 MySQL 保存教师、树洞、培训、问卷等数据。

## 一、本地开发测试

适合开发阶段快速验证页面和接口。

### 1. 准备数据库

本项目使用 MySQL 8.x。进入 `server` 目录后执行：

```powershell
.\scripts\init-db.ps1 -HostName 127.0.0.1 -Port 3306 -User root -Password password
```

也可以手动导入：

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p --default-character-set=utf8mb4 < schema.sql
```

默认后端连接：

```text
root:password@tcp(127.0.0.1:3306)/teacher_platform?charset=utf8mb4&parseTime=True&loc=Local
```

如果你的 MySQL 密码不是 `password`，启动后端前设置 `MYSQL_DSN`。

### 2. 启动后端

```powershell
cd server
go run ./cmd/api
```

启动成功后访问：

```text
http://127.0.0.1:8090/healthz
```

返回 `status: up` 说明后端和数据库都正常。

### 3. 启动管理后台

```powershell
cd admin
pnpm install
pnpm run dev
```

浏览器打开：

```text
http://127.0.0.1:5173
```

本地默认账号：

```text
school-admin / admin123456
college-admin / admin123456
```

### 4. 打开小程序

用微信开发者工具打开 `miniprogram` 目录。

当前 `miniprogram/config.js` 中的接口地址是：

```js
http://192.168.110.189:8090/api/v1
```

如果你的电脑 IP 不是 `192.168.110.189`，需要改成当前电脑的局域网 IP，例如：

```js
http://192.168.1.23:8090/api/v1
```

开发者工具中可以使用 `touristappid` 或真实 AppID。项目配置已经设置 `urlCheck: false`，开发阶段可以请求局域网 HTTP 地址。

## 二、手机真机预览测试

适合验证真实微信环境下的交互、登录、上传、页面兼容性。

手机真机预览时，手机必须能访问电脑上的后端服务。

检查项：

- 手机和电脑连接同一个 Wi-Fi。
- `miniprogram/config.js` 使用电脑局域网 IP，不能写 `127.0.0.1` 或 `localhost`。
- Windows 防火墙允许 `8090` 端口被局域网访问。
- 后端启动时监听 `:8090`，不是只监听 `127.0.0.1:8090`。
- 微信开发者工具点击“预览”，用手机微信扫码打开。

如果手机提示网络请求失败，优先在手机浏览器访问：

```text
http://电脑局域网IP:8090/healthz
```

如果手机浏览器也打不开，问题在网络、防火墙或后端监听地址，不在小程序页面。

## 三、体验版和线上真实环境

真实小程序上线后，运行方式和本地开发不同。

线上要求：

- 必须使用真实小程序 AppID，不能使用 `touristappid`。
- 后端 API 必须有公网 HTTPS 域名。
- 微信公众平台后台必须配置 request 合法域名。
- 小程序代码里的 API 地址必须改成线上 HTTPS 地址，例如：

```js
https://api.example.edu.cn/api/v1
```

- 后端应配置正式环境变量：

```text
DEV_AUTH_ENABLED=false
WECHAT_APP_ID=真实小程序AppID
WECHAT_APP_SECRET=真实小程序AppSecret
AUTH_TOKEN_SECRET=足够长的随机密钥
ADMIN_LOGIN_PASSWORD=正式后台密码
MYSQL_DSN=正式数据库连接
```

本地开发可以用 HTTP 和开发登录；体验版、审核版、正式版都应该按线上要求配置 HTTPS 域名和真实微信登录。

## 四、推荐测试顺序

### 1. 后端基础测试

- 打开 `/healthz`，确认数据库可用。
- 后端启动时没有数据库连接错误。
- 上传目录 `server/uploads` 存在并可写。

### 2. 管理后台测试

- 使用 `school-admin / admin123456` 登录。
- 查看教师列表。
- 新增、编辑、删除教师。
- 发布培训。
- 查看培训报名名单。
- 查看树洞诉求并处理反馈。
- 创建问卷，配置题目和选项。

### 3. 小程序教师端测试

- 首次打开小程序，确认能自动登录。
- 查看和编辑个人信息。
- 提交树洞诉求。
- 上传图片附件。
- 查看诉求进度。
- 查看培训列表。
- 报名培训。
- 取消报名。
- 填写问卷并提交。

### 4. 联动测试

- 小程序提交树洞后，后台能看到该诉求。
- 后台处理树洞后，小程序能看到状态变化。
- 后台发布培训后，小程序能看到培训。
- 小程序报名培训后，后台能看到报名记录。
- 后台发布问卷后，小程序能看到问卷。
- 小程序提交问卷后，后台能看到统计或答卷。

## 五、常见问题

### 小程序登录失败：wechat app is not configured

开发阶段确认后端环境变量 `DEV_AUTH_ENABLED=true`。当前代码默认就是 `true`，但如果你手动设置成了 `false`，就必须配置真实 `WECHAT_APP_ID` 和 `WECHAT_APP_SECRET`。

### 开发者工具能用，手机预览不能用

通常是手机访问不到电脑后端。确认手机和电脑在同一网络，并检查 Windows 防火墙和 `miniprogram/config.js` 中的 IP。

### 线上审核或体验版请求失败

确认 API 地址是 HTTPS，并且域名已经在微信公众平台配置为合法 request 域名。线上版本不能依赖开发者工具里的“不校验合法域名”。

### 后台能登录，小程序没有数据

小程序开发登录会按设备生成新的微信 openid，后端可能自动创建了一个新教师账号。可以在后台教师列表里查看该账号，或在数据库中检查 `teacher.wechat_openid`。

## 六、最小验收清单

一次完整验收至少覆盖：

- 后端 `/healthz` 正常。
- 后台管理员能登录。
- 小程序能自动登录。
- 小程序提交一条树洞，后台能处理，处理结果能回到小程序。
- 后台发布一个培训，小程序能报名，后台能看到报名。
- 后台发布一份问卷，小程序能提交，后台能看到结果。
- 手机真机预览至少跑通一次，不只依赖开发者工具模拟器。
