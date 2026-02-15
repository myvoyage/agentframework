import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import {resolve} from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  // 构建优化
  build: {
    // 启用源码映射
    sourcemap: false,
    // 代码压缩选项
    minify: 'esbuild',
    // 分块策略
    rollupOptions: {
      output: {
        manualChunks: {
          'vue-vendor': ['vue', 'vue-router', 'pinia'],
          'element-plus': ['element-plus', '@element-plus/icons-vue'],
          'g6': ['@antv/g6'],
          'axios': ['axios']
        },
        // 优化分包大小
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
        // 启用gzip压缩
        compact: true
      }
    },
    // 启用gzip压缩
    compress: true,
    // 生成manifest.json文件
    manifest: true,
    // 启用CSS代码分割
    cssCodeSplit: true,
    // 启用预构建
    prefetch: { 
      include: { 
        type: 'asyncChunks' 
      } 
    },
    // 启用预加载
    preload: { 
      include: { 
        type: 'asyncChunks' 
      } 
    }
  },
  // 开发服务器配置
  server: {
    port: 3000,
    open: true,
    // 代理配置
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    },
    // 启用HMR
    hmr: {
      overlay: true
    }
  },
  // 路径别名配置
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@/wailsjs': resolve(__dirname, 'wailsjs')
    },
    // 启用缓存
    dedupe: ['vue']
  },
  // 性能优化
  optimizeDeps: {
    include: ['vue', 'vue-router', 'pinia', 'element-plus', '@element-plus/icons-vue', 'axios'],
    exclude: ['@antv/g6'],
    // 启用依赖预构建缓存
    cache: true,
    // 启用esbuild优化
    esbuildOptions: {
      target: 'es2020',
      // 启用ts编译优化
      tsconfigRaw: {
        compilerOptions: {
          module: 'esnext',
          target: 'es2020',
          strict: true
        }
      }
    }
  },
  // CSS优化
  css: {
    preprocessorOptions: {
      css: {
        charset: false
      }
    },
    devSourcemap: false,
    // 启用CSS模块化
    modules: {
      localsConvention: 'camelCaseOnly'
    }
  },
  // 预览服务器配置
  preview: {
    port: 4173,
    open: true,
    // 启用gzip压缩
    compress: true
  }
})
