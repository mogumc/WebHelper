<template>
  <div class="lang-selector">
    <el-select v-model="currentLang" placeholder="选择语言" @change="handleLangChange">
      <el-option
        v-for="lang in languages"
        :key="lang.code"
        :label="lang.name"
        :value="lang.code"
      />
    </el-select>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { GetALLLang, SetLanguage, GetCurrentLang, GetLangTextMap } from '../../wailsjs/go/service/App'
import { useI18n } from '../composables/useI18n'

const { textMap } = useI18n()
const currentLang = ref('')
const languages = ref([])

onMounted(async () => {
  try {
    const langInfos = await GetALLLang()
    languages.value = langInfos.map(info => ({
      code: info.language_code,
      name: info.language_name
    }))
    currentLang.value = await GetCurrentLang() || 'zh-CN'
  } catch (error) {
    console.error('获取语言列表失败:', error)
  }
})

const handleLangChange = async (langCode) => {
  try {
    await SetLanguage(langCode)
    // 重新加载语言映射，无需刷新页面
    const map = await GetLangTextMap()
    textMap.value = map || {}
  } catch (error) {
    console.error('设置语言失败:', error)
  }
}
</script>

<style scoped>
.lang-selector {
  display: inline-block;
}
</style>