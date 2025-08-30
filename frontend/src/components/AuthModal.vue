<template>
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-xl shadow-2xl p-8 max-w-md w-full mx-4">
      <!-- 头部 -->
      <div class="text-center mb-8">
        <div class="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <span class="text-2xl">🔐</span>
        </div>
        <h2 class="text-2xl font-bold text-gray-800 mb-2">系统授权验证</h2>
        <p class="text-gray-600">请输入您的授权码以使用六合彩智能解析系统</p>
      </div>

      <!-- 授权表单 -->
      <form @submit.prevent="handleSubmit" class="space-y-6">
        <div>
          <label for="authCode" class="block text-sm font-medium text-gray-700 mb-2">
            授权码
          </label>
          <textarea
            id="authCode"
            v-model="authCode"
            rows="4"
            placeholder="请粘贴您的授权码..."
            class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none font-mono text-sm"
            :disabled="loading"
            required
          ></textarea>
        </div>

        <!-- 错误信息 -->
        <div v-if="error" class="bg-red-50 border border-red-200 rounded-lg p-4">
          <div class="flex items-center">
            <span class="text-red-500 mr-2">❌</span>
            <span class="text-red-700 text-sm">{{ error }}</span>
          </div>
        </div>

        <!-- 成功信息 -->
        <div v-if="success" class="bg-green-50 border border-green-200 rounded-lg p-4">
          <div class="flex items-center">
            <span class="text-green-500 mr-2">✅</span>
            <span class="text-green-700 text-sm">{{ success }}</span>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="flex space-x-4">
          <button
            type="submit"
            :disabled="loading || !authCode.trim()"
            class="flex-1 bg-blue-600 text-white py-3 px-4 rounded-lg font-medium hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200"
          >
            <span v-if="loading" class="flex items-center justify-center">
              <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              验证中...
            </span>
            <span v-else>开始验证</span>
          </button>
          
          <button
            type="button"
            @click="clearInput"
            :disabled="loading"
            class="px-4 py-3 border border-gray-300 text-gray-700 rounded-lg font-medium hover:bg-gray-50 focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200"
          >
            清空
          </button>
        </div>
      </form>

      <!-- 帮助信息 -->
      <div class="mt-8 pt-6 border-t border-gray-200">
        <div class="text-xs text-gray-500 space-y-2">
          <p>• 授权码由系统管理员提供</p>
          <p>• 请确保授权码完整且未过期</p>
          <p>• 如有问题请联系技术支持</p>
        </div>
      </div>

      <!-- 版本信息 -->
      <div class="mt-4 text-center">
        <p class="text-xs text-gray-400">
          六合彩智能解析系统 v1.0.0
        </p>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { Authorize } from '../../wailsjs/go/main/App'

// 定义 emits
const emit = defineEmits<{
  authorize: []
}>()

// 响应式数据
const authCode = ref('')
const loading = ref(false)
const error = ref('')
const success = ref('')

// 方法
const handleSubmit = async () => {
  if (!authCode.value.trim()) {
    error.value = '请输入授权码'
    return
  }

  loading.value = true
  error.value = ''
  success.value = ''

  try {
    const result = await Authorize(authCode.value.trim())
    
    if (result) {
      success.value = '授权验证成功！正在进入系统...'
      
      // 延迟一下让用户看到成功消息
      setTimeout(() => {
        emit('authorize')
      }, 1000)
    } else {
      error.value = '授权验证失败，请检查授权码是否正确'
    }
  } catch (err: any) {
    console.error('授权验证错误:', err)
    
    // 根据错误类型显示不同的错误信息
    if (err.message) {
      if (err.message.includes('expired')) {
        error.value = '授权码已过期，请联系管理员获取新的授权码'
      } else if (err.message.includes('invalid')) {
        error.value = '授权码格式不正确，请检查是否完整复制'
      } else if (err.message.includes('mismatch')) {
        error.value = '授权码与当前设备不匹配'
      } else if (err.message.includes('debug')) {
        error.value = '检测到调试环境，无法在调试模式下运行'
      } else {
        error.value = `授权验证失败: ${err.message}`
      }
    } else {
      error.value = '授权验证失败，请稍后重试'
    }
  } finally {
    loading.value = false
  }
}

const clearInput = () => {
  authCode.value = ''
  error.value = ''
  success.value = ''
}

// 键盘事件处理
const handleKeydown = (event: KeyboardEvent) => {
  if (event.ctrlKey && event.key === 'v') {
    // 允许粘贴
    setTimeout(() => {
      error.value = ''
      success.value = ''
    }, 100)
  }
}

// 自动聚焦到输入框
import { onMounted } from 'vue'
onMounted(() => {
  const textarea = document.getElementById('authCode') as HTMLTextAreaElement
  if (textarea) {
    textarea.focus()
  }
})
</script>

<style scoped>
/* 组件内特定样式 */
.modal-overlay {
  backdrop-filter: blur(4px);
}

/* 输入框样式增强 */
textarea:focus {
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

/* 按钮悬停效果 */
button:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* 成功状态下的特殊样式 */
.success-state {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

/* 加载动画 */
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.animate-spin {
  animation: spin 1s linear infinite;
}

/* 错误信息抖动动画 */
.error-shake {
  animation: shake 0.5s ease-in-out;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-5px); }
  75% { transform: translateX(5px); }
}
</style>
