<script setup>
import { ref } from 'vue'
import { useI18n } from '../composables/useI18n'
import { ExecuteJs } from '../../wailsjs/go/service/App'

const { t } = useI18n()

const code = ref(`// 在这里输入 JavaScript 代码
function hello() {
  console.log('Hello, WebHelper!')
  return '执行完成'
}

hello()`)

const output = ref('')
const result = ref('')
const duration = ref('')
const success = ref(null)
const running = ref(false)

const runCode = async () => {
  if (!code.value.trim()) {
    ElMessage.warning(t('jsdebug.msg.code_required'))
    return
  }

  running.value = true
  output.value = ''
  result.value = ''
  duration.value = ''
  success.value = null

  try {
    const res = await ExecuteJs(code.value)
    output.value = res.output || ''
    result.value = res.result || ''
    duration.value = res.duration || ''
    success.value = res.success

    if (!res.success) {
      ElMessage.error(t('jsdebug.msg.execute_failed') + res.error)
    }
  } catch (e) {
    success.value = false
    ElMessage.error(t('jsdebug.msg.execute_failed') + e)
  } finally {
    running.value = false
  }
}

const clearAll = () => {
  code.value = ''
  output.value = ''
  result.value = ''
  duration.value = ''
  success.value = null
}

const copyResult = async () => {
  const text = output.value + (result.value ? '\n' + t('jsdebug.msg.copy_prefix') + result.value : '')
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(t('jsdebug.msg.copied'))
  } catch (e) {
    ElMessage.error(t('jsdebug.msg.copy_failed'))
  }
}
</script>

<template>
  <div class="js-debug">
    <div class="toolbar">
      <div class="toolbar-left">
        <span class="toolbar-title">{{ t('jsdebug.title') }}</span>
        <el-tag v-if="duration" type="info" size="small">{{ t('jsdebug.tag.duration') }}{{ duration }}</el-tag>
        <el-tag v-if="success === true" type="success" size="small">{{ t('jsdebug.tag.success') }}</el-tag>
        <el-tag v-if="success === false" type="danger" size="small">{{ t('jsdebug.tag.failed') }}</el-tag>
      </div>
      <div class="toolbar-right">
        <el-button type="primary" :loading="running" @click="runCode">{{ t('jsdebug.btn.run') }}</el-button>
        <el-button @click="clearAll">{{ t('jsdebug.btn.clear') }}</el-button>
      </div>
    </div>

    <div class="code-section">
      <div class="section-label">{{ t('jsdebug.section.code') }}</div>
      <el-input
        v-model="code"
        type="textarea"
        :rows="16"
        :placeholder="t('jsdebug.code_placeholder')"
        class="code-editor"
      />
    </div>

    <div class="output-section">
      <div class="section-label">
        <span>{{ t('jsdebug.section.output') }}</span>
        <el-button size="small" @click="copyResult" :disabled="!output && !result">{{ t('jsdebug.btn.copy') }}</el-button>
      </div>
      <el-input
        :value="output"
        type="textarea"
        :rows="8"
        readonly
        :placeholder="t('jsdebug.output_placeholder')"
      />
    </div>

    <div class="result-section">
      <div class="section-label">{{ t('jsdebug.section.return_value') }}</div>
      <el-input
        :value="result"
        type="textarea"
        :rows="3"
        readonly
        :placeholder="t('jsdebug.return_placeholder')"
      />
    </div>
  </div>
</template>

<style scoped>
.js-debug {
  display: flex;
  flex-direction: column;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  background-color: #f8f9fa;
  border-radius: 4px;
  margin-bottom: 16px;
  flex-shrink: 0;
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
  gap: 8px;
}

.code-section {
  margin-bottom: 16px;
  flex-shrink: 0;
}

.section-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 500;
  color: #606266;
  font-size: 14px;
  margin-bottom: 8px;
}

.output-section {
  margin-bottom: 16px;
  flex-shrink: 0;
}

.result-section {
  flex-shrink: 0;
}

.code-editor :deep(textarea) {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  background-color: #1e1e1e;
  color: #d4d4d4;
}
</style>
