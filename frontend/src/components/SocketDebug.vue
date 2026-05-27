<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useI18n } from '../composables/useI18n'
import { 
  ConnectSocket, 
  SendSocketMessage, 
  DisconnectSocket,
  IsSocketConnected
} from '../../wailsjs/go/service/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { Connection, SwitchButton, Promotion, Delete } from '@element-plus/icons-vue'

const { t } = useI18n()

const protocol = ref('ws')
const host = ref('localhost')
const port = ref(8080)
const path = ref('/')
const headers = ref('')
const message = ref('')
const messages = ref([])
const connected = ref(false)
const connecting = ref(false)
const autoScroll = ref(true)
const messageListHeight = ref(0)

// 模板引用
const messageListRef = ref(null)

const protocols = [
  { label: 'WebSocket (ws://)', value: 'ws' },
  { label: 'WebSocket Secure (wss://)', value: 'wss' },
  { label: 'TCP', value: 'tcp' }
]

// 获取当前时间
const getTime = () => {
  return new Date().toLocaleTimeString()
}

// 添加消息
const addMessage = (type, content) => {
  messages.value.push({
    type,
    content,
    time: getTime()
  })
  
  // 自动滚动到底部
  if (autoScroll.value) {
    nextTick(() => {
      scrollToBottom()
    })
  }
}

// 滚动到底部
const scrollToBottom = () => {
  if (messageListRef.value) {
    const list = messageListRef.value
    list.scrollTop = list.scrollHeight
  }
}

// 处理滚动事件
const handleScroll = (e) => {
  // 可用于检测用户是否滚动到顶部/底部
}

// 处理端口输入，只允许数字
const handlePortInput = (value) => {
  const numStr = value.replace(/\D/g, '')
  const num = parseInt(numStr)
  if (numStr === '') {
    port.value = ''
  } else if (num > 65535) {
    port.value = 65535
  } else {
    port.value = num
  }
}

// 连接
const connect = async () => {
  if (!host.value) {
    ElMessage.warning(t('socket.msg.host_required'))
    return
  }

  const portNum = parseInt(port.value)
  if (isNaN(portNum) || portNum < 1 || portNum > 65535) {
    ElMessage.warning(t('socket.msg.port_invalid'))
    return
  }

  connecting.value = true

  try {
    const result = await ConnectSocket(
      protocol.value,
      host.value,
      portNum,
      path.value,
      headers.value
    )
  } catch (e) {
    addMessage('error', t('socket.msg.connect_failed') + e)
  } finally {
    connecting.value = false
  }
}

// 断开连接
const disconnect = async () => {
  try {
    await DisconnectSocket(protocol.value)
    connected.value = false
  } catch (e) {
    addMessage('error', t('socket.msg.disconnect_failed') + e)
  }
}

// 发送消息
const send = async () => {
  if (!message.value.trim()) return
  if (!connected.value) {
    ElMessage.warning(t('socket.msg.not_connected'))
    return
  }

  try {
    const result = await SendSocketMessage(protocol.value, message.value)
    if (result) {
      addMessage('sent', message.value)
      message.value = ''
    }
  } catch (e) {
    addMessage('error', t('socket.msg.send_failed') + e)
  }
}

// 清空消息
const clearMessages = () => {
  messages.value = []
}

// 监听 Socket 事件
const setupEventListeners = () => {
  EventsOn('socket-message', (data) => {
    addMessage(data.type, data.content)
  })

  EventsOn('socket-status', (data) => {
    if (data.type === 'connected') {
      connected.value = true
      addMessage('system', data.content)
    } else if (data.type === 'disconnected') {
      connected.value = false
      addMessage('system', data.content)
    } else if (data.type === 'connecting') {
      addMessage('system', data.content)
    }
  })
}

// 清理事件监听
const cleanupEventListeners = () => {
  EventsOff('socket-message')
  EventsOff('socket-status')
}

// 监听 messages 变化自动滚动
watch(() => messages.value.length, () => {
  if (autoScroll.value) {
    nextTick(() => {
      scrollToBottom()
    })
  }
})

onMounted(() => {
  setupEventListeners()
  nextTick(() => {
    if (messageListRef.value) {
      messageListHeight.value = messageListRef.value.clientHeight
    }
  })
  IsSocketConnected(protocol.value).then(res => {
    connected.value = res
  })
})

onUnmounted(() => {
  cleanupEventListeners()
})
</script>

