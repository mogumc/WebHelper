<script setup>
// 递归渲染 JSON 树节点，支持任意深度
import { ref } from 'vue'

const props = defineProps({
  value: { required: true },
  path: { type: String, required: true },
  depth: { type: Number, default: 0 },
  expandedKeys: { type: Object, required: true },
  selectedPath: { type: String, default: '' }
})

const emit = defineEmits(['toggle-expand', 'select-node'])

const isExpanded = (key) => props.expandedKeys.has(key)

const getValueType = (value) => {
  if (value === null) return 'null'
  if (Array.isArray(value)) return 'array'
  return typeof value
}

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

const formatValue = (value) => {
  if (value === null) return 'null'
  if (typeof value === 'string') return `"${value}"`
  return String(value)
}

// 切分子路径给下级
const childPath = (parentPath, key, index) => {
  if (index !== null && index !== undefined) {
    return `${parentPath}[${index}]`
  }
  // key 含特殊字符时用 ['key'] 格式
  if (typeof key === 'string' && /[.\s\[\]]/.test(key)) {
    return `${parentPath}['${key}']`
  }
  // 首层根键不加点
  if (parentPath === '$') {
    return `$.${key}`
  }
  return `${parentPath}.${key}`
}

const isSelected = (nodePath) => props.selectedPath === nodePath

const onToggle = (key) => {
  emit('toggle-expand', key)
}

const onSelect = (nodePath) => {
  emit('select-node', nodePath)
}
</script>

<template>
  <div class="tree-node" :style="{ marginLeft: depth > 0 ? '18px' : '0' }">
    <!-- 节点头 -->
    <div
      class="node-header"
      :class="{ 'node-selected': isSelected(path) }"
      @click="onSelect(path)"
    >
      <span
        v-if="typeof value === 'object' && value !== null"
        class="expand-icon"
        :class="{ expanded: isExpanded(path) }"
        @click.stop="onToggle(path)"
      ></span>
      <span v-else class="expand-icon spacer"></span>

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
    </div>

    <!-- 子节点 -->
    <div v-if="isExpanded(path) && typeof value === 'object' && value !== null" class="node-children">
      <template v-if="Array.isArray(value)">
        <JsonTreeNode
          v-for="(item, index) in value"
          :key="index"
          :value="item"
          :path="childPath(path, `[${index}]`, index)"
          :depth="depth + 1"
          :expandedKeys="expandedKeys"
          :selectedPath="selectedPath"
          @toggle-expand="emit('toggle-expand', $event)"
          @select-node="emit('select-node', $event)"
        />
      </template>
      <template v-else>
        <JsonTreeNode
          v-for="(v, k) in value"
          :key="k"
          :value="v"
          :path="childPath(path, k)"
          :depth="depth + 1"
          :expandedKeys="expandedKeys"
          :selectedPath="selectedPath"
          @toggle-expand="emit('toggle-expand', $event)"
          @select-node="emit('select-node', $event)"
        />
      </template>
    </div>
  </div>
</template>

<style scoped>
.tree-node {
  font-size: 13px;
}

.node-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 6px;
  cursor: pointer;
  border-radius: 3px;
  transition: background-color 0.15s;
  white-space: nowrap;
}

.node-header:hover {
  background-color: #e9ecef;
}

.node-selected {
  background-color: #d9ecff !important;
}

.expand-icon {
  width: 16px;
  height: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: #909399;
  font-size: 10px;
  transition: transform 0.15s;
}

.expand-icon::before {
  content: '▶';
}

.expand-icon.expanded::before {
  content: '▼';
}

.expand-icon.spacer {
  visibility: hidden;
}

.node-type {
  font-size: 11px;
  font-weight: 500;
  padding: 1px 5px;
  background-color: rgba(0, 0, 0, 0.06);
  border-radius: 4px;
  flex-shrink: 0;
}

.node-preview {
  color: #606266;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-children {
  border-left: 1px dashed #dcdfe6;
  margin-left: 7px;
}
</style>
