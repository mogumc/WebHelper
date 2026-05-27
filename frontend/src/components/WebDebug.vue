<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from '../composables/useI18n'
import { 
  SendHttpRequest, 
  OpenFileSelect, 
  GetRequestLogs, 
  DeleteRequestLog, 
  ClearRequestLogs,
  SearchRequestLogs
} from '../../wailsjs/go/service/App'
import { Delete, Upload, Notebook, Search } from '@element-plus/icons-vue'

const { t } = useI18n()

// 请求相关状态
const url = ref('')
const response = ref({})
const loading = ref(false)
const method = ref('GET')
const headers = ref([{ key: '', value: '' }])
const cookies = ref('')
const bodyType = ref('text')
const body = ref('')
const filePath = ref('')
const contentType = ref('')
const proxy = ref('')
const timeout = ref(30)
const insecure = ref(false)
const saveLog = ref(true)

// 响应展示相关
const responseTab = ref('body')

// 日志相关状态
const logs = ref([])
const selectedLog = ref(null)
const logSearchKeyword = ref('')
const showLogDialog = ref(false)

const methods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']

// 获取方法颜色
const getMethodColor = (method) => {
  const colors = {
    'GET': '#67c23a',
    'POST': '#409eff',
    'PUT': '#e6a23c',
    'DELETE': '#f56c6c',
    'PATCH': '#909399',
    'HEAD': '#909399',
    'OPTIONS': '#909399'
  }
  return colors[method] || '#909399'
}

// 获取状态码类型
const getStatusType = (code) => {
  if (code >= 200 && code < 300) return 'success'
  if (code >= 300 && code < 400) return 'warning'
  if (code >= 400) return 'danger'
  return 'info'
}

// 检查是否是 JSON
const isJson = computed(() => {
  if (!response.value || !response.value.body) return false
  try {
    JSON.parse(response.value.body)
    return true
  } catch {
    return false
  }
})

// 检查是否有响应数据
const hasResponse = computed(() => {
  return response.value && response.value.statusCode
})

// 格式化响应头为文本
const formatResponseHeaders = computed(() => {
  if (!response.value || !response.value.headers) return ''
  return Object.entries(response.value.headers)
    .map(([key, value]) => `${key}: ${value}`)
    .join('\n')
})

