import axios from "axios"

export const http = axios.create({
  baseURL: "/api/v1",
  timeout: 8000
})

http.interceptors.request.use(config => {
  const token = localStorage.getItem("admin_token")
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

function clearExpiredSession() {
  localStorage.removeItem("admin_token")
  localStorage.removeItem("admin_user")
  if (window.location.pathname !== "/") {
    window.history.replaceState({}, "", "/")
  }
  window.dispatchEvent(new Event("admin-session-expired"))
}

http.interceptors.response.use(response => {
  const body = response.data
  if (body.code !== 200 && body.code !== 0) {
    if (body.code === 401) {
      clearExpiredSession()
    }
    return Promise.reject(new Error(body.msg || body.message || "请求失败"))
  }
  return body.data
}, error => {
  if (error.response && error.response.status === 401) {
    clearExpiredSession()
  }
  return Promise.reject(error)
})
