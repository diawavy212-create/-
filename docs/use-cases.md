# 用例说明

## 当前开发范围

本项目当前只做三个端都能用的三个功能：个人信息管理、教师树洞、思政培训活动。三个端分别为教师端、二级党委管理员端、校级管理员端。其他功能暂不开发。

## 三端用例

```mermaid
flowchart TD
  Teacher["教师端"]
  PartyAdmin["二级党委管理员端"]
  SchoolAdmin["校级管理员端"]
  Profile["个人信息管理"]
  Treehole["教师树洞"]
  Training["思政培训活动"]

  Teacher --> Profile
  Teacher --> Treehole
  Teacher --> Training

  PartyAdmin --> Profile
  PartyAdmin --> Treehole
  PartyAdmin --> Training

  SchoolAdmin --> Profile
  SchoolAdmin --> Treehole
  SchoolAdmin --> Training

  Profile --> ViewInfo["查看基础信息"]
  Treehole --> Track["查看诉求与办理状态"]
  Treehole --> Submit["教师端提交诉求"]
  Treehole --> Satisfaction["教师端满意度评价"]
  Training --> List["查看培训列表"]
  Training --> Enroll["教师端报名参与"]
  Training --> Ledger["教师端查看个人学习台账"]
```

## 页面和接口映射

| 功能 | 教师端页面 | 管理后台页面 | 后端模块 |
| --- | --- | --- | --- |
| 个人信息管理 | `miniprogram/pages/profile` | `admin/src/views/profile` | `profile` |
| 教师树洞 | `miniprogram/pages/treehole` | `admin/src/views/treehole` | `treehole` |
| 思政培训活动 | `miniprogram/pages/training` | `admin/src/views/training` | `training` |

## 暂不包含

思想状况调研、师德建设季报、学子说、数据统计与导出、系统运维、角色权限配置、通用支撑模块、管理员审批、管理员分派办理和报表生成功能。
