<script lang="ts" setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
</script>

<template>
  <div class="app-container">
    <!-- 侧边栏导航 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <img 
          id="logo" 
          alt="AgentFramework logo" 
          src="./assets/images/logo-universal.png" 
        />
        <h2>AgentFramework</h2>
      </div>
      <nav class="sidebar-nav">
        <el-menu
          :default-active="activeIndex"
          class="el-menu-vertical-demo"
          @select="handleNavigation"
          router
        >
          <el-menu-item index="/">
            <el-icon><House /></el-icon>
            <span>首页</span>
          </el-menu-item>
          <el-menu-item index="/workflow">
            <el-icon><DataAnalysis /></el-icon>
            <span>工作流编辑器</span>
          </el-menu-item>
          <el-menu-item index="/skills">
            <el-icon><Grid /></el-icon>
            <span>技能管理</span>
          </el-menu-item>
          <el-menu-item index="/config">
            <el-icon><Setting /></el-icon>
            <span>配置管理</span>
          </el-menu-item>
          <el-menu-item index="/files">
            <el-icon><Folder /></el-icon>
            <span>文件浏览器</span>
          </el-menu-item>
        </el-menu>
      </nav>
    </aside>

    <!-- 主内容区域 -->
    <main class="main-content">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<script lang="ts">
export default {
  data() {
    return {
      activeIndex: '/'
    }
  },
  methods: {
    handleNavigation(path: string) {
      this.activeIndex = path
    }
  }
}
</script>

<style>
#logo {
  display: block;
  width: 80%;
  height: auto;
  margin: auto;
  background-position: center;
  background-repeat: no-repeat;
  background-size: 100% 100%;
  background-origin: content-box;
}
</style>

<style scoped>
.app-container {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}

.sidebar {
  width: 240px;
  background-color: #f5f7fa;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid #e4e7ed;
  text-align: center;
}

.sidebar-header h2 {
  margin: 10px 0 0 0;
  font-size: 18px;
  color: #303133;
}

.sidebar-nav {
  flex: 1;
  padding: 20px 0;
}

.main-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background-color: #ffffff;
}

/* 过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

