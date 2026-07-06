const { request } = require("../../utils/request")
const app = getApp()

function buildForm(teacher) {
  return {
    name: teacher.name || "",
    college: teacher.college || "",
    department: teacher.department || "",
    phone: teacher.phone || "",
    email: teacher.email || ""
  }
}

Page({
  data: {
    teacher: {},
    form: buildForm({}),
    editing: false,
    saving: false
  },

  onLoad() {
    this.loadProfile()
  },

  loadProfile() {
    request({ url: "/profile/me" }).then(data => {
      this.setData({
        teacher: data,
        form: buildForm(data)
      })
    }).catch(() => {
      wx.showToast({ title: "信息加载失败", icon: "none" })
    })
  },

  startEdit() {
    this.setData({
      editing: true,
      form: buildForm(this.data.teacher)
    })
  },

  cancelEdit() {
    this.setData({
      editing: false,
      form: buildForm(this.data.teacher)
    })
  },

  onInput(event) {
    const field = event.currentTarget.dataset.field
    if (!field) return
    this.setData({
      [`form.${field}`]: event.detail.value
    })
  },

  saveProfile() {
    if (this.data.saving) return
    const form = this.data.form
    if (!form.name.trim()) {
      wx.showToast({ title: "请输入姓名", icon: "none" })
      return
    }
    this.setData({ saving: true })

    request({
      url: "/profile/me",
      method: "PUT",
      data: form
    }).then(data => {
      const cachedUser = app.globalData.user || wx.getStorageSync("user") || {}
      const nextUser = {
        ...cachedUser,
        name: data.name,
        realName: data.name,
        employeeNo: data.employeeNo,
        college: data.college,
        department: data.department
      }
      app.globalData.user = nextUser
      wx.setStorageSync("user", nextUser)
      this.setData({
        teacher: data,
        form: buildForm(data),
        editing: false
      })
      wx.showToast({ title: "保存成功", icon: "success" })
    }).catch(() => {
      wx.showToast({ title: "保存失败", icon: "none" })
    }).finally(() => {
      this.setData({ saving: false })
    })
  }
})
