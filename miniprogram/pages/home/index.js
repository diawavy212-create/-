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
    { title: "校级思政培训报名", desc: `${month}月培训已开放报名`, time: formatDate(now), level: "normal" },
    { title: "培训学习台账更新", desc: "报名审核和学习记录可查看", time: formatDate(now), level: "done" }
  ]
}

function buildTreeholeTodos(items) {
  return (items || []).slice(0, 2).map(item => ({
    title: "树洞诉求办理提醒",
    desc: `诉求 ${item.id} ${item.statusText || appealStatusText(item.status)}`,
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

function buildStats(treeholes, trainings) {
  const appeals = treeholes || []
  const openTrainings = trainings || []
  const feedbackCount = appeals.filter(item => item.status === 2 || item.handleContent).length
  const followCount = appeals.filter(item => item.status === 0 || item.status === 1).length
  return [
    { value: feedbackCount, label: "树洞反馈" },
    { value: followCount, label: "待跟进" },
    { value: openTrainings.length, label: "开放培训" }
  ]
}

Page({
  data: {
    greeting: "您好",
    teacherTitle: "老师",
    avatarUrl: "",
    todos: [],
    features: [
      { title: "个人信息管理", icon: "人", color: "gold", path: "/pages/profile/index" },
      { title: "教师树洞", icon: "诉", color: "red", path: "/pages/treehole/index" },
      { title: "思政培训活动", icon: "学", color: "dark-red", path: "/pages/training/index" }
    ],
    stats: [
      { value: 0, label: "树洞反馈" },
      { value: 0, label: "待跟进" },
      { value: 0, label: "开放培训" }
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
      request({ url: "/trainings/ledgers", data: { page: 1, size: 20 } }).catch(() => ({ list: [] }))
    ]).then(([treeholes, trainings, ledgers]) => {
      const todos = [
        ...buildTreeholeTodos(treeholes.list),
        ...buildTrainingTodos(trainings.list, ledgers.list, now)
      ].slice(0, 3)
      const stats = buildStats(treeholes.list, trainings.list)

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
    if (path) {
      wx.switchTab({ url: path })
    }
  }
})
