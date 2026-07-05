const app = getApp()

function request(options) {
  return app.ensureLogin().then(token => {
    return new Promise((resolve, reject) => {
      wx.request({
        url: `${app.globalData.apiBaseURL}${options.url}`,
        method: options.method || "GET",
        data: options.data || {},
        header: {
          Authorization: `Bearer ${token}`,
          "content-type": "application/json"
        },
        success(res) {
          if (res.statusCode >= 200 && res.statusCode < 300 && (res.data.code === 200 || res.data.code === 0)) {
            resolve(res.data.data)
            return
          }
          if (res.statusCode === 401) {
            wx.removeStorageSync("token")
            app.globalData.token = ""
          }
          reject(new Error(res.data && (res.data.msg || res.data.message) ? (res.data.msg || res.data.message) : "请求失败"))
        },
        fail: reject
      })
    })
  })
}

module.exports = { request }
