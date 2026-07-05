<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">思政培训</h1>
      <el-button type="primary" @click="openCreate">发布培训</el-button>
    </div>
    <div class="panel">
      <el-table v-loading="loading" :data="items" stripe>
        <el-table-column prop="title" label="培训名称" min-width="220" />
        <el-table-column prop="location" label="地点" min-width="140" />
        <el-table-column prop="quota" label="名额" width="100" />
        <el-table-column prop="hours" label="学时" width="100" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">{{ trainingStatusText(row.statusText) }}</template>
        </el-table-column>
        <el-table-column label="已报名" width="110">
          <template #default="{ row }">{{ row.enrolledCount || 0 }} 人</template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openRecords(row)">报名名单</el-button>
            <el-button type="primary" link @click="openEdit(row)">编辑</el-button>
            <el-button type="danger" link @click="deleteTraining(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="formVisible" :title="form.id ? '编辑培训' : '发布培训'" width="720px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="培训名称" required>
          <el-input v-model="form.title" placeholder="请输入培训名称" />
        </el-form-item>
        <el-form-item label="培训类型">
          <el-input v-model="form.type" placeholder="training / workshop / seminar" />
        </el-form-item>
        <el-form-item label="培训层级">
          <el-select v-model="form.level" class="full-width">
            <el-option label="校级" :value="0" />
            <el-option label="院级" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item label="主办单位">
          <el-input v-model="form.sponsorUnit" />
        </el-form-item>
        <el-form-item label="承办单位">
          <el-input v-model="form.organizerUnit" />
        </el-form-item>
        <el-form-item label="开始时间">
          <el-date-picker v-model="form.startTime" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" class="full-width" />
        </el-form-item>
        <el-form-item label="结束时间">
          <el-date-picker v-model="form.endTime" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" class="full-width" />
        </el-form-item>
        <el-form-item label="地点/链接">
          <el-input v-model="form.location" />
        </el-form-item>
        <el-form-item label="名额">
          <el-input-number v-model="form.quota" :min="0" class="full-width" />
        </el-form-item>
        <el-form-item label="报名要求">
          <el-input v-model="form.requirements" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="成果要求">
          <el-input v-model="form.achievementRequire" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" class="full-width">
            <el-option label="草稿" :value="0" />
            <el-option label="开放报名" :value="1" />
            <el-option label="进行中" :value="2" />
            <el-option label="已结束" :value="3" />
            <el-option label="已归档" :value="4" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveTraining">{{ form.id ? "保存" : "发布" }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="recordsVisible" :title="`${currentTrainingTitle} 报名名单`" width="900px">
      <el-table v-loading="recordsLoading" :data="records" stripe empty-text="暂无报名记录">
        <el-table-column prop="teacherName" label="姓名" width="130" />
        <el-table-column prop="employeeNo" label="工号" width="130" />
        <el-table-column prop="department" label="部门" min-width="150" />
        <el-table-column prop="phone" label="手机" width="150" />
        <el-table-column prop="email" label="邮箱" min-width="190" />
        <el-table-column label="报名状态" width="120">
          <template #default="{ row }">{{ applyStatusText(row.applyStatus) }}</template>
        </el-table-column>
        <el-table-column prop="applyTime" label="报名时间" width="170" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ElMessage, ElMessageBox } from "element-plus"
import { onMounted, ref } from "vue"
import { http } from "../../api/http"

const items = ref([])
const loading = ref(false)
const records = ref([])
const recordsVisible = ref(false)
const recordsLoading = ref(false)
const currentTrainingTitle = ref("")
const formVisible = ref(false)
const saving = ref(false)
const form = ref(defaultForm())

function defaultForm() {
  return {
    title: "",
    type: "training",
    level: 0,
    sponsorUnit: "",
    organizerUnit: "",
    startTime: "",
    endTime: "",
    location: "",
    quota: 0,
    requirements: "",
    achievementRequire: "",
    status: 1
  }
}

function loadTrainings() {
  loading.value = true
  http.get("/trainings", { params: { page: 1, size: 50 } })
    .then(data => {
      items.value = data.list || []
    })
    .finally(() => {
      loading.value = false
    })
}

function openCreate() {
  form.value = defaultForm()
  formVisible.value = true
}

function openEdit(row) {
  form.value = {
    ...defaultForm(),
    ...row,
    sponsorUnit: row.sponsorUnit || "",
    organizerUnit: row.organizerUnit || "",
    requirements: row.requirements || "",
    achievementRequire: row.achievementRequire || ""
  }
  formVisible.value = true
}

function saveTraining() {
  if (!form.value.title) {
    ElMessage.warning("请填写培训名称")
    return
  }
  saving.value = true
  const action = form.value.id
    ? http.put(`/trainings/${form.value.id}`, form.value)
    : http.post("/trainings", form.value)
  action.then(() => {
    ElMessage.success(form.value.id ? "保存成功" : "发布成功")
    formVisible.value = false
    loadTrainings()
  }).finally(() => {
    saving.value = false
  })
}

function deleteTraining(row) {
  ElMessageBox.confirm(`确定删除“${row.title}”吗？相关报名记录也会删除。`, "删除培训", {
    type: "warning"
  }).then(() => {
    return http.delete(`/trainings/${row.id}`)
  }).then(() => {
    ElMessage.success("删除成功")
    loadTrainings()
  }).catch(() => {})
}

function openRecords(row) {
  currentTrainingTitle.value = row.title
  recordsVisible.value = true
  recordsLoading.value = true
  records.value = []
  http.get(`/trainings/${row.id}/records`, { params: { page: 1, size: 50 } })
    .then(data => {
      records.value = data.list || []
    })
    .finally(() => {
      recordsLoading.value = false
    })
}

function trainingStatusText(status) {
  const map = {
    draft: "草稿",
    open: "开放报名",
    in_progress: "进行中",
    ended: "已结束",
    archived: "已归档"
  }
  return map[status] || status || "-"
}

function applyStatusText(status) {
  const map = {
    0: "待审核",
    1: "已通过",
    2: "已驳回"
  }
  return map[status] || "-"
}

onMounted(loadTrainings)
</script>

<style scoped>
.full-width {
  width: 100%;
}
</style>