<template>
  <div class="socket-debug">
    <div class="connection-section">
      <div class="connection-header">
        <div class="connection-title">
          <span>{{ t('socket.connection.title') }}</span>
          <el-tag v-if="connected" type="success" size="small">{{ t('socket.connection.status.connected') }}</el-tag>
          <el-tag v-else type="info" size="small">{{ t('socket.connection.status.disconnected') }}</el-tag>
        </div>
        <div class="connection-actions">
          <el-button
            v-if="!connected"
            type="primary"
            :loading="connecting"
            :icon="Connection"
            @click="connect"
          >
            {{ t('socket.connection.btn.connect') }}
          </el-button>
          <el-button v-else type="danger" :icon="SwitchButton" @click="disconnect">
            {{ t('socket.connection.btn.disconnect') }}
          </el-button>
        </div>
      </div>

      <div class="connection-form">
        <el-form label-width="80px">
          <el-row :gutter="16">
            <el-col :span="6">
              <el-form-item :label="t('socket.connection.label.protocol')">
                <el-select v-model="protocol" style="width: 100%" :disabled="connected">
                  <el-option
                    v-for="p in protocols"
                    :key="p.value"
                    :label="p.label"
                    :value="p.value"
                  />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item :label="t('socket.connection.label.address')">
                <el-input v-model="host" :placeholder="t('socket.connection.placeholder.host')" :disabled="connected" />
              </el-form-item>
            </el-col>
            <el-col :span="5">
              <el-form-item :label="t('socket.connection.label.port')">
                <el-input 
                  v-model="port" 
                  :placeholder="t('socket.connection.placeholder.port')" 
                  :disabled="connected"
                  @input="handlePortInput"
                />
              </el-form-item>
            </el-col>
            <el-col :span="5">
              <el-form-item :label="t('socket.connection.label.path')">
                <el-input v-model="path" :placeholder="t('socket.connection.placeholder.path')" :disabled="connected" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row>
            <el-col :span="24">
              <el-form-item :label="t('socket.connection.label.headers')">
                <el-input v-model="headers" type="textarea" :rows="2" :placeholder="t('socket.connection.placeholder.headers')" :disabled="connected" />
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </div>
    </div>

    <div class="message-section">
      <div class="message-header">
        <span>{{ t('socket.message.title') }} ({{ messages.length }})</span>
        <div class="message-actions">
          <el-checkbox v-model="autoScroll">{{ t('socket.message.auto_scroll') }}</el-checkbox>
          <el-button size="small" :icon="Delete" @click="clearMessages">{{ t('socket.message.btn.clear') }}</el-button>
        </div>
      </div>
      
      <div 
        ref="messageListRef"
        class="message-list"
      >
        <div class="message-list-content">
          <div
            v-for="(msg, index) in messages"
            :key="index"
            :class="['message-item', msg.type]"
          >
            <div class="message-meta">
              <span class="message-time">{{ msg.time }}</span>
              <el-tag 
                v-if="msg.type === 'sent'" 
                type="primary" 
                size="small"
              >{{ t('socket.message.tag.sent') }}</el-tag>
              <el-tag 
                v-else-if="msg.type === 'received'" 
                type="success" 
                size="small"
              >{{ t('socket.message.tag.received') }}</el-tag>
              <el-tag 
                v-else-if="msg.type === 'error'" 
                type="danger" 
                size="small"
              >{{ t('socket.message.tag.error') }}</el-tag>
              <el-tag 
                v-else 
                type="info" 
                size="small"
              >{{ t('socket.message.tag.system') }}</el-tag>
            </div>
            <div class="message-content">{{ msg.content }}</div>
          </div>
        </div>
        <el-empty v-if="messages.length === 0" :description="t('socket.message.empty')" />
      </div>

      <div class="message-input">
        <el-input
          v-model="message"
          :placeholder="t('socket.message.placeholder')"
          :disabled="!connected"
          @keyup.enter="send"
        >
          <template #append>
            <el-button 
              :icon="Promotion" 
              :disabled="!connected || !message.trim()"
              @click="send"
            >
              {{ t('socket.message.btn.send') }}
            </el-button>
          </template>
        </el-input>
      </div>
    </div>
  </div>
</template>

<style scoped>
.socket-debug {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.connection-section {
  background: #fff;
  border-radius: 4px;
  border: 1px solid #ebeef5;
  overflow: hidden;
  flex-shrink: 0;
}

.connection-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: #f8f9fa;
  border-bottom: 1px solid #ebeef5;
}

.connection-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-weight: 600;
  color: #303133;
}

.connection-form {
  padding: 16px;
}

.message-section {
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 4px;
  border: 1px solid #ebeef5;
  overflow: hidden;
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #ebeef5;
  font-weight: 600;
  color: #303133;
  flex-shrink: 0;
}

.message-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.message-list {
  height: 360px;
  overflow-y: scroll;
  scrollbar-width: none;
  background: #fafafa;
}

.message-list::-webkit-scrollbar {
  display: none;
}

.message-item {
  padding: 10px 14px;
  border-bottom: 1px solid #ebeef5;
  background: #fff;
}

.message-item.sent {
  background: #ecf5ff;
  border-left: 3px solid #409eff;
}

.message-item.received {
  background: #f0f9eb;
  border-left: 3px solid #67c23a;
}

.message-item.error {
  background: #fef0f0;
  border-left: 3px solid #f56c6c;
}

.message-item.system {
  background: #f4f4f5;
  border-left: 3px solid #909399;
}

.message-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.message-time {
  font-size: 12px;
  color: #909399;
}

.message-content {
  word-break: break-all;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
}

.message-input {
  padding: 16px;
  border-top: 1px solid #ebeef5;
}
</style>
