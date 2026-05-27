<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../composables/useI18n'
import { GetALLLang, GetCurrentLang, SetLanguage, GetLangTextMap } from '../../wailsjs/go/service/App'

const { t, textMap } = useI18n()

const settings = ref({
  proxyEnabled: false,
  proxyHost: '',
  proxyPort: 7890
})

const languages = ref([])
const currentLang = ref('')

// 加载语言列表
onMounted(async () => {
  try {
    const langInfos = await GetALLLang()
    languages.value = langInfos.map(info => ({
      code: info.language_code,
      name: info.language_name
    }))
    currentLang.value = await GetCurrentLang() || 'zh-CN'
  } catch (e) {
    console.error('获取语言列表失败:', e)
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

const saveSettings = () => {
  localStorage.setItem('webhelper-settings', JSON.stringify(settings.value))
  ElMessage.success(t('settings.msg.saved'))
}

const resetSettings = () => {
  settings.value = {
    proxyEnabled: false,
    proxyHost: '',
    proxyPort: 7890
  }
  localStorage.removeItem('webhelper-settings')
  ElMessage.info(t('settings.msg.reset'))
}
</script>

<template>
  <div class="settings">
    <div class="settings-header">
      <span>{{ t('settings.title') }}</span>
    </div>

    <el-form :model="settings" label-width="100px" class="settings-form">
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

      <el-divider content-position="left">{{ t('settings.form.proxy_divider') }}</el-divider>

      <el-form-item :label="t('settings.form.enable_proxy_label')">
        <el-switch v-model="settings.proxyEnabled" />
      </el-form-item>

      <el-form-item v-if="settings.proxyEnabled" :label="t('settings.form.proxy_address_label')">
        <el-input v-model="settings.proxyHost" :placeholder="t('settings.form.proxy_address_placeholder')" style="width: 200px" />
      </el-form-item>

      <el-form-item v-if="settings.proxyEnabled" :label="t('settings.form.proxy_port_label')">
        <el-input-number v-model="settings.proxyPort" :min="1" :max="65535" />
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
</style>
