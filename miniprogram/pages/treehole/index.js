const { request } = require("../../utils/request")
const app = getApp()

const anonymousOptions = [
  { label: "实名", value: 0 },
  { label: "匿名", value: 1 },
  { label: "匿名可回访", value: 2 }
]

const categoryOptions = ["工作支持", "教学支持", "心理支持", "后勤保障", "其他"]
const emergencyOptions = [
  { label: "普通", value: 0 },
  { label: "较急", value: 1 },
  { label: "紧急", value: 2 }
]
const satisfactionOptions = [
  { label: "非常满意", value: 3 },
  { label: "满意", value: 2 },
  { label: "基本满意", value: 1 },
  { label: "不满意", value: 0 }
]

function satisfactionLabel(value) {
  const option = satisfactionOptions.find(item => item.value === Number(value))
  return option ? option.label : ""
}

function fullURL(path) {
  if (!path) return ""
  if (/^https?:\/\//.test(path)) return path
  return app.globalData.apiBaseURL.replace(/\/api\/v1$/, "") + path
}

function summary(text, fallback) {
  const value = (text || "").trim()
  if (!value) return fallback
  return value.length > 18 ? `${value.slice(0, 18)}...` : value
}

function normalizeAppeal(item) {
  const content = item.content || item.description || ""
  const displayTitle = item.title || item.subCategory || summary(content, "未命名诉求")
  const rawSatisfaction = item.satisfaction
  const satisfaction = Number(rawSatisfaction)
  const hasSatisfaction = rawSatisfaction !== undefined
    && rawSatisfaction !== null
    && rawSatisfaction !== ""
    && satisfaction >= 0
    && satisfaction <= 3
  const status = Number(item.status)
  const statusText = item.statusText || ""
  const isPendingOrProcessing = status === 0
    || status === 1
    || statusText === "待受理"
    || statusText === "处理中"
  const hasHandledStatus = !Number.isNaN(status) || Boolean(statusText)
  const canEvaluate = !hasSatisfaction && hasHandledStatus && !isPendingOrProcessing
  return {
    ...item,
    content,
    displayTitle,
    satisfactionText: hasSatisfaction ? satisfactionLabel(satisfaction) : "未评价",
    hasSatisfaction,
    canEvaluate,
    attachmentPreviewUrl: fullURL(item.attachmentUrl)
  }
}

Page({
  data: {
    title: "",
    content: "",
    attachmentUrl: "",
    attachmentPreviewUrl: "",
    items: [],
    anonymousOptions,
    categoryOptions,
    emergencyOptions,
    satisfactionOptions,
    anonymousType: 2,
    categoryIndex: 0,
    emergencyIndex: 0,
    detailVisible: false,
    currentDetail: {},
    satisfactionVisible: false,
    satisfactionValue: 3,
    evaluatingItem: {},
    evaluating: false
  },

  onLoad() {
    this.loadItems()
  },

  onShow() {
    this.loadItems()
  },

  onTitleInput(event) {
    this.setData({ title: event.detail.value })
  },

  onInput(event) {
    this.setData({ content: event.detail.value })
  },

  selectAnonymous(event) {
    this.setData({ anonymousType: Number(event.currentTarget.dataset.value) })
  },

  onCategoryChange(event) {
    this.setData({ categoryIndex: Number(event.detail.value) })
  },

  onEmergencyChange(event) {
    this.setData({ emergencyIndex: Number(event.detail.value) })
  },

  loadItems() {
    request({ url: "/treeholes", data: { page: 1, size: 10 } })
      .then(data => {
        const items = (data.list || []).map(normalizeAppeal)
        this.setData({ items })
      })
      .catch(() => {
        wx.showToast({ title: "列表加载失败", icon: "none" })
      })
  },

  chooseAttachment() {
    wx.chooseMedia({
      count: 1,
      mediaType: ["image"],
      sourceType: ["album", "camera"],
      success: res => {
        const file = res.tempFiles && res.tempFiles[0]
        if (!file || !file.tempFilePath) return
        this.uploadAttachment(file.tempFilePath)
      }
    })
  },

  uploadAttachment(filePath) {
    app.ensureLogin().then(token => {
      wx.uploadFile({
        url: `${app.globalData.apiBaseURL}/treeholes/uploads`,
        filePath,
        name: "file",
        header: { Authorization: `Bearer ${token}` },
        success: res => {
          let body = {}
          try {
            body = JSON.parse(res.data || "{}")
          } catch (error) {
            wx.showToast({ title: "附件上传失败", icon: "none" })
            return
          }
          if (res.statusCode >= 200 && res.statusCode < 300 && (body.code === 200 || body.code === 0)) {
            const attachmentUrl = body.data.url
            this.setData({
              attachmentUrl,
              attachmentPreviewUrl: fullURL(attachmentUrl)
            })
            wx.showToast({ title: "附件已上传" })
            return
          }
          wx.showToast({ title: "附件上传失败", icon: "none" })
        },
        fail: () => wx.showToast({ title: "附件上传失败", icon: "none" })
      })
    })
  },

  previewAttachment() {
    if (!this.data.attachmentPreviewUrl) return
    wx.previewImage({ urls: [this.data.attachmentPreviewUrl] })
  },

  submit() {
    const title = this.data.title.trim()
    const content = this.data.content.trim()
    if (!title) {
      wx.showToast({ title: "请输入标题", icon: "none" })
      return
    }
    if (!content) {
      wx.showToast({ title: "请输入内容", icon: "none" })
      return
    }

    const category = categoryOptions[this.data.categoryIndex]
    const emergency = emergencyOptions[this.data.emergencyIndex]
    request({
      url: "/treeholes",
      method: "POST",
      data: {
        subCategory: title,
        description: content,
        attachmentUrl: this.data.attachmentUrl,
        anonymousType: this.data.anonymousType,
        category,
        emergencyLevel: emergency.value,
        influenceScope: 0,
        expectedMethod: 1
      }
    })
      .then(() => {
        wx.showToast({ title: "已提交" })
        this.setData({ title: "", content: "", attachmentUrl: "", attachmentPreviewUrl: "" })
        this.loadItems()
      })
      .catch(err => {
        wx.showToast({ title: err.message || "提交失败", icon: "none" })
      })
  },

  openDetail(event) {
    const item = this.data.items[Number(event.currentTarget.dataset.index)]
    if (!item) return
    this.setData({
      currentDetail: item,
      detailVisible: true
    })
  },

  closeDetail() {
    this.setData({
      detailVisible: false,
      currentDetail: {}
    })
  },

  openSatisfaction(event) {
    const item = this.data.items[Number(event.currentTarget.dataset.index)]
    if (!item) return
    this.setData({
      evaluatingItem: item,
      satisfactionValue: 3,
      satisfactionVisible: true
    })
  },

  openSatisfactionById(event) {
    const id = Number(event.currentTarget.dataset.id)
    const item = this.data.items.find(row => Number(row.id) === id)
    if (!item) return
    this.setData({
      evaluatingItem: item,
      satisfactionValue: 3,
      satisfactionVisible: true
    })
  },

  closeSatisfaction() {
    if (this.data.evaluating) return
    this.setData({
      satisfactionVisible: false,
      evaluatingItem: {}
    })
  },

  selectSatisfaction(event) {
    this.setData({ satisfactionValue: Number(event.detail.value) })
  },

  submitSatisfactionValue(event) {
    const id = Number(event.currentTarget.dataset.id)
    const value = Number(event.currentTarget.dataset.value)
    const item = this.data.items.find(row => Number(row.id) === id)
    if (!item || !item.id || this.data.evaluating) return
    this.saveSatisfaction(item, value)
  },

  submitSatisfaction() {
    const item = this.data.evaluatingItem
    if (!item || !item.id || this.data.evaluating) return
    const satisfaction = Number(this.data.satisfactionValue)
    this.saveSatisfaction(item, satisfaction)
  },

  saveSatisfaction(item, satisfaction) {
    if (!satisfactionLabel(satisfaction)) {
      wx.showToast({ title: "请选择满意度", icon: "none" })
      return
    }

    this.setData({ evaluating: true })
    request({
      url: `/treeholes/${item.id}/satisfaction`,
      method: "POST",
      data: {
        satisfaction,
        satisfactionScore: satisfaction,
        score: satisfaction,
        remark: ""
      }
    }).then(data => {
      const saved = Number(data && data.satisfaction)
      if (saved !== satisfaction) {
        throw new Error("评分未保存，请重启后端后重试")
      }
      const index = this.data.items.findIndex(row => Number(row.id) === Number(item.id))
      if (index >= 0) {
        this.setData({
          [`items[${index}].status`]: 3,
          [`items[${index}].statusText`]: "已评价",
          [`items[${index}].satisfaction`]: saved,
          [`items[${index}].satisfactionText`]: satisfactionLabel(saved),
          [`items[${index}].hasSatisfaction`]: true,
          [`items[${index}].canEvaluate`]: false,
          [`dismissedEvaluationIds.${item.id}`]: false
        })
      }
      wx.showToast({ title: satisfactionLabel(satisfaction) })
      this.setData({
        satisfactionVisible: false,
        evaluatingItem: {}
      })
    }).catch(err => {
      wx.showToast({ title: err.message || "评价失败", icon: "none" })
    }).finally(() => {
      this.setData({ evaluating: false })
    })
  },

  previewDetailAttachment() {
    const url = this.data.currentDetail && this.data.currentDetail.attachmentPreviewUrl
    if (!url) return
    wx.previewImage({ urls: [url] })
  },

  noop() {
  }
})
