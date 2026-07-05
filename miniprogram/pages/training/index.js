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
          statusText: statusText(item.statusText)
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
    const title = event.currentTarget.dataset.title
    request({ url: "/trainings/ledgers", data: { page: 1, size: 50 } })
      .then(data => {
        const record = (data.list || []).find(item => Number(item.trainingId) === trainingId)
        if (!record) {
          wx.showModal({
            title: "暂无台账",
            content: "当前培训还没有你的报名或学习记录，请先报名。",
            showCancel: false
          })
          return
        }
        wx.showModal({
          title: "培训台账",
          content: `${title}\n报名状态：${applyStatusText(record.applyStatus)}\n学时：${record.learningHour || 0}\n成果状态：${applyStatusText(record.achievementStatus)}`,
          showCancel: false
        })
      })
      .catch(() => {
        wx.showToast({ title: "台账加载失败", icon: "none" })
      })
  }
})
