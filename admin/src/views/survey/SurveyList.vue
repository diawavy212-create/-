<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">思想状况调研</h1>
      <el-button type="primary" @click="openCreate">发布问卷</el-button>
    </div>

    <div class="panel">
      <el-table v-loading="loading" :data="items" stripe class="survey-table">
        <el-table-column prop="title" label="问卷标题" min-width="180" show-overflow-tooltip />
        <el-table-column prop="typeText" label="类型" width="100" />
        <el-table-column prop="scope" label="投放范围" width="100" />
        <el-table-column prop="college" label="学院" width="120" show-overflow-tooltip>
          <template #default="{ row }">{{ row.college || "全校" }}</template>
        </el-table-column>
        <el-table-column prop="group" label="群体" width="110" show-overflow-tooltip>
          <template #default="{ row }">{{ row.group || "全部" }}</template>
        </el-table-column>
        <el-table-column label="周期" min-width="230">
          <template #default="{ row }">{{ timeText(row) }}</template>
        </el-table-column>
        <el-table-column prop="questionCount" label="题数" width="70" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">{{ row.statusText }}</template>
        </el-table-column>
        <el-table-column label="操作" width="210" fixed="right">
          <template #default="{ row }">
            <div class="action-row">
              <el-button type="primary" link @click="openReport(row)">报告</el-button>
              <el-button type="primary" link @click="openEdit(row)">编辑</el-button>
              <el-button type="danger" link @click="deleteSurvey(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="formVisible" :title="form.id ? '编辑问卷' : '发布问卷'" width="820px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="问卷标题" required>
          <el-input v-model="form.title" placeholder="请输入问卷标题" />
        </el-form-item>
        <el-form-item label="问卷类型">
          <el-select v-model="form.type" class="full-width">
            <el-option label="常态短测" :value="0" />
            <el-option label="年度长测" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item label="投放范围">
          <el-select v-model="form.scope" class="full-width">
            <el-option label="全校" value="全校" />
            <el-option label="学院" value="学院" />
            <el-option label="特定群体" value="特定群体" />
          </el-select>
        </el-form-item>
        <el-form-item label="投放学院">
          <el-input v-model="form.college" placeholder="全校范围可留空" />
        </el-form-item>
        <el-form-item label="投放群体">
          <el-input v-model="form.group" placeholder="如：青年教师/管理员" />
        </el-form-item>
        <el-form-item label="开始时间">
          <el-date-picker v-model="form.startTime" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" class="full-width" />
        </el-form-item>
        <el-form-item label="截止时间">
          <el-date-picker v-model="form.endTime" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" class="full-width" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" class="full-width">
            <el-option label="未发布" :value="0" />
            <el-option label="进行中" :value="1" />
            <el-option label="已结束" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="题库配置">
          <el-input v-model="questionsText" type="textarea" :rows="12" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveSurvey">{{ form.id ? "保存" : "发布" }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="reportVisible" title="调研分析报告" width="920px">
      <div v-if="report" class="report-view">
        <div class="report-title">{{ report.title }}</div>
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-value">{{ report.participationRate }}%</div>
            <div class="stat-label">参与率</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ report.validTotal }}</div>
            <div class="stat-label">有效问卷</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ report.invalidTotal }}</div>
            <div class="stat-label">过滤无效</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ report.targetTotal }}</div>
            <div class="stat-label">目标教师</div>
          </div>
        </div>

        <div class="report-section">
          <h3>选项占比</h3>
          <el-table :data="report.optionStats || []" size="small" empty-text="暂无选项数据">
            <el-table-column prop="question" label="题目" min-width="180" show-overflow-tooltip />
            <el-table-column prop="option" label="选项" width="140" />
            <el-table-column prop="count" label="数量" width="90" />
          </el-table>
        </div>

        <div class="report-section two-columns">
          <div>
            <h3>学院对比</h3>
            <div class="pill-row" v-for="item in report.collegeCompare || []" :key="item.college">
              <span>{{ item.college }}</span>
              <strong>{{ item.count }}</strong>
            </div>
          </div>
          <div>
            <h3>开放题高频主题</h3>
            <div class="pill-row" v-for="item in report.openTopics || []" :key="item.topic">
              <span>{{ item.topic }}</span>
              <strong>{{ item.count }}</strong>
            </div>
          </div>
        </div>

        <div class="report-section">
          <h3>风险提示清单</h3>
          <el-alert
            v-for="(risk, index) in report.riskList || []"
            :key="index"
            :title="risk.content"
            :type="risk.level === 'high' ? 'error' : risk.level === 'medium' ? 'warning' : 'success'"
            show-icon
            :closable="false"
            class="risk-alert"
          />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ElMessage, ElMessageBox } from "element-plus"
