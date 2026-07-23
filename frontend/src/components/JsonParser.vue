<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from '../composables/useI18n'
import JsonTreeNode from './JsonTreeNode.vue'

const { t } = useI18n()

const input = ref('')
const parsedData = ref(null)
const error = ref('')
const viewMode = ref('tree')
const expandedKeys = ref(new Set())
const searchKeyword = ref('')
const selectedPath = ref('')
const selectedValue = ref(null)

// 解析 JSON
const parseJson = () => {
  error.value = ''
  parsedData.value = null
  selectedPath.value = ''
  selectedValue.value = null

  if (!input.value.trim()) {
    return
  }

  try {
    parsedData.value = JSON.parse(input.value)
    expandFirstLevel()
  } catch (e) {
    error.value = t('jsonparser.msg.parse_error') + e.message
  }
}

// 自动展开第一层
const expandFirstLevel = () => {
  expandedKeys.value.clear()
  if (parsedData.value && typeof parsedData.value === 'object') {
    expandedKeys.value.add('$')
    if (Array.isArray(parsedData.value)) {
      parsedData.value.forEach((item, index) => {
        if (typeof item === 'object' && item !== null) {
          expandedKeys.value.add(`$[${index}]`)
        }
      })
    } else {
      Object.keys(parsedData.value).forEach(key => {
        const val = parsedData.value[key]
        if (typeof val === 'object' && val !== null) {
          expandedKeys.value.add(`$.${key}`)
        }
      })
    }
  }
}

// 切换展开/折叠
const toggleExpand = (key) => {
  if (expandedKeys.value.has(key)) {
    expandedKeys.value.delete(key)
  } else {
    expandedKeys.value.add(key)
  }
  // 强制响应式更新
  expandedKeys.value = new Set(expandedKeys.value)
}

// 选中节点
const onSelectNode = (path) => {
  selectedPath.value = path
  selectedValue.value = resolvePath(parsedData.value, path)
}

// 根据路径字符串解析值
const resolvePath = (data, path) => {
  if (path === '$') return data
  const segments = path
    .replace(/^\$\./, '')
    .replace(/\[(\d+)\]/g, '.$1')
    .replace(/\['(.+?)'\]/g, '.$1')
    .split('.')
  let current = data
  for (const seg of segments) {
    if (current == null) return null
    if (Array.isArray(current)) {
      current = current[parseInt(seg)]
    } else {
      current = current[seg]
    }
  }
  return current
}

// 获取值类型
const getValueType = (value) => {
  if (value === null) return 'null'
  if (Array.isArray(value)) return 'array'
  return typeof value
}

// 格式化值显示
const formatValue = (value) => {
  if (value === null) return 'null'
  if (typeof value === 'object') return JSON.stringify(value, null, 2)
  if (typeof value === 'string') return value
  return String(value)
}

// 复��到剪贴板
const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(t('jsonparser.msg.copied'))
  } catch (e) {
    ElMessage.error(t('jsonparser.msg.copy_failed'))
  }
}

// 复制整个 JSON
const copyJson = () => {
  if (parsedData.value) {
    copyToClipboard(JSON.stringify(parsedData.value, null, 2))
  }
}

// 复制选中路径
const copyPath = () => {
  if (selectedPath.value) {
    copyToClipboard(selectedPath.value)
  }
}

// 复制选中值
const copySelectedValue = () => {
  if (selectedValue.value !== null) {
    const text = typeof selectedValue.value === 'object'
      ? JSON.stringify(selectedValue.value, null, 2)
      : String(selectedValue.value)
    copyToClipboard(text)
  }
}

// 清空
const clearAll = () => {
  input.value = ''
  parsedData.value = null
  error.value = ''
  expandedKeys.value.clear()
  selectedPath.value = ''
  selectedValue.value = null
}

