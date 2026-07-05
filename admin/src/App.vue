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
      <p>生产环境通过学校 CAS 回跳并携带 ticket 后自动登录。</p>
      <el-input v-model="ticket" placeholder="本地联调用 CAS ticket 或 college-admin" />
      <el-select v-model="role" class="login-role">
        <el-option label="二级党委管理员" value="party_admin" />
        <el-option label="校级管理员" value="school_admin" />
      </el-select>
      <el-button type="primary" :loading="loading" @click="login">登录</el-button>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue"
import { useSessionStore } from "./stores/session"
import { http } from "./api/http"

const session = useSessionStore()
const ticket = ref("")
const role = ref("party_admin")
const loading = ref(false)

onMounted(() => {
  const params = new URLSearchParams(window.location.search)
  const casTicket = params.get("ticket")
  if (casTicket && !session.token) {
    ticket.value = casTicket
    login()
  }
})

function login() {
  if (!ticket.value) return
  loading.value = true
  http.post("/auth/cas-login", {
    ticket: ticket.value,
    role: role.value,
    service: `${window.location.origin}${window.location.pathname}`
  }).then(data => {
    session.save(data)
    window.history.replaceState({}, "", window.location.pathname)
  }).finally(() => {
    loading.value = false
  })
}

function logout() {
  session.clear()
}
</script>
