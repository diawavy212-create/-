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

http.interceptors.response.use(response => {
  const body = response.data
  if (body.code !== 200 && body.code !== 0) {
    return Promise.reject(new Error(body.msg || body.message || "请求失败"))
  }
  return body.data
})
