const { request } = require("../../utils/request")

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

Page({
  data: {
    items: [],
    enrollingId: 0
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
  }
})
