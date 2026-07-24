const { request } = require("../../utils/request")

function timeRange(startTime, endTime) {
  const start = startTime ? String(startTime).replace(/:\d{2}$/, "") : ""
  const end = endTime ? String(endTime).replace(/:\d{2}$/, "") : ""
  if (start && end) return `${start} 至 ${end}`
  return start || end || "未设置"
}

function buildAnswerMap(questions) {
  const answerMap = {}
  ;(questions || []).forEach(question => {
    answerMap[question.id] = question.type === "text" ? "" : 0
  })
  return answerMap
}

Page({
  data: {
    items: [],
    current: null,
    answerMap: {},
    startedAt: 0,
    submitting: false
  },

  onLoad() {
    this.loadSurveys()
  },

  onShow() {
    this.loadSurveys()
  },

  loadSurveys() {
    request({ url: "/survey/list", data: { page: 1, size: 20, status: 1 } })
      .then(data => {
        const items = (data.list || []).map(item => ({
          ...item,
          timeText: timeRange(item.startTime, item.endTime),
          submittedText: item.submitted ? "已提交" : "待填写"
        }))
        this.setData({ items })
      })
      .catch(() => {
        this.setData({ items: [] })
        wx.showToast({ title: "调研列表加载失败", icon: "none" })
      })
  },

  openSurvey(event) {
    const surveyId = Number(event.currentTarget.dataset.id)
    const current = this.data.items.find(item => Number(item.id) === surveyId)
    if (!current) return
    if (current.submitted) {
      wx.showToast({ title: "该问卷已提交", icon: "none" })
      return
    }
    this.setData({
      current,
      answerMap: buildAnswerMap(current.questions),
      startedAt: Date.now()
    })
  },

  closeSurvey() {
    this.setData({ current: null, answerMap: {}, startedAt: 0 })
  },

  chooseOption(event) {
    const questionId = event.currentTarget.dataset.questionId
    const optionId = Number(event.detail.value)
    this.setData({ [`answerMap.${questionId}`]: optionId })
  },

  inputText(event) {
    const questionId = event.currentTarget.dataset.questionId
    this.setData({ [`answerMap.${questionId}`]: event.detail.value })
  },

  submitSurvey() {
    const survey = this.data.current
    if (!survey || this.data.submitting) return
    const questions = survey.questions || []
    const answers = []
    for (const question of questions) {
      const value = this.data.answerMap[question.id]
      if (question.required && (value === 0 || value === "" || value === undefined)) {
        wx.showToast({ title: "请完成必答题", icon: "none" })
        return
      }
      answers.push({
        questionId: question.id,
        optionId: question.type === "text" ? 0 : Number(value || 0),
        content: question.type === "text" ? String(value || "") : ""
      })
    }

    this.setData({ submitting: true })
    request({
      url: "/survey/answer/submit",
      method: "POST",
      data: {
        surveyId: survey.id,
        durationSeconds: Math.max(Math.floor((Date.now() - this.data.startedAt) / 1000), 0),
        answers
      }
    }).then(data => {
      wx.showToast({ title: data.valid ? "提交成功" : "已提交，质量待复核", icon: "none" })
      this.closeSurvey()
      return this.loadSurveys()
    }).catch(err => {
      wx.showToast({ title: err.message || "提交失败", icon: "none" })
    }).finally(() => {
      this.setData({ submitting: false })
    })
  },

  viewRecords() {
    request({ url: "/survey/records", data: { page: 1, size: 20 } })
      .then(data => {
        const records = data.list || []
        if (!records.length) {
          wx.showModal({ title: "我的调研记录", content: "暂无提交记录", showCancel: false })
          return
        }
        wx.showModal({
          title: "我的调研记录",
          content: records.map(item => `${item.title}\n${item.submitTime} ${item.valid ? "有效" : "已过滤"}`).join("\n\n"),
          showCancel: false
        })
      })
      .catch(() => {
        wx.showToast({ title: "记录加载失败", icon: "none" })
      })
  }
})
