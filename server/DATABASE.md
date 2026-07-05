# 数据库开发说明

后端数据库使用 MySQL 8.x，建库脚本位于 `server/schema.sql`。当前版本只保留需求文档要求的个人信息、教师树洞、思政培训活动相关数据表。

## 初始化数据库

在本机已安装 MySQL Server 和 MySQL Client 的情况下，进入 `server` 目录后执行：

```powershell
.\scripts\init-db.ps1 -HostName 127.0.0.1 -Port 3306 -User root -Password password
```

也可以直接使用 MySQL 客户端导入：

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p --default-character-set=utf8mb4 < schema.sql
```

导入完成后，后端默认读取以下 DSN：

```text
root:password@tcp(127.0.0.1:3306)/teacher_platform?charset=utf8mb4&parseTime=True&loc=Local
```

如果数据库账号或密码不同，请在启动后端前设置环境变量 `MYSQL_DSN`。

## 当前数据表

- `teacher`：教师、二级党委管理员、校级管理员的基础账号信息，并保存微信 `openid`、CAS 账号绑定
- `appeal`：教师树洞诉求、办理进度与满意度评价
- `training`：思政培训活动
- `training_record`：培训报名、学习留痕和学习成果记录

## 主要约束

- `teacher.user_id` 唯一，避免同一工号或账号重复建档。
- `teacher.wechat_openid` 唯一，用于教师端微信登录绑定。
- `teacher.cas_account` 唯一，用于管理后台 CAS 登录绑定。
- `training_record.training_id + teacher_id` 唯一，避免同一教师重复报名同一培训。
- 树洞、培训、培训记录的核心关联字段设置外键，并为状态、时间、单位等常用查询字段建立索引。
