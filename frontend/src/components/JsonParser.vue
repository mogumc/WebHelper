<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from '../composables/useI18n'

const { t } = useI18n()

const input = ref('')
const parsedData = ref(null)
const error = ref('')
const viewMode = ref('tree') // tree, table, raw
const expandedKeys = ref(new Set())
const searchKeyword = ref('')

// 解析 JSON
const parseJson = () => {
  error.value = ''
  parsedData.value = null
  
  if (!input.value.trim()) {
    return
  }

  try {
    parsedData.value = JSON.parse(input.value)
    // 自动展开第一层
    expandFirstLevel()
  } catch (e) {
    error.value = t('jsonparser.msg.parse_error') + e.message
  }
}

// 自动展开第一层
const expandFirstLevel = () => {
  expandedKeys.value.clear()
  if (parsedData.value && typeof parsedData.value === 'object') {
    if (Array.isArray(parsedData.value)) {
      parsedData.value.forEach((_, index) => {
        expandedKeys.value.add(`root-${index}`)
      })
    } else {
      Object.keys(parsedData.value).forEach(key => {
        expandedKeys.value.add(`root-${key}`)
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
}

// 检查是否展开
const isExpanded = (key) => {
  return expandedKeys.value.has(key)
}

// 获取值类型
const getValueType = (value) => {
  if (value === null) return 'null'
  if (Array.isArray(value)) return 'array'
  return typeof value
}

// 获取类型颜色
const getTypeColor = (type) => {
  const colors = {
    'string': '#67c23a',
    'number': '#e6a23c',
    'boolean': '#409eff',
    'null': '#909399',
    'object': '#f56c6c',
    'array': '#9b59b6'
  }
  return colors[type] || '#909399'
}

// 格式化值显示
const formatValue = (value) => {
  if (value === null) return 'null'
  if (typeof value === 'string') return `"${value}"`
  return String(value)
}

// 复制到剪贴板
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

// 复制单个值
const copyValue = (value) => {
  const text = typeof value === 'object' ? JSON.stringify(value, null, 2) : String(value)
  copyToClipboard(text)
}

// 清空
const clearAll = () => {
  input.value = ''
  parsedData.value = null
  error.value = ''
  expandedKeys.value.clear()
}

// 统计信息
const stats = computed(() => {
  if (!parsedData.value) return null
  
  const count = {
    objects: 0,
    arrays: 0,
    strings: 0,
    numbers: 0,
    booleans: 0,
    nulls: 0
  }

  const traverse = (value) => {
    if (value === null) {
      count.nulls++
      return
    }
    if (Array.isArray(value)) {
      count.arrays++
      value.forEach(traverse)
      return
    }
    if (typeof value === 'object') {
      count.objects++
      Object.values(value).forEach(traverse)
      return
    }
    switch (typeof value) {
      case 'string': count.strings++; break
      case 'number': count.numbers++; break
      case 'boolean': count.booleans++; break
    }
  }

  traverse(parsedData.value)
  return count
})

// 监听输入变化
watch(input, () => {
  if (input.value.trim()) {
    parseJson()
  } else {
    parsedData.value = null
    error.value = ''
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
        
        <div v-if="viewMode === 'tree' && parsedData" class="tree-view">
          <div class="tree-node root">
            <div class="node-header" @click="toggleExpand('root')">
              <span class="expand-icon">{{ isExpanded('root') ? '▼' : '▶' }}</span>
              <span class="node-type" :style="{ color: getTypeColor(getValueType(parsedData)) }">
                {{ getValueType(parsedData) }}
              </span>
              <span class="node-info" v-if="Array.isArray(parsedData)">
                Array[{{ parsedData.length }}]
              </span>
            </div>
            <div class="node-children" v-if="isExpanded('root')">
              <template v-if="Array.isArray(parsedData)">
                <div v-for="(item, index) in parsedData" :key="index" class="tree-node">
                  <div class="node-header" @click="toggleExpand(`root-${index}`)">
                    <span class="expand-icon">{{ isExpanded(`root-${index}`) ? '▼' : '▶' }}</span>
                    <span class="node-index">[{{ index }}]</span>
                    <span class="node-type" :style="{ color: getTypeColor(getValueType(item)) }">
                      {{ getValueType(item) }}
                    </span>
                    <span class="node-preview" v-if="typeof item !== 'object' || item === null">
                      {{ formatValue(item) }}
                    </span>
                    <span class="node-preview" v-else-if="Array.isArray(item)">
                      Array[{{ item.length }}]
                    </span>
                    <span class="node-preview" v-else>
                      Object{ {{ Object.keys(item).length }} }
                    </span>
                    <el-button size="small" text @click.stop="copyValue(item)">{{ t('jsonparser.btn.copy') }}</el-button>
                  </div>
                  <div class="node-children" v-if="isExpanded(`root-${index}`) && typeof item === 'object' && item !== null">
                    <template v-if="Array.isArray(item)">
                      <div v-for="(subItem, subIndex) in item" :key="subIndex" class="tree-node leaf">
                        <div class="node-header">
                          <span class="node-index">[{{ subIndex }}]</span>
                          <span class="node-type" :style="{ color: getTypeColor(getValueType(subItem)) }">
                            {{ getValueType(subItem) }}
                          </span>
                          <span class="node-value">{{ formatValue(subItem) }}</span>
                        </div>
                      </div>
                    </template>
                    <template v-else>
                      <div v-for="(subValue, subKey) in item" :key="subKey" class="tree-node leaf">
                        <div class="node-header">
                          <span class="node-key">{{ subKey }}:</span>
                          <span class="node-type" :style="{ color: getTypeColor(getValueType(subValue)) }">
                            {{ getValueType(subValue) }}
                          </span>
                          <span class="node-value">{{ formatValue(subValue) }}</span>
                        </div>
                      </div>
                    </template>
                  </div>
                </div>
              </template>
              <template v-else>
                <div v-for="(value, key) in parsedData" :key="key" class="tree-node">
                  <div class="node-header" @click="toggleExpand(`root-${key}`)">
                    <span class="expand-icon">{{ isExpanded(`root-${key}`) ? '▼' : '▶' }}</span>
                    <span class="node-key">{{ key }}:</span>
                    <span class="node-type" :style="{ color: getTypeColor(getValueType(value)) }">
                      {{ getValueType(value) }}
                    </span>
                    <span class="node-preview" v-if="typeof value !== 'object' || value === null">
                      {{ formatValue(value) }}
                    </span>
                    <span class="node-preview" v-else-if="Array.isArray(value)">
                      Array[{{ value.length }}]
                    </span>
                    <span class="node-preview" v-else>
                      Object{ {{ Object.keys(value).length }} }
                    </span>
                    <el-button size="small" text @click.stop="copyValue(value)">{{ t('jsonparser.btn.copy') }}</el-button>
                  </div>
                  <div class="node-children" v-if="isExpanded(`root-${key}`) && typeof value === 'object' && value !== null">
                    <template v-if="Array.isArray(value)">
                      <div v-for="(subItem, subIndex) in value" :key="subIndex" class="tree-node leaf">
                        <div class="node-header">
                          <span class="node-index">[{{ subIndex }}]</span>
                          <span class="node-type" :style="{ color: getTypeColor(getValueType(subItem)) }">
                            {{ getValueType(subItem) }}
                          </span>
                          <span class="node-value">{{ formatValue(subItem) }}</span>
                        </div>
                      </div>
                    </template>
                    <template v-else>
                      <div v-for="(subValue, subKey) in value" :key="subKey" class="tree-node leaf">
                        <div class="node-header">
                          <span class="node-key">{{ subKey }}:</span>
                          <span class="node-type" :style="{ color: getTypeColor(getValueType(subValue)) }">
                            {{ getValueType(subValue) }}
                          </span>
                          <span class="node-value">{{ formatValue(subValue) }}</span>
                        </div>
                      </div>
                    </template>
                  </div>
                </div>
              </template>
            </div>
          </div>
        </div>

        <div v-if="viewMode === 'table' && parsedData" class="table-view">
          <el-table :data="Array.isArray(parsedData) ? parsedData.map((item, index) => ({ _index: index, _value: item })) : Object.entries(parsedData).map(([key, value]) => ({ _key: key, _value: value }))" border stripe max-height="500">
            <el-table-column v-if="!Array.isArray(parsedData)" prop="_key" :label="t('jsonparser.table.column.key')" width="200" />
            <el-table-column v-else prop="_index" :label="t('jsonparser.table.column.index')" width="100" />
            <el-table-column prop="_value" :label="t('jsonparser.table.column.value')">
              <template #default="{ row }">
                <div class="table-value">
                  <el-tag :color="getTypeColor(getValueType(row._value))" effect="dark" size="small" style="margin-right: 8px">
                    {{ getValueType(row._value) }}
                  </el-tag>
                  <span>{{ typeof row._value === 'object' ? JSON.stringify(row._value) : formatValue(row._value) }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column :label="t('jsonparser.table.column.action')" width="80">
              <template #default="{ row }">
                <el-button size="small" text @click="copyValue(row._value)">{{ t('jsonparser.btn.copy') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div v-if="viewMode === 'raw'" class="raw-view">
          <el-input
            :value="parsedData ? JSON.stringify(parsedData, null, 2) : ''"
            type="textarea"
            :rows="20"
            readonly
            :placeholder="t('jsonparser.raw_placeholder')"
          />
        </div>

        <el-empty v-if="!parsedData && !error && viewMode !== 'raw'" :description="t('jsonparser.empty')" />
      </div>
    </div>

    <div v-if="error" class="error-message">
      <el-alert :title="error" type="error" show-icon :closable="false" />
    </div>
  </div>
</template>

<style scoped>
.json-parser {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background-color: #f8f9fa;
  border-radius: 4px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
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
  gap: 16px;
  overflow: hidden;
}

.input-section,
.output-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}

.section-header {
  font-weight: 500;
  color: #606266;
  padding: 0 4px;
}

/* 树形视图样式 */
.tree-view {
  flex: 1;
  overflow: auto;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 12px;
  background: #fafafa;
}

.tree-node {
  margin-left: 20px;
}

.tree-node.root {
  margin-left: 0;
}

.tree-node.leaf {
  margin-left: 20px;
}

.node-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.node-header:hover {
  background-color: #e9ecef;
}

.expand-icon {
  width: 16px;
  font-size: 12px;
  color: #909399;
}

.node-type {
  font-size: 12px;
  font-weight: 500;
  padding: 2px 6px;
  background-color: rgba(0, 0, 0, 0.05);
  border-radius: 4px;
}

.node-key {
  font-weight: 500;
  color: #303133;
}

.node-index {
  color: #909399;
}

.node-preview {
  color: #606266;
  font-size: 13px;
}

.node-value {
  color: #606266;
  word-break: break-all;
}

.node-children {
  margin-left: 20px;
  border-left: 1px dashed #dcdfe6;
  padding-left: 8px;
}

/* 表格视图样式 */
.table-view {
  flex: 1;
}

.table-value {
  display: flex;
  align-items: center;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

/* 原始视图样式 */
.raw-view {
  flex: 1;
}

.raw-view :deep(textarea) {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.error-message {
  margin-top: 8px;
}
</style>
