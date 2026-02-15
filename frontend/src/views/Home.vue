<template>
  <div class="home-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>欢迎使用 AgentFramework</span>
        </div>
      </template>
      <div class="card-content">
        <p>AgentFramework 是一个高性能、企业级的 AI 代理框架，支持工作流编排、技能管理和文件系统操作。</p>
        <div class="feature-grid">
          <el-card shadow="hover" class="feature-card">
            <div class="feature-icon">
              <el-icon><DataAnalysis /></el-icon>
            </div>
            <h3>可视化工作流编辑器</h3>
            <p>基于 G6 图可视化库，提供直观的工作流设计界面</p>
            <el-button type="primary" @click="$router.push('/workflow')">
              进入工作流编辑器
            </el-button>
          </el-card>
          
          <el-card shadow="hover" class="feature-card">
            <div class="feature-icon">
              <el-icon><Grid /></el-icon>
            </div>
            <h3>技能管理</h3>
            <p>管理和配置各种 AI 技能，支持技能的启用、禁用和配置</p>
            <el-button type="primary" @click="$router.push('/skills')">
              进入技能管理
            </el-button>
          </el-card>
          
          <el-card shadow="hover" class="feature-card">
            <div class="feature-icon">
              <el-icon><Setting /></el-icon>
            </div>
            <h3>配置管理</h3>
            <p>全局配置、模型配置和缓存配置的图形化管理界面</p>
            <el-button type="primary" @click="$router.push('/config')">
              进入配置管理
            </el-button>
          </el-card>
          
          <el-card shadow="hover" class="feature-card">
            <div class="feature-icon">
              <el-icon><Folder /></el-icon>
            </div>
            <h3>文件系统浏览器</h3>
            <p>直观的文件系统浏览和操作界面，支持拖拽文件到工作区</p>
            <el-button type="primary" @click="$router.push('/files')">
              进入文件浏览器
            </el-button>
          </el-card>
        </div>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.home-container {
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-content {
  padding: 20px 0;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-top: 30px;
}

.feature-card {
  text-align: center;
  transition: transform 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-5px);
}

.feature-icon {
  font-size: 48px;
  color: #409eff;
  margin-bottom: 15px;
}

.feature-card h3 {
  margin: 10px 0;
  font-size: 18px;
  color: #303133;
}

.feature-card p {
  color: #606266;
  margin-bottom: 20px;
}
</style>
