<template>
  <div class="about-container">
    <el-card>
      <h1>📖 关于这个项目</h1>
      
      <el-divider content-position="left">项目说明</el-divider>
      <p>
        这是一个用于测试 <strong>Grape Registry</strong> 功能的 Vue 3 基础项目。
        通过这个项目，您可以验证 Grape 作为私有 npm 仓库的完整功能。
      </p>

      <el-divider content-position="left">技术栈</el-divider>
      <el-table :data="techStack" stripe style="width: 100%">
        <el-table-column prop="name" label="技术" width="200" />
        <el-table-column prop="version" label="版本" width="150" />
        <el-table-column prop="description" label="说明" />
      </el-table>

      <el-divider content-position="left">测试步骤</el-divider>
      <el-timeline>
        <el-timeline-item timestamp="步骤 1" type="primary">
          配置 .npmrc 文件，指向 Grape Registry
        </el-timeline-item>
        <el-timeline-item timestamp="步骤 2" type="success">
          运行 <code>npm install</code> 安装依赖
        </el-timeline-item>
        <el-timeline-item timestamp="步骤 3" type="warning">
          运行 <code>npm run dev</code> 启动开发服务器
        </el-timeline-item>
        <el-timeline-item timestamp="步骤 4" type="danger">
          验证所有功能正常工作
        </el-timeline-item>
      </el-timeline>

      <el-divider content-position="left">常用命令</el-divider>
      <el-table :data="commands" style="width: 100%">
        <el-table-column prop="command" label="命令" width="250" />
        <el-table-column prop="description" label="说明" />
      </el-table>

      <div class="back-btn">
        <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
const techStack = [
  { name: 'Vue 3', version: '3.5.25', description: '渐进式 JavaScript 框架' },
  { name: 'Vite', version: '7.3.1', description: '下一代前端构建工具' },
  { name: 'Pinia', version: '3.0.3', description: 'Vue 官方状态管理库' },
  { name: 'Vue Router', version: '4.5.1', description: 'Vue 官方路由管理器' },
  { name: 'Element Plus', version: '2.9.10', description: '基于 Vue 3 的组件库' },
  { name: 'Axios', version: '1.11.0', description: 'HTTP 请求库' },
  { name: 'TypeScript', version: '5.9.3', description: 'JavaScript 超集' },
  { name: 'Sass', version: '1.90.0', description: 'CSS 预处理器' },
]

const commands = [
  { command: 'npm install', description: '安装所有依赖（从 Grape Registry 下载）' },
  { command: 'npm run dev', description: '启动开发服务器（http://localhost:3000）' },
  { command: 'npm run build', description: '构建生产版本' },
  { command: 'npm run preview', description: '预览生产构建' },
]
</script>

<style lang="scss" scoped>
.about-container {
  max-width: 1000px;
  margin: 0 auto;
}

h1 {
  color: #333;
  margin-bottom: 20px;
}

p {
  color: #666;
  line-height: 1.8;
  margin: 15px 0;
}

code {
  background: #f5f7fa;
  padding: 2px 8px;
  border-radius: 4px;
  color: #409eff;
  font-family: 'Courier New', monospace;
}

.back-btn {
  margin-top: 30px;
  text-align: center;
}
</style>
