<template>
  <div id="app" class="min-h-screen bg-gray-100 flex flex-col font-inter">
    
    <template v-if="isAuthorized">
      <!-- 导航栏 -->
      <nav class="bg-white shadow-sm border-b border-gray-200">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div class="flex justify-between h-16">
            <div class="flex space-x-8">
              <button 
                @click="currentView = 'home'"
                :class="[
                  'inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors',
                  currentView === 'home' 
                    ? 'border-blue-500 text-blue-600' 
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                ]"
              >
                智能解析
              </button>
              <button   
                @click="currentView = 'config'"
                :class="[
                  'inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors',
                  currentView === 'config' 
                    ? 'border-blue-500 text-blue-600' 
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                ]"
              >
                系统配置
              </button>
            </div>
            <div class="flex items-center">
              <span class="text-sm text-gray-500 mr-4">六合彩智能解析机器人</span>
              <button 
                @click="handleLogout"
                class="bg-red-500 hover:bg-red-600 text-white text-sm px-3 py-1 rounded transition-colors"
              >
                退出登录
              </button>
            </div>
          </div>
        </div>
      </nav>

      <!-- 主要内容区域 -->
      <div class="flex-1 flex flex-col">
        <HomeView v-if="currentView === 'home'" :isAuthorized="isAuthorized" />
        <ConfigView v-if="currentView === 'config'" :isAuthorized="isAuthorized" />
      </div>
    </template>
    <template v-else>
      <div class="flex flex-col items-center justify-center min-h-screen bg-gray-900 text-white">
        <h1 class="text-4xl font-bold mb-8">六合彩智能解析机器人 - 授权登录</h1>
        <div class="bg-gray-800 p-8 rounded-lg shadow-xl w-96">
          <p class="mb-4 text-center text-gray-300">请输入授权码以启动应用。</p>
          <div class="mb-4">
            <label for="authCode" class="block text-gray-400 text-sm font-bold mb-2">授权码:</label>
            <input
                type="password"
                id="authCode"
                v-model="authCodeInput"
                @keyup.enter="handleAuthorize"
                class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline bg-gray-700 border-gray-600 placeholder-gray-400 text-white"
                placeholder="请输入授权码"
            />
          </div>
          <button
              @click="handleAuthorize"
              class="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-lg w-full focus:outline-none focus:shadow-outline transition duration-200"
          >
            授权
          </button>
          <p v-if="authError" class="text-red-500 text-sm mt-4 text-center">{{ authError }}</p>
          
        </div>
      </div>
    </template>

    <MessageBox ref="messageBox" />
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue';
import { goApi } from './api/goApi';
import HomeView from './views/HomeView.vue';
import ConfigView from './views/ConfigView.vue';
import MessageBox from './components/MessageBox.vue';

const isAuthorized = ref(false);
const authCodeInput = ref('');
const authError = ref('');
const messageBox = ref(null);
const currentView = ref('home');
let authCheckInterval = null;

const checkAuthorization = async () => {
  try {
    const authorized = await goApi.isAuthorized();
    isAuthorized.value = Boolean(authorized);
    
    if (!authorized) {
      authError.value = '授权已过期或未授权。请重新授权。';
    } else {
      authError.value = '';
    }
  } catch (error) {
    console.error("检查授权状态失败:", error);
    authError.value = "检查授权状态时发生错误。";
    isAuthorized.value = false;
  }
};

const startAuthCheckInterval = () => {
  if (authCheckInterval) {
    clearInterval(authCheckInterval);
  }
  authCheckInterval = setInterval(checkAuthorization, 60 * 1000);
};

const handleAuthorize = async () => {
  authError.value = '';
  const success = await goApi.authorize(authCodeInput.value);

  if (success) {
    isAuthorized.value = true;
    
    await nextTick();
    
    authCodeInput.value = '';
    // 获取授权到期时间并弹窗显示
    const expiry = await goApi.getAuthorizationExpiry();
    const expiryStr = expiry ? new Date(expiry).toLocaleString() : '未知';
    messageBox.value.show('授权成功', `应用已成功授权，有效期至：${expiryStr}`, 'success');
    startAuthCheckInterval();
  } else {
    authError.value = '授权码不正确或已过期。';
    messageBox.value.show('授权失败', '授权码不正确或已过期。', 'error');
    startAuthCheckInterval();
  }
};

const handleLogout = () => {
  if (confirm('确定要退出登录吗？')) {
    isAuthorized.value = false;
    currentView.value = 'home';
    authCodeInput.value = '';
    authError.value = '';
  }
};

onMounted(async () => {
  await checkAuthorization();
  startAuthCheckInterval();
});

onUnmounted(() => {
  if (authCheckInterval) {
    clearInterval(authCheckInterval);
  }
});
</script>

<style scoped>
.font-inter {
  font-family: 'Inter', sans-serif;
}
</style>