const { request } = require("../../utils/request")
const app = getApp()

function statusText(status) {
  const map = {
    open: "报名中",
    in_progress: "进行中",
    ended: "已结束",
    archived: "已归档"
  }
  return map[status] || status || "报名中"
}

function applyStatusText(status) {
  const map = {
    0: "待审核",
    1: "已通过",
    2: "已驳回"
  }
  return map[status] || "待审核"
}

function displayValue(value) {
  return value || "未填写"
}

function shortTime(value) {
  if (!value) return ""
  return String(value).replace(/:\d{2}$/, "")
}

function timeRange(startTime, endTime) {
  const start = shortTime(startTime)
  const end = shortTime(endTime)
  if (start && end) return `${start} 至 ${end}`
  return start || end || "未填写"
}

function quotaText(value) {
  const quota = Number(value || 0)
  return quota > 0 ? `${quota} 人` : "不限"
}

function detailLines(training, title) {
  return [
    title,
    `时间：${training.timeText || timeRange(training.startTime, training.endTime)}`,
    `地点：${training.locationText || displayValue(training.location)}`,
    `主办单位：${training.sponsorText || displayValue(training.sponsorUnit)}`,
    `承办学院：${training.organizerText || displayValue(training.organizerUnit)}`,
    `名额：${training.quotaText || quotaText(training.quota)}`
  ]
}

