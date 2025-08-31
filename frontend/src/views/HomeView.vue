<template>
  <div class="flex flex-col h-full bg-gray-100 text-gray-800 font-inter overflow-hidden">
    <!-- 主体内容 -->
    <div class="flex flex-1 w-full h-full">
      <!-- 左侧：智能解析区域 -->
      <div class="h-full w-[35%] min-w-0 flex flex-col">
        <!-- 标题栏 -->
        <div class="bg-white shadow-sm p-4 border-b">
          <h1 class="text-2xl font-bold text-blue-700 text-center">
            六合彩智能解析机器人
          </h1>
        </div>

        <!-- 智能解析区域 -->
        <div class="flex-1 flex flex-col bg-white shadow-md transition-all duration-300 ease-in-out min-h-0">
          <!-- 解析区域标题 -->
          <div class="sticky top-0 z-10 flex justify-between items-center p-3 border-b bg-white shadow-sm">
            <h2 class="text-xl font-bold text-blue-600">智能解析器</h2>
            <!-- 彩种勾选 -->
            <div class="flex items-center space-x-4 text-sm">
              <label class="flex items-center cursor-pointer">
                <input type="checkbox" v-model="enabledLotteries.new_macau" class="mr-2 text-blue-600 focus:ring-blue-500 rounded">
                <span class="text-blue-600">新澳</span>
              </label>
              <label class="flex items-center cursor-pointer">
                <input type="checkbox" v-model="enabledLotteries.old_macau" class="mr-2 text-green-600 focus:ring-green-500 rounded">
                <span class="text-green-600">老澳</span>
              </label>
              <label class="flex items-center cursor-pointer">
                <input type="checkbox" v-model="enabledLotteries.hongkong" class="mr-2 text-red-600 focus:ring-red-500 rounded">
                <span class="text-red-600">香港</span>
              </label>
            </div>
          </div>

          <!-- 解析内容 -->
          <div class="p-4 flex-1 flex flex-col min-h-0">
            <!-- 下注字符串输入 -->
            <div class="mb-4">
              <label class="block text-sm font-medium text-gray-700 mb-2">下注字符串输入</label>
              <textarea 
                v-model="betInput" 
                @input="onBetInputChange"
                class="w-full h-40 p-3 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 resize-none font-mono text-sm"
                placeholder="请输入下注字符串，支持复杂格式智能解析...&#10;&#10;示例：&#10;01,02,03,04/200&#10;大,单/300&#10;特碰01/500"
              ></textarea>
            </div>

            <!-- 操作按钮 -->
            <div class="flex space-x-3 mb-4">
              <button @click="parseInput" class="btn-primary text-white px-4 py-2 rounded-md hover:opacity-90 transition-opacity">
                手动解析
              </button>
              <button @click="clearInput" class="bg-gray-500 text-white px-4 py-2 rounded-md hover:bg-gray-600 transition-colors">
                清空
              </button>
              <button @click="addToParsedList" class="bg-green-600 text-white px-4 py-2 rounded-md hover:bg-green-700 transition-colors disabled:bg-gray-400 disabled:cursor-not-allowed" :disabled="!betInput.trim()">
                添加到已解析
              </button>
            </div>

            <!-- 开奖结果设置 -->
            <div class="mb-4 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
              <h3 class="text-sm font-semibold text-gray-700 mb-3">开奖结果设置</h3>
              <div class="space-y-2">
                <div v-for="result in lotteryResults" :key="result.type" class="grid grid-cols-3 gap-2 text-sm">
                  <span class="text-sm font-medium flex items-center" :class="{
                    'text-blue-600': result.type === 'new_macau',
                    'text-red-600': result.type === 'hongkong', 
                    'text-green-600': result.type === 'old_macau'
                  }">{{ result.name }}:</span>
                  <input v-model="result.numbers" type="text" placeholder="开奖号码(逗号分隔,最后为特码)" class="p-2 border rounded focus:ring-blue-500 focus:border-blue-500 col-span-2">
                </div>
              </div>
              <div class="flex space-x-2 mt-3">
                <button @click="settleBets" class="btn-success text-white px-4 py-1 rounded text-sm">
                  结算所有下注
                </button>
                <button @click="exportToExcel" class="bg-indigo-600 text-white px-4 py-1 rounded text-sm hover:bg-indigo-700 transition-colors">
                  导出Excel
                </button>
                <button @click="clearLotteryResults" class="bg-gray-500 text-white px-4 py-1 rounded text-sm hover:bg-gray-600 transition-colors">
                  清空开奖
                </button>
              </div>
            </div>

            <!-- 实时预览 -->
            <div class="flex-1 border border-gray-200 rounded-md p-3 bg-gray-50 overflow-y-auto">
              <h3 class="text-sm font-semibold text-gray-700 mb-2">实时预览</h3>
              

              <div v-if="previewResult && !previewResult.hasError" class="text-sm">
                <!-- 成功解析的表格显示 -->
                <div class="bg-white rounded border overflow-hidden">
                  <table class="w-full text-xs">
                    <thead class="bg-gray-100">
                      <tr>
                        <th class="px-2 py-1 text-left border-r">体彩</th>
                        <th class="px-2 py-1 text-left border-r">下注类别</th>
                        <th class="px-2 py-1 text-center border-r">组数</th>
                        <th class="px-2 py-1 text-center border-r">金额</th>
                        <th class="px-2 py-1 text-center">详情</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(lotteryStats, lottery) in getLotteryStats(previewResult)" :key="lottery" class="border-t">
                        <td class="px-2 py-1 border-r font-medium" :class="{
                          'text-blue-600': lottery === '新澳',
                          'text-green-600': lottery === '老澳', 
                          'text-red-600': lottery === '香港'
                        }">{{ lottery }}</td>
                        <td class="px-2 py-1 border-r">
                          <div v-for="(betType, index) in lotteryStats.betTypes" :key="index" class="text-xs">
                            {{ betType.type }} ({{ betType.groups }}组)
                          </div>
                        </td>
                        <td class="px-2 py-1 text-center border-r font-mono">{{ lotteryStats.totalGroups }}</td>
                        <td class="px-2 py-1 text-center border-r font-mono">{{ lotteryStats.totalAmount }}元</td>
                        <td class="px-2 py-1 text-center">
                          <div v-for="(betType, index) in lotteryStats.betTypes" :key="index" class="mb-1">
                            <button @click="showBetDetail(lottery, betType.type, betType.info)" 
                                    class="px-2 py-1 bg-blue-500 text-white text-xs rounded hover:bg-blue-600 transition-colors">
                              详情
                            </button>
                          </div>
                        </td>
                      </tr>
                      <tr class="border-t bg-blue-50 font-semibold">
                        <td colspan="2" class="px-2 py-1 border-r">总计</td>
                        <td class="px-2 py-1 text-center border-r font-mono">{{ getTotalGroups(previewResult) }}</td>
                        <td class="px-2 py-1 text-center border-r font-mono text-blue-600">{{ getTotalAmount(previewResult) }}元</td>
                        <td class="px-2 py-1 text-center">-</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                
                <!-- 解析时间显示 -->
                <div class="mt-2 text-xs text-gray-500 text-center">
                  解析时间: {{ formatParseTime(previewResult) }}
                </div>
              </div>
              
              <div v-else-if="previewResult && previewResult.hasError" class="text-sm">
                <!-- 解析失败的错误表格显示 -->
                <div class="bg-red-50 border border-red-200 rounded overflow-hidden">
                  <div class="bg-red-100 px-3 py-2 border-b border-red-200">
                    <h4 class="text-red-800 font-semibold text-sm">⚠️ 解析失败</h4>
                  </div>
                  <table class="w-full text-xs">
                    <thead class="bg-red-100">
                      <tr>
                        <th class="px-2 py-1 text-left border-r border-red-200">笔数</th>
                        <th class="px-2 py-1 text-left border-r border-red-200">错误类型</th>
                        <th class="px-2 py-1 text-left">原始内容</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(error, index) in getErrorDetails(previewResult)" :key="index" class="border-t border-red-200">
                        <td class="px-2 py-1 border-r border-red-200 font-mono">第{{ index + 1 }}笔</td>
                        <td class="px-2 py-1 border-r border-red-200 text-red-600">{{ error.type }}</td>
                        <td class="px-2 py-1 font-mono text-gray-700">{{ error.content }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                
                <!-- 错误详情 -->
                <div class="mt-2 p-2 bg-yellow-50 border border-yellow-200 rounded">
                  <div class="text-xs text-gray-600">
                    <strong>建议：</strong>请检查输入格式，确保包含有效的号码和金额信息
                  </div>
                </div>
              </div>
              
              <div v-else class="text-gray-500 text-center py-8">
                输入下注内容查看实时预览
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧：已解析下注列表 -->
      <div class="h-full w-[65%] flex flex-col border-l bg-white">
        <!-- 已解析列表标题 -->
        <div class="p-4 border-b bg-gray-50">
          <div class="flex justify-between items-center">
            <h3 class="text-lg font-bold text-gray-800">已解析下注列表</h3>
            <div class="flex space-x-2">
              <span class="text-sm text-gray-600">总计: {{ parsedBets.length }} 笔</span>
              <button @click="clearParsedBets" class="text-red-600 hover:text-red-800 text-sm">
                清空列表
              </button>
            </div>
          </div>
        </div>

        <!-- 表格头部 -->
        <div class="bg-gray-100 border-b text-sm font-semibold text-gray-700 p-3">
          <div class="grid grid-cols-12 gap-3">
            <div class="col-span-3">原内容</div>
            <div class="col-span-3">格式化内容</div>
            <div class="col-span-2">下注总额</div>
            <div class="col-span-1">水钱</div>
            <div class="col-span-2">下注详情</div>
            <div class="col-span-1">操作</div>
          </div>
        </div>

        <!-- 表格内容 -->
        <div class="flex-1 overflow-y-auto">
          <div v-for="(bet, index) in parsedBets" :key="index" class="border-b text-sm p-3 hover:bg-gray-50">
            <div class="grid grid-cols-12 gap-3 items-center">
              <div class="col-span-3 truncate" :title="bet.original">{{ bet.original }}</div>
              <div class="col-span-3 truncate" :title="bet.formatted">{{ bet.formatted }}</div>
              <div class="col-span-2 text-right font-semibold text-blue-600">{{ bet.totalAmount }}元</div>
              <div class="col-span-1 text-right text-gray-600">{{ bet.commission }}元</div>
              <div class="col-span-2">
                <div class="text-xs space-y-1">
                  <div v-if="bet.details?.newMacau?.totalGroups > 0" class="text-blue-600">
                    新澳: {{ bet.details.newMacau.totalGroups }}组 / {{ bet.details.newMacau.totalAmount }}分
                  </div>
                  <div v-if="bet.details?.oldMacau?.totalGroups > 0" class="text-green-600">
                    老澳: {{ bet.details.oldMacau.totalGroups }}组 / {{ bet.details.oldMacau.totalAmount }}分
                  </div>
                  <div v-if="bet.details?.hongkong?.totalGroups > 0" class="text-red-600">
                    香港: {{ bet.details.hongkong.totalGroups }}组 / {{ bet.details.hongkong.totalAmount }}分
                  </div>
                  <div v-if="!bet.details || (!bet.details.newMacau?.totalGroups && !bet.details.oldMacau?.totalGroups && !bet.details.hongkong?.totalGroups)" class="text-gray-500">
                    总积分: {{ Object.values(bet.amounts).reduce((sum, v) => sum + (v || 0), 0) }}
                  </div>
                </div>
              </div>
              <div class="col-span-1">
                <button @click="showBetDetails(bet)" class="bg-blue-500 text-white px-2 py-1 rounded text-xs hover:bg-blue-600 transition-colors">
                  详情
                </button>
              </div>
            </div>
          </div>
          <div v-if="parsedBets.length === 0" class="text-center text-gray-500 py-12">
            暂无已解析的下注记录
          </div>
        </div>

        <!-- 统计信息 -->
        <div class="border-t bg-gray-50 p-3">
          <div class="grid grid-cols-4 gap-2 text-sm">
            <div class="text-center p-2 bg-blue-100 rounded">
              <div class="font-bold text-blue-600">{{ betStatistics.totalBets }}</div>
              <div class="text-gray-600 text-xs">总笔数</div>
            </div>
            <div class="text-center p-2 bg-green-100 rounded">
              <div class="font-bold text-green-600">{{ betStatistics.totalAmount }}</div>
              <div class="text-gray-600 text-xs">总金额</div>
            </div>
            <div class="text-center p-2 bg-yellow-100 rounded">
              <div class="font-bold text-yellow-600">{{ betStatistics.totalCommission }}</div>
              <div class="text-gray-600 text-xs">总水钱</div>
            </div>
            <div class="text-center p-2 bg-purple-100 rounded">
              <div class="font-bold text-purple-600">{{ betStatistics.avgAmount }}</div>
              <div class="text-gray-600 text-xs">平均金额</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部免责声明 -->
    <div class="bg-gray-800 text-white p-2 text-xs text-center">
      <span>⚠️ 本软件仅供学习和个人研究使用，请勿用于非法用途 ⚠️</span>
    </div>

    <!-- 下注详情弹窗 -->
    <div v-if="showDetailsModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click="closeDetailsModal">
      <div class="bg-white rounded-lg p-6 m-4 max-w-4xl w-full max-h-[80vh] overflow-y-auto" @click.stop>
        <div class="flex justify-between items-center mb-4">
          <h3 class="text-xl font-bold text-gray-800">下注详情</h3>
          <button @click="closeDetailsModal" class="text-gray-500 hover:text-gray-700 text-2xl">&times;</button>
        </div>
        
        <div v-if="selectedBet" class="space-y-4">
          <!-- 基本信息 -->
          <div class="bg-gray-50 p-4 rounded-md">
            <h4 class="font-semibold text-gray-700 mb-2">基本信息</h4>
            <div class="grid grid-cols-2 gap-4 text-sm">
              <div><span class="text-gray-600">原内容：</span>{{ selectedBet.original }}</div>
              <div><span class="text-gray-600">格式化内容：</span>{{ selectedBet.formatted }}</div>
              <div><span class="text-gray-600">下注总额：</span><span class="font-semibold text-blue-600">{{ selectedBet.totalAmount }}元</span></div>
              <div><span class="text-gray-600">水钱：</span><span class="font-semibold text-yellow-600">{{ selectedBet.commission }}元</span></div>
            </div>
          </div>

          <!-- 详细下注分析 -->
          <div class="grid grid-cols-3 gap-4">
            <!-- 新澳门 -->
            <div class="bg-blue-50 border border-blue-200 rounded-md p-4">
              <h5 class="font-semibold text-blue-700 mb-3 text-center">新澳门</h5>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                  <span>3中3：</span>
                  <span class="font-semibold">{{ selectedBet.details?.newMacau?.groups['3of3'] || 0 }}组 / {{ selectedBet.amounts.new_3of3 || 0 }}积分</span>
                </div>
                <div class="flex justify-between">
                  <span>3中2：</span>
                  <span class="font-semibold">{{ selectedBet.details?.newMacau?.groups['3of2'] || 0 }}组 / {{ selectedBet.amounts.new_3of2 || 0 }}积分</span>
                </div>
                <div class="flex justify-between">
                  <span>特碰：</span>
                  <span class="font-semibold">{{ selectedBet.details?.newMacau?.groups.special || 0 }}组 / {{ selectedBet.amounts.new_special || 0 }}积分</span>
                </div>
                <div class="border-t pt-2 flex justify-between font-semibold text-blue-700">
                  <span>小计：</span>
                  <span>{{ selectedBet.details?.newMacau?.totalGroups || 0 }}组 / {{ selectedBet.details?.newMacau?.totalAmount || 0 }}积分</span>
                </div>
              </div>
            </div>

            <!-- 老澳门 -->
            <div class="bg-green-50 border border-green-200 rounded-md p-4">
              <h5 class="font-semibold text-green-700 mb-3 text-center">老澳门</h5>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                  <span>2中2：</span>
                  <span class="font-semibold">{{ selectedBet.details?.oldMacau?.groups['2of2'] || 0 }}组 / {{ selectedBet.amounts.old_2of2 || 0 }}积分</span>
                </div>
                <div class="flex justify-between">
                  <span>3中3：</span>
                  <span class="font-semibold">{{ selectedBet.details?.oldMacau?.groups['3of3'] || 0 }}组 / {{ selectedBet.amounts.old_3of3 || 0 }}积分</span>
                </div>
                <div class="flex justify-between">
                  <span>3中2：</span>
                  <span class="font-semibold">{{ selectedBet.details?.oldMacau?.groups['3of2'] || 0 }}组 / {{ selectedBet.amounts.old_3of2 || 0 }}积分</span>
                </div>
                <div class="flex justify-between">
                  <span>特碰：</span>
                  <span class="font-semibold">{{ selectedBet.details?.oldMacau?.groups.special || 0 }}组 / {{ selectedBet.amounts.old_special || 0 }}积分</span>
                </div>
                <div class="border-t pt-2 flex justify-between font-semibold text-green-700">
                  <span>小计：</span>
                  <span>{{ selectedBet.details?.oldMacau?.totalGroups || 0 }}组 / {{ selectedBet.details?.oldMacau?.totalAmount || 0 }}积分</span>
                </div>
              </div>
            </div>

            <!-- 香港 -->
            <div class="bg-red-50 border border-red-200 rounded-md p-4">
              <h5 class="font-semibold text-red-700 mb-3 text-center">香港</h5>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                  <span>2中2：</span>
                  <span class="font-semibold">{{ selectedBet.details?.hongkong?.groups['2of2'] || 0 }}组 / {{ selectedBet.amounts.hk_2of2 || 0 }}积分</span>
                </div>
                <div class="flex justify-between">
                  <span>3中3：</span>
                  <span class="font-semibold">{{ selectedBet.details?.hongkong?.groups['3of3'] || 0 }}组 / {{ selectedBet.amounts.hk_3of3 || 0 }}积分</span>
                </div>
                <div class="flex justify-between">
                  <span>3中2：</span>
                  <span class="font-semibold">{{ selectedBet.details?.hongkong?.groups['3of2'] || 0 }}组 / {{ selectedBet.amounts.hk_3of2 || 0 }}积分</span>
                </div>
                <div class="flex justify-between">
                  <span>特碰：</span>
                  <span class="font-semibold">{{ selectedBet.details?.hongkong?.groups.special || 0 }}组 / {{ selectedBet.amounts.hk_special || 0 }}积分</span>
                </div>
                <div class="border-t pt-2 flex justify-between font-semibold text-red-700">
                  <span>小计：</span>
                  <span>{{ selectedBet.details?.hongkong?.totalGroups || 0 }}组 / {{ selectedBet.details?.hongkong?.totalAmount || 0 }}积分</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 总计 -->
          <div class="bg-gray-100 p-4 rounded-md">
            <div class="flex justify-between items-center text-lg font-bold">
              <span>总积分：</span>
              <span class="text-blue-600">{{ Object.values(selectedBet.amounts).reduce((sum, v) => sum + (v || 0), 0) }}积分</span>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 通知组件 -->
    <Notification ref="notification" />
    
    <!-- 下注详情弹窗 -->
    <BetDetailModal 
      :isVisible="detailModal.isVisible"
      :lottery="detailModal.lottery"
      :betType="detailModal.betType"
      :betTypeInfo="detailModal.betTypeInfo"
      @close="closeDetailModal"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { goApi } from '../api/goApi';
import Notification from '../components/Notification.vue';
import BetDetailModal from '../components/BetDetailModal.vue';

defineProps({
  isAuthorized: Boolean
});

// 数据定义
const betInput = ref('');
const parseResult = ref(null);
const previewResult = ref(null);
const parsedBets = ref([]);

// 下注详情弹窗状态
const detailModal = ref({
  isVisible: false,
  lottery: '',
  betType: '',
  betTypeInfo: {}
});

// 彩种启用状态
const enabledLotteries = ref({
  new_macau: true,
  old_macau: true,
  hongkong: true
});

// 开奖结果（固定三个彩种）
const lotteryResults = ref([
  { type: 'new_macau', name: '新澳门', numbers: '' },
  { type: 'hongkong', name: '香港', numbers: '' },
  { type: 'old_macau', name: '澳门', numbers: '' }
]);

// 弹窗状态
const showDetailsModal = ref(false);
const selectedBet = ref(null);

// 通知组件引用
const notification = ref(null);

// 统计数据计算
const betStatistics = computed(() => {
  const totalBets = parsedBets.value.length;
  const totalAmount = parsedBets.value.reduce((sum, bet) => sum + (bet.totalAmount || 0), 0);
  const totalCommission = parsedBets.value.reduce((sum, bet) => sum + (bet.commission || 0), 0);
  const avgAmount = totalBets > 0 ? Math.round(totalAmount / totalBets) : 0;

  return {
    totalBets,
    totalAmount,
    totalCommission,
    avgAmount
  };
});

// 监听输入变化进行实时预览
watch(betInput, (newValue) => {
  onBetInputChange();
}, { immediate: true });

// 方法定义
// 通用解析函数 - 复用解析逻辑
const performBetParsing = async (showNotification = false) => {
  if (!betInput.value.trim()) {
    previewResult.value = null;
    return null;
  }

  try {
    // 获取启用的彩种
    const enabledTypes = Object.keys(enabledLotteries.value).filter(key => enabledLotteries.value[key]);
    
    // 调用智能解析API
    const result = await goApi.parseBetInputIntelligent(betInput.value, enabledTypes);
    
    if (result && !result.hasError) {
      // 生成格式化预览
      const formatted = result.parsedBets.map((bet, index) => {
        if (bet.hasError) {
          return `❌ 第${index + 1}笔: ${bet.errorMessage}`;
        }
        
        const lotteryInfo = [];
        for (const [lottery, info] of Object.entries(bet.lotteryBets)) {
          const betTypeInfo = [];
          for (const [betType, typeInfo] of Object.entries(info.betTypeDetails)) {
            betTypeInfo.push(`${betType}: ${typeInfo.totalGroups}组/${typeInfo.totalAmount}元`);
          }
          lotteryInfo.push(`${lottery}(${betTypeInfo.join(', ')})`);
        }
        
        return `✅ 第${index + 1}笔: ${lotteryInfo.join(' | ')}`;
      }).join('\n');
      
      // 生成详细统计
      const stats = result.roundStatistics;
      let details = `解析结果: ${stats.totalBets}笔下注
总金额: ${stats.totalAmount}元
总组数: ${stats.totalGroups}组

各体彩统计:`;
      
      for (const [lottery, total] of Object.entries(stats.lotteryTotals || {})) {
        details += `\n${lottery}: ${total.groups}组, ${total.amount}元`;
        
        if (stats.lotteryBetTypeStats && stats.lotteryBetTypeStats[lottery]) {
          for (const [betType, stat] of Object.entries(stats.lotteryBetTypeStats[lottery])) {
            details += `\n  - ${betType}: ${stat.groups}组, ${stat.amount}元`;
          }
        }
      }

      previewResult.value = {
        formatted,
        details,
        parseData: result,
        hasError: false
      };
      
      // 显示通知（如果需要）
      if (showNotification && notification.value) {
        notification.value.show('智能解析成功', 
          `成功解析${stats.totalBets}笔下注，总金额${stats.totalAmount}元`, 'success');
      }
      
      return result;
    } else {
      const errorMsg = result?.errorMessages?.join('; ') || '解析失败';
      previewResult.value = {
        formatted: `❌ 解析失败: ${errorMsg}`,
        details: errorMsg,
        parseData: result,
        hasError: true
      };
      
      if (showNotification && notification.value) {
        notification.value.show('智能解析失败', errorMsg, 'error');
      }
      
      return result;
    }
  } catch (error) {
    const errorMsg = showNotification ? '智能解析错误: ' + error.message : '正在分析输入格式...';
    
    previewResult.value = {
      formatted: showNotification ? `❌ 解析失败: ${error.message}` : '解析中...',
      details: errorMsg,
      hasError: true
    };
    
    if (showNotification && notification.value) {
      notification.value.show('智能解析失败', '智能解析错误: ' + error.message, 'error');
    }
    
    return null;
  }
};

// 实时预览（静默）
const onBetInputChange = async () => {
  await performBetParsing(false); // 不显示通知
};

// 移除了autoDetectLotteryTypes函数 - 用户要求删除自动检测逻辑

// 手动解析（显示通知）
const parseInput = async () => {
  if (!betInput.value.trim()) {
    return;
  }
  
  const result = await performBetParsing(true); // 显示通知
  parseResult.value = result;
};

const addToParsedList = async () => {
  // 检查是否有输入内容
  if (!betInput.value.trim()) {
    if (notification.value) {
      notification.value.show('提示', '请先输入下注内容', 'warning');
    }
    return;
  }

  try {
    // 确保有解析结果，如果没有则重新解析
    let result = parseResult.value;
    if (!result) {
      const enabledTypes = Object.keys(enabledLotteries.value).filter(key => enabledLotteries.value[key]);
      result = await goApi.parseBetInputIntelligent(betInput.value, enabledTypes);
    }

    // 严格检查解析是否成功
    if (!result || result.hasError || !result.roundStatistics || result.roundStatistics.totalAmount <= 0) {
      // 解析失败，显示详细错误信息
      let errorMsg = '解析失败：';
      
      if (result?.errorMessages && result.errorMessages.length > 0) {
        errorMsg += result.errorMessages.join('; ');
      } else if (result?.parsedBets) {
        const errorBets = result.parsedBets.filter(bet => bet.hasError);
        if (errorBets.length > 0) {
          errorMsg += `发现${errorBets.length}笔无效下注。`;
          errorBets.forEach((bet, index) => {
            errorMsg += `\n第${index + 1}笔: ${bet.errorMessage} (内容: ${bet.originalText})`;
          });
        } else {
          errorMsg += '未能识别有效的下注格式';
        }
      } else {
        errorMsg += '无法识别下注格式，请检查输入内容';
      }
      
      if (notification.value) {
        notification.value.show('添加失败', errorMsg, 'error');
      }
      return;
    }

    // 解析成功，添加到列表
      const newBet = {
        id: Date.now(),
        original: betInput.value,
      formatted: `轮次${result.roundID}: ${result.roundStatistics.totalBets}笔下注`,
      totalAmount: result.roundStatistics.totalAmount,
      commission: Math.round(result.roundStatistics.totalAmount * 0.05),
      amounts: extractBetAmountsFromIntelligentAPI(result),
      details: analyzeBetDetailsFromIntelligentAPI(result),
        timestamp: new Date().toISOString(),
        parseData: result
      };

      parsedBets.value.push(newBet);
      
      if (notification.value) {
      notification.value.show('添加成功', 
        `已添加${result.roundStatistics.totalBets}笔下注，总金额${result.roundStatistics.totalAmount}元`, 'success');
    }
    
    // 清空输入
    betInput.value = '';
    parseResult.value = null;
    previewResult.value = null;
    
  } catch (error) {
    console.error('添加下注失败:', error);
    if (notification.value) {
      notification.value.show('添加失败', '处理下注时发生错误: ' + error.message, 'error');
    }
  }
};

// 删除了不再使用的旧版解析方法：
// - createDefaultBetData
// - calculateTotalAmount  
// - calculateCommission
// - extractBetAmounts

// 清理了旧版API结果提取方法：
// - extractBetAmountsFromAPI (旧版本，兼容旧解析API)
// - analyzeBetDetails (旧版本)  
// - analyzeBetDetailsFromAPI (旧版本，兼容旧解析API)

// 从智能解析API结果提取下注金额
const extractBetAmountsFromIntelligentAPI = (result) => {
  const amounts = {
    new_3of3: 0, new_3of2: 0, new_special: 0,
    old_2of2: 0, old_3of3: 0, old_3of2: 0, old_special: 0,
    hk_2of2: 0, hk_3of3: 0, hk_3of2: 0, hk_special: 0
  };

  if (!result?.roundStatistics?.lotteryBetTypeStats) {
    return amounts;
  }
  
  const stats = result.roundStatistics.lotteryBetTypeStats;
  
  // 新澳
  if (stats['新澳']) {
    amounts.new_3of3 = stats['新澳']['三中三']?.amount || 0;
    amounts.new_3of2 = stats['新澳']['三中二']?.amount || 0;
    amounts.new_special = stats['新澳']['特碰']?.amount || 0;
  }
  
  // 老澳
  if (stats['老澳']) {
    amounts.old_2of2 = stats['老澳']['二中二']?.amount || 0;
    amounts.old_3of3 = stats['老澳']['三中三']?.amount || 0;
    amounts.old_3of2 = stats['老澳']['三中二']?.amount || 0;
    amounts.old_special = stats['老澳']['特碰']?.amount || 0;
  }
  
  // 香港
  if (stats['香港']) {
    amounts.hk_2of2 = stats['香港']['二中二']?.amount || 0;
    amounts.hk_3of3 = stats['香港']['三中三']?.amount || 0;
    amounts.hk_3of2 = stats['香港']['三中二']?.amount || 0;
    amounts.hk_special = stats['香港']['特碰']?.amount || 0;
  }
  
  return amounts;
};

// 从智能解析API结果分析下注详情
const analyzeBetDetailsFromIntelligentAPI = (result) => {
  if (!result?.roundStatistics) {
    return {
      newMacau: { totalAmount: 0, totalGroups: 0, groups: {} },
      oldMacau: { totalAmount: 0, totalGroups: 0, groups: {} },
      hongkong: { totalAmount: 0, totalGroups: 0, groups: {} }
    };
  }

  const stats = result.roundStatistics;
  
  return {
    newMacau: {
      totalAmount: stats.lotteryTotals['新澳']?.amount || 0,
      totalGroups: stats.lotteryTotals['新澳']?.groups || 0,
      groups: {
        '3of3': stats.lotteryBetTypeStats['新澳']?.['三中三']?.groups || 0,
        '3of2': stats.lotteryBetTypeStats['新澳']?.['三中二']?.groups || 0,
        'special': stats.lotteryBetTypeStats['新澳']?.['特碰']?.groups || 0
      }
    },
    oldMacau: {
      totalAmount: stats.lotteryTotals['老澳']?.amount || 0,
      totalGroups: stats.lotteryTotals['老澳']?.groups || 0,
      groups: {
        '2of2': stats.lotteryBetTypeStats['老澳']?.['二中二']?.groups || 0,
        '3of3': stats.lotteryBetTypeStats['老澳']?.['三中三']?.groups || 0,
        '3of2': stats.lotteryBetTypeStats['老澳']?.['三中二']?.groups || 0,
        'special': stats.lotteryBetTypeStats['老澳']?.['特碰']?.groups || 0
      }
    },
    hongkong: {
      totalAmount: stats.lotteryTotals['香港']?.amount || 0,
      totalGroups: stats.lotteryTotals['香港']?.groups || 0,
      groups: {
        '2of2': stats.lotteryBetTypeStats['香港']?.['二中二']?.groups || 0,
        '3of3': stats.lotteryBetTypeStats['香港']?.['三中三']?.groups || 0,
        '3of2': stats.lotteryBetTypeStats['香港']?.['三中二']?.groups || 0,
        'special': stats.lotteryBetTypeStats['香港']?.['特碰']?.groups || 0
      }
    }
  };
};

// 表格化预览的辅助方法
const getLotteryStats = (result) => {
  if (!result?.parseData?.roundStatistics?.lotteryTotals) {
    return {};
  }
  
  const stats = {};
  const roundStats = result.parseData.roundStatistics;
  const lotteryBetTypeStats = roundStats.lotteryBetTypeStats || {};
  
  for (const [lottery, total] of Object.entries(roundStats.lotteryTotals)) {
    const betTypes = [];
    if (lotteryBetTypeStats[lottery]) {
      for (const [betType, betStat] of Object.entries(lotteryBetTypeStats[lottery])) {
        // 从第一笔下注中获取详细的betTypeInfo
        let betTypeInfo = null;
        if (result.parseData.parsedBets) {
          for (const bet of result.parseData.parsedBets) {
            if (bet.lotteryBets?.[lottery]?.betTypeDetails?.[betType]) {
              betTypeInfo = bet.lotteryBets[lottery].betTypeDetails[betType];
              break;
            }
          }
        }
        
        betTypes.push({
          type: betType,
          groups: betStat.groups,
          amount: betStat.amount,
          info: betTypeInfo
        });
      }
    }
    
    stats[lottery] = {
      totalGroups: total.groups,
      totalAmount: total.amount,
      betTypes: betTypes
    };
  }

  return stats;
};

const getTotalGroups = (result) => {
  return result?.parseData?.roundStatistics?.totalGroups || 0;
};

const getTotalAmount = (result) => {
  return result?.parseData?.roundStatistics?.totalAmount || 0;
};

const getErrorDetails = (result) => {
  if (!result?.parseData?.parsedBets) {
    return [{
      type: "系统错误",
      content: result?.details || "未知错误"
    }];
  }
  
  const errors = [];
  result.parseData.parsedBets.forEach((bet, index) => {
    if (bet.hasError) {
      errors.push({
        type: bet.errorMessage || "解析错误",
        content: bet.originalText || "无内容"
      });
    }
  });
  
  // 如果没有具体的bet错误，显示整体错误
  if (errors.length === 0 && result.parseData.errorMessages?.length > 0) {
    result.parseData.errorMessages.forEach(msg => {
      errors.push({
        type: "格式错误",
        content: msg
      });
    });
  }
  
  return errors.length > 0 ? errors : [{
    type: "解析失败",
    content: "无法识别下注格式"
  }];
};

const formatParseTime = (result) => {
  const parseTime = result?.parseData?.parseTime || result?.parseTime;
  if (!parseTime) {
    return "未知";
  }
  
  try {
    const parseTime = new Date(result.parseTime);
    return parseTime.toLocaleTimeString();
  } catch (e) {
    return "未知";
  }
};

// 弹窗相关方法
const showBetDetail = (lottery, betType, betTypeInfo) => {
  if (!betTypeInfo) {
    console.warn('缺少下注类型信息', { lottery, betType });
    return;
  }
  
  detailModal.value = {
    isVisible: true,
    lottery: lottery,
    betType: betType,
    betTypeInfo: betTypeInfo
  };
};

const closeDetailModal = () => {
  detailModal.value.isVisible = false;
};

const clearInput = () => {
  betInput.value = '';
  parseResult.value = null;
  previewResult.value = null;
};

const clearParsedBets = async () => {
  if (notification.value) {
    const confirmed = await notification.value.show('确认清空', '确定要清空所有已解析的下注记录吗？', 'warning', true);
    if (confirmed) {
      parsedBets.value = [];
      notification.value.show('清空完成', '已清空所有下注记录', 'success');
    }
  }
};

const showBetDetails = (bet) => {
  selectedBet.value = bet;
  showDetailsModal.value = true;
};

const closeDetailsModal = () => {
  showDetailsModal.value = false;
  selectedBet.value = null;
};

const clearLotteryResults = () => {
  lotteryResults.value.forEach(result => {
    result.numbers = '';
  });
};

const settleBets = async () => {
  // 检查是否有有效的开奖结果
  const validResults = lotteryResults.value.filter(result => result.numbers.trim());
  if (validResults.length === 0) {
    if (notification.value) {
      notification.value.show('提示', '请至少设置一个开奖结果', 'warning');
    }
    return;
  }

  if (parsedBets.value.length === 0) {
    if (notification.value) {
      notification.value.show('提示', '没有需要结算的下注记录', 'warning');
    }
    return;
  }

  try {
    // 处理多个开奖结果
    const processedResults = [];
    for (const result of validResults) {
      const numbers = result.numbers.split(',').map(n => n.trim());
      if (numbers.length < 7) {
        if (notification.value) {
          notification.value.show('格式错误', `${result.name} 开奖号码不完整，请输入7个号码（前6个为平码，最后1个为特码）`, 'error');
        }
        return;
      }

      processedResults.push({
        type: result.type,
        name: result.name,
        special: numbers[numbers.length - 1], // 最后一个为特码
        regular: numbers.slice(0, -1), // 前面的为平码
      });
    }

    const settlementData = {
      lotteryResults: processedResults,
      bets: parsedBets.value
    };

    const result = await goApi.settleBets(settlementData);
    
    // 更新下注记录的结算结果
    parsedBets.value = parsedBets.value.map(bet => ({
      ...bet,
      settled: true,
      winAmount: result.settlements[bet.id]?.winAmount || 0,
      winLoss: result.settlements[bet.id]?.winLoss || 0,
      winDetails: result.settlements[bet.id]?.details || {}
    }));

    if (notification.value) {
      notification.value.show('结算完成', `总中奖金额: ${result.totalWinAmount || 0}元\n结算彩种数: ${processedResults.length}个`, 'success');
    }
  } catch (error) {
    console.error('结算失败:', error);
    if (notification.value) {
      notification.value.show('结算失败', '结算失败: ' + error.message, 'error');
    }
  }
};

const exportToExcel = async () => {
  if (parsedBets.value.length === 0) {
    if (notification.value) {
      notification.value.show('提示', '没有数据可以导出', 'warning');
    }
    return;
  }

  try {
    // 处理多个开奖结果
    const validResults = lotteryResults.value.filter(result => result.type && result.numbers);
    const processedResults = validResults.map(result => {
      const numbers = result.numbers.split(',').map(n => n.trim());
      return {
        type: result.type,
        special: numbers.length > 0 ? numbers[numbers.length - 1] : '',
        regular: numbers.length > 1 ? numbers.slice(0, -1) : [],
      };
    });
    
    const exportData = {
      lotteryResults: processedResults,
      bets: parsedBets.value,
      statistics: betStatistics.value,
      exportTime: new Date().toISOString()
    };

    await goApi.exportToExcel(exportData);
    if (notification.value) {
      notification.value.show('导出成功', '数据已成功导出到Excel文件', 'success');
    }
  } catch (error) {
    console.error('导出失败:', error);
    if (notification.value) {
      notification.value.show('导出失败', '导出失败: ' + error.message, 'error');
    }
  }
};
</script>

<style scoped>
.font-inter {
  font-family: "Inter", sans-serif;
}

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