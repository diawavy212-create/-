const { request } = require("../../utils/request")
const app = getApp()

function padNumber(value) {
  return String(value).padStart(2, "0")
}

function formatDate(date) {
  return `${padNumber(date.getMonth() + 1)}-${padNumber(date.getDate())}`
}

function getGreeting(date) {
  const hour = date.getHours()
  if (hour < 6) return "凌晨好"
  if (hour < 12) return "上午好"
  if (hour < 14) return "中午好"
  if (hour < 18) return "下午好"
  return "晚上好"
}

function appealStatusText(status) {
  const map = {
    0: "待受理",
    1: "处理中",
    2: "已反馈",
    3: "已评价",
    4: "已归档"
  }
  return map[status] || "处理中"
}

function applyStatusText(status) {
  const map = {
    0: "待审核",
    1: "已通过",
    2: "已驳回"
  }
  return map[status] || "待审核"
}

function buildFallbackTodos(now) {
  const month = now.getMonth() + 1
  return [
    { title: "树洞诉求办理提醒", desc: "你的树洞诉求正在办理中", time: "今天", level: "high" },
    { title: "思想状况调研填写提醒", desc: `${month}月调研问卷已开放填写`, time: formatDate(now), level: "normal" },
    { title: "培训学习台账更新", desc: "报名审核和学习记录可查看", time: formatDate(now), level: "done" }
  ]
}

function summary(text, fallback) {
  const value = (text || "").trim()
  if (!value) return fallback
  return value.length > 12 ? `${value.slice(0, 12)}...` : value
}

function appealTitle(item) {
  return summary(item.title || item.subCategory || item.content || item.description, "树洞诉求")
}

function buildTreeholeTodos(items) {
  return (items || []).slice(0, 1).map(item => ({
    title: "树洞诉求办理提醒",
    desc: `${appealTitle(item)}：${item.statusText || appealStatusText(item.status)}`,
    time: "今天",
    level: item.status === 0 || item.status === 1 ? "high" : "done"
  }))
}

function buildTrainingTodos(trainings, ledgers, now) {
  const todos = []
  const latestLedger = (ledgers || [])[0]
  if (latestLedger) {
    todos.push({
      title: "培训报名状态更新",
      desc: `${latestLedger.title}：${applyStatusText(latestLedger.applyStatus)}`,
      time: formatDate(now),
      level: latestLedger.applyStatus === 2 ? "high" : "normal"
    })
  }

  const openTraining = (trainings || [])[0]
  if (openTraining) {
    todos.push({
      title: "思政培训报名通知",
      desc: `${openTraining.title} 已开放报名`,
      time: formatDate(now),
      level: "normal"
    })
  }

  return todos
}

function buildSurveyTodos(surveys, now) {
  return (surveys || []).filter(item => !item.submitted).slice(0, 1).map(item => ({
    title: "思想状况调研填写提醒",
    desc: `${item.title} 待填写`,
    time: formatDate(now),
    level: "high"
  }))
}

function buildStats(treeholes, trainings, surveys) {
  const appeals = treeholes || []
  const openTrainings = trainings || []
  const openSurveys = surveys || []
  const feedbackCount = appeals.filter(item => item.status === 2 || item.handleContent).length
  const followCount = appeals.filter(item => item.status === 0 || item.status === 1).length
  return [
    { value: feedbackCount, label: "树洞反馈" },
    { value: followCount, label: "待跟进" },
    { value: openTrainings.length, label: "开放培训" },
    { value: openSurveys.filter(item => !item.submitted).length, label: "待填调研" }
  ]
}

Page({
  data: {
    greeting: "您好",
    teacherTitle: "老师",
    avatarUrl: "",
    todos: [],
    features: [
      { title: "个人信息管理", icon: "人", color: "gold", path: "/pages/profile/index", tab: true },
      { title: "教师树洞", icon: "诉", color: "red", path: "/pages/treehole/index", tab: true },
      { title: "思政培训活动", icon: "学", color: "dark-red", path: "/pages/training/index", tab: true },
      { title: "思想状况调研", icon: "调", color: "gold", path: "/pages/survey/index", tab: false }
    ],
    stats: [
      { value: 0, label: "树洞反馈" },
      { value: 0, label: "待跟进" },
      { value: 0, label: "开放培训" },
      { value: 0, label: "待填调研" }
    ]
  },

  onLoad() {
    this.refreshHome()
  },

  onShow() {
    this.refreshHome()
  },

  refreshHome() {
    const now = new Date()
    const storedAvatar = wx.getStorageSync("teacherAvatarUrl")
    const user = app.globalData.user || wx.getStorageSync("user") || {}
    const rawName = user.name || user.realName || user.nickname || ""
    const teacherTitle = rawName ? `${rawName.replace(/老师$/, "")}老师` : "老师"

    this.setData({
      greeting: getGreeting(now),
      teacherTitle,
      avatarUrl: storedAvatar || user.avatarUrl || user.avatar || "",
      todos: buildFallbackTodos(now)
    })
    this.loadTodos(now)
  },

  loadTodos(now) {
    Promise.all([
      request({ url: "/treeholes", data: { page: 1, size: 20 } }).catch(() => ({ list: [] })),
      request({ url: "/trainings", data: { page: 1, size: 20, status: 1 } }).catch(() => ({ list: [] })),
      request({ url: "/trainings/ledgers", data: { page: 1, size: 20 } }).catch(() => ({ list: [] })),
      request({ url: "/survey/list", data: { page: 1, size: 20, status: 1 } }).catch(() => ({ list: [] }))
    ]).then(([treeholes, trainings, ledgers, surveys]) => {
      const todos = [
        ...buildSurveyTodos(surveys.list, now),
        ...buildTreeholeTodos(treeholes.list),
        ...buildTrainingTodos(trainings.list, ledgers.list, now)
      ].slice(0, 3)
      const stats = buildStats(treeholes.list, trainings.list, surveys.list)

      if (todos.length > 0) {
        this.setData({ todos })
      }
      this.setData({ stats })
    })
  },

  onChooseAvatar(event) {
    const avatarUrl = event.detail && event.detail.avatarUrl
    if (!avatarUrl) return
    wx.setStorageSync("teacherAvatarUrl", avatarUrl)
    this.setData({ avatarUrl })
  },

  chooseAvatarFallback() {
    if (wx.canIUse && wx.canIUse("button.open-type.chooseAvatar")) return
    wx.chooseMedia({
      count: 1,
      mediaType: ["image"],
      sourceType: ["album", "camera"],
      success: res => {
        const file = res.tempFiles && res.tempFiles[0]
        if (!file || !file.tempFilePath) return
        wx.setStorageSync("teacherAvatarUrl", file.tempFilePath)
        this.setData({ avatarUrl: file.tempFilePath })
      }
    })
  },

  openFeature(event) {
    const path = event.currentTarget.dataset.path
    const feature = this.data.features.find(item => item.path === path)
    if (!path) return
    if (feature && feature.tab) {
      wx.switchTab({ url: path })
      return
    }
    wx.navigateTo({ url: path })
  }
})
