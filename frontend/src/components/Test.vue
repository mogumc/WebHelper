<script setup>
import { ref, onMounted, nextTick } from 'vue'
import {
  Greet, Flashtime, Gettestjson,
  GetALLLang, GetCurrentLang, SetLanguage, GetLangTextMap,
  GetLogFiles, GetLogFileContent, SetLogLevel, GetLogLevel,
  GetSystemInfo, OpenFileSelect, OpenFolderSelect, SaveFileSelect,
  ReadFileContent, WriteFileContent, Notify
} from '../../wailsjs/go/service/App'
import { ElMessage } from 'element-plus'
import { EventsOn } from '../../wailsjs/runtime'
import { useI18n } from '../composables/useI18n'
import {
  ChatDotRound, Timer, DataLine, Document,
  Setting, Monitor, View, Cpu, FolderOpened, Edit, Bell
} from '@element-plus/icons-vue'

const { t, textMap } = useI18n()

// ==================== 1. 前后端数据交互 (Greet) ====================
const greetName = ref('')
const greetResult = ref('')
const greetLoading = ref(false)

function doGreet() {
  if (!greetName.value.trim()) {
    ElMessage.warning(t('input_name', '输入名字'))
    return
  }
  greetLoading.value = true
  Greet(greetName.value)
    .then((result) => {
      greetResult.value = result
      greetLoading.value = false
    })
    .catch((err) => {
      ElMessage.error(String(err))
      greetLoading.value = false
    })
}

// ==================== 2. 事件系统 (实时时间) ====================
const systime = ref('Loading...')
Flashtime()
EventsOn('time', (time) => {
  systime.value = time
})

// ==================== 3. JSON 数据交互 ====================
const jsonData = ref(null)
const dataLoading = ref(false)

