<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../composables/useI18n'
import {
  GetALLLang, GetCurrentLang, SetLanguage, GetLangTextMap,
  GetProxy, SetProxy, GetTimeout, SetTimeout,
  GetLogLevel, SetLogLevel
} from '../../wailsjs/go/service/App'

const { t, textMap } = useI18n()

const settings = ref({
  proxyURL: '',
  timeout: 30,
  logLevel: 'info'
})

const languages = ref([])
const currentLang = ref('')

const logLevels = [
  { label: 'Debug', value: 'debug' },
  { label: 'Info', value: 'info' },
  { label: 'Warn', value: 'warn' },
  { label: 'Error', value: 'error' }
]

// 加载所有设置
onMounted(async () => {
  try {
    const [langInfos, proxy, timeoutCfg, logLevel] = await Promise.all([
      GetALLLang(),
      GetProxy(),
      GetTimeout(),
      GetLogLevel()
    ])

    languages.value = langInfos.map(info => ({
      code: info.language_code,
      name: info.language_name
    }))
    currentLang.value = await GetCurrentLang() || 'zh-CN'

    if (proxy) settings.value.proxyURL = proxy.url || ''
    if (timeoutCfg) settings.value.timeout = timeoutCfg.timeout || 30
    settings.value.logLevel = logLevel || 'info'
  } catch (e) {
    console.error('获取设置失败:', e)
  }
})

// 切换语言
const handleLangChange = async (langCode) => {
  try {
    await SetLanguage(langCode)
    const map = await GetLangTextMap()
    textMap.value = map || {}
    ElMessage.success(t('settings.msg.saved'))
  } catch (e) {
    console.error('设置语言失败:', e)
  }
}

// 保存所有设置
const saveSettings = async () => {
  try {
    await Promise.all([
      SetProxy(settings.value.proxyURL),
      SetTimeout(settings.value.timeout),
      SetLogLevel(settings.value.logLevel)
    ])
    ElMessage.success(t('settings.msg.saved'))
  } catch (e) {
    ElMessage.error('保存设置失败: ' + e)
  }
}

// 恢复默认
const resetSettings = async () => {
  settings.value = {
    proxyURL: '',
    timeout: 30,
    logLevel: 'info'
  }

  const defaultLang = 'zh-CN'

  try {
    await Promise.all([
      SetProxy(''),
      SetTimeout(30),
      SetLogLevel('info'),
      SetLanguage(defaultLang)
    ])

    currentLang.value = defaultLang
    const map = await GetLangTextMap()
    textMap.value = map || {}

    ElMessage.info(t('settings.msg.reset'))
  } catch (e) {
    ElMessage.error('恢复默认失败: ' + e)
  }
}
</script>

<template>
  <div class="settings">
    <div class="settings-header">
      <span>{{ t('settings.title') }}</span>
    </div>

    <el-form :model="settings" label-width="120px" class="settings-form">
      <!-- 外观 -->
      <el-divider content-position="left">{{ t('settings.form.appearance_divider') }}</el-divider>

      <el-form-item :label="t('settings.form.language_label')">
        <el-select v-model="currentLang" style="width: 200px" @change="handleLangChange">
          <el-option
            v-for="lang in languages"
            :key="lang.code"
            :label="lang.name"
            :value="lang.code"
          />
        </el-select>
      </el-form-item>

      <!-- 网络 -->
      <el-divider content-position="left">{{ t('settings.form.network_divider') }}</el-divider>

      <el-form-item :label="t('settings.form.proxy_label')">
        <el-input
          v-model="settings.proxyURL"
          :placeholder="t('settings.form.proxy_placeholder')"
          style="width: 320px"
          clearable
        />
        <div class="form-tip">{{ t('settings.form.proxy_tip') }}</div>
      </el-form-item>

      <el-form-item :label="t('settings.form.timeout_label')">
        <el-input-number v-model="settings.timeout" :min="1" :max="300" />
        <span class="form-unit">{{ t('settings.form.timeout_unit') }}</span>
      </el-form-item>

      <!-- 日志 -->
      <el-divider content-position="left">{{ t('settings.form.log_divider') }}</el-divider>

      <el-form-item :label="t('settings.form.log_level_label')">
        <el-select v-model="settings.logLevel" style="width: 200px">
          <el-option
            v-for="level in logLevels"
            :key="level.value"
            :label="level.label"
            :value="level.value"
          />
        </el-select>
        <span class="form-tip-inline">{{ t('settings.form.log_level_tip') }}</span>
      </el-form-item>

      <el-divider />

      <el-form-item>
        <el-button type="primary" @click="saveSettings">{{ t('settings.form.btn.save') }}</el-button>
        <el-button @click="resetSettings">{{ t('settings.form.btn.reset') }}</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<style scoped>
.settings {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  padding: 24px;
}

.settings-header {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 24px;
}

.settings-form {
  max-width: 600px;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}

.form-unit {
  margin-left: 8px;
  color: #606266;
}

.form-tip-inline {
  margin-left: 8px;
  font-size: 12px;
  color: #909399;
}
</style>
