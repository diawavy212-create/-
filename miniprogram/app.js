App({
  globalData: {
    apiBaseURL: "http://127.0.0.1:8090/api/v1",
    token: wx.getStorageSync("token") || "",
    user: wx.getStorageSync("user") || null,
    loginPromise: null
  },
  onLaunch() {
    this.ensureLogin()
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
            data: { code: code || "dev" },
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
              reject(new Error(res.data && (res.data.msg || res.data.message) ? (res.data.msg || res.data.message) : "登录失败"))
            },
            fail: reject,
            complete: () => {
              this.globalData.loginPromise = null
            }
          })
        },
        fail: reject
      })
    })
    return this.globalData.loginPromise
  }
})