function convertBytes(byteSize) {
  byteSize = Number(byteSize)
  if (byteSize < 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let index = 0
  while (byteSize >= 1024 && index < units.length - 1) {
    byteSize /= 1024
    index++
  }
  return `${byteSize.toFixed(2)} ${units[index]}`
}

function fetchJsonData() {
  dataLoading.value = true
  Gettestjson()
    .then((result) => {
      const parsed = JSON.parse(result)
      jsonData.value = parsed.data.map((item) => ({
        name: item.name,
        type: item.isdir ? t('folder', '文件夹') : t('file', '文件'),
        size: item.isdir ? '---' : convertBytes(item.size),
        time: '---',
        path: item.name,
      }))
      dataLoading.value = false
    })
    .catch((err) => {
      ElMessage.error(String(err))
      dataLoading.value = false
    })
}

// ==================== 4. 语言切换 ====================
const langList = ref([])
const currentLang = ref('')

onMounted(async () => {
  try {
    const langInfos = await GetALLLang()
    langList.value = langInfos.map(info => ({
      code: info.language_code,
      name: info.language_name
    }))
    // 获取当前语言
    currentLang.value = await GetCurrentLang() || 'zh-CN'
  } catch (e) {
    console.error('初始化失败:', e)
  }
})

async function handleLangChange(langCode) {
  try {
    await SetLanguage(langCode)
    // 重新加载语言映射，无需刷新页面
    const map = await GetLangTextMap()
    textMap.value = map || {}
    ElMessage.success(t('lang_switched', '语言已切换'))
  } catch (err) {
    ElMessage.error(String(err))
  }
}

// ==================== 5. 日志等级设置 ====================
const currentLogLevel = ref('INFO')
const logLevelOptions = [
  { value: 'DEBUG', label: 'DEBUG' },
  { value: 'INFO', label: 'INFO' },
  { value: 'WARN', label: 'WARN' },
  { value: 'ERROR', label: 'ERROR' },
]

onMounted(async () => {
  try {
    const level = await GetLogLevel()
    currentLogLevel.value = level || 'INFO'
  } catch (e) {
    console.error('获取日志等级失败:', e)
  }
})

async function handleLogLevelChange(level) {
  try {
    const ok = await SetLogLevel(level)
    if (ok) {
      ElMessage.success(t('log_level_changed', '日志等级已切换为') + ' ' + level)
    }
  } catch (err) {
    ElMessage.error(String(err))
  }
}

// ==================== 6. 日志文件查看 ====================
const logFiles = ref([])
const selectedLogFile = ref('')
const logContent = ref('')
const logLoading = ref(false)

async function loadLogFiles() {
  try {
    logFiles.value = await GetLogFiles()
  } catch (err) {
    console.error('获取日志列表失败:', err)
  }
}

async function handleLogFileChange(filename) {
  if (!filename) {
    logContent.value = ''
    return
  }
  logLoading.value = true
  try {
    logContent.value = await GetLogFileContent(filename)
    await nextTick()
    // 滚动到底部
    const el = document.getElementById('log-viewer')
    if (el) el.scrollTop = el.scrollHeight
  } catch (err) {
    ElMessage.error(String(err))
  } finally {
    logLoading.value = false
  }
}

onMounted(() => {
  loadLogFiles()
})

// ==================== 7. 系统信息 ====================
const sysInfo = ref(null)
const sysInfoLoading = ref(false)

async function fetchSystemInfo() {
  sysInfoLoading.value = true
  try {
    sysInfo.value = await GetSystemInfo()
  } catch (err) {
    ElMessage.error(String(err))
  } finally {
    sysInfoLoading.value = false
  }
}

onMounted(() => { fetchSystemInfo() })

// ==================== 8. 文件对话框 ====================
const dialogFilePath = ref('')

async function handleOpenFile() {
  const path = await OpenFileSelect()
  if (path) {
    dialogFilePath.value = path
  }
}

async function handleOpenFolder() {
  const path = await OpenFolderSelect()
  if (path) {
    dialogFilePath.value = path
  }
}

async function handleSaveFile() {
  const path = await SaveFileSelect()
  if (path) {
    dialogFilePath.value = path
  }
}

// ==================== 9. 文件读写 ====================
const rwFilePath = ref('')
const rwFileContent = ref('')
const rwLoading = ref(false)

async function handleReadFile() {
  if (!rwFilePath.value.trim()) {
    ElMessage.warning(t('file_path_empty', '请输入文件路径'))
    return
  }
  rwLoading.value = true
  try {
    const content = await ReadFileContent(rwFilePath.value)
    rwFileContent.value = content
    ElMessage.success(t('file_read_success', '文件读取成功'))
  } catch (err) {
    ElMessage.error(String(err))
  } finally {
    rwLoading.value = false
  }
}

async function handleWriteFile() {
  if (!rwFilePath.value.trim()) {
    ElMessage.warning(t('file_path_empty', '请输入文件路径'))
    return
  }
  rwLoading.value = true
  try {
    const ok = await WriteFileContent(rwFilePath.value, rwFileContent.value)
    if (ok) {
      ElMessage.success(t('file_write_success', '文件写入成功'))
    } else {
      ElMessage.error(t('file_write_failed', '文件写入失败'))
    }
  } catch (err) {
    ElMessage.error(String(err))
  } finally {
    rwLoading.value = false
  }
}

// ==================== 10. 事件通知 ====================
const notifyTitle = ref('')
const notifyMessage = ref('')
const notifyLog = ref([])

EventsOn('notification', (data) => {
  notifyLog.value.unshift({
    time: new Date().toLocaleTimeString(),
    title: data.title,
    message: data.message,
  })
  ElMessage.info(`${t('notification_received', '收到通知')}: ${data.title}`)
})

function handleSendNotify() {
  if (!notifyTitle.value.trim()) {
    ElMessage.warning(t('notify_title', '通知标题'))
    return
  }
  Notify(notifyTitle.value, notifyMessage.value)
}

// ==================== 功能列表 ====================
const features = [
  { key: 'feature_i18n', desc: 'feature_i18n_desc', icon: '🌐', color: '#409EFF' },
  { key: 'feature_backend', desc: 'feature_backend_desc', icon: '🔗', color: '#67C23A' },
  { key: 'feature_event', desc: 'feature_event_desc', icon: '📡', color: '#E6A23C' },
  { key: 'feature_window', desc: 'feature_window_desc', icon: '🪟', color: '#F56C6C' },
  { key: 'feature_logger', desc: 'feature_logger_desc', icon: '📝', color: '#909399' },
  { key: 'feature_json', desc: 'feature_json_desc', icon: '📦', color: '#9B59B6' },
  { key: 'feature_ui', desc: 'feature_ui_desc', icon: '🎨', color: '#1ABC9C' },
  { key: 'feature_embed', desc: 'feature_embed_desc', icon: '💾', color: '#3498DB' },
  { key: 'feature_dialog', desc: 'feature_dialog_desc', icon: '📂', color: '#E74C3C' },
  { key: 'feature_fileio', desc: 'feature_fileio_desc', icon: '✏️', color: '#F39C12' },
]
</script>

<template>
  <div class="demo-container">
    <el-card class="section-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="card-title">{{ t('feature_overview', '功能概览') }}</span>
        </div>
      </template>
      <p class="overview-desc">{{ t('feature_overview_desc', '本模板集成了以下核心功能，可作为 Wails 项目的开发起点。') }}</p>
      <div class="feature-grid">
        <div v-for="feature in features" :key="feature.key" class="feature-item">
          <div class="feature-icon" :style="{ backgroundColor: feature.color + '15', color: feature.color }">
            {{ feature.icon }}
          </div>
          <div class="feature-info">
            <div class="feature-name">{{ t(feature.key, feature.key) }}</div>
            <div class="feature-desc">{{ t(feature.desc, feature.desc) }}</div>
          </div>
        </div>
      </div>
    </el-card>

    <el-card class="section-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon><ChatDotRound /></el-icon>
          <span>{{ t('demo_data_interaction', '前后端数据交互') }}</span>
        </div>
      </template>
      <p class="section-desc">{{ t('demo_data_interaction_desc', '通过 Wails 绑定调用 Go 后端函数，获取返回值并展示。') }}</p>
      <div class="demo-row">
        <el-input
          v-model="greetName"
          :placeholder="t('greet_placeholder', '请输入你的名字')"
          style="max-width: 300px;"
          @keyup.enter="doGreet"
        />
        <el-button @click="doGreet" type="primary" :loading="greetLoading">
          {{ t('send_greet', '发送问候') }}
        </el-button>
      </div>
      <el-alert
        v-if="greetResult"
        :title="greetResult"
        type="success"
        show-icon
        :closable="false"
        style="margin-top: 12px;"
      />
    </el-card>

    <el-card class="section-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon><Timer /></el-icon>
          <span>{{ t('feature_event', '事件系统') }} - {{ t('realtime_time', '实时时间') }}</span>
        </div>
      </template>
      <p class="section-desc">{{ t('demo_event_desc', 'Go 后端通过 EventsEmit 推送数据，前端通过 EventsOn 监听接收。') }}</p>
      <div class="time-display">
        <span class="time-label">{{ t('time_label', '当前时间') }}</span>
        <span class="time-value">{{ systime }}</span>
      </div>
    </el-card>

    <el-card class="section-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon><DataLine /></el-icon>
          <span>{{ t('feature_json', 'JSON 数据处理') }}</span>
        </div>
      </template>
      <p class="section-desc">{{ t('demo_json_desc', '后端返回 JSON 字符串，前端解析后以表格形式展示。') }}</p>
      <div class="demo-row" style="margin-bottom: 16px;">
        <el-button @click="fetchJsonData" type="primary" :loading="dataLoading">
          {{ t('fetch_data', '获取数据') }}
        </el-button>
      </div>
      <el-table v-if="jsonData" :data="jsonData" stripe border style="width: 100%" max-height="300">
        <el-table-column prop="name" :label="t('file_name', '文件名')" min-width="150" />
        <el-table-column prop="type" :label="t('file_type', '文件类型')" min-width="100" />
        <el-table-column prop="size" :label="t('file_size', '文件大小')" min-width="100" />
        <el-table-column prop="time" :label="t('create_time', '创建时间')" min-width="160" />
      </el-table>
      <el-empty v-else :description="t('click_to_fetch', '点击按钮获取数据')" />
    </el-card>

    <el-card class="section-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon><Monitor /></el-icon>
          <span>{{ t('demo_lang_switch', '语言切换 (i18n)') }}</span>
        </div>
      </template>
      <p class="section-desc">{{ t('demo_lang_switch_desc', '切换应用语言，支持嵌入语言包和本地语言文件。') }}</p>
      <div class="demo-row">
        <span class="demo-label">{{ t('current_lang', '当前语言') }}:</span>
        <el-select v-model="currentLang" style="width: 200px;" @change="handleLangChange">
          <el-option
            v-for="lang in langList"
            :key="lang.code"
            :label="lang.name"
            :value="lang.code"
          />
        </el-select>
      </div>
    </el-card>

    <el-card class="section-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon><Setting /></el-icon>
          <span>{{ t('demo_log_level', '日志等级设置') }}</span>
        </div>
      </template>
      <p class="section-desc">{{ t('demo_log_level_desc', '动态调整日志输出等级。') }}</p>
      <div class="demo-row">
        <span class="demo-label">{{ t('log_level', '日志等级') }}:</span>
        <el-select v-model="currentLogLevel" style="width: 200px;" @change="handleLogLevelChange">
          <el-option
            v-for="opt in logLevelOptions"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </div>
    </el-card>

    <el-card class="section-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon><Document /></el-icon>
          <span>{{ t('demo_log_viewer', '日志文件查看') }}</span>
        </div>
      </template>
      <p class="section-desc">{{ t('demo_log_viewer_desc', '获取日志文件列表并查看内容，展示前后端数据交互的实际应用。') }}</p>
      <div class="demo-row" style="margin-bottom: 12px;">
        <el-select
          v-model="selectedLogFile"
          :placeholder="t('select_log_file', '选择日志文件')"
          style="width: 300px;"
          @change="handleLogFileChange"
        >
          <el-option
            v-for="file in logFiles"
            :key="file"
            :label="file"
            :value="file"
          />
        </el-select>
        <el-button @click="loadLogFiles" :icon="View" circle />
      </div>
      <div
        v-if="logContent"
        id="log-viewer"
        class="log-viewer"
        v-loading="logLoading"
      >
        <pre class="log-pre">{{ logContent }}</pre>
      </div>
      <el-empty v-else :description="t('select_log_to_view', '选择一个日志文件查看内容')" />
    </el-card>

    <el-card class="section-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon><Cpu /></el-icon>
          <span>{{ t('demo_system_info', '系统信息') }}</span>
        </div>
      </template>
      <p class="section-desc">{{ t('demo_system_info_desc', '获取当前操作系统、架构、CPU 核心数等系统信息。') }}</p>
      <div class="demo-row" style="margin-bottom: 12px;">
        <el-button @click="fetchSystemInfo" type="primary" :loading="sysInfoLoading">
          {{ t('fetch_data', '获取数据') }}
        </el-button>
      </div>
      <el-descriptions v-if="sysInfo" :column="2" border>
        <el-descriptions-item :label="t('os', '操作系统')">{{ sysInfo.os }}</el-descriptions-item>
        <el-descriptions-item :label="t('arch', '架构')">{{ sysInfo.arch }}</el-descriptions-item>
        <el-descriptions-item :label="t('cpu_count', 'CPU 核心数')">{{ sysInfo.num_cpu }}</el-descriptions-item>
        <el-descriptions-item :label="t('hostname', '主机名')">{{ sysInfo.hostname }}</el-descriptions-item>
        <el-descriptions-item :label="t('go_version', 'Go 版本')">{{ sysInfo.go_ver }}</el-descriptions-item>
        <el-descriptions-item :label="t('time_label', '当前时间')">{{ sysInfo.time }}</el-descriptions-item>
        <el-descriptions-item :label="t('process_name', '进程名')">{{ sysInfo.process_name }}</el-descriptions-item>
      </el-descriptions>
      <el-empty v-else :description="t('click_to_fetch', '点击按钮获取数据')" />
    </el-card>

    <el-card class="section-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon><FolderOpened /></el-icon>
          <span>{{ t('demo_file_dialog', '文件对话框') }}</span>
        </div>
      </template>
      <p class="section-desc">{{ t('demo_file_dialog_desc', '调用系统原生文件对话框，支持打开文件、选择目录和保存文件。') }}</p>
      <div class="demo-row" style="margin-bottom: 12px;">
        <el-button @click="handleOpenFile" type="primary">{{ t('open_file', '打开文件') }}</el-button>
        <el-button @click="handleOpenFolder" type="success">{{ t('open_folder', '选择目录') }}</el-button>
        <el-button @click="handleSaveFile" type="warning">{{ t('save_file', '保存文件') }}</el-button>
      </div>
      <el-alert
        v-if="dialogFilePath"
        :title="t('selected_path', '已选择路径') + ': ' + dialogFilePath"
        type="info"
        show-icon
        :closable="false"
      />
      <el-empty v-else :description="t('click_to_fetch', '点击按钮获取数据')" />
    </el-card>

    <el-card class="section-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon><Edit /></el-icon>
          <span>{{ t('demo_file_read_write', '文件读写') }}</span>
        </div>
      </template>
      <p class="section-desc">{{ t('demo_file_read_write_desc', '读取和写入本地文件，展示文件系统交互能力。') }}</p>
      <div class="demo-row" style="margin-bottom: 12px;">
        <el-input
          v-model="rwFilePath"
          :placeholder="t('file_path', '文件路径')"
          style="max-width: 400px;"
        />
        <el-button @click="handleReadFile" type="primary" :loading="rwLoading">
          {{ t('read_file', '读取文件') }}
        </el-button>
        <el-button @click="handleWriteFile" type="success" :loading="rwLoading">
          {{ t('write_file', '写入文件') }}
        </el-button>
      </div>
      <el-input
        v-model="rwFileContent"
        type="textarea"
        :rows="6"
        :placeholder="t('file_content', '文件内容')"
        style="width: 100%;"
      />
    </el-card>

    <el-card class="section-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon><Bell /></el-icon>
          <span>{{ t('demo_notification', '事件通知') }}</span>
        </div>
      </template>
      <p class="section-desc">{{ t('demo_notification_desc', '后端通过 EventsEmit 向前端发送自定义通知事件。') }}</p>
      <div class="demo-row" style="margin-bottom: 12px;">
        <el-input
          v-model="notifyTitle"
          :placeholder="t('notify_title', '通知标题')"
          style="max-width: 200px;"
        />
        <el-input
          v-model="notifyMessage"
          :placeholder="t('notify_message', '通知内容')"
          style="max-width: 300px;"
        />
        <el-button @click="handleSendNotify" type="primary">
          {{ t('send_notify', '发送通知') }}
        </el-button>
      </div>
      <div v-if="notifyLog.length" class="notify-log">
        <div v-for="(item, idx) in notifyLog" :key="idx" class="notify-item">
          <span class="notify-time">[{{ item.time }}]</span>
          <span class="notify-title">{{ item.title }}</span>
          <span class="notify-msg" v-if="item.message"> - {{ item.message }}</span>
        </div>
      </div>
      <el-empty v-else :description="t('click_to_fetch', '点击按钮获取数据')" />
    </el-card>
  </div>
</template>

<style scoped>
.demo-container {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.section-card {
  border-radius: 8px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.card-title {
  font-size: 18px;
  font-weight: 700;
}

.overview-desc,
.section-desc {
  color: #666;
  margin-bottom: 16px;
  font-size: 13px;
  line-height: 1.6;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.feature-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #f0f0f0;
  transition: all 0.3s;
}

.feature-item:hover {
  border-color: #409EFF;
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.1);
}

.feature-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  flex-shrink: 0;
}

.feature-info {
  flex: 1;
  min-width: 0;
}

.feature-name {
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 4px;
}

.feature-desc {
  font-size: 12px;
  color: #999;
  line-height: 1.5;
}

.demo-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.demo-label {
  font-size: 14px;
  color: #606266;
  white-space: nowrap;
}

.time-display {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.time-label {
  font-size: 14px;
  color: #666;
}

.time-value {
  font-size: 20px;
  font-weight: 600;
  font-family: 'Courier New', monospace;
  color: #409EFF;
}

.log-viewer {
  max-height: 400px;
  overflow: auto;
  background: #1e1e1e;
  border-radius: 8px;
  border: 1px solid #333;
}

.log-pre {
  margin: 0;
  padding: 16px;
  font-family: 'Cascadia Code', 'Fira Code', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  color: #d4d4d4;
  white-space: pre-wrap;
  word-break: break-all;
}

.notify-log {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 12px;
  max-height: 200px;
  overflow-y: auto;
}

.notify-item {
  padding: 6px 0;
  font-size: 13px;
  border-bottom: 1px solid #ebeef5;
}

.notify-item:last-child {
  border-bottom: none;
}

.notify-time {
  color: #909399;
  margin-right: 8px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
}

.notify-title {
  font-weight: 600;
  color: #409EFF;
}

.notify-msg {
  color: #606266;
}

@media (max-width: 768px) {
  .feature-grid {
    grid-template-columns: 1fr;
  }
}
</style>