import { onMounted, ref } from "vue"
import { http } from "../../api/http"

const items = ref([])
const loading = ref(false)
const formVisible = ref(false)
const saving = ref(false)
const reportVisible = ref(false)
const report = ref(null)
const form = ref(defaultForm())
const questionsText = ref("")

function defaultQuestions() {
  return [
    {
      title: "近期工作压力感受",
      type: "single",
      required: true,
      options: [
        { label: "较轻", score: 1 },
        { label: "适中", score: 2 },
        { label: "压力较大", score: 3 }
      ]
    },
    {
      title: "对学院支持保障的满意度",
      type: "single",
      required: true,
      options: [
        { label: "满意", score: 1 },
        { label: "基本满意", score: 2 },
        { label: "不满意", score: 3 }
      ]
    },
    { title: "希望学校重点改进的问题", type: "text", required: false, options: [] }
  ]
}

function defaultForm() {
  return {
    title: "",
    type: 0,
    scope: "全校",
    college: "",
    group: "",
    startTime: "",
    endTime: "",
    status: 1,
    questions: defaultQuestions()
  }
}

function load() {
  loading.value = true
  http.get("/survey/list", { params: { page: 1, size: 50 } })
    .then(data => {
      items.value = data.list || []
    })
    .finally(() => {
      loading.value = false
    })
}

function openCreate() {
  form.value = defaultForm()
  questionsText.value = JSON.stringify(form.value.questions, null, 2)
  formVisible.value = true
}

function openEdit(row) {
  form.value = {
    ...defaultForm(),
    ...row,
    college: row.college || "",
    group: row.group || "",
    questions: row.questions || defaultQuestions()
  }
  questionsText.value = JSON.stringify(form.value.questions, null, 2)
  formVisible.value = true
}

function saveSurvey() {
  if (!form.value.title) {
    ElMessage.warning("请填写问卷标题")
    return
  }
  let questions
  try {
    questions = JSON.parse(questionsText.value || "[]")
  } catch (error) {
    ElMessage.error("题库配置不是合法 JSON")
    return
  }
  saving.value = true
  const payload = { ...form.value, questions }
  const action = form.value.id
    ? http.put(`/survey/${form.value.id}`, payload)
    : http.post("/survey/create", payload)
  action.then(() => {
    ElMessage.success(form.value.id ? "保存成功" : "发布成功")
    formVisible.value = false
    load()
  }).catch(error => {
    ElMessage.error(error.message || "问卷发布失败")
  }).finally(() => {
    saving.value = false
  })
}

function deleteSurvey(row) {
  ElMessageBox.confirm(`确定删除“${row.title}”吗？相关答卷也会删除。`, "删除问卷", {
    type: "warning"
  }).then(() => {
    return http.delete(`/survey/${row.id}`)
  }).then(() => {
    ElMessage.success("删除成功")
    load()
  }).catch(() => {})
}

function openReport(row) {
  report.value = null
  reportVisible.value = true
  http.get(`/survey/report/${row.id}`).then(data => {
    report.value = data
  }).catch(error => {
    ElMessage.error(error.message || "报告加载失败")
  })
}

function shortTime(value) {
  if (!value) return ""
  return String(value).replace(/:\d{2}$/, "")
}

function timeText(row) {
  const start = shortTime(row.startTime)
  const end = shortTime(row.endTime)
  if (start && end) return `${start} 至 ${end}`
  return start || end || "-"
}

onMounted(load)
</script>

<style scoped>
.full-width {
  width: 100%;
}

.action-row {
  display: flex;
  gap: 10px;
}

:deep(.survey-table .cell) {
  white-space: nowrap;
}

.report-title {
  margin-bottom: 18px;
  font-size: 20px;
  font-weight: 800;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.stat-card {
  padding: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #f8fafc;
}

.stat-value {
  color: #c51624;
  font-size: 26px;
  font-weight: 900;
}

.stat-label {
  margin-top: 4px;
  color: #64748b;
}

.report-section {
  margin-top: 22px;
}

.report-section h3 {
  margin: 0 0 12px;
  font-size: 16px;
}

.two-columns {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}

.pill-row {
  display: flex;
  justify-content: space-between;
  padding: 10px 12px;
  margin-bottom: 8px;
  border-radius: 8px;
  background: #f8fafc;
}

.risk-alert {
  margin-bottom: 10px;
}
</style>
