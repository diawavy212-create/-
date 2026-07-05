<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">个人信息管理</h1>
      <div>
        <el-button @click="load">刷新</el-button>
        <el-button type="primary" @click="openCreate">新增教师</el-button>
      </div>
    </div>

    <div class="panel">
      <el-alert
        v-if="errorText"
        class="page-alert"
        type="error"
        :title="errorText"
        show-icon
        :closable="false"
      />
      <el-table v-loading="loading" :data="items" stripe empty-text="暂无教师信息，请点击右上角新增教师">
        <el-table-column prop="name" label="姓名" width="130" />
        <el-table-column prop="employeeNo" label="工号" width="140" />
        <el-table-column prop="college" label="单位/学院" min-width="150" />
        <el-table-column prop="department" label="部门" min-width="150" />
        <el-table-column prop="phone" label="手机" width="150" />
        <el-table-column prop="email" label="邮箱" min-width="190" />
        <el-table-column label="角色" width="150">
          <template #default="{ row }">{{ roleText(row.role) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openEdit(row)">编辑</el-button>
            <el-button type="danger" link @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="visible" :title="form.id ? '编辑教师信息' : '新增教师信息'" width="560px">
      <el-form :model="form" label-width="92px">
        <el-form-item label="姓名" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="工号" required>
          <el-input v-model="form.employeeNo" />
        </el-form-item>
        <el-form-item label="单位/学院" required>
          <el-input v-model="form.college" />
        </el-form-item>
        <el-form-item label="部门">
          <el-input v-model="form.department" />
        </el-form-item>
        <el-form-item label="手机">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role" class="full-width">
            <el-option label="教师" value="teacher" />
            <el-option label="二级党委管理员" value="party_admin" />
            <el-option label="校级管理员" value="school_admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ElMessage, ElMessageBox } from "element-plus"
import { onMounted, ref } from "vue"
import { http } from "../../api/http"

const items = ref([])
const loading = ref(false)
const visible = ref(false)
const saving = ref(false)
const errorText = ref("")
const form = ref(defaultForm())

function defaultForm() {
  return {
    name: "",
    employeeNo: "",
    college: "",
    department: "",
    phone: "",
    email: "",
    role: "teacher"
  }
}

function load() {
  loading.value = true
  errorText.value = ""
  http.get("/profile/teachers", { params: { page: 1, size: 50 } })
    .then(data => {
      items.value = data.list || []
    })
    .catch(error => {
      errorText.value = error.message || "教师信息加载失败，请确认当前账号是管理员"
      items.value = []
    })
    .finally(() => {
      loading.value = false
    })
}

function openCreate() {
  form.value = defaultForm()
  visible.value = true
}

function openEdit(row) {
  form.value = { ...defaultForm(), ...row }
  visible.value = true
}

function save() {
  if (!form.value.name || !form.value.employeeNo || !form.value.college) {
    ElMessage.warning("请填写姓名、工号和单位/学院")
    return
  }
  saving.value = true
  const action = form.value.id
    ? http.put(`/profile/teachers/${form.value.id}`, form.value)
    : http.post("/profile/teachers", form.value)
  action.then(() => {
    ElMessage.success("保存成功")
    visible.value = false
    load()
  }).catch(error => {
    ElMessage.error(error.message || "保存失败")
  }).finally(() => {
    saving.value = false
  })
}

function remove(row) {
  ElMessageBox.confirm(`确定删除“${row.name}”吗？关联的报名记录会一并删除，树洞记录会保留为匿名。`, "删除教师", {
    type: "warning"
  }).then(() => {
    return http.delete(`/profile/teachers/${row.id}`)
  }).then(() => {
    ElMessage.success("删除成功")
    load()
  }).catch(error => {
    if (error !== "cancel") {
      ElMessage.error(error.message || "删除失败")
    }
  })
}

function roleText(role) {
  const map = {
    teacher: "教师",
    party_admin: "二级党委管理员",
    school_admin: "校级管理员"
  }
  return map[role] || role || "-"
}

onMounted(load)
</script>

<style scoped>
.full-width {
  width: 100%;
}

.page-alert {
  margin-bottom: 16px;
}
</style>
