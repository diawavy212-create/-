<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">教师树洞</h1>
      <el-button type="primary" @click="load">刷新</el-button>
    </div>
    <div class="panel">
      <el-table v-loading="loading" :data="items" stripe class="treehole-table">
        <el-table-column prop="displayId" label="ID" width="60" sortable />
        <el-table-column prop="teacherName" label="教师" width="90" show-overflow-tooltip />
        <el-table-column prop="anonymousText" label="实名/匿名" width="95" />
        <el-table-column prop="college" label="学院" width="90" show-overflow-tooltip />
        <el-table-column
          prop="category"
          label="类目"
          width="100"
          :filters="categoryFilters"
          :filter-method="filterCategory"
        />
        <el-table-column
          prop="emergencyText"
          label="紧急程度"
          width="90"
          :filters="emergencyFilters"
          :filter-method="filterEmergency"
        />
        <el-table-column
          prop="statusText"
          label="状态"
          width="85"
          :filters="statusFilters"
          :filter-method="filterStatus"
        />
        <el-table-column label="满意度" width="100">
          <template #default="{ row }">
            <el-tag v-if="satisfactionDisplay(row) !== '未评价'" :type="satisfactionTagType(row)">
              {{ satisfactionDisplay(row) }}
            </el-tag>
            <span v-else>未评价</span>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
        <el-table-column label="附件" width="65">
          <template #default="{ row }">
            <el-tag v-if="row.attachmentUrl" type="success">有</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" class-name="action-column">
          <template #default="{ row }">
            <div class="action-row">
              <el-button type="primary" link @click="openDetail(row)">详情</el-button>
              <el-button type="primary" link :disabled="row.status !== 0" @click="accept(row)">受理</el-button>
              <el-button type="primary" link :disabled="row.status === 4" @click="feedback(row)">反馈</el-button>
              <el-button type="success" link :disabled="row.status === 4" @click="complete(row)">处理</el-button>
              <el-button type="danger" link @click="remove(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="detailVisible" title="诉求详情" width="760px">
      <div v-if="current" class="detail-view">
        <div class="detail-title">{{ current.title }}</div>
        <div class="detail-meta">
          <el-tag>{{ current.anonymousText }}</el-tag>
          <el-tag type="warning">{{ current.category || "未分类" }}</el-tag>
          <el-tag type="danger">{{ current.emergencyText }}</el-tag>
          <el-tag type="info">{{ current.statusText }}</el-tag>
        </div>
        <div class="detail-block">
          <div class="detail-label">内容</div>
          <div class="detail-content">{{ current.content || current.description || "-" }}</div>
        </div>
        <div class="detail-block">
          <div class="detail-label">反馈</div>
          <div class="detail-content">{{ current.handleContent || "暂无" }}</div>
        </div>
        <div class="detail-block">
          <div class="detail-label">满意度评价</div>
          <div class="detail-content">{{ satisfactionDisplay(current) }}</div>
        </div>
        <div class="detail-block">
          <div class="detail-label">附件</div>
          <el-image
            v-if="current.attachmentUrl"
            class="attachment-image"
            :src="fileURL(current.attachmentUrl)"
            :preview-src-list="[fileURL(current.attachmentUrl)]"
            fit="cover"
          />
          <span v-else>暂无</span>
        </div>
      </div>
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
  { text: "已评价", value: 3 },
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

function hasSatisfaction(row) {
  return row && Number(row.satisfaction) >= 0
}

function satisfactionText(value) {
  const map = {
    0: "不满意",
    1: "基本满意",
    2: "满意",
    3: "非常满意"
  }
  return map[value] || "未评价"
}

function satisfactionDisplay(row) {
  if (!row) return "未评价"
  if (hasSatisfaction(row)) return row.satisfactionText || satisfactionText(Number(row.satisfaction))
  if (row.evaluated || row.status === 3) return "评分缺失"
  return "未评价"
}

function satisfactionTagType(row) {
  if (!hasSatisfaction(row)) return "warning"
  const value = Number(row.satisfaction)
  const map = {
    0: "danger",
    1: "warning",
    2: "success",
    3: "success"
  }
  return map[value] || "info"
}

function load() {
  loading.value = true
  http.get("/treeholes", { params: { page: 1, size: 20 } }).then(data => {
    items.value = (data.list || []).map((item, index) => ({
      ...item,
      displayId: index + 1
    }))
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

.action-row {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 10px;
}

.action-row :deep(.el-button + .el-button) {
  margin-left: 0;
}

:deep(.treehole-table .cell) {
  white-space: nowrap;
}

:deep(.action-column .cell) {
  overflow: visible;
}

.detail-view {
  color: #1f2937;
}

.detail-title {
  margin-bottom: 12px;
  font-size: 20px;
  font-weight: 700;
  word-break: break-word;
}

.detail-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 18px;
}

.detail-block {
  margin-bottom: 18px;
}

.detail-label {
  margin-bottom: 8px;
  color: #64748b;
  font-weight: 700;
}

.detail-content {
  padding: 12px;
  min-height: 48px;
  line-height: 1.8;
  white-space: pre-wrap;
  word-break: break-word;
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}
</style>
