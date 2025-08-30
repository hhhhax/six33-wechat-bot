<template>
  <div class="card">
    <div class="card-header">
      <h2 class="card-title">历史记录</h2>
      <p class="text-gray-600">查看解析历史和统计信息</p>
    </div>
    
    <div class="space-y-4">
      <div class="flex justify-between items-center">
        <h3 class="text-lg font-semibold">解析历史</h3>
        <button class="btn btn-outline">清空历史</button>
      </div>
      
      <div class="overflow-x-auto">
        <table class="table">
          <thead>
            <tr>
              <th>时间</th>
              <th>输入内容</th>
              <th>识别格式</th>
              <th>置信度</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td colspan="5" class="text-center text-gray-500">暂无历史记录</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
// 历史记录逻辑
</script>