<template>
  <div>
    <div class="dashboard-hero">
      <div>
        <h2>教师服务后台工作台</h2>
      </div>
      <el-button type="primary" @click="load">刷新数据</el-button>
    </div>

    <div class="stats">
      <div class="stat">
        <div class="stat-label">树洞待受理</div>
        <div class="stat-value">{{ treehole.pending }}</div>
        <div class="stat-hint">需要尽快受理分派</div>
      </div>
      <div class="stat">
        <div class="stat-label">树洞处理中</div>
        <div class="stat-value">{{ treehole.processing }}</div>
        <div class="stat-hint">跟进办理进展和反馈</div>
      </div>
      <div class="stat">
        <div class="stat-label">开放培训</div>
        <div class="stat-value">{{ training.open }}</div>
        <div class="stat-hint">可报名或正在开展</div>
      </div>
      <div class="stat">
        <div class="stat-label">累计学时</div>
        <div class="stat-value">{{ training.hours }}</div>
        <div class="stat-hint">来自教师培训台账</div>
      </div>
    </div>

    <div class="module-grid">
      <div class="panel module-card">
        <h3 class="module-title">教师树洞</h3>
        <div class="module-item">
          <span>待受理诉求</span>
          <strong>{{ treehole.pending }}</strong>
        </div>
        <div class="module-item">
          <span>已反馈诉求</span>
          <strong>{{ treehole.feedback }}</strong>
        </div>
        <el-button type="primary" link @click="$router.push('/treeholes')">进入树洞管理</el-button>
      </div>

      <div class="panel module-card">
        <h3 class="module-title">思政培训</h3>
        <div class="module-item">
          <span>开放/进行中</span>
          <strong>{{ training.open }}</strong>
        </div>
        <div class="module-item">
          <span>已结束/归档</span>
          <strong>{{ training.completed }}</strong>
        </div>
        <el-button type="primary" link @click="$router.push('/trainings')">进入培训管理</el-button>
      </div>

      <div class="panel module-card">
        <h3 class="module-title">教师信息</h3>
        <div class="module-item">
          <span>管理重点</span>
          <strong>实名、工号、联系方式</strong>
        </div>
        <div class="module-item">
          <span>关联业务</span>
          <strong>报名校验、诉求回显</strong>
        </div>
        <el-button type="primary" link @click="$router.push('/profile')">进入信息管理</el-button>
      </div>
    </div>

    <div class="panel capacity-panel">
      <h3 class="module-title">系统容量</h3>
      <div class="capacity-grid">
        <div class="capacity-item">
          <span>教师账号</span>
          <strong>{{ system.teachers }}</strong>
        </div>
        <div class="capacity-item">
          <span>树洞诉求</span>
          <strong>{{ system.appeals }}</strong>
        </div>
        <div class="capacity-item">
          <span>培训活动</span>
          <strong>{{ system.trainings }}</strong>
        </div>
        <div class="capacity-item">
          <span>报名记录</span>
          <strong>{{ system.trainingRecords }}</strong>
        </div>
      </div>
    </div>

    <div class="panel">
      <h3 class="module-title">今日处理建议</h3>
      <el-timeline>
        <el-timeline-item type="danger" :timestamp="treehole.pending ? '优先' : '正常'">
          {{ treehole.pending ? `有 ${treehole.pending} 条树洞诉求待受理，建议先完成受理或分派。` : "暂无待受理树洞诉求，保持日常巡检。" }}
        </el-timeline-item>
        <el-timeline-item type="primary" timestamp="跟进">
          {{ treehole.processing ? `有 ${treehole.processing} 条诉求处理中，注意补充办理反馈。` : "处理中诉求为 0，可关注已反馈事项是否需要标记已处理。" }}
        </el-timeline-item>
        <el-timeline-item type="success" timestamp="运营">
          {{ training.open ? `当前有 ${training.open} 个培训开放或进行中，关注报名人数和台账完整性。` : "暂无开放培训，可按计划发布新的思政培训活动。" }}
        </el-timeline-item>
      </el-timeline>
    </div>
  </div>
</template>

<script setup>
import { ElMessage } from "element-plus"
import { onMounted, ref } from "vue"
import { http } from "../../api/http"

const treehole = ref({
  pending: 0,
  processing: 0,
  feedback: 0,
  archived: 0
})

const training = ref({
  open: 0,
  completed: 0,
  hours: 0
})

const system = ref({
  teachers: 0,
  appeals: 0,
  trainings: 0,
  trainingRecords: 0
})

function load() {
  Promise.all([
    http.get("/treeholes/statistics"),
    http.get("/trainings/statistics"),
    http.get("/system/summary")
  ]).then(([treeholeData, trainingData, systemData]) => {
    treehole.value = {
      pending: treeholeData.pending || 0,
      processing: treeholeData.processing || 0,
      feedback: treeholeData.feedback || 0,
      archived: treeholeData.archived || 0
    }
    training.value = {
      open: trainingData.open || 0,
      completed: trainingData.completed || 0,
      hours: trainingData.hours || 0
    }
    system.value = {
      teachers: systemData.teachers || 0,
      appeals: systemData.appeals || 0,
      trainings: systemData.trainings || 0,
      trainingRecords: systemData.trainingRecords || 0
    }
  }).catch(error => {
    ElMessage.error(error.message || "工作台数据加载失败")
  })
}

onMounted(load)
</script>

<style scoped>
.dashboard-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 22px;
  margin-bottom: 18px;
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.dashboard-hero h2 {
  margin: 0;
  font-size: 22px;
}

.stat-label {
  color: #475569;
  font-weight: 700;
}

.stat-hint {
  margin-top: 6px;
  color: #94a3b8;
  font-size: 13px;
}

.module-card {
  margin-bottom: 18px;
}

.capacity-panel {
  margin-bottom: 18px;
}

.capacity-grid {
  display: grid;
  grid-template-columns: repeat(6, minmax(120px, 1fr));
  gap: 12px;
}

.capacity-item {
  padding: 14px;
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.capacity-item span {
  display: block;
  color: #64748b;
  font-size: 13px;
}

.capacity-item strong {
  display: block;
  margin-top: 8px;
  color: #0f172a;
  font-size: 20px;
}

@media (max-width: 1200px) {
  .capacity-grid {
    grid-template-columns: repeat(3, minmax(120px, 1fr));
  }
}
</style>
