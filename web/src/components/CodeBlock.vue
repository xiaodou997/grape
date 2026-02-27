<template>
  <div class="code-block">
    <pre><code :class="languageClass">{{ code }}</code></pre>
    <el-button 
      class="copy-btn" 
      :icon="DocumentCopy" 
      circle 
      size="small" 
      @click="handleCopy" 
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { DocumentCopy } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  code: string
  language?: string
}>()

const languageClass = computed(() => {
  return props.language || 'bash'
})

const handleCopy = async () => {
  try {
    await navigator.clipboard.writeText(props.code)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}
</script>

<style scoped>
.code-block {
  position: relative;
  background: #1e1e1e;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 12px;
}

.code-block pre {
  margin: 0;
  padding: 16px;
  padding-right: 48px;
  overflow-x: auto;
}

.code-block code {
  font-family: 'SF Mono', Monaco, Consolas, 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #d4d4d4;
  white-space: pre;
}

.copy-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  background: transparent !important;
  border: none !important;
  color: #6b6b6b !important;
}

.copy-btn:hover {
  color: #d4d4d4 !important;
}
</style>
