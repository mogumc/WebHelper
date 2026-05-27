import { ref, onMounted } from 'vue'
import { GetLangTextMap } from '../../wailsjs/go/service/App'

// 全局单例状态
const textMap = ref({})
const isLoaded = ref(false)

export function useI18n() {
  const loadTextMap = async () => {
    try {
      const map = await GetLangTextMap()
      textMap.value = map || {}
      isLoaded.value = true
    } catch (error) {
      console.error('加载语言包失败:', error)
      textMap.value = {}
      isLoaded.value = true
    }
  }

  const t = (key, defaultValue = '') => {
    return textMap.value[key] || defaultValue || key
  }

  onMounted(() => {
    if (!isLoaded.value) {
      loadTextMap()
    }
  })

  return {
    t,
    textMap,
    isLoaded,
    loadTextMap
  }
}