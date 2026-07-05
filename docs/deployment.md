# 上线步骤清单

当前项目包含三个部分：`server` 后端 API、`admin` 管理后台、`miniprogram` 微信小程序。生产环境推荐部署在 Linux 服务器上：Nginx 负责 HTTPS、静态后台页面和 `/api/v1` 反向代理，systemd 守护 Go API 服务。

## 1. 服务器目录

推荐目录约定：

```text
/opt/teacher-platform/
  server/teacher-platform-api
  admin/dist/
/etc/teacher-platform/teacher-platform.env
/var/log/teacher-platform/
```

创建用户和目录：

```bash
sudo useradd --system --home /opt/teacher-platform --shell /usr/sbin/nologin teacher-platform
sudo mkdir -p /opt/teacher-platform/server /opt/teacher-platform/admin /etc/teacher-platform /var/log/teacher-platform
sudo chown -R teacher-platform:teacher-platform /opt/teacher-platform /var/log/teacher-platform
```

## 2. 准备数据库

安装 MySQL 8.x 后导入：

```bash
cd server
mysql -h 127.0.0.1 -P 3306 -u root -p --default-character-set=utf8mb4 < schema.sql
```

建议创建最小权限账号：

```sql
CREATE USER 'teacher_platform_user'@'%' IDENTIFIED BY 'REPLACE_PASSWORD';
GRANT SELECT, INSERT, UPDATE, DELETE ON teacher_platform.* TO 'teacher_platform_user'@'%';
FLUSH PRIVILEGES;
```

上线前需要为教师账号绑定 `teacher.wechat_openid`，为管理员账号绑定 `teacher.cas_account`。

## 3. 生产环境变量

复制模板：

```bash
sudo cp deploy/env/teacher-platform.env.example /etc/teacher-platform/teacher-platform.env
sudo chmod 600 /etc/teacher-platform/teacher-platform.env
```

必须修改：

```text
HTTP_ADDR=127.0.0.1:8090
MYSQL_DSN=teacher_platform_user:密码@tcp(数据库地址:3306)/teacher_platform?charset=utf8mb4&parseTime=True&loc=Local
CAS_ENDPOINT=学校 CAS 地址
CAS_SERVICE_URL=https://你的后台域名
WECHAT_APP_ID=微信小程序 AppID
WECHAT_APP_SECRET=微信小程序 AppSecret
AUTH_TOKEN_SECRET=至少 32 位随机字符串
DEV_AUTH_ENABLED=false
```

`DEV_AUTH_ENABLED=false` 很重要，它会关闭本地开发 token 兼容分支。

## 4. 构建后端

在构建机或服务器上执行：

```bash
cd server
go test ./...
go build -o teacher-platform-api ./cmd/api
```

上传到：

```bash
sudo install -o teacher-platform -g teacher-platform -m 0755 teacher-platform-api /opt/teacher-platform/server/teacher-platform-api
```

## 5. systemd 服务守护

复制服务文件：

```bash
sudo cp deploy/systemd/teacher-platform-api.service /etc/systemd/system/teacher-platform-api.service
sudo systemctl daemon-reload
sudo systemctl enable --now teacher-platform-api
```

查看状态和日志：

```bash
sudo systemctl status teacher-platform-api
sudo tail -f /var/log/teacher-platform/api.log
sudo tail -f /var/log/teacher-platform/api.err.log
```

健康检查：

```bash
curl http://127.0.0.1:8090/healthz
```

## 6. 构建管理后台

```bash
cd admin
pnpm install --frozen-lockfile
pnpm run build
```

部署构建产物：

```bash
sudo rsync -a --delete dist/ /opt/teacher-platform/admin/dist/
sudo chown -R teacher-platform:teacher-platform /opt/teacher-platform/admin/dist
```

后台使用相对路径 `/api/v1` 访问 API，因此生产环境由 Nginx 同域反向代理即可。

## 7. Nginx 与 HTTPS

准备证书后，复制模板：

```bash
sudo cp deploy/nginx/teacher-platform.conf /etc/nginx/sites-available/teacher-platform.conf
sudo ln -s /etc/nginx/sites-available/teacher-platform.conf /etc/nginx/sites-enabled/teacher-platform.conf
```

修改模板中的：

- `teacher.example.edu`
- `ssl_certificate`
- `ssl_certificate_key`
- `root /opt/teacher-platform/admin/dist`

检查并重载：

```bash
sudo nginx -t
sudo systemctl reload nginx
```

验证：

```bash
curl https://你的域名/healthz
curl https://你的域名/api/v1/profile/me
```

第二个请求需要携带登录 token；浏览器访问后台时 CAS 登录会自动获取。

## 8. 日志轮转

复制 logrotate 模板：

```bash
sudo cp deploy/logrotate/teacher-platform /etc/logrotate.d/teacher-platform
sudo logrotate -d /etc/logrotate.d/teacher-platform
```

API 标准输出写入 `/var/log/teacher-platform/api.log`，错误输出写入 `/var/log/teacher-platform/api.err.log`。

## 9. 小程序合法域名

微信公众平台需要配置：

```text
request 合法域名：https://你的域名
```

小程序端修改 `miniprogram/app.js`：

```js
apiBaseURL: "https://你的域名/api/v1"
```

然后使用微信开发者工具：

```text
真机预览 -> 上传代码 -> 提交审核 -> 发布
```

## 10. 上线前检查

- 后端：`go test ./...`
- 后台：`pnpm run build`
- Nginx：`nginx -t`
- systemd：`systemctl status teacher-platform-api`
- HTTPS：证书有效，`https://你的域名/healthz` 返回 `status: up`
- 数据库：只开放最小权限账号
- 安全：`AUTH_TOKEN_SECRET` 已替换，`DEV_AUTH_ENABLED=false`
- 登录：教师账号已绑定 `wechat_openid`，管理员账号已绑定 `cas_account`
- 小程序：合法域名已配置，真机预览通过

## 11. 回滚

保留上一版二进制和后台 `dist`：

```bash
sudo cp /opt/teacher-platform/server/teacher-platform-api /opt/teacher-platform/server/teacher-platform-api.bak
sudo cp -a /opt/teacher-platform/admin/dist /opt/teacher-platform/admin/dist.bak
```

如果新版本异常：

```bash
sudo mv /opt/teacher-platform/server/teacher-platform-api.bak /opt/teacher-platform/server/teacher-platform-api
sudo rm -rf /opt/teacher-platform/admin/dist
sudo mv /opt/teacher-platform/admin/dist.bak /opt/teacher-platform/admin/dist
sudo systemctl restart teacher-platform-api
sudo systemctl reload nginx
```
