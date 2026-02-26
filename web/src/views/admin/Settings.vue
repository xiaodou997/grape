<template>
  <div class="settings-page">
    <el-form :model="settings" label-width="120px" style="max-width: 600px">
      <el-form-item label="上游源">
        <el-input v-model="settings.upstream" placeholder="https://registry.npmjs.org" />
      </el-form-item>

      <el-form-item label="允许注册">
        <el-switch v-model="settings.allowRegistration" />
      </el-form-item>

      <el-form-item label="包保留天数">
        <el-input-number v-model="settings.retentionDays" :min="1" :max="365" />
      </el-form-item>

      <el-form-item label="最大包大小">
        <el-input v-model="settings.maxPackageSize">
          <template #append>MB</template>
        </el-input>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" @click="handleSave">保存设置</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { ElMessage } from 'element-plus'

const settings = reactive({
  upstream: 'https://registry.npmjs.org',
  allowRegistration: false,
  retentionDays: 30,
  maxPackageSize: 100,
})

const handleSave = () => {
  // TODO: Save settings to API
  ElMessage.success('设置已保存')
}
</script>

<style scoped>
.settings-page {
  padding: 20px 0;
}
</style>
