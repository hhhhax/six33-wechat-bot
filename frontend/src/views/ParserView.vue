<template>
  <div class="card">
    <div class="card-header">
      <h2 class="card-title">智能解析器</h2>
      <p class="text-gray-600">输入下注字符串，智能解析为标准格式</p>
    </div>
    
    <div class="space-y-4">
      <div class="form-group">
        <label class="form-label">彩种类型</label>
        <select v-model="lotteryType" class="form-input form-select">
          <option value="new_macau">新澳门</option>
          <option value="old_macau">老澳门</option>
          <option value="hongkong">香港</option>
        </select>
      </div>
      
      <div class="form-group">
        <label class="form-label">下注字符串</label>
        <textarea 
          v-model="inputText" 
          class="form-input form-textarea" 
          placeholder="请输入下注字符串..."
          rows="4"
        ></textarea>
      </div>
      
      <div class="flex space-x-2">
        <button @click="parseInput" class="btn btn-primary">解析</button>
        <button @click="clearInput" class="btn btn-outline">清空</button>
      </div>
      
      <div v-if="parseResult" class="mt-6">
        <h3 class="text-lg font-semibold mb-3">解析结果</h3>
        <div class="bg-gray-50 p-4 rounded-lg">
          <pre class="text-sm">{{ JSON.stringify(parseResult, null, 2) }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { goApi } from '../api/goApi';

const lotteryType = ref('new_macau');
const inputText = ref('');
const parseResult = ref(null);

const parseInput = async () => {
  if (!inputText.value.trim()) return;
  
  try {
    const result = await goApi.parseBetInput(inputText.value, lotteryType.value);
    parseResult.value = result;
  } catch (error) {
    console.error('解析失败:', error);
    parseResult.value = { error: '解析失败: ' + error.message };
  }
};

const clearInput = () => {
  inputText.value = '';
  parseResult.value = null;
};
</script>