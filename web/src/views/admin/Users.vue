<template>
  <div class="users-page">
    <div class="page-header">
      <h3>用户管理</h3>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        添加用户
      </el-button>
    </div>

    <el-table :data="users" stripe>
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="role" label="角色">
        <template #default="{ row }">
          <el-tag :type="row.role === 'admin' ? 'danger' : 'info'">
            {{ row.role === 'admin' ? '管理员' : '开发者' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="lastLogin" label="最后登录" width="180" />
      <el-table-column prop="createdAt" label="创建时间" width="180" />
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button text type="primary" @click="handleEdit(row)">编辑</el-button>
          <el-button text type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Create User Dialog -->
    <el-dialog v-model="showCreateDialog" title="添加用户" width="400px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="form.role" style="width: 100%">
            <el-option label="开发者" value="developer" />
            <el-option label="管理员" value="admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'

const showCreateDialog = ref(false)
const formRef = ref<FormInstance>()

const users = ref([
  { username: 'admin', email: 'admin@example.com', role: 'admin', lastLogin: '2024-01-15 10:30', createdAt: '2024-01-01' },
  { username: 'developer', email: 'dev@example.com', role: 'developer', lastLogin: '2024-01-14 15:00', createdAt: '2024-01-10' },
])

const form = reactive({
  username: '',
  email: '',
  password: '',
  role: 'developer',
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' },
  ],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }],
}

const handleCreate = async () => {
  const valid = await formRef.value?.validate()
  if (!valid) return

  // TODO: Call API to create user
  users.value.push({
    username: form.username,
    email: form.email,
    role: form.role,
    lastLogin: '-',
    createdAt: new Date().toLocaleDateString('zh-CN'),
  })

  ElMessage.success('创建成功')
  showCreateDialog.value = false
  formRef.value?.resetFields()
}

const handleEdit = (row: any) => {
  ElMessage.info(`编辑用户: ${row.username}`)
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定删除用户 ${row.username}？`, '警告', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
    
    users.value = users.value.filter(u => u.username !== row.username)
    ElMessage.success('删除成功')
  } catch {
    // User cancelled
  }
}

onMounted(() => {
  // TODO: Load users from API
})
</script>

<style scoped>
.users-page {
  padding: 20px 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h3 {
  margin: 0;
}
</style>
