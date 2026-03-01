<template>
  <div class="header-actions">
    <!-- 语言切换 -->
    <el-dropdown trigger="click" @command="handleLocaleChange">
      <el-button text class="locale-btn">
        <span class="locale-icon">🌐</span>
        <span class="locale-text">{{ currentLocaleLabel }}</span>
      </el-button>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="zh" :class="{ active: currentLocale === 'zh' }">
            <span class="locale-option">🇨🇳 中文</span>
          </el-dropdown-item>
          <el-dropdown-item command="en" :class="{ active: currentLocale === 'en' }">
            <span class="locale-option">🇺🇸 English</span>
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <!-- GitHub 链接 -->
    <el-tooltip :content="t('nav.github')" placement="bottom">
      <a
        href="https://github.com/xiaodou997/grape"
        target="_blank"
        rel="noopener noreferrer"
        class="github-link"
      >
        <svg
          height="20"
          viewBox="0 0 16 16"
          version="1.1"
          width="20"
          aria-hidden="true"
          fill="currentColor"
        >
          <path
            d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"
          ></path>
        </svg>
      </a>
    </el-tooltip>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLocale, type Locale } from '@/locales'

const { t, locale } = useI18n()

const currentLocale = computed(() => locale.value)

const currentLocaleLabel = computed(() => {
  return locale.value === 'zh' ? '中文' : 'EN'
})

const handleLocaleChange = (newLocale: string) => {
  setLocale(newLocale as Locale)
}
</script>

<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.locale-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--el-text-color-regular);
}

.locale-btn:hover {
  color: var(--grape-primary);
}

.locale-icon {
  font-size: 18px;
}

.locale-text {
  font-size: 14px;
}

.locale-option {
  font-size: 14px;
}

.github-link {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  color: var(--el-text-color-regular);
  transition: color 0.2s;
}

.github-link:hover {
  color: var(--grape-primary);
}

:deep(.el-dropdown-menu__item.active) {
  color: var(--grape-primary);
  background-color: var(--el-dropdown-menu-item-hover-fill);
}
</style>
