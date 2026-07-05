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
  return {
    ...item,
    content,
    displayTitle,
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
    anonymousType: 2,
    categoryIndex: 0,
    emergencyIndex: 0
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
    const detail = [
      `内容：${item.content || "-"}`,
      `类目：${item.category || "-"}`,
      `紧急程度：${item.emergencyText || "-"}`,
      `实名/匿名：${item.anonymousText || "-"}`,
      `状态：${item.statusText || "-"}`,
      item.handleContent ? `反馈：${item.handleContent}` : "反馈：暂无"
    ].join("\n")

    wx.showModal({
      title: item.displayTitle,
      content: detail,
      confirmText: item.attachmentPreviewUrl ? "看附件" : "知道了",
      cancelText: "关闭",
      showCancel: Boolean(item.attachmentPreviewUrl),
      success: res => {
        if (res.confirm && item.attachmentPreviewUrl) {
          wx.previewImage({ urls: [item.attachmentPreviewUrl] })
        }
      }
    })
  }
})
