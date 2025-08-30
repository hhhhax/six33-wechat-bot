<template>
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center">
    <!-- 背景遮罩 -->
    <div class="fixed inset-0 bg-black bg-opacity-50" @click="hide"></div>
    
    <!-- 消息框 -->
    <div class="relative bg-white rounded-lg shadow-xl p-6 max-w-md w-full mx-4 z-10">
      <!-- 图标和标题 -->
      <div class="flex items-center mb-4">
        <div :class="iconClass" class="w-8 h-8 mr-3 flex items-center justify-center rounded-full">
          <svg v-if="type === 'success'" class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"/>
          </svg>
          <svg v-else-if="type === 'error'" class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"/>
          </svg>
          <svg v-else-if="type === 'warning'" class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
          </svg>
          <svg v-else class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"/>
          </svg>
        </div>
        <h3 class="text-lg font-semibold text-gray-900">{{ title }}</h3>
      </div>
      
      <!-- 消息内容 -->
      <p class="text-gray-600 mb-6">{{ message }}</p>
      
      <!-- 按钮 -->
      <div class="flex justify-end">
        <button
          @click="hide"
          class="px-4 py-2 bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300 transition-colors"
        >
          确定
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';

const visible = ref(false);
const title = ref('');
const message = ref('');
const type = ref('info');

const iconClass = computed(() => {
  const baseClass = 'flex items-center justify-center w-8 h-8 rounded-full';
  switch (type.value) {
    case 'success':
      return `${baseClass} bg-green-500`;
    case 'error':
      return `${baseClass} bg-red-500`;
    case 'warning':
      return `${baseClass} bg-yellow-500`;
    default:
      return `${baseClass} bg-blue-500`;
  }
});

const show = (titleText, messageText, typeValue = 'info') => {
  title.value = titleText;
  message.value = messageText;
  type.value = typeValue;
  visible.value = true;
};

const hide = () => {
  visible.value = false;
};

// 暴露方法给父组件
defineExpose({
  show,
  hide
});
</script>
