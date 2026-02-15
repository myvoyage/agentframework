<template>
  <div class="config-manager">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>配置管理</span>
          <div class="header-actions">
            <el-button type="primary" size="small" @click="saveConfig">
              <el-icon><Save /></el-icon> 保存配置
            </el-button>
            <el-button size="small" @click="reloadConfig">
              <el-icon><Refresh /></el-icon> 重新加载
            </el-button>
          </div>
        </div>
      </template>
      
      <div class="config-content">
        <!-- 配置标签页 -->
        <el-tabs v-model="activeTab" type="border-card">
          <!-- 全局配置 -->
          <el-tab-pane label="全局配置" name="global">
            <el-form label-position="top" size="small">
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="应用名称">
                    <el-input v-model="config.appName" placeholder="请输入应用名称"></el-input>
                  </el-form-item>
                  <el-form-item label="日志级别">
                    <el-select v-model="config.logLevel" placeholder="请选择日志级别">
                      <el-option label="调试" value="debug"></el-option>
                      <el-option label="信息" value="info"></el-option>
                      <el-option label="警告" value="warn"></el-option>
                      <el-option label="错误" value="error"></el-option>
                    </el-select>
                  </el-form-item>
                  <el-form-item label="工作目录">
                    <el-input v-model="config.workDir" placeholder="请输入工作目录"></el-input>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="最大并发数">
                    <el-input-number v-model="config.maxConcurrency" :min="1" :max="100"></el-input-number>
                  </el-form-item>
                  <el-form-item label="超时时间（秒）">
                    <el-input-number v-model="config.timeout" :min="1" :max="3600"></el-input-number>
                  </el-form-item>
                  <el-form-item label="启用监控">
                    <el-switch v-model="config.enableMonitoring"></el-switch>
                  </el-form-item>
                </el-col>
              </el-row>
            </el-form>
          </el-tab-pane>
          
          <!-- 模型配置 -->
          <el-tab-pane label="模型配置" name="model">
            <el-form label-position="top" size="small">
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="默认模型">
                    <el-input v-model="config.defaultModel" placeholder="请输入默认模型名称"></el-input>
                  </el-form-item>
                  <el-form-item label="API 密钥">
                    <el-input v-model="config.apiKey" type="password" placeholder="请输入API密钥" show-password></el-input>
                  </el-form-item>
                  <el-form-item label="API 端点">
                    <el-input v-model="config.apiEndpoint" placeholder="请输入API端点"></el-input>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="模型温度">
                    <el-slider v-model="config.temperature" :min="0" :max="1" :step="0.1"></el-slider>
                    <div class="slider-value">{{ config.temperature }}</div>
                  </el-form-item>
                  <el-form-item label="最大 tokens">
                    <el-input-number v-model="config.maxTokens" :min="100" :max="32000"></el-input-number>
                  </el-form-item>
                  <el-form-item label="启用缓存">
                    <el-switch v-model="config.enableCache"></el-switch>
                  </el-form-item>
                </el-col>
              </el-row>
            </el-form>
          </el-tab-pane>
          
          <!-- 缓存配置 -->
          <el-tab-pane label="缓存配置" name="cache">
            <el-form label-position="top" size="small">
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="缓存类型">
                    <el-select v-model="config.cache.type" placeholder="请选择缓存类型">
                      <el-option label="内存" value="memory"></el-option>
                      <el-option label="Redis" value="redis"></el-option>
                    </el-select>
                  </el-form-item>
                  <el-form-item label="缓存过期时间（秒）">
                    <el-input-number v-model="config.cache.expiration" :min="60" :max="86400"></el-input-number>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item v-if="config.cache.type === 'redis'" label="Redis 地址">
                    <el-input v-model="config.cache.redis.address" placeholder="请输入Redis地址"></el-input>
                  </el-form-item>
                  <el-form-item v-if="config.cache.type === 'redis'" label="Redis 密码">
                    <el-input v-model="config.cache.redis.password" type="password" placeholder="请输入Redis密码" show-password></el-input>
                  </el-form-item>
                </el-col>
              </el-row>
            </el-form>
          </el-tab-pane>
          
          <!-- 高级配置 -->
          <el-tab-pane label="高级配置" name="advanced">
            <el-form label-position="top" size="small">
              <el-form-item label="原始配置（JSON）">
                <el-input 
                  v-model="configJson" 
                  type="textarea" 
                  rows="15" 
                  placeholder="JSON格式的配置" 
                  @input="validateConfigJson"
                ></el-input>
                <div v-if="configJsonError" class="config-error">
                  {{ configJsonError }}
                </div>
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import * as main from '../../wailsjs/go/main/App'

const activeTab = ref('global')
const configJson = ref('')
const configJsonError = ref('')