// 格式化内容大小
const formatContentSize = computed(() => {
  if (!response.value) return '-'
  const size = response.value.size || response.value.contentLength
  if (!size || size < 0) return '-'
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(2)} KB`
  return `${(size / (1024 * 1024)).toFixed(2)} MB`
})

// 添加请求头
const addHeader = () => {
  headers.value.push({ key: '', value: '' })
}

// 删除请求头
const removeHeader = (index) => {
  headers.value.splice(index, 1)
}

// 格式化请求头
const formatHeaders = computed(() => {
  return headers.value
    .filter(h => h.key && h.value)
    .map(h => `${h.key}: ${h.value}`)
    .join('\n')
})

// 选择文件
const selectFile = async () => {
  const path = await OpenFileSelect()
  if (path) {
    filePath.value = path
    bodyType.value = 'file'
  }
}

// 发送请求
const sendRequest = async () => {
  if (!url.value) {
    ElMessage.warning(t('webdebug.msg.url_required'))
    return
  }

  loading.value = true
  response.value = {}

  try {
    const result = await SendHttpRequest(
      method.value,
      url.value,
      formatHeaders.value,
      cookies.value,
      proxy.value,
      bodyType.value,
      body.value,
      filePath.value,
      contentType.value,
      timeout.value,
      insecure.value,
      saveLog.value
    )
    response.value = result || {}
    responseTab.value = 'body'
  } catch (e) {
    ElMessage.error(t('webdebug.msg.request_failed') + e)
    response.value = { error: e.toString() }
  } finally {
    loading.value = false
  }
}

// 加载日志列表
const loadLogs = async () => {
  try {
    if (logSearchKeyword.value) {
      logs.value = await SearchRequestLogs(logSearchKeyword.value) || []
    } else {
      logs.value = await GetRequestLogs() || []
    }
  } catch (e) {
    ElMessage.error(t('webdebug.msg.load_logs_failed') + e)
  }
}

// 打开日志窗口
const openLogDialog = async () => {
  await loadLogs()
  showLogDialog.value = true
}

// 选择日志
const selectLog = (log) => {
  selectedLog.value = log
}

// 双击日志 - 添加到调试页面
const applyLog = (log) => {
  method.value = log.method
  url.value = log.url
  
  // 恢复请求头
  if (log.headers) {
    headers.value = Object.entries(log.headers).map(([key, value]) => ({ key, value }))
  } else {
    headers.value = [{ key: '', value: '' }]
  }
  
  // 恢复 Cookie
  if (log.cookies) {
    cookies.value = Object.entries(log.cookies).map(([key, value]) => `${key}=${value}`).join('; ')
  }
  
  // 恢复请求体
  if (log.body) {
    body.value = log.body
    bodyType.value = log.bodyType || 'text'
  }
  
  showLogDialog.value = false
}

// 删除日志
const deleteLog = async (id) => {
  try {
    await DeleteRequestLog(id)
    await loadLogs()
    if (selectedLog.value && selectedLog.value.id === id) {
      selectedLog.value = null
    }
  } catch (e) {
    ElMessage.error(t('webdebug.msg.delete_failed') + e)
  }
}

// 清空日志
const clearLogs = async () => {
  try {
    await ClearRequestLogs()
    logs.value = []
    selectedLog.value = null
  } catch (e) {
    ElMessage.error(t('webdebug.msg.clear_failed') + e)
  }
}

// 搜索日志
const searchLogs = async () => {
  await loadLogs()
}

// 复制到 JSON 解析器
const copyToJsonParser = () => {
  if (response.value && response.value.body) {
    const event = new CustomEvent('copy-to-json', { detail: response.value.body })
    window.dispatchEvent(event)
  }
}
</script>

<template>
  <div class="web-debug">
    <div class="request-section">
      <div class="url-bar">
        <el-select v-model="method" class="method-select">
          <el-option v-for="m in methods" :key="m" :label="m" :value="m" />
        </el-select>
        <el-input v-model="url" :placeholder="t('webdebug.url_placeholder')" class="url-input" />
        <el-button type="primary" :loading="loading" @click="sendRequest">
          {{ t('webdebug.btn_send') }}
        </el-button>
      </div>

      <el-tabs type="border-card" class="request-tabs">
        <el-tab-pane :label="t('webdebug.tab.headers')">
          <div class="headers-section">
            <div v-for="(header, index) in headers" :key="index" class="header-row">
              <el-input v-model="header.key" :placeholder="t('webdebug.header.key_placeholder')" />
              <el-input v-model="header.value" :placeholder="t('webdebug.header.value_placeholder')" />
              <el-button type="danger" :icon="Delete" circle size="small" @click="removeHeader(index)" />
            </div>
            <el-button type="primary" plain @click="addHeader">{{ t('webdebug.header.add') }}</el-button>
          </div>
        </el-tab-pane>
        <el-tab-pane :label="t('webdebug.tab.body')">
          <div class="body-section">
            <el-radio-group v-model="bodyType" class="body-type-group">
              <el-radio-button label="text">{{ t('webdebug.body.type_text') }}</el-radio-button>
              <el-radio-button label="json">{{ t('webdebug.body.type_json') }}</el-radio-button>
              <el-radio-button label="form">{{ t('webdebug.body.type_form') }}</el-radio-button>
              <el-radio-button label="file">{{ t('webdebug.body.type_file') }}</el-radio-button>
            </el-radio-group>
            
            <template v-if="bodyType === 'file'">
              <div class="file-section">
                <el-input v-model="filePath" :placeholder="t('webdebug.body.file_placeholder')" readonly>
                  <template #append>
                    <el-button :icon="Upload" @click="selectFile">{{ t('webdebug.body.file_btn') }}</el-button>
                  </template>
                </el-input>
              </div>
            </template>
            <template v-else>
              <el-input v-model="body" type="textarea" :rows="8" :placeholder="bodyType === 'json' ? t('webdebug.body.json_placeholder') : t('webdebug.body.text_placeholder')" />
            </template>
          </div>
        </el-tab-pane>
        <el-tab-pane :label="t('webdebug.tab.cookie')">
          <el-input v-model="cookies" type="textarea" :rows="5" :placeholder="t('webdebug.body.cookie_placeholder')" />
        </el-tab-pane>
        <el-tab-pane :label="t('webdebug.tab.advanced')">
          <div class="advanced-settings">
            <el-form label-width="120px">
              <el-form-item :label="t('webdebug.advanced.proxy_label')">
                <el-input v-model="proxy" :placeholder="t('webdebug.advanced.proxy_placeholder')" />
              </el-form-item>
              <el-form-item :label="t('webdebug.advanced.timeout_label')">
                <el-input-number v-model="timeout" :min="1" :max="300" /> {{ t('webdebug.advanced.timeout_unit') }}
              </el-form-item>
              <el-form-item :label="t('webdebug.advanced.content_type_label')">
                <el-input v-model="contentType" :placeholder="t('webdebug.advanced.content_type_placeholder')" />
              </el-form-item>
              <el-form-item :label="t('webdebug.advanced.insecure_label')">
                <el-switch v-model="insecure" />
              </el-form-item>
              <el-divider content-position="left">{{ t('webdebug.advanced.log_divider') }}</el-divider>
              <el-form-item :label="t('webdebug.advanced.save_log_label')">
                <el-switch v-model="saveLog" />
              </el-form-item>
              <el-form-item :label="t('webdebug.advanced.log_manage_label')">
                <el-button type="primary" :icon="Notebook" @click="openLogDialog">
                  {{ t('webdebug.advanced.view_logs_btn') }}
                </el-button>
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <div class="response-section">
      <div class="response-header">
        <div class="response-title">
          <span>{{ t('webdebug.response.title') }}</span>
          <div class="response-info" v-if="hasResponse">
            <el-tag v-if="response.statusCode" :type="getStatusType(response.statusCode)">
              {{ response.statusCode }} {{ response.status }}
            </el-tag>
            <el-tag v-if="response.time" type="info">
              {{ response.time }}
            </el-tag>
            <el-tag v-if="response.size" type="info">
              {{ (response.size / 1024).toFixed(2) }} KB
            </el-tag>
          </div>
        </div>
        <el-radio-group v-model="responseTab" size="small">
          <el-radio-button label="body">{{ t('webdebug.response.tab.body') }}</el-radio-button>
          <el-radio-button label="headers">{{ t('webdebug.response.tab.headers') }}</el-radio-button>
          <el-radio-button label="more">{{ t('webdebug.response.tab.more') }}</el-radio-button>
        </el-radio-group>
      </div>
      
      <div class="response-content">
        <div v-if="responseTab === 'body'" class="response-body-container">
          <el-input 
            :value="response.body || ''" 
            type="textarea" 
            :rows="18" 
            readonly 
            :placeholder="t('webdebug.response.body_placeholder')"
          />
          <div class="response-actions" v-if="hasResponse && response.body && isJson">
            <el-button size="small" @click="copyToJsonParser">
              {{ t('webdebug.response.send_to_json') }}
            </el-button>
          </div>
        </div>
        
        <div v-if="responseTab === 'headers'" class="response-headers-container">
          <el-input 
            :value="formatResponseHeaders" 
            type="textarea" 
            :rows="18" 
            readonly 
            :placeholder="t('webdebug.response.headers_placeholder')"
          />
        </div>
        
        <div v-if="responseTab === 'more'" class="more-info">
          <el-descriptions :column="2" border>
            <el-descriptions-item :label="t('webdebug.response.status_code_label')">{{ response.statusCode || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('webdebug.response.status_label')">{{ response.status || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('webdebug.response.content_type_label')">{{ response.contentType || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('webdebug.response.content_length_label')">{{ formatContentSize }}</el-descriptions-item>
            <el-descriptions-item :label="t('webdebug.response.request_time_label')">{{ response.time || '-' }}</el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
    </div>

    <el-dialog 
      v-model="showLogDialog" 
      :title="t('webdebug.log.title')" 
      width="90%"
      :style="{ maxWidth: '900px' }"
      top="5vh"
      :close-on-click-modal="false"
    >
      <div class="log-dialog-content">
        <div class="log-list-panel">
          <div class="log-list-header">
            <el-input 
              v-model="logSearchKeyword" 
              :placeholder="t('webdebug.log.search_placeholder')" 
              clearable
              @keyup.enter="searchLogs"
            >
              <template #append>
                <el-button :icon="Search" @click="searchLogs" />
              </template>
            </el-input>
            <el-button type="danger" size="small" @click="clearLogs">{{ t('webdebug.log.clear') }}</el-button>
          </div>
          
          <div class="log-list">
            <div 
              v-for="log in logs" 
              :key="log.id"
              :class="['log-item', { selected: selectedLog && selectedLog.id === log.id }]"
              @click="selectLog(log)"
              @dblclick="applyLog(log)"
            >
              <div class="log-item-header">
                <span class="log-id">#{{ log.id }}</span>
                <el-tag :color="getMethodColor(log.method)" effect="dark" size="small">
                  {{ log.method }}
                </el-tag>
              </div>
              <div class="log-item-info">
                <div class="log-host">{{ log.host }}</div>
                <div class="log-path">{{ log.path }}</div>
              </div>
              <div class="log-item-meta">
                <el-tag :type="getStatusType(log.statusCode)" size="small">
                  {{ log.statusCode }}
                </el-tag>
                <span class="log-time">{{ log.createdAt }}</span>
              </div>
              <el-button 
                type="danger" 
                size="small" 
                :icon="Delete" 
                circle 
                class="log-delete-btn"
                @click.stop="deleteLog(log.id)"
              />
            </div>
            <el-empty v-if="logs.length === 0" :description="t('webdebug.log.empty')" />
          </div>
        </div>

        <div class="log-detail-panel" v-if="selectedLog">
          <el-tabs type="border-card">
            <el-tab-pane :label="t('webdebug.log.tab.request_info')">
              <el-descriptions :column="1" border>
                <el-descriptions-item :label="t('webdebug.log.request.method_label')">{{ selectedLog.method }}</el-descriptions-item>
                <el-descriptions-item :label="t('webdebug.log.request.url_label')">{{ selectedLog.url }}</el-descriptions-item>
                <el-descriptions-item :label="t('webdebug.log.request.host_label')">{{ selectedLog.host }}</el-descriptions-item>
                <el-descriptions-item :label="t('webdebug.log.request.path_label')">{{ selectedLog.path }}</el-descriptions-item>
                <el-descriptions-item :label="t('webdebug.log.request.body_label')" v-if="selectedLog.body">
                  <el-input :value="selectedLog.body" type="textarea" :rows="5" readonly />
                </el-descriptions-item>
              </el-descriptions>
            </el-tab-pane>
            <el-tab-pane :label="t('webdebug.log.tab.response_info')">
              <el-descriptions :column="2" border>
                <el-descriptions-item :label="t('webdebug.log.response.status_code_label')">{{ selectedLog.statusCode }}</el-descriptions-item>
                <el-descriptions-item :label="t('webdebug.log.response.status_label')">{{ selectedLog.status }}</el-descriptions-item>
                <el-descriptions-item :label="t('webdebug.log.response.size_label')">{{ selectedLog.respSize ? (selectedLog.respSize / 1024).toFixed(2) + ' KB' : '-' }}</el-descriptions-item>
                <el-descriptions-item :label="t('webdebug.log.response.time_label')">{{ selectedLog.time }}</el-descriptions-item>
              </el-descriptions>
              <div class="log-resp-body">
                <div class="log-resp-header">{{ t('webdebug.log.response.body_label') }}</div>
                <el-input :value="selectedLog.respBody" type="textarea" :rows="10" readonly />
              </div>
            </el-tab-pane>
          </el-tabs>
          <div class="log-detail-actions">
            <el-button type="primary" @click="applyLog(selectedLog)">{{ t('webdebug.log.load_to_debug') }}</el-button>
          </div>
        </div>
        <div class="log-detail-panel empty" v-else>
          <el-empty :description="t('webdebug.log.select_hint')" />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.web-debug {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 100%;
}

.request-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex-shrink: 0;
}

.url-bar {
  display: flex;
  gap: 8px;
}

.method-select {
  width: 120px;
}

.url-input {
  flex: 1;
}

.request-tabs {
  flex-shrink: 0;
}

.request-tabs :deep(.el-tabs) {
  overflow: visible;
}

.request-tabs :deep(.el-tabs__header) {
  margin: 0;
  border-bottom: none;
}

.request-tabs :deep(.el-tabs__nav-wrap) {
  padding-left: 12px;
}

.request-tabs :deep(.el-tabs__item) {
  padding: 0 24px;
  height: 42px;
  line-height: 42px;
  font-size: 14px;
}

.request-tabs :deep(.el-tabs__content) {
  padding: 0;
  overflow: visible;
}

.request-tabs :deep(.el-tab-pane) {
  padding: 0;
  overflow: visible;
}

.headers-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px;
}

.header-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.header-row .el-input {
  flex: 1;
}

.body-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px;
}

.body-type-group {
  margin-bottom: 8px;
}

.file-section {
  margin-top: 8px;
}

.advanced-settings {
  padding: 16px;
}

.response-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex-shrink: 0;
}

.response-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background-color: #f8f9fa;
  border-radius: 4px;
}

.response-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-weight: 600;
  color: #303133;
}

.response-info {
  display: flex;
  gap: 8px;
}

.response-content {
  flex: 1;
}

.response-body-container,
.response-headers-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.response-actions {
  display: flex;
  justify-content: flex-end;
}

.more-info {
  padding: 16px 0;
}

/* 日志对话框样式 */
.log-dialog-content {
  display: flex;
  gap: 16px;
  height: 50vh;
  max-height: 500px;
}

.log-list-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  border: 1px solid #ebeef5;
  border-radius: 4px;
}

.log-list-header {
  display: flex;
  gap: 8px;
  padding: 12px;
  border-bottom: 1px solid #ebeef5;
}

.log-list {
  flex: 1;
  overflow-y: scroll;
  scrollbar-width: none;
  padding: 8px;
}

.log-list::-webkit-scrollbar {
  display: none;
}

.log-item {
  padding: 12px;
  margin-bottom: 8px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
}

.log-item:hover {
  border-color: #409eff;
  background-color: #f5f7fa;
}

.log-item.selected {
  border-color: #409eff;
  background-color: #ecf5ff;
}

.log-item-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.log-id {
  font-weight: 600;
  color: #606266;
}

.log-item-info {
  margin-bottom: 8px;
}

.log-host {
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.log-path {
  font-size: 12px;
  color: #909399;
  word-break: break-all;
}

.log-item-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #909399;
}

.log-time {
  margin-left: auto;
}

.log-delete-btn {
  position: absolute;
  right: 8px;
  top: 8px;
  opacity: 0;
  transition: opacity 0.2s;
}

.log-item:hover .log-delete-btn {
  opacity: 1;
}

.log-detail-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  overflow: hidden;
}

.log-detail-panel.empty {
  justify-content: center;
}

.log-resp-body {
  margin-top: 16px;
}

.log-resp-header {
  font-weight: 500;
  margin-bottom: 8px;
  color: #606266;
}

.log-detail-actions {
  padding: 12px;
  border-top: 1px solid #ebeef5;
  display: flex;
  justify-content: flex-end;
}
</style>
