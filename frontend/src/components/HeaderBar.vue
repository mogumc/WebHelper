<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../composables/useI18n'
import { Minus, FullScreen, Close } from '@element-plus/icons-vue'
import { WindowMinimise, WindowToggleMaximise, WindowClose, GetVersion } from '../../wailsjs/go/service/App'

const { t } = useI18n()
const version = ref('dev')

onMounted(async () => {
  try {
    version.value = await GetVersion()
  } catch (e) {
    // 保持默认值
  }
})

const handleDoubleClick = (e) => {
  if (e.target.closest('.right-panel')) {
    return
  }
  WindowToggleMaximise()
}
</script>

<template>
  <div class="top-bar" @dblclick="handleDoubleClick">
    <div class="left-panel">
      <img src="/appicon.png" alt="logo" class="app-icon" />
      <span class="app-title">{{ t('header.app_name') }}</span>
      <span class="app-version">{{ version }}</span>
    </div>

    <div class="right-panel">
      <button type="button" class="window-btn" @click.stop="WindowMinimise">
        <el-icon><Minus /></el-icon>
      </button>
      <button type="button" class="window-btn" @click.stop="WindowToggleMaximise">
        <el-icon><FullScreen /></el-icon>
      </button>
      <button type="button" class="window-btn close-btn" @click.stop="WindowClose">
        <el-icon><Close /></el-icon>
      </button>
    </div>
  </div>
</template>

<style scoped>
.top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px;
  height: 32px;
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
  --wails-draggable: drag;
  user-select: none;
  flex-shrink: 0;
}

.left-panel {
  display: flex;
  align-items: center;
  gap: 8px;
  pointer-events: none;
}

.app-icon {
  width: 20px;
  height: 20px;
  object-fit: contain;
}

.app-title {
  font-weight: 600;
  font-size: 13px;
  color: #303133;
}

.app-version {
  font-size: 11px;
  color: #909399;
}

.right-panel {
  display: flex;
  align-items: center;
  gap: 4px;
  -webkit-app-region: no-drag;
}

.window-btn {
  width: 28px;
  height: 28px;
  padding: 0;
  border: none;
  background-color: transparent;
  color: #606266;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.15s ease;
}

.window-btn:hover {
  background-color: rgba(0, 0, 0, 0.08);
}

.window-btn:active {
  background-color: rgba(0, 0, 0, 0.12);
  transform: scale(0.95);
}

.close-btn:hover {
  background-color: #f56c6c;
  color: #fff;
}

.close-btn:active {
  background-color: #e64242;
}
</style>