// 统计信息
const stats = computed(() => {
  if (!parsedData.value) return null

  const count = { objects: 0, arrays: 0, strings: 0, numbers: 0, booleans: 0, nulls: 0 }

  const traverse = (value) => {
    if (value === null) { count.nulls++; return }
    if (Array.isArray(value)) { count.arrays++; value.forEach(traverse); return }
    if (typeof value === 'object') { count.objects++; Object.values(value).forEach(traverse); return }
    switch (typeof value) {
      case 'string': count.strings++; break
      case 'number': count.numbers++; break
      case 'boolean': count.booleans++; break
    }
  }

  traverse(parsedData.value)
  return count
})

// 监听输入��化
watch(input, () => {
  if (input.value.trim()) {
    parseJson()
  } else {
    parsedData.value = null
    error.value = ''
    selectedPath.value = ''
    selectedValue.value = null
  }
})

// 监听从 WebDebug 发送过来的 JSON 数据
const handleCopyToJson = (event) => {
  if (event.detail) {
    input.value = event.detail
    ElMessage.success(t('jsonparser.msg.received'))
  }
}

onMounted(() => {
  window.addEventListener('copy-to-json', handleCopyToJson)
})

onUnmounted(() => {
  window.removeEventListener('copy-to-json', handleCopyToJson)
})
</script>

<template>
  <div class="json-parser-wrap">
    <div class="json-parser">
      <div class="toolbar">
      <div class="toolbar-left">
        <span class="toolbar-title">{{ t('jsonparser.title') }}</span>
        <template v-if="stats">
          <el-tag type="info" size="small">{{ t('jsonparser.stats.objects') }}{{ stats.objects }}</el-tag>
          <el-tag type="info" size="small">{{ t('jsonparser.stats.arrays') }}{{ stats.arrays }}</el-tag>
          <el-tag type="success" size="small">{{ t('jsonparser.stats.strings') }}{{ stats.strings }}</el-tag>
          <el-tag type="warning" size="small">{{ t('jsonparser.stats.numbers') }}{{ stats.numbers }}</el-tag>
        </template>
      </div>
      <div class="toolbar-right">
        <el-radio-group v-model="viewMode" size="small">
          <el-radio-button label="tree">{{ t('jsonparser.view.tree') }}</el-radio-button>
          <el-radio-button label="table">{{ t('jsonparser.view.table') }}</el-radio-button>
          <el-radio-button label="raw">{{ t('jsonparser.view.raw') }}</el-radio-button>
        </el-radio-group>
        <el-button type="primary" @click="copyJson" :disabled="!parsedData">
          {{ t('jsonparser.btn.copy') }}
        </el-button>
        <el-button @click="clearAll">{{ t('jsonparser.btn.clear') }}</el-button>
      </div>
    </div>

    <div class="parser-content">
      <div class="input-section">
        <div class="section-header">{{ t('jsonparser.section.input') }}</div>
        <el-input
          v-model="input"
          type="textarea"
          :rows="20"
          :placeholder="t('jsonparser.input_placeholder')"
        />
      </div>

      <div class="output-section">
        <div class="section-header">{{ t('jsonparser.section.result') }}</div>

        <!-- 树形视图 -->
        <div v-if="viewMode === 'tree' && parsedData" class="tree-view">
          <JsonTreeNode
            :value="parsedData"
            path="$"
            :depth="0"
            :expandedKeys="expandedKeys"
            :selectedPath="selectedPath"
            @toggle-expand="toggleExpand"
            @select-node="onSelectNode"
          />
        </div>

        <!-- ���格视图 -->
        <div v-if="viewMode === 'table' && parsedData" class="table-view">
          <el-table
            :data="Array.isArray(parsedData)
              ? parsedData.map((item, i) => ({ _key: i, _value: item }))
              : Object.entries(parsedData).map(([k, v]) => ({ _key: k, _value: v }))"
            border stripe max-height="500"
          >
            <el-table-column prop="_key" :label="Array.isArray(parsedData) ? t('jsonparser.table.column.index') : t('jsonparser.table.column.key')" width="180" />
            <el-table-column prop="_value" :label="t('jsonparser.table.column.value')" min-width="200" />
          </el-table>
        </div>

        <!-- 原始视图 -->
        <div v-if="viewMode === 'raw' && parsedData" class="raw-view">
          <el-input
            :model-value="JSON.stringify(parsedData, null, 2)"
            type="textarea"
            :rows="20"
            readonly
            :placeholder="t('jsonparser.raw_placeholder')"
          />
        </div>

        <el-empty v-if="!parsedData && !error" :description="t('jsonparser.empty')" />
      </div>
    </div>

    <!-- 键值选择面板 -->
    <div v-if="selectedPath" class="key-value-panel">
      <div class="kv-box">
        <div class="kv-label">{{ t('jsonparser.panel.path') }}</div>
        <div class="kv-content mono">
          {{ selectedPath }}
          <el-button size="small" text class="kv-copy-btn" @click="copyPath">{{ t('jsonparser.btn.copy') }}</el-button>
        </div>
      </div>
      <div class="kv-box">
        <div class="kv-label">{{ t('jsonparser.panel.value') }}</div>
        <div class="kv-content mono">
          {{ formatValue(selectedValue) }}
          <el-button size="small" text class="kv-copy-btn" @click="copySelectedValue">{{ t('jsonparser.btn.copy') }}</el-button>
        </div>
      </div>
    </div>

      <div v-if="error" class="error-message">
        <el-alert :title="error" type="error" show-icon :closable="false" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.json-parser-wrap {
  position: absolute;
  top: 16px;
  right: 16px;
  bottom: 16px;
  left: 16px;
  overflow: hidden;
}

