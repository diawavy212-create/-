# 功能模块落地清单

## 当前开发范围

只做三个端都能用的三个功能：个人信息管理、教师树洞、思政培训活动。三个端分别为教师端、二级党委管理员端、校级管理员端。其他功能暂不开发。

## 三端功能

| 使用端 | 个人信息管理 | 教师树洞 | 思政培训活动 |
| --- | --- | --- | --- |
| 教师端 | 查看本人基础信息，后续维护非敏感联系方式 | 提交诉求，查看进度和结果，评价满意度 | 查看培训，提交报名申请，签到学习，上传学习成果，查看个人台账 |
| 二级党委管理员端 | 查看本端账号基础信息 | 查看本院相关诉求和办理状态 | 发起院级培训，审核报名与成果，查看院级培训统计 |
| 校级管理员端 | 查看本端账号基础信息 | 查看全校相关诉求和办理状态 | 发起校级培训，审核报名与成果，查看校级培训统计 |

## API 对照

| 功能 | 接口 |
| --- | --- |
| 微信登录占位 | `POST /api/v1/auth/wechat-login` |
| 后台登录占位 | `POST /api/v1/auth/cas-login` |
| 个人信息 | `GET /api/v1/profile/me` |
| 诉求列表 | `GET /api/v1/treeholes` |
| 诉求提交 | `POST /api/v1/treeholes` |
| 诉求受理 | `POST /api/v1/treeholes/:id/accept` |
| 诉求分派 | `POST /api/v1/treeholes/:id/assign` |
| 办理反馈 | `POST /api/v1/treeholes/:id/feedback` |
| 满意度评价 | `POST /api/v1/treeholes/:id/satisfaction` |
| 诉求统计 | `GET /api/v1/treeholes/statistics` |
| 培训列表 | `GET /api/v1/trainings` |
| 培训发布 | `POST /api/v1/trainings` |
| 培训报名 | `POST /api/v1/trainings/:id/enroll` |
| 报名审核 | `POST /api/v1/trainings/:id/audit` |
| 学习留痕 | `POST /api/v1/trainings/:id/learning-records` |
| 学习成果审核 | `POST /api/v1/trainings/:id/audit`（当前与报名审核共用占位接口，后续可拆分） |
| 个人学习台账 | `GET /api/v1/trainings/ledgers` |
| 培训统计 | `GET /api/v1/trainings/statistics` |

## 暂不落地

思想状况调研、师德建设季报、学子说、数据统计与导出、系统运维、角色权限配置、通用支撑模块、管理员审批、管理员分派办理和报表生成功能均不在当前版本范围内。
