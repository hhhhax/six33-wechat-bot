<template>
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-xl shadow-2xl p-6 max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
      <!-- 头部 -->
      <div class="flex items-center justify-between mb-6">
        <h3 class="text-xl font-bold text-gray-800">提供正确的解析结果</h3>
        <button
          @click="$emit('close')"
          class="text-gray-400 hover:text-gray-600 transition-colors duration-200"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <!-- 原始输入显示 -->
      <div class="mb-6">
        <label class="block text-sm font-medium text-gray-700 mb-2">
          原始输入
        </label>
        <div class="bg-gray-50 rounded-lg p-3 text-sm font-mono border">
          {{ input }}
        </div>
      </div>

      <!-- 系统解析结果 -->
      <div class="mb-6">
        <label class="block text-sm font-medium text-gray-700 mb-2">
          系统解析结果
        </label>
        <div class="bg-red-50 border border-red-200 rounded-lg p-3">
          <div v-if="result && result.bet_groups && result.bet_groups.length > 0">
            <div
              v-for="(group, index) in result.bet_groups"
              :key="index"
              class="mb-2 last:mb-0"
            >
              <span class="text-sm text-red-800">
                {{ formatBetType(group.type) }}: {{ group.numbers.join('.') }} - {{ group.amount }}元
              </span>
            </div>
          </div>
          <div v-else class="text-red-600 text-sm">
            系统未能正确解析
          </div>
        </div>
      </div>

      <!-- 正确结果输入 -->
      <div class="mb-6">
        <label class="block text-sm font-medium text-gray-700 mb-2">
          请描述正确的解析结果
        </label>
        <textarea
          v-model="correctionText"
          rows="4"
          placeholder="请详细描述正确的解析结果，例如：&#10;应该解析为：三中三 1.2.3 100元&#10;或者：二中二 4.5 50元 + 三中三 6.7.8 80元"
          class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
        ></textarea>
      </div>

      <!-- 错误类型选择 -->
      <div class="mb-6">
        <label class="block text-sm font-medium text-gray-700 mb-2">
          错误类型
        </label>
        <div class="grid grid-cols-2 gap-2">
          <label
            v-for="errorType in errorTypes"
            :key="errorType.value"
            class="flex items-center p-3 border border-gray-200 rounded-lg cursor-pointer hover:bg-gray-50 transition-colors duration-200"
            :class="{ 'border-blue-500 bg-blue-50': selectedErrorType === errorType.value }"
          >
            <input
              type="radio"
              :value="errorType.value"
              v-model="selectedErrorType"
              class="mr-3 text-blue-600"
            >
            <div>
              <div class="text-sm font-medium text-gray-800">{{ errorType.label }}</div>
              <div class="text-xs text-gray-600">{{ errorType.description }}</div>
            </div>
          </label>
        </div>
      </div>

      <!-- 附加说明 -->
      <div class="mb-6">
        <label class="block text-sm font-medium text-gray-700 mb-2">
          附加说明（可选）
        </label>
        <textarea
          v-model="additionalComment"
          rows="3"
          placeholder="请提供更多详细信息或改进建议..."
          class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
        ></textarea>
      </div>

      <!-- 操作按钮 -->
      <div class="flex space-x-4">
        <button
          @click="$emit('close')"
          class="flex-1 px-6 py-3 border border-gray-300 text-gray-700 rounded-lg font-medium hover:bg-gray-50 focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors duration-200"
        >
          取消
        </button>
        <button
          @click="submitCorrection"
          :disabled="!correctionText.trim()"
          class="flex-1 px-6 py-3 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200"
        >
          提交纠正
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue'

// 定义 props
interface Props {
  input: string
  result: any
}

const props = defineProps<Props>()

// 定义 emits
const emit = defineEmits<{
  close: []
  submit: [data: any]
}>()

// 响应式数据
const correctionText = ref('')
const selectedErrorType = ref('PARSE_FAILED')
const additionalComment = ref('')

// 错误类型选项
const errorTypes = [
  {
    value: 'PARSE_FAILED',
    label: '完全解析失败',
    description: '系统无法识别输入内容'
  },
  {
    value: 'BET_TYPE_ERROR',
    label: '下注类型错误',
    description: '识别的玩法类型不正确'
  },
  {
    value: 'NUMBER_ERROR',
    label: '号码识别错误',
    description: '号码组合识别不正确'
  },
  {
    value: 'AMOUNT_ERROR',
    label: '金额解析错误',
    description: '下注金额计算不正确'
  },
  {
    value: 'LOTTERY_TYPE_ERROR',
    label: '彩种识别错误',
    description: '彩种类型识别不正确'
  },
  {
    value: 'GROUP_COUNT_ERROR',
    label: '组数计算错误',
    description: '下注组数计算不正确'
  },
  {
    value: 'FORMAT_ERROR',
    label: '格式理解错误',
    description: '对输入格式的理解有误'
  },
  {
    value: 'OTHER_ERROR',
    label: '其他错误',
    description: '其他类型的解析问题'
  }
]

// 方法
const formatBetType = (type: string) => {
  const typeMap: Record<string, string> = {
    '2_in_2': '二中二',
    '3_in_3': '三中三',
    '3_in_2': '三中二',
    'special': '特碰'
  }
  return typeMap[type] || type
}

const submitCorrection = () => {
  if (!correctionText.value.trim()) {
    return
  }

  const correctionData = {
    input: props.input,
    system_output: props.result?.bet_groups || [],
    correction_text: correctionText.value.trim(),
    error_type: selectedErrorType.value,
    additional_comment: additionalComment.value.trim(),
    is_correct: false,
    timestamp: new Date().toISOString()
  }

  emit('submit', correctionData)
}
</script>

<style scoped>
/* 滚动条样式 */
::-webkit-scrollbar {
  width: 6px;
}

::-webkit-scrollbar-track {
  background: #f1f5f9;
}

::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

/* 模态框遮罩层 */
.modal-overlay {
  backdrop-filter: blur(4px);
}

/* 单选按钮样式 */
input[type="radio"]:checked {
  background-color: #2563eb;
  border-color: #2563eb;
}

/* 过渡动画 */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
