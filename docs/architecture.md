# 系统架构与功能结构

## 总体架构

```mermaid
flowchart TD
  Mini["微信小程序（教师端）"]
  Admin["管理后台（二级党委管理员端 / 校级管理员端）"]
  Gateway["API 网关（Gin）"]
  Services["业务逻辑层"]
  MySQL[("MySQL 数据库")]
  Auth["登录认证\n微信登录 / 后台账号密码 / CAS 预留"]

  Mini --> Gateway
  Admin --> Gateway
  Gateway --> Services
  Services --> MySQL
  Services --> Auth
```

## 功能结构

```mermaid
flowchart TD
  System["西电教师综合服务系统"]
  Teacher["教师端"]
  PartyAdmin["二级党委管理员端"]
  SchoolAdmin["校级管理员端"]
  Profile["个人信息管理"]
  Treehole["教师树洞"]
  Training["思政培训活动"]

  System --> Teacher
  System --> PartyAdmin
  System --> SchoolAdmin

  Teacher --> Profile
  Teacher --> Treehole
  Teacher --> Training

  PartyAdmin --> Profile
  PartyAdmin --> Treehole
  PartyAdmin --> Training

  SchoolAdmin --> Profile
  SchoolAdmin --> Treehole
  SchoolAdmin --> Training
```

## API 路径

| 模块 | 路径 |
| --- | --- |
| 登录认证 | `/api/v1/auth` |
| 个人信息管理 | `/api/v1/profile` |
| 教师树洞 | `/api/v1/treeholes` |
| 思政培训 | `/api/v1/trainings` |

## 接口约定

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

除登录接口外，请求需携带：

```text
Authorization: Bearer <token>
```

当前版本只保留三个端共用的个人信息管理、教师树洞、思政培训活动。其他业务模块暂不接入。
