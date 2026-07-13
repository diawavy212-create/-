<template>
  <el-container v-if="session.token" class="layout">
    <el-aside width="236px" class="sidebar">
      <div class="brand">西电教师综合服务系统</div>
      <el-menu router :default-active="$route.path" background-color="#111827" text-color="#d1d5db" active-text-color="#ffffff">
        <el-menu-item index="/">工作台</el-menu-item>
        <el-menu-item index="/profile">个人信息管理</el-menu-item>
        <el-menu-item index="/treeholes">教师树洞</el-menu-item>
        <el-menu-item index="/trainings">思政培训</el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="topbar">
        <span>{{ $route.meta.title }}</span>
        <div class="topbar-actions">
          <el-tag type="success">{{ session.user?.name || "管理员" }}</el-tag>
          <el-button text @click="logout">退出</el-button>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>

  <div v-else class="login-page">
    <div class="login-panel">
      <h1>管理后台登录</h1>
      <p>请输入管理员账号和密码登录后台。</p>
      <el-input v-model="username" placeholder="账号：school-admin 或 college-admin" />
      <el-input v-model="password" type="password" show-password placeholder="密码：admin123456" @keyup.enter="login" />
      <el-button type="primary" :loading="loading" @click="login">登录</el-button>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue"
import { ElMessage } from "element-plus"
import { useSessionStore } from "./stores/session"
import { http } from "./api/http"

const session = useSessionStore()
const ticket = ref("")
const username = ref("school-admin")
const password = ref("")
const loading = ref(false)

onMounted(() => {
  window.addEventListener("admin-session-expired", logout)
  const params = new URLSearchParams(window.location.search)
  const casTicket = params.get("ticket")
  if (casTicket && !session.token) {
    ticket.value = casTicket
    casLogin()
  }
})

function login() {
  if (!username.value || !password.value) {
    ElMessage.warning("请填写账号和密码")
    return
  }
  loading.value = true
  http.post("/auth/admin-login", {
    username: username.value,
    password: password.value
  }).then(data => {
    session.save(data)
    window.history.replaceState({}, "", window.location.pathname)
  }).catch(error => {
    ElMessage.error(error.message || "登录失败")
  }).finally(() => {
    loading.value = false
  })
}

function casLogin() {
  if (!ticket.value) return
  loading.value = true
  http.post("/auth/cas-login", {
    ticket: ticket.value,
    service: `${window.location.origin}${window.location.pathname}`
  }).then(data => {
    session.save(data)
    window.history.replaceState({}, "", window.location.pathname)
  }).catch(error => {
    ElMessage.error(error.message || "登录失败")
  }).finally(() => {
    loading.value = false
  })
}

function logout() {
  session.clear()
}
</script>
