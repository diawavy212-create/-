const { DEFAULT_API_BASE_URL } = require("./config")
const AUTH_VERSION = "auto-login-v1"

function createClientId() {
  return `mp-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function buildError(res, fallback) {
  const error = new Error(res.data && (res.data.msg || res.data.message) ? (res.data.msg || res.data.message) : fallback)
  error.statusCode = res.statusCode
  return error
}

function buildNetworkError(err, fallback) {
  return new Error(err && err.errMsg ? err.errMsg : fallback)
}

App({
  globalData: {
    apiBaseURL: wx.getStorageSync("apiBaseURL") || DEFAULT_API_BASE_URL,
    token: wx.getStorageSync("token") || "",
    user: wx.getStorageSync("user") || null,
    clientId: wx.getStorageSync("clientId") || "",
    loginPromise: null
  },

  onLaunch() {
    if (wx.getStorageSync("authVersion") !== AUTH_VERSION) {
      this.globalData.token = ""
      this.globalData.user = null
      this.globalData.apiBaseURL = DEFAULT_API_BASE_URL
      wx.removeStorageSync("token")
      wx.removeStorageSync("user")
      wx.removeStorageSync("apiBaseURL")
      wx.setStorageSync("authVersion", AUTH_VERSION)
    }
    if (!this.globalData.clientId) {
      this.globalData.clientId = createClientId()
      wx.setStorageSync("clientId", this.globalData.clientId)
    }
    this.ensureLogin().catch(() => {})
  },

  setAPIBaseURL(url) {
    const value = (url || DEFAULT_API_BASE_URL).trim().replace(/\/$/, "")
    this.globalData.apiBaseURL = value
    wx.setStorageSync("apiBaseURL", value)
  },

  clearLogin() {
    this.globalData.token = ""
    this.globalData.user = null
    wx.removeStorageSync("token")
    wx.removeStorageSync("user")
  },

  ensureLogin() {
    if (this.globalData.token) {
      return Promise.resolve(this.globalData.token)
    }
    if (this.globalData.loginPromise) {
      return this.globalData.loginPromise
    }

    this.globalData.loginPromise = new Promise((resolve, reject) => {
      wx.login({
        success: ({ code }) => {
          wx.request({
            url: `${this.globalData.apiBaseURL}/auth/wechat-login`,
            method: "POST",
            data: { code: code || "dev", clientId: this.globalData.clientId },
            header: { "content-type": "application/json" },
            success: res => {
              if (res.statusCode >= 200 && res.statusCode < 300 && (res.data.code === 200 || res.data.code === 0)) {
                const data = res.data.data
                this.globalData.token = data.token
                this.globalData.user = data.user
                wx.setStorageSync("token", data.token)
                wx.setStorageSync("user", data.user)
                resolve(data.token)
                return
              }
              reject(buildError(res, "登录失败"))
            },
            fail: err => reject(buildNetworkError(err, "登录请求失败")),
            complete: () => {
              this.globalData.loginPromise = null
            }
          })
        },
        fail: err => reject(buildNetworkError(err, "微信登录失败"))
      })
    })
    return this.globalData.loginPromise
  }
})
