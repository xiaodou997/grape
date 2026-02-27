<template>
  <div class="users-page">
    <div class="page-header">
      <h3>用户管理 <span class="user-count">共 {{ users.length }} 个用户</span></h3>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        添加用户
      </el-button>
    </div>

    <el-table :data="users" stripe v-loading="loading">
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="role" label="角色" width="120">
        <template #default="{ row }">
          <el-tag :type="roleTagType(row.role)">
            {{ roleLabel(row.role) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="lastLogin" label="最后登录" width="180">
        <template #default="{ row }">
          {{ row.lastLogin ? formatTime(row.lastLogin) : '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button text type="primary" @click="handleEdit(row)" :disabled="row.username === 'admin'">
            编辑
          </el-button>
          <el-button text type="danger" @click="handleDelete(row)" :disabled="row.username === 'admin'">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && users.length === 0" description="暂无用户" />

    <!-- Create User Dialog -->
    <el-dialog v-model="showCreateDialog" title="添加用户" width="420px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="用户名" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="form.role" style="width: 100%">
            <el-option label="管理员" value="admin" />
            <el-option label="开发者" value="developer" />
            <el-option label="只读" value="readonly" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="creating">创建</el-button>
      </template>
    </el-dialog>

    <!-- Edit User Dialog -->
    <el-dialog v-model="showEditDialog" title="编辑用户" width="420px">
      <el-form ref="editFormRef" :model="editForm" label-width="80px">
        <el-form-item label="用户名">
          <el-input :value="editForm.username" disabled />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="editForm.email" />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="editForm.password" type="password" show-password placeholder="留空则不修改" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="editForm.role" style="width: 100%">
            <el-option label="管理员" value="admin" />
            <el-option label="开发者" value="developer" />
            <el-option label="只读" value="readonly" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import { adminApi } from '@/api'

interface User {
  username: string
  email: string
  role: string
  lastLogin?: string
  createdAt?: string
}

const loading = ref(false)
const creating = ref(false)
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const formRef = ref<FormInstance>()

const users = ref<User[]>([])

const form = reactive({
  name: '',
  email: '',
  password: '',
  role: 'developer',
})

const editForm = reactive({
  username: '',
  email: '',
  password: '',
  role: 'developer',
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度 3-20 个字符', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' },
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' },
  ],
}

const roleTagType = (role: string): 'danger' | 'primary' | 'info' => {
  if (role === 'admin') return 'danger'
  if (role === 'developer') return 'primary'
  return 'info'
}

const roleLabel = (role: string): string => {
  if (role === 'admin') return '管理员'
  if (role === 'developer') return '开发者'
  if (role === 'readonly') return '只读'
  return role
}

const formatTime = (time?: string): string => {
  if (!time) return '-'
  try {
    return new Date(time).toLocaleString('zh-CN')
  } catch {
    return time
  }
}

const loadUsers = async () => {
  loading.value = true
  try {
    const res = await adminApi.getUsers()
    users.value = res.data.users || []
  } catch {
    ElMessage.error('加载用户列表失败')
  } finally {
    loading.value = false
  }
}

const handleCreate = async () => {
  const valid = await formRef.value?.validate()
  if (!valid) return

  creating.value = true
  try {
    await adminApi.createUser({
      name: form.name,
      email: form.email,
      password: form.password,
      role: form.role,
    })
    ElMessage.success('创建成功')
    showCreateDialog.value = false
    formRef.value?.resetFields()
    form.role = 'developer'
    loadUsers()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '创建失败')
  } finally {
    creating.value = false
  }
}

const handleEdit = (row: User) => {
  editForm.username = row.username
  editForm.email = row.email
  editForm.password = ''
  editForm.role = row.role || 'developer'
  showEditDialog.value = true
}

const handleSaveEdit = async () => {
  const payload: { email?: string; password?: string; role?: string } = {}
  if (editForm.email) payload.email = editForm.email
  if (editForm.password) payload.password = editForm.password
  if (editForm.role) payload.role = editForm.role

  try {
    await adminApi.updateUser(editForm.username, payload)
    ElMessage.success('保存成功')
    showEditDialog.value = false
    loadUsers()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  }
}

const handleDelete = async (row: User) => {
  if (row.username === 'admin') {
    ElMessage.warning('不能删除管理员账户')
    return
  }

  try {
    await ElMessageBox.confirm(`确定删除用户 ${row.username}？`, '警告', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await adminApi.deleteUser(row.username)
    ElMessage.success('删除成功')
    loadUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

onMounted(() => {
  loadUsers()
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

.user-count {
  font-size: 14px;
  color: #909399;
  font-weight: normal;
  margin-left: 8px;
}
</style>
