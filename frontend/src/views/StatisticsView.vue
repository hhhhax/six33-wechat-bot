<template>
  <div class="card">
    <div class="card-header">
      <h2 class="card-title">统计信息</h2>
      <p class="text-gray-600">解析器性能和使用统计</p>
    </div>
    
    <div class="space-y-6">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div class="bg-blue-50 p-4 rounded-lg text-center">
          <div class="text-2xl font-bold text-blue-600">0</div>
          <div class="text-sm text-gray-600">总解析次数</div>
        </div>
        <div class="bg-green-50 p-4 rounded-lg text-center">
          <div class="text-2xl font-bold text-green-600">0</div>
          <div class="text-sm text-gray-600">成功解析</div>
        </div>
        <div class="bg-red-50 p-4 rounded-lg text-center">
          <div class="text-2xl font-bold text-red-600">0</div>
          <div class="text-sm text-gray-600">解析失败</div>
        </div>
        <div class="bg-yellow-50 p-4 rounded-lg text-center">
          <div class="text-2xl font-bold text-yellow-600">0%</div>
          <div class="text-sm text-gray-600">平均置信度</div>
        </div>
      </div>
      
      <div class="bg-gray-50 p-4 rounded-lg">
        <h3 class="text-lg font-semibold mb-3">支持的格式</h3>
        <div class="text-sm text-gray-600">
          <div>• 三中三组合</div>
          <div>• 特码直选</div>
          <div>• 复式投注</div>
          <div>• 拖码格式</div>
          <div>• 生肖投注</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// 统计信息逻辑
</script>