.json-parser {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 8px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background-color: #f8f9fa;
  border-radius: 4px;
  flex-shrink: 0;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.toolbar-title {
  font-weight: 600;
  color: #303133;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.parser-content {
  flex: 1;
  display: flex;
  gap: 12px;
  overflow: hidden;
  min-height: 0;
}

.input-section,
.output-section {
  flex: 1;
  min-width: 0;
  width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.section-header {
  font-weight: 500;
  color: #606266;
  padding: 0 4px;
  flex-shrink: 0;
}

/* 输入区滚动 */
.input-section :deep(.el-textarea) {
  flex: 1;
  min-height: 0;
}

.input-section :deep(.el-textarea__inner) {
  height: 100% !important;
  resize: none;
  white-space: pre-wrap;
  word-break: break-all;
  overflow-wrap: break-word;
}

/* 树形视图 */
.tree-view {
  flex: 1;
  overflow: auto;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 8px 10px;
  background: #fafafa;
  min-height: 0;
}

/* 表格视图 */
.table-view {
  flex: 1;
  overflow: auto;
  min-height: 0;
}

/* 原始视图 */
.raw-view {
  flex: 1;
  min-height: 0;
}

.raw-view :deep(.el-textarea) {
  height: 100%;
}

.raw-view :deep(.el-textarea__inner) {
  height: 100% !important;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  white-space: pre-wrap;
  word-break: break-all;
}

/* 键值选择面板 */
.key-value-panel {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
  min-height: 0;
}

.kv-box {
  flex: 1;
  min-width: 0;
  width: 0;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  background: #fafafa;
  display: flex;
  flex-direction: column;
}

.kv-label {
  font-size: 12px;
  color: #909399;
  padding: 4px 10px;
  border-bottom: 1px solid #ebeef5;
  flex-shrink: 0;
  background: #f5f7fa;
}

.kv-content {
  flex: 1;
  padding: 6px 10px;
  overflow: auto;
  position: relative;
  font-size: 13px;
  color: #303133;
}

.kv-content:hover .kv-copy-btn {
  opacity: 1;
}

.kv-copy-btn {
  position: absolute;
  top: 4px;
  right: 6px;
  opacity: 0;
  transition: opacity 0.15s;
}

.mono {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  white-space: pre-wrap;
  word-break: break-all;
}

.error-message {
  flex-shrink: 0;
}
</style>
