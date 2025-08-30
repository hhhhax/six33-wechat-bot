<template>
  <div class="fixed top-0 left-0 w-full h-full flex items-center justify-center z-[9999] pointer-events-none">
    <!-- 遮罩层 -->
    <div 
      v-if="visible" 
      class="fixed inset-0 bg-black bg-opacity-50 pointer-events-auto"
      @click="close"
    ></div>
    
    <!-- 通知弹窗 -->
    <transition
      enter-active-class="transition duration-300 ease-out"
      enter-from-class="transform scale-95 opacity-0"
      enter-to-class="transform scale-100 opacity-100"
      leave-active-class="transition duration-200 ease-in"
      leave-from-class="transform scale-100 opacity-100"
      leave-to-class="transform scale-95 opacity-0"
    >
      <div 
        v-if="visible"
        class="bg-white rounded-lg shadow-2xl max-w-md w-full mx-4 pointer-events-auto relative z-10"
      >
        <!-- 头部 -->
        <div class="flex items-center px-6 py-4 border-b border-gray-200">
          <div class="flex items-center">
            <!-- 图标 -->
            <div 
              class="flex items-center justify-center w-8 h-8 rounded-full mr-3"
              :class="iconClasses"
            >
              <svg class="w-5 h-5" :class="iconTextClasses" fill="currentColor" viewBox="0 0 20 20">
                <!-- 成功图标 -->
                <path v-if="type === 'success'" fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"></path>
                <!-- 错误图标 -->
                <path v-else-if="type === 'error'" fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path>
                <!-- 警告图标 -->
                <path v-else-if="type === 'warning'" fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"></path>
                <!-- 信息图标 -->
                <path v-else fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"></path>
              </svg>
            </div>
            <!-- 标题 -->
            <h3 class="text-lg font-semibold text-gray-900">{{ title }}</h3>
          </div>
          <!-- 关闭按钮 -->
          <button 
            @click="close"
            class="ml-auto text-gray-400 hover:text-gray-600 transition-colors"
          >
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path>
            </svg>
          </button>
        </div>
        
        <!-- 内容 -->
        <div class="px-6 py-4">
          <p class="text-gray-700 leading-relaxed">{{ message }}</p>
        </div>
        
        <!-- 底部按钮 -->
        <div class="flex justify-end px-6 py-4 border-t border-gray-200 space-x-3">
          <button 
            v-if="showCancel"
            @click="handleCancel"
            class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
          >
            取消
          </button>
          <button 
            @click="handleConfirm"
            class="px-4 py-2 text-sm font-medium text-white rounded-md transition-colors"
            :class="confirmButtonClasses"
          >
            确定
          </button>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';

const visible = ref(false);
const title = ref('');
const message = ref('');
const type = ref('info'); // 'success', 'error', 'warning', 'info'
const showCancel = ref(false);
const resolveCallback = ref(null);

// 图标样式
const iconClasses = computed(() => {
  switch (type.value) {
    case 'success':
      return 'bg-green-100';
    case 'error':
      return 'bg-red-100';
    case 'warning':
      return 'bg-yellow-100';
    default:
      return 'bg-blue-100';
  }
});

const iconTextClasses = computed(() => {
  switch (type.value) {
    case 'success':
      return 'text-green-600';
    case 'error':
      return 'text-red-600';
    case 'warning':
      return 'text-yellow-600';
    default:
      return 'text-blue-600';
  }
});

// 确认按钮样式
const confirmButtonClasses = computed(() => {
  switch (type.value) {
    case 'success':
      return 'bg-green-600 hover:bg-green-700';
    case 'error':
      return 'bg-red-600 hover:bg-red-700';
    case 'warning':
      return 'bg-yellow-600 hover:bg-yellow-700';
    default:
      return 'bg-blue-600 hover:bg-blue-700';
  }
});

// 显示通知
const show = (titleText, messageText, notificationType = 'info', showCancelButton = false) => {
  title.value = titleText;
  message.value = messageText;
  type.value = notificationType;
  showCancel.value = showCancelButton;
  visible.value = true;
  
  return new Promise((resolve) => {
    resolveCallback.value = resolve;
  });
};

// 关闭通知
const close = () => {
  visible.value = false;
  if (resolveCallback.value) {
    resolveCallback.value(false);
    resolveCallback.value = null;
  }
};

// 确认
const handleConfirm = () => {
  visible.value = false;
  if (resolveCallback.value) {
    resolveCallback.value(true);
    resolveCallback.value = null;
  }
};

// 取消
const handleCancel = () => {
  visible.value = false;
  if (resolveCallback.value) {
    resolveCallback.value(false);
    resolveCallback.value = null;
  }
};

// 暴露方法
defineExpose({
  show,
  close
});
</script>

<style scoped>
/* 确保弹窗在最顶层 */
.fixed {
  position: fixed !important;
}
</style>