function fullURL(path) {
  if (!path) return ""
  if (/^https?:\/\//.test(path)) return path
  return app.globalData.apiBaseURL.replace(/\/api\/v1$/, "") + path
}

Page({
  data: {
    items: [],
    enrollingId: 0,
    traceVisible: false,
    traceItem: {},
    traceRecord: {},
    achievementUrl: "",
    achievementFileName: "",
    studyHours: "",
    tracing: false,
    uploading: false
  },

  onLoad() {
    this.loadTrainings()
  },

  onShow() {
    this.loadTrainings()
  },

  loadTrainings() {
    return request({ url: "/trainings", data: { page: 1, size: 10, status: 1 } })
      .then(data => {
        const items = (data.list || []).map(item => ({
          ...item,
          enrolled: Boolean(item.enrolled),
          enrolledCount: Number(item.enrolledCount || 0),
          statusText: statusText(item.statusText),
          timeText: timeRange(item.startTime, item.endTime),
          locationText: displayValue(item.location),
          sponsorText: displayValue(item.sponsorUnit),
          organizerText: displayValue(item.organizerUnit),
          quotaText: quotaText(item.quota)
        }))
        this.setData({ items })
      })
      .catch(() => {
        this.setData({ items: [] })
        wx.showToast({ title: "培训列表加载失败", icon: "none" })
      })
  },

  toggleEnroll(event) {
    const trainingId = Number(event.currentTarget.dataset.id)
    const index = this.data.items.findIndex(item => Number(item.id) === trainingId)
    if (this.data.enrollingId || index < 0) return

    if (this.data.items[index].enrolled) {
      this.confirmCancel(trainingId, index)
      return
    }
    this.enroll(trainingId, index)
  },

  enroll(trainingId, index) {
    this.setData({ enrollingId: trainingId })
    request({
      url: `/trainings/${trainingId}/enroll`,
      method: "POST",
      data: {}
    })
      .then(data => {
        wx.showToast({ title: "报名成功" })
        const oldCount = Number(this.data.items[index].enrolledCount || 0)
        this.setData({
          [`items[${index}].enrolled`]: true,
          [`items[${index}].enrolledCount`]: Number(data.enrolledCount || 0) || oldCount + 1
        })
        return this.loadTrainings()
      })
      .catch(err => {
        wx.showToast({ title: err.message || "报名失败", icon: "none" })
        return this.loadTrainings()
      })
      .finally(() => {
        this.setData({ enrollingId: 0 })
      })
  },

  confirmCancel(trainingId, index) {
    wx.showModal({
      title: "取消报名",
      content: "确定取消该培训报名吗？",
      confirmText: "取消报名",
      confirmColor: "#c51624",
      success: res => {
        if (res.confirm) {
          this.cancelEnroll(trainingId, index)
        }
      }
    })
  },

  cancelEnroll(trainingId, index) {
    this.setData({ enrollingId: trainingId })
    request({
      url: `/trainings/${trainingId}/enroll`,
      method: "DELETE",
      data: {}
    })
      .then(data => {
        wx.showToast({ title: "已取消" })
        const oldCount = Number(this.data.items[index].enrolledCount || 0)
        this.setData({
          [`items[${index}].enrolled`]: false,
          [`items[${index}].enrolledCount`]: Number(data.enrolledCount || 0)
            || Math.max(oldCount - 1, 0)
        })
        return this.loadTrainings()
      })
      .catch(err => {
        wx.showToast({ title: err.message || "取消失败", icon: "none" })
        return this.loadTrainings()
      })
      .finally(() => {
        this.setData({ enrollingId: 0 })
      })
  },

  viewLedger(event) {
    const trainingId = Number(event.currentTarget.dataset.id)
    const training = this.data.items.find(item => Number(item.id) === trainingId) || {}
    const title = training.title || event.currentTarget.dataset.title || "培训"
    request({ url: "/trainings/ledgers", data: { page: 1, size: 50 } })
      .then(data => {
        const record = (data.list || []).find(item => Number(item.trainingId) === trainingId)
        if (!record) {
          wx.showModal({
            title: "培训详情",
            content: `${detailLines(training, title).join("\n")}\n报名状态：未报名`,
            showCancel: false
          })
          return
        }
        const detail = {
          ...record,
          ...training,
          timeText: training.timeText || timeRange(record.startTime, record.endTime),
          locationText: training.locationText || displayValue(record.location),
          sponsorText: training.sponsorText || displayValue(record.sponsorUnit),
          organizerText: training.organizerText || displayValue(record.organizerUnit),
          quotaText: training.quotaText || quotaText(record.quota)
        }
        wx.showModal({
          title: "培训台账",
          content: `${detailLines(detail, title).join("\n")}\n报名状态：${applyStatusText(record.applyStatus)}\n成果状态：${applyStatusText(record.achievementStatus)}`,
          showCancel: false
        })
      })
      .catch(() => {
        wx.showToast({ title: "台账加载失败", icon: "none" })
      })
  },

  openTrace(event) {
    const trainingId = Number(event.currentTarget.dataset.id)
    const training = this.data.items.find(item => Number(item.id) === trainingId)
    if (!training) return
    if (!training.enrolled) {
      wx.showToast({ title: "请先报名培训", icon: "none" })
      return
    }
    this.setData({
      traceVisible: true,
      traceItem: training,
      traceRecord: {},
      achievementUrl: "",
      achievementFileName: "",
      studyHours: "",
      tracing: true
    })
    this.loadTraceRecord(trainingId)
  },

  loadTraceRecord(trainingId) {
    request({ url: "/trainings/ledgers", data: { page: 1, size: 50 } })
      .then(data => {
        const record = (data.list || []).find(item => Number(item.trainingId) === trainingId) || {}
        this.setData({
          traceRecord: {
            ...record,
            achievementPreviewUrl: fullURL(record.achievementUrl)
          },
          achievementUrl: record.achievementUrl || "",
          achievementFileName: record.achievementUrl ? "已上传成果文件" : "",
          studyHours: record.learningHour ? String(record.learningHour) : ""
        })
      })
      .catch(() => {
        wx.showToast({ title: "留痕记录加载失败", icon: "none" })
      })
      .finally(() => {
        this.setData({ tracing: false })
      })
  },

  closeTrace() {
    if (this.data.tracing || this.data.uploading) return
    this.setData({
      traceVisible: false,
      traceItem: {},
      traceRecord: {},
      achievementUrl: "",
      achievementFileName: "",
      studyHours: ""
    })
  },

  onStudyHoursInput(event) {
    this.setData({ studyHours: event.detail.value })
  },

  onAchievementUrlInput(event) {
    this.setData({
      achievementUrl: event.detail.value,
      achievementFileName: event.detail.value ? "外部成果链接" : ""
    })
  },

  chooseAchievement() {
    if (this.data.uploading) return
    wx.chooseMessageFile({
      count: 1,
      type: "file",
      success: res => {
        const file = res.tempFiles && res.tempFiles[0]
        if (!file || !file.path) return
        this.uploadAchievement(file.path, file.name || "成果文件")
      }
    })
  },

  uploadAchievement(filePath, fileName) {
    this.setData({ uploading: true })
    app.ensureLogin().then(token => {
      wx.uploadFile({
        url: `${app.globalData.apiBaseURL}/trainings/learning-records/uploads`,
        filePath,
        name: "file",
        header: { Authorization: `Bearer ${token}` },
        success: res => {
          let body = {}
          try {
            body = JSON.parse(res.data || "{}")
          } catch (error) {
            wx.showToast({ title: "成果上传失败", icon: "none" })
            return
          }
          if (res.statusCode >= 200 && res.statusCode < 300 && (body.code === 200 || body.code === 0)) {
            this.setData({
              achievementUrl: body.data.url,
              achievementFileName: body.data.name || fileName
            })
            wx.showToast({ title: "成果已上传" })
            return
          }
          wx.showToast({ title: body.msg || body.message || "成果上传失败", icon: "none" })
        },
        fail: () => wx.showToast({ title: "成果上传失败", icon: "none" }),
        complete: () => this.setData({ uploading: false })
      })
    }).catch(() => {
      this.setData({ uploading: false })
    })
  },

  submitTrace() {
    const trainingId = this.data.traceItem && this.data.traceItem.id
    if (!trainingId || this.data.tracing) return
    const studyHours = Number(this.data.studyHours || 0)
    if (this.data.studyHours && (Number.isNaN(studyHours) || studyHours < 0)) {
      wx.showToast({ title: "请输入正确学时", icon: "none" })
      return
    }

    this.setData({ tracing: true })
    request({
      url: `/trainings/${trainingId}/learning-records`,
      method: "POST",
      data: {
        signIn: true,
        studyHours,
        achievementUrl: this.data.achievementUrl
      }
    }).then(() => {
      wx.showToast({ title: "留痕已提交" })
      this.setData({
        traceVisible: false,
        traceItem: {},
        traceRecord: {},
        achievementUrl: "",
        achievementFileName: "",
        studyHours: ""
      })
      return this.loadTrainings()
    }).catch(err => {
      wx.showToast({ title: err.message || "留痕提交失败", icon: "none" })
    }).finally(() => {
      this.setData({ tracing: false })
    })
  },

  noop() {
  }
})
