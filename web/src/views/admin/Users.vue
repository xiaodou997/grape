<template>
  <div class="users-page fade-in">
    <div class="page-header-modern">
      <div class="header-info">
        <h3>{{ t('users.title') }}</h3>
        <span class="count-badge">{{ users.length }} {{ t('common.all') }}</span>
      </div>
      <el-button type="primary" @click="showCreateDialog = true" class="btn-with-shadow">
        <el-icon><Plus /></el-icon>
        {{ t('users.createUser') }}
      </el-button>
    </div>

    <div class="table-container">
      <el-table :data="users" v-loading="loading" class="modern-table">
        <el-table-column prop="username" :label="t('users.username')" min-width="120">
          <template #default="{ row }">
            <span class="username-cell">{{ row.username }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="email" :label="t('users.email')" min-width="180" />
        <el-table-column prop="role" :label="t('users.role')" width="120">
          <template #default="{ row }">
            <el-tag :type="roleTagType(row.role)" size="small" effect="light" round>
              {{ roleLabel(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="lastLogin" :label="t('tokens.lastUsed')" width="160">
          <template #default="{ row }">
            <span class="time-cell">{{ row.lastLogin ? formatTime(row.lastLogin) : '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="160" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button text type="primary" size="small" @click="handleEdit(row)" :disabled="row.username === 'admin'">
                {{ t('common.edit') }}
              </el-button>
              <el-button text type="danger" size="small" @click="handleDelete(row)" :disabled="row.username === 'admin'">
                {{ t('common.delete') }}
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && users.length === 0" :description="t('common.empty')" />
    </div>

    <!-- Create User Dialog -->
    <el-dialog v-model="showCreateDialog" :title="t('users.createUser')" width="440px" class="modern-dialog">
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item :label="t('users.username')" prop="name">
          <el-input v-model="form.name" :placeholder="t('users.username')" />
        </el-form-item>
        <el-form-item :label="t('users.email')" prop="email">
          <el-input v-model="form.email" placeholder="email@example.com" />
        </el-form-item>
        <el-form-item :label="t('users.password')" prop="password">
          <el-input v-model="form.password" type="password" show-password :placeholder="t('users.passwordMinLength')" />
        </el-form-item>
        <el-form-item :label="t('users.role')" prop="role">
          <el-select v-model="form.role" style="width: 100%">
            <el-option :label="t('users.roleAdmin')" value="admin" />
            <el-option :label="t('users.roleUser')" value="developer" />
            <el-option :label="t('common.default')" value="readonly" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showCreateDialog = false">{{ t('common.cancel') }}</el-button>
          <el-button type="primary" @click="handleCreate" :loading="creating">{{ t('common.confirm') }}</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- Edit User Dialog -->
    <el-dialog v-model="showEditDialog" :title="t('users.editUser')" width="440px" class="modern-dialog">
      <el-form ref="editFormRef" :model="editForm" label-position="top">
        <el-form-item :label="t('users.username')">
          <el-input :value="editForm.username" disabled />
        </el-form-item>
        <el-form-item :label="t('users.email')">
          <el-input v-model="editForm.email" />
        </el-form-item>
        <el-form-item :label="t('users.newPassword')">
          <el-input v-model="editForm.password" type="password" show-password :placeholder="t('settings.jwtSecretPlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('users.role')">
          <el-select v-model="editForm.role" style="width: 100%">
            <el-option :label="t('users.roleAdmin')" value="admin" />
            <el-option :label="t('users.roleUser')" value="developer" />
            <el-option :label="t('common.default')" value="readonly" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showEditDialog = false">{{ t('common.cancel') }}</el-button>
          <el-button type="primary" @click="handleSaveEdit">{{ t('common.save') }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import { adminApi } from '@/api'

const { t } = useI18n()

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
    { required: true, message: t('login.usernameRequired'), trigger: 'blur' },
    { min: 3, max: 20, message: '3-20 characters', trigger: 'blur' },
  ],
  email: [
    { required: true, message: 'Email required', trigger: 'blur' },
    { type: 'email', message: 'Invalid email', trigger: 'blur' },
  ],
  password: [
    { required: true, message: t('login.passwordRequired'), trigger: 'blur' },
    { min: 5, message: t('users.passwordMinLength'), trigger: 'blur' },
  ],
  role: [
    { required: true, message: 'Role required', trigger: 'change' },
  ],
}

const roleTagType = (role: string): 'danger' | 'primary' | 'info' => {
  if (role === 'admin') return 'danger'
  if (role === 'developer') return 'primary'
  return 'info'
}

const roleLabel = (role: string): string => {
  if (role === 'admin') return t('users.roleAdmin')
  if (role === 'developer') return t('users.roleUser')
  return t('common.default')
}

const formatTime = (time?: string): string => {
  if (!time) return '-'
  return new Date(time).toLocaleDateString(undefined, {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit'
  })
}

const loadUsers = async () => {
  loading.value = true
  try {
    const res = await adminApi.getUsers()
    users.value = res.data.users || []
  } catch {
    ElMessage.error(t('errors.loadFailed'))
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
    ElMessage.success(t('users.userCreated'))
    showCreateDialog.value = false
    formRef.value?.resetFields()
    form.role = 'developer'
    loadUsers()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || t('errors.saveFailed'))
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
    ElMessage.success(t('users.userUpdated'))
    showEditDialog.value = false
    loadUsers()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || t('errors.saveFailed'))
  }
}

const handleDelete = async (row: User) => {
  try {
    await ElMessageBox.confirm(`${t('users.deleteUserConfirm')} ${row.username}?`, t('common.warning'), {
      confirmButtonText: t('common.delete'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    })

    await adminApi.deleteUser(row.username)
    ElMessage.success(t('users.userDeleted'))
    loadUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || t('errors.deleteFailed'))
    }
  }
}

onMounted(loadUsers)
</script>

<style scoped>
.users-page {
  padding: 0;
}

.page-header-modern {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-info h3 {
  font-size: 20px;
  font-weight: 700;
  margin: 0 0 4px 0;
}

.count-badge {
  font-size: 12px;
  color: var(--g-text-muted);
  background: var(--g-bg);
  padding: 2px 8px;
  border-radius: 10px;
}

.username-cell {
  font-weight: 600;
  color: var(--g-text-primary);
}

.time-cell {
  font-size: 13px;
  color: var(--g-text-secondary);
}

.action-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
}

.btn-with-shadow {
  box-shadow: 0 4px 6px -1px rgba(124, 58, 237, 0.2);
}
</style>
