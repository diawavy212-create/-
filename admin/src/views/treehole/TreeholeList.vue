<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">教师树洞</h1>
      <el-button type="primary" @click="load">刷新</el-button>
    </div>
    <div class="panel">
      <el-table v-loading="loading" :data="items" stripe>
        <el-table-column prop="id" label="ID" width="80" sortable />
        <el-table-column prop="teacherName" label="教师" width="130" />
        <el-table-column prop="anonymousText" label="实名/匿名" width="120" />
        <el-table-column prop="college" label="学院" width="130" />
        <el-table-column
          prop="category"
          label="类目"
          width="130"
          :filters="categoryFilters"
          :filter-method="filterCategory"
        />
        <el-table-column
          prop="emergencyText"
          label="紧急程度"
          width="130"
          :filters="emergencyFilters"
          :filter-method="filterEmergency"
        />
        <el-table-column
          prop="statusText"
          label="状态"
          width="130"
          :filters="statusFilters"
          :filter-method="filterStatus"
        />
        <el-table-column prop="title" label="标题" min-width="220" />
        <el-table-column label="附件" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.attachmentUrl" type="success">有</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="310" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openDetail(row)">详情</el-button>
            <el-button size="small" :disabled="row.status !== 0" @click="accept(row)">受理</el-button>
            <el-button size="small" type="primary" :disabled="row.status === 4" @click="feedback(row)">反馈</el-button>
            <el-button size="small" type="success" :disabled="row.status === 4" @click="complete(row)">已处理</el-button>
            <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="detailVisible" title="诉求详情" width="680px">
      <el-descriptions v-if="current" border :column="1">
        <el-descriptions-item label="标题">{{ current.title }}</el-descriptions-item>
        <el-descriptions-item label="内容">{{ current.content }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ current.statusText }}</el-descriptions-item>
        <el-descriptions-item label="反馈">{{ current.handleContent || "暂无" }}</el-descriptions-item>
        <el-descriptions-item label="附件">
          <el-image
            v-if="current.attachmentUrl"
            class="attachment-image"
            :src="fileURL(current.attachmentUrl)"
            :preview-src-list="[fileURL(current.attachmentUrl)]"
            fit="cover"
          />
          <span v-else>暂无</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ElMessage, ElMessageBox } from "element-plus"
import { computed, onMounted, ref } from "vue"
import { http } from "../../api/http"

const items = ref([])
const loading = ref(false)
const detailVisible = ref(false)
const current = ref(null)

const categoryFilters = computed(() => {
  return [...new Set(items.value.map(item => item.category).filter(Boolean))]
    .map(value => ({ text: value, value }))
})

const emergencyFilters = [
  { text: "普通", value: 0 },
  { text: "较急", value: 1 },
  { text: "紧急", value: 2 }
]

const statusFilters = [
  { text: "待受理", value: 0 },
  { text: "处理中", value: 1 },
  { text: "已反馈", value: 2 },
  { text: "已处理", value: 4 }
]

function filterCategory(value, row) {
  return row.category === value
}

function filterEmergency(value, row) {
  return row.emergencyLevel === value
}

function filterStatus(value, row) {
  return row.status === value
}

function load() {
  loading.value = true
  http.get("/treeholes", { params: { page: 1, size: 20 } }).then(data => {
    items.value = data.list || []
  }).catch(error => {
    ElMessage.error(error.message || "树洞列表加载失败")
  }).finally(() => {
    loading.value = false
  })
}

function fileURL(path) {
  if (!path) return ""
  if (/^https?:\/\//.test(path)) return path
  return path
}

function openDetail(row) {
  current.value = row
  detailVisible.value = true
}

function accept(row) {
  http.post(`/treeholes/${row.id}/accept`).then(() => {
    ElMessage.success("已受理")
    load()
  }).catch(error => {
    ElMessage.error(error.message || "受理失败")
  })
}

function feedback(row) {
  ElMessageBox.prompt("请输入办理反馈", "反馈诉求", {
    confirmButtonText: "提交",
    cancelButtonText: "取消",
    inputType: "textarea",
    inputValue: row.handleContent || "已收到诉求，正在跟进处理。"
  }).then(({ value }) => {
    return http.post(`/treeholes/${row.id}/feedback`, {
      handlerUnit: row.college || "管理端",
      handleContent: value
    })
  }).then(() => {
    ElMessage.success("已反馈")
    load()
  }).catch(error => {
    if (error !== "cancel") {
      ElMessage.error(error.message || "反馈失败")
    }
  })
}

function complete(row) {
  http.post(`/treeholes/${row.id}/complete`).then(() => {
    ElMessage.success("已标记处理")
    load()
  }).catch(error => {
    ElMessage.error(error.message || "标记失败")
  })
}

function remove(row) {
  ElMessageBox.confirm(`确定删除“${row.title}”吗？`, "删除诉求", {
    type: "warning"
  }).then(() => {
    return http.delete(`/treeholes/${row.id}`)
  }).then(() => {
    ElMessage.success("删除成功")
    load()
  }).catch(() => {})
}

onMounted(load)
</script>

<style scoped>
.attachment-image {
  width: 160px;
  height: 120px;
  border-radius: 8px;
}
</style>
