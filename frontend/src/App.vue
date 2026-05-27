<script setup>
import { ref } from 'vue'
import { useI18n } from './composables/useI18n'
import HeaderBar from './components/HeaderBar.vue'
import WebDebug from './components/WebDebug.vue'
import JsDebug from './components/JsDebug.vue'
import JsonParser from './components/JsonParser.vue'
import SocketDebug from './components/SocketDebug.vue'
import Settings from './components/Settings.vue'

const { t } = useI18n()
const activeTab = ref('webdebug')

const tabs = [
  { name: 'webdebug', key: 'app.tab.webdebug' },
  { name: 'jsdebug', key: 'app.tab.jsdebug' },
  { name: 'jsonparser', key: 'app.tab.jsonparser' },
  { name: 'socketdebug', key: 'app.tab.socketdebug' },
  { name: 'settings', key: 'app.tab.settings' }
]
</script>

<template>
  <div class="app-layout">
    <HeaderBar />
    
    <div class="tab-bar">
      <div 
        v-for="tab in tabs"
        :key="tab.name"
        :class="['tab-item', { active: activeTab === tab.name }]"
        @click="activeTab = tab.name"
      >
        {{ t(tab.key) }}
      </div>
    </div>
    
    <div class="app-content">
      <WebDebug v-show="activeTab === 'webdebug'" class="page-component hide-scrollbar" />
      <JsDebug v-show="activeTab === 'jsdebug'" class="page-component hide-scrollbar" />
      <JsonParser v-show="activeTab === 'jsonparser'" class="page-component hide-scrollbar" />
      <SocketDebug v-show="activeTab === 'socketdebug'" class="page-component hide-scrollbar" />
      <Settings v-show="activeTab === 'settings'" class="page-component hide-scrollbar" />
    </div>
  </div>
</template>

<style scoped>
.app-layout {
  width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #f5f7fa;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  user-select: none;
}

.tab-bar {
  display: flex;
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
  padding: 0 12px;
  flex-shrink: 0;
}

.tab-item {
  padding: 12px 24px;
  cursor: pointer;
  color: #606266;
  font-size: 14px;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.tab-item:hover {
  color: #409eff;
}

.tab-item.active {
  color: #409eff;
  border-bottom-color: #409eff;
}

.app-content {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.page-component {
  position: absolute;
  inset: 0;
  padding: 16px;
  overflow-y: scroll;
  scrollbar-width: none;
}

.page-component::-webkit-scrollbar {
  display: none;
}
</style>