// 配置数据
const config = reactive({
  appName: 'AgentFramework',
  logLevel: 'info',
  workDir: './',
  maxConcurrency: 10,
  timeout: 30,
  enableMonitoring: true,
  defaultModel: 'gpt-4',
  apiKey: '',
  apiEndpoint: 'https://api.openai.com/v1',
  temperature: 0.7,
  maxTokens: 4096,
  enableCache: true,
  cache: {
    type: 'memory',
    expiration: 3600,
    redis: {
      address: 'localhost:6379',
      password: ''
    }
  }
})

// 监听配置变化，更新JSON
watch(config, (newConfig) => {
  configJson.value = JSON.stringify(newConfig, null, 2)
  validateConfigJson()
}, { deep: true })

// 获取配置
const getConfig = async () => {
  try {
    const result = await main.GetConfig()
    if (result) {
      // 将后端配置映射到前端配置
      config.appName = result.Name || 'AgentFramework'
      config.defaultModel = result.DefaultModel || 'gpt-4'

      // 从 Models (Record<string, ModelConfig>) 获取配置
      if (result.Models) {
        const modelKeys = Object.keys(result.Models)
        if (modelKeys.length > 0) {
          // 获取默认模型或第一个模型
          const defaultModelName = config.defaultModel
          let modelConfig = result.Models[defaultModelName]

          // 如果默认模型不存在，使用第一个模型
          if (!modelConfig && modelKeys.length > 0) {
            modelConfig = result.Models[modelKeys[0]]
            config.defaultModel = modelKeys[0]
          }

          if (modelConfig) {
            config.apiKey = modelConfig.api_key || ''
            config.apiEndpoint = modelConfig.base_url || 'https://api.openai.com/v1'
            config.temperature = modelConfig.temperature || 0.7
            config.maxTokens = modelConfig.max_tokens || 4096
          }
        }
      }

      // 从 Memory.ModelCache 获取缓存配置
      if (result.Memory && result.Memory.ModelCache) {
        const modelCache = result.Memory.ModelCache
        config.enableCache = modelCache.Enabled || false
        // ModelCacheSpec 没有 Type 字段，根据 Enabled 判断
        config.cache.expiration = modelCache.TTL || 3600
      }

      // 从 ThreadStore 获取配置
      if (result.ThreadStore) {
        config.workDir = result.ThreadStore.Dir || './'
      }
    }

    configJson.value = JSON.stringify(result || config, null, 2)
    validateConfigJson()
  } catch (error) {
    console.error('获取配置失败:', error)
    ElMessage.error('获取配置失败: ' + (error as Error).message)
  }
}

// 保存配置
const saveConfig = async () => {
  try {
    // 首先获取当前配置
    const currentConfig = await main.GetConfig()
    if (!currentConfig) {
      throw new Error('无法获取当前配置')
    }

    // 更新配置
    currentConfig.Name = config.appName
    currentConfig.DefaultModel = config.defaultModel

    // 更新模型配置
    if (currentConfig.Models && currentConfig.Models[config.defaultModel]) {
      const modelConfig = currentConfig.Models[config.defaultModel]
      modelConfig.api_key = config.apiKey
      modelConfig.base_url = config.apiEndpoint
      modelConfig.temperature = config.temperature
      modelConfig.max_tokens = config.maxTokens
    }

    // 更新缓存配置
    if (currentConfig.Memory && currentConfig.Memory.ModelCache) {
      currentConfig.Memory.ModelCache.Enabled = config.enableCache
      // 注意：缓存类型和TTL的更新需要更复杂的处理，这里暂时只更新启用状态
    }

    // 更新 ThreadStore 配置
    if (currentConfig.ThreadStore) {
      currentConfig.ThreadStore.Dir = config.workDir
    }

    await main.UpdateConfig(currentConfig)
    ElMessage.success('配置已保存')

    // 重新加载配置以确保同步
    await getConfig()
  } catch (error) {
    console.error('保存配置失败:', error)
    ElMessage.error('保存配置失败: ' + (error as Error).message)
  }
}

// 重新加载配置
const reloadConfig = async () => {
  try {
    await main.ReloadConfig()
    await getConfig()
    ElMessage.success('配置已重新加载')
  } catch (error) {
    console.error('重新加载配置失败:', error)
    ElMessage.error('重新加载配置失败: ' + (error as Error).message)
  }
}

// 验证配置JSON格式
const validateConfigJson = () => {
  try {
    if (configJson.value) {
      const parsedConfig = JSON.parse(configJson.value)
      // 更新配置对象
      Object.assign(config, parsedConfig)
    }
    configJsonError.value = ''
  } catch (error) {
    configJsonError.value = 'JSON格式错误: ' + (error as Error).message
  }
}

// 生命周期钩子
onMounted(() => {
  getConfig()
})
</script>

<style scoped>
.config-manager {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.config-content {
  margin-top: 20px;
}

.slider-value {
  text-align: center;
  margin-top: 10px;
  color: #606266;
  font-size: 14px;
}

.config-error {
  color: #f56c6c;
  font-size: 12px;
  margin-top: 5px;
}
</style>
