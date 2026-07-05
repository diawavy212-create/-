import { createRouter, createWebHistory } from "vue-router"

const routes = [
  { path: "/", component: () => import("../views/dashboard/Dashboard.vue"), meta: { title: "工作台" } },
  { path: "/profile", component: () => import("../views/profile/ProfileView.vue"), meta: { title: "个人信息管理" } },
  { path: "/treeholes", component: () => import("../views/treehole/TreeholeList.vue"), meta: { title: "教师树洞" } },
  { path: "/trainings", component: () => import("../views/training/TrainingList.vue"), meta: { title: "思政培训" } }
]

export default createRouter({
  history: createWebHistory(),
  routes
})
