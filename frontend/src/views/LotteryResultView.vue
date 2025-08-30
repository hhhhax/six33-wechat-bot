<template>
  <div class="card">
    <div class="card-header">
      <h2 class="card-title">开奖结果管理</h2>
      <p class="text-gray-600">设置和查看各彩种开奖结果</p>
    </div>
    
    <div class="space-y-6">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div v-for="lottery in lotteryTypes" :key="lottery.key" class="bg-gray-50 p-4 rounded-lg">
          <h3 class="text-lg font-semibold mb-3">{{ lottery.name }}</h3>
          
          <div class="space-y-3">
            <div class="form-group">
              <label class="form-label">期数</label>
              <input v-model="lottery.period" type="text" class="form-input" placeholder="期数">
            </div>
            
            <div class="form-group">
              <label class="form-label">主号码 (6个)</label>
              <input v-model="lottery.mainNumbers" type="text" class="form-input" placeholder="用逗号分隔">
            </div>
            
            <div class="form-group">
              <label class="form-label">特码</label>
              <input v-model="lottery.specialNumber" type="number" class="form-input" placeholder="特码">
            </div>
            
            <button @click="setResult(lottery)" class="btn btn-primary w-full">设置开奖结果</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue';
import { goApi } from '../api/goApi';

const lotteryTypes = reactive([
  { key: 'new_macau', name: '新澳门', period: '', mainNumbers: '', specialNumber: '' },
  { key: 'old_macau', name: '老澳门', period: '', mainNumbers: '', specialNumber: '' },
  { key: 'hongkong', name: '香港', period: '', mainNumbers: '', specialNumber: '' }
]);

const setResult = async (lottery) => {
  try {
    const mainNumbers = lottery.mainNumbers.split(',').map(n => parseInt(n.trim())).filter(n => !isNaN(n));
    const specialNumber = parseInt(lottery.specialNumber);
    
    if (mainNumbers.length !== 6 || isNaN(specialNumber)) {
      alert('请输入正确的号码格式');
      return;
    }
    
    await goApi.setLotteryResult(lottery.key, mainNumbers, specialNumber, lottery.period);
    alert('开奖结果设置成功');
  } catch (error) {
    console.error('设置开奖结果失败:', error);
    alert('设置失败: ' + error.message);
  }
};
</script>