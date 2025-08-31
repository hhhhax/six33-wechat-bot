<template>
  <div v-if="isVisible" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="close">
    <div class="bg-white rounded-lg shadow-xl max-w-4xl max-h-[80vh] w-full mx-4 overflow-hidden">
      <!-- 标题栏 -->
      <div class="bg-blue-600 text-white px-6 py-4 flex justify-between items-center">
        <h3 class="text-lg font-semibold">
          {{ betTypeDisplay }} - 下注详情
        </h3>
        <button @click="close" class="text-white hover:text-gray-200 text-xl">
          ×
        </button>
      </div>

      <!-- 内容区域 -->
      <div class="p-6 overflow-y-auto max-h-[calc(80vh-120px)]">
        <!-- 总览信息 -->
        <div class="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
          <div class="grid grid-cols-3 gap-4 text-center">
            <div>
              <div class="text-sm text-gray-600">总组数</div>
              <div class="text-xl font-bold text-blue-600">{{ betTypeInfo?.totalGroups || 0 }}组</div>
            </div>
            <div>
              <div class="text-sm text-gray-600">模式数量</div>
              <div class="text-xl font-bold text-purple-600">{{ getModeCount() }}种</div>
            </div>
            <div>
              <div class="text-sm text-gray-600">总金额</div>
              <div class="text-xl font-bold text-red-600">{{ betTypeInfo?.totalAmount || 0 }}元</div>
            </div>
          </div>
        </div>

        <!-- 按模式分组的下注详情 -->
        <div v-for="(mode, modeName) in betTypeInfo?.modes || {}" :key="modeName" 
             class="bg-white border border-gray-200 rounded-lg overflow-hidden mb-4">
          <div class="bg-gray-100 px-4 py-3 border-b flex justify-between items-center">
            <h4 class="font-semibold text-gray-700">{{ getModeDisplayName(modeName) }}</h4>
            <div class="text-sm text-gray-600">
              {{ mode.groups }}组 / {{ mode.amount }}元
            </div>
          </div>
          
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-4 py-3 text-left border-r">序号</th>
                  <th class="px-4 py-3 text-left border-r">下注号码</th>
                  <th class="px-4 py-3 text-center border-r">金额</th>
                  <th class="px-4 py-3 text-center">描述</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(detail, index) in mode.betDetails || []" :key="index" 
                    class="border-t hover:bg-gray-50 transition-colors">
                  <td class="px-4 py-3 border-r font-mono">{{ index + 1 }}</td>
                  <td class="px-4 py-3 border-r">
                    <div class="flex flex-wrap gap-1">
                      <span v-for="number in detail.numbers" :key="number" 
                            class="inline-block px-2 py-1 bg-blue-100 text-blue-800 rounded text-xs font-mono">
                        {{ String(number).padStart(2, '0') }}
                      </span>
                    </div>
                  </td>
                  <td class="px-4 py-3 text-center border-r font-mono font-semibold text-red-600">
                    {{ detail.amount }}元
                  </td>
                  <td class="px-4 py-3 text-center text-gray-600">
                    {{ detail.description }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- 模式说明信息 -->
        <div v-if="hasModeInfo" class="mt-4 bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <h4 class="font-semibold text-yellow-800 mb-2">下注模式说明</h4>
          <div class="text-sm text-yellow-700 space-y-1">
            <div v-if="hasMode('complex')" class="flex items-center">
              <span class="inline-block w-2 h-2 bg-yellow-500 rounded-full mr-2"></span>
              复式下注：从选择的号码中生成所有可能的组合
            </div>
            <div v-if="hasMode('drag')" class="flex items-center">
              <span class="inline-block w-2 h-2 bg-yellow-500 rounded-full mr-2"></span>
              拖码下注：多组号码的笛卡尔积组合
            </div>
            <div v-if="hasMode('multiple')" class="flex items-center">
              <span class="inline-block w-2 h-2 bg-yellow-500 rounded-full mr-2"></span>
              多组下注：每组独立下注
            </div>
            <div v-if="hasMode('single')" class="flex items-center">
              <span class="inline-block w-2 h-2 bg-yellow-500 rounded-full mr-2"></span>
              单组下注：单一号码组合
            </div>
            <div class="flex items-center">
              <span class="inline-block w-2 h-2 bg-yellow-500 rounded-full mr-2"></span>
              下注类型：{{ betTypeDisplay }}
            </div>
          </div>
        </div>
      </div>

      <!-- 底部按钮 -->
      <div class="bg-gray-50 px-6 py-4 border-t flex justify-end">
        <button @click="close" 
                class="px-6 py-2 bg-gray-500 text-white rounded hover:bg-gray-600 transition-colors">
          关闭
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  isVisible: {
    type: Boolean,
    default: false
  },
  lottery: {
    type: String,
    default: ''
  },
  betType: {
    type: String,
    default: ''
  },
  betTypeInfo: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['close'])

// 计算属性
const betTypeDisplay = computed(() => {
  const typeMap = {
    '三中三': '三中三',
    '二中二': '二中二', 
    '三中二': '三中二',
    '特碰': '特碰'
  }
  return `${props.lottery} ${typeMap[props.betType] || props.betType}`
})

// 方法
const close = () => {
  emit('close')
}

const getModeCount = () => {
  return Object.keys(props.betTypeInfo?.modes || {}).length
}

const getModeDisplayName = (modeName) => {
  const modeNames = {
    'complex': '复式模式',
    'drag': '拖码模式',  
    'multiple': '多组模式',
    'single': '单组模式'
  }
  return modeNames[modeName] || modeName
}

const hasMode = (modeName) => {
  return props.betTypeInfo?.modes && props.betTypeInfo.modes[modeName]
}

const hasModeInfo = computed(() => {
  return Object.keys(props.betTypeInfo?.modes || {}).length > 0
})
</script>

<style scoped>
/* 自定义滚动条样式 */
.overflow-y-auto::-webkit-scrollbar {
  width: 6px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}
</style>
