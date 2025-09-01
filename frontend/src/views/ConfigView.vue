<template>
  <div class="flex flex-col h-full bg-gray-100 text-gray-800 font-inter overflow-hidden">
    <!-- 配置内容 -->
    <div class="flex-1 p-6 overflow-y-auto">
      <div class="max-w-6xl mx-auto">
        <h1 class="text-3xl font-bold text-gray-800 mb-6">系统配置</h1>
        
        <!-- 配置选项卡 -->
        <div class="bg-white rounded-lg shadow-md">
          <!-- 选项卡导航 -->
          <div class="border-b border-gray-200">
            <nav class="flex space-x-8 px-6">
              <button 
                v-for="tab in configTabs" 
                :key="tab.key"
                @click="activeTab = tab.key"
                :class="[
                  'py-4 px-1 border-b-2 font-medium text-sm transition-colors',
                  activeTab === tab.key 
                    ? 'border-blue-500 text-blue-600' 
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                ]"
              >
                {{ tab.name }}
              </button>
            </nav>
          </div>

          <!-- 选项卡内容 -->
          <div class="p-6">
            <!-- 12生肖配置 -->
            <div v-if="activeTab === 'zodiac'" class="space-y-6">
              <h2 class="text-xl font-semibold text-gray-800 mb-4">12生肖数字配置</h2>
              <div class="grid grid-cols-3 gap-6">
                <div v-for="zodiac in zodiacConfig" :key="zodiac.name" class="bg-gray-50 p-4 rounded-md">
                  <label class="block text-sm font-medium text-gray-700 mb-2">{{ zodiac.name }}</label>
                  <input 
                    v-model="zodiac.numbers" 
                    type="text" 
                    placeholder="用逗号分隔数字" 
                    :class="[
                      'w-full p-2 border rounded-md focus:ring-blue-500 focus:border-blue-500',
                      getFieldError('zodiac', zodiac.key) ? 'border-red-500 bg-red-50' : 'border-gray-300'
                    ]"
                    @blur="validateOnBlur('zodiac', zodiac.key, zodiac.numbers, zodiac.name)"
                  >
                  <div v-if="getFieldError('zodiac', zodiac.key)" class="text-xs text-red-600 mt-1">
                    {{ getFieldError('zodiac', zodiac.key) }}
                  </div>
                  <div v-else class="text-xs text-gray-500 mt-1">
                    当前: {{ zodiac.numbers || '未设置' }}
                  </div>
                </div>
              </div>
              <div class="flex space-x-3">
                <button @click="saveZodiacConfig" class="btn-primary text-white px-6 py-2 rounded-md">
                  保存生肖配置
                </button>
                <button @click="resetZodiacConfig" class="bg-gray-500 text-white px-6 py-2 rounded-md hover:bg-gray-600">
                  重置为默认
                </button>
              </div>
            </div>

            <!-- 颜色波段配置 -->
            <div v-if="activeTab === 'colors'" class="space-y-6">
              <h2 class="text-xl font-semibold text-gray-800 mb-4">红绿蓝波数字配置</h2>
              <div class="grid grid-cols-3 gap-6">
                <div v-for="color in colorConfig" :key="color.name" class="border rounded-md p-4" :class="color.bgClass">
                  <label class="block text-sm font-medium mb-2" :class="color.textClass">{{ color.name }}</label>
                  <input 
                    v-model="color.numbers" 
                    type="text" 
                    placeholder="用逗号分隔数字" 
                    :class="[
                      'w-full p-2 border rounded-md focus:ring-blue-500 focus:border-blue-500',
                      getFieldError('colors', color.key) ? 'border-red-500 bg-red-50' : 'border-gray-300'
                    ]"
                    @blur="validateOnBlur('colors', color.key, color.numbers, color.name)"
                  >
                  <div v-if="getFieldError('colors', color.key)" class="text-xs text-red-600 mt-1">
                    {{ getFieldError('colors', color.key) }}
                  </div>
                  <div v-else class="text-xs mt-1 opacity-75" :class="color.textClass">
                    当前: {{ color.numbers || '未设置' }}
                  </div>
                </div>
              </div>
              <div class="flex space-x-3">
                <button @click="saveColorConfig" class="btn-primary text-white px-6 py-2 rounded-md">
                  保存颜色配置
                </button>
                <button @click="resetColorConfig" class="bg-gray-500 text-white px-6 py-2 rounded-md hover:bg-gray-600">
                  重置为默认
                </button>
              </div>
            </div>

            <!-- 尾数配置 -->
            <div v-if="activeTab === 'tails'" class="space-y-6">
              <h2 class="text-xl font-semibold text-gray-800 mb-4">0~9尾数字配置</h2>
              <div class="grid grid-cols-5 gap-4">
                <div v-for="tail in tailConfig" :key="tail.tail" class="bg-gray-50 p-4 rounded-md">
                  <label class="block text-sm font-medium text-gray-700 mb-2">{{ tail.tail }}尾</label>
                  <input 
                    v-model="tail.numbers" 
                    type="text" 
                    placeholder="逗号分隔" 
                    :class="[
                      'w-full p-2 border rounded-md focus:ring-blue-500 focus:border-blue-500',
                      getFieldError('tails', tail.key) ? 'border-red-500 bg-red-50' : 'border-gray-300'
                    ]"
                    @blur="validateOnBlur('tails', tail.key, tail.numbers, tail.tail + '尾')"
                  >
                  <div v-if="getFieldError('tails', tail.key)" class="text-xs text-red-600 mt-1">
                    {{ getFieldError('tails', tail.key) }}
                  </div>
                  <div v-else class="text-xs text-gray-500 mt-1">
                    {{ tail.numbers || '未设置' }}
                  </div>
                </div>
              </div>
              <div class="flex space-x-3">
                <button @click="saveTailConfig" class="btn-primary text-white px-6 py-2 rounded-md">
                  保存尾数配置
                </button>
                <button @click="resetTailConfig" class="bg-gray-500 text-white px-6 py-2 rounded-md hover:bg-gray-600">
                  重置为默认
                </button>
              </div>
            </div>

            <!-- 下注类型别名配置 -->
            <div v-if="activeTab === 'bet_types'" class="space-y-6">
              <h2 class="text-xl font-semibold text-gray-800 mb-4">下注类型别名配置</h2>
              <div class="grid grid-cols-2 gap-6">
                <div v-for="betType in betTypeConfig" :key="betType.type" class="bg-gray-50 p-4 rounded-md">
                  <label class="block text-sm font-medium text-gray-700 mb-2">{{ betType.name }}</label>
                  <input 
                    v-model="betType.aliases" 
                    type="text" 
                    placeholder="用逗号分隔别名" 
                    :class="[
                      'w-full p-2 border rounded-md focus:ring-blue-500 focus:border-blue-500',
                      getFieldError('betTypes', betType.type) ? 'border-red-500 bg-red-50' : 'border-gray-300'
                    ]"
                    @blur="validateOnBlur('betTypes', betType.type, betType.aliases, betType.name)"
                  >
                  <div v-if="getFieldError('betTypes', betType.type)" class="text-xs text-red-600 mt-1">
                    {{ getFieldError('betTypes', betType.type) }}
                  </div>
                  <div v-else class="text-xs text-gray-500 mt-1">
                    别名: {{ betType.aliases || '未设置' }}
                  </div>
                </div>
              </div>
              <div class="flex space-x-3">
                <button @click="saveBetTypeConfig" class="btn-primary text-white px-6 py-2 rounded-md">
                  保存别名配置
                </button>
                <button @click="resetBetTypeConfig" class="bg-gray-500 text-white px-6 py-2 rounded-md hover:bg-gray-600">
                  重置为默认
                </button>
              </div>
            </div>

            <!-- 赔率设置配置 -->
            <div v-if="activeTab === 'odds'" class="space-y-6">
              <h2 class="text-xl font-semibold text-gray-800 mb-4">赔率设置配置</h2>
              
              <!-- 三中三赔率 -->
              <div class="bg-gray-50 p-6 rounded-lg">
                <h3 class="text-lg font-medium text-gray-800 mb-4">三中三赔率</h3>
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">赔率</label>
                    <input 
                      v-model.number="oddsConfig.threeOfThree.oddsRatio" 
                      type="number" 
                      step="0.1" 
                      min="0" 
                      placeholder="请输入赔率" 
                      class="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                    >
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">回水率 (%)</label>
                    <input 
                      v-model.number="oddsConfig.threeOfThree.rebate" 
                      type="number" 
                      step="0.01" 
                      min="0" 
                      max="1" 
                      placeholder="请输入回水率"
                      class="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                    >
                  </div>
                </div>
              </div>
              
              <!-- 三中二赔率 -->
              <div class="bg-gray-50 p-6 rounded-lg">
                <h3 class="text-lg font-medium text-gray-800 mb-4">三中二赔率</h3>
                
                <!-- 中二个时的赔率 -->
                <div class="mb-6">
                  <h4 class="text-md font-medium text-gray-700 mb-3">中二个时的赔率</h4>
                  <div class="grid grid-cols-2 gap-4">
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-2">赔率</label>
                      <input 
                        v-model.number="oddsConfig.threeOfTwo.hitTwoOdds.oddsRatio" 
                        type="number" 
                        step="0.1" 
                        min="0" 
                        placeholder="请输入赔率" 
                        class="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                      >
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-2">回水率 (%)</label>
                      <input 
                        v-model.number="oddsConfig.threeOfTwo.hitTwoOdds.rebate" 
                        type="number" 
                        step="0.01" 
                        min="0" 
                        max="1" 
                        placeholder="请输入回水率"
                        class="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                      >
                    </div>
                  </div>
                </div>
                
                <!-- 中三个时的赔率 -->
                <div>
                  <h4 class="text-md font-medium text-gray-700 mb-3">中三个时的赔率</h4>
                  <div class="grid grid-cols-2 gap-4">
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-2">赔率</label>
                      <input 
                        v-model.number="oddsConfig.threeOfTwo.hitThreeOdds.oddsRatio" 
                        type="number" 
                        step="0.1" 
                        min="0" 
                        placeholder="请输入赔率" 
                        class="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                      >
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-2">回水率 (%)</label>
                      <input 
                        v-model.number="oddsConfig.threeOfTwo.hitThreeOdds.rebate" 
                        type="number" 
                        step="0.01" 
                        min="0" 
                        max="1" 
                        placeholder="请输入回水率"
                        class="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                      >
                    </div>
                  </div>
                </div>
              </div>
              
              <!-- 二中二赔率 -->
              <div class="bg-gray-50 p-6 rounded-lg">
                <h3 class="text-lg font-medium text-gray-800 mb-4">二中二赔率</h3>
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">赔率</label>
                    <input 
                      v-model.number="oddsConfig.twoOfTwo.oddsRatio" 
                      type="number" 
                      step="0.1" 
                      min="0" 
                      placeholder="请输入赔率" 
                      class="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                    >
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">回水率 (%)</label>
                    <input 
                      v-model.number="oddsConfig.twoOfTwo.rebate" 
                      type="number" 
                      step="0.01" 
                      min="0" 
                      max="1" 
                      placeholder="请输入回水率"
                      class="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                    >
                  </div>
                </div>
              </div>
              
              <!-- 特碰赔率 -->
              <div class="bg-gray-50 p-6 rounded-lg">
                <h3 class="text-lg font-medium text-gray-800 mb-4">特碰赔率</h3>
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">赔率</label>
                    <input 
                      v-model.number="oddsConfig.special.oddsRatio" 
                      type="number" 
                      step="0.1" 
                      min="0" 
                      placeholder="请输入赔率" 
                      class="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                    >
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">回水率 (%)</label>
                    <input 
                      v-model.number="oddsConfig.special.rebate" 
                      type="number" 
                      step="0.01" 
                      min="0" 
                      max="1" 
                      placeholder="请输入回水率"
                      class="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                    >
                  </div>
                </div>
              </div>
              
              <div class="flex space-x-3">
                <button @click="saveOddsConfig" class="btn-primary text-white px-6 py-2 rounded-md">
                  保存赔率配置
                </button>
                <button @click="resetOddsConfig" class="bg-gray-500 text-white px-6 py-2 rounded-md hover:bg-gray-600">
                  重置为默认
                </button>
              </div>
            </div>

            <!-- 关键字别名配置 -->
            <div v-if="activeTab === 'keywords'" class="space-y-6">
              <h2 class="text-xl font-semibold text-gray-800 mb-4">关键字别名配置</h2>
              <div class="grid grid-cols-2 gap-6">
                <div v-for="keyword in keywordConfig" :key="keyword.type" class="bg-gray-50 p-4 rounded-md">
                  <label class="block text-sm font-medium text-gray-700 mb-2">{{ keyword.name }}</label>
                  <input 
                    v-model="keyword.aliases" 
                    type="text" 
                    placeholder="用逗号分隔别名" 
                    :class="[
                      'w-full p-2 border rounded-md focus:ring-blue-500 focus:border-blue-500',
                      getFieldError('keywords', keyword.type) ? 'border-red-500 bg-red-50' : 'border-gray-300'
                    ]"
                    @blur="validateOnBlur('keywords', keyword.type, keyword.aliases, keyword.name)"
                  >
                  <div v-if="getFieldError('keywords', keyword.type)" class="text-xs text-red-600 mt-1">
                    {{ getFieldError('keywords', keyword.type) }}
                  </div>
                  <div v-else class="text-xs text-gray-500 mt-1">
                    别名: {{ keyword.aliases || '未设置' }}
                  </div>
                </div>
              </div>
              <div class="flex space-x-3">
                <button @click="saveKeywordConfig" class="btn-primary text-white px-6 py-2 rounded-md">
                  保存关键字配置
                </button>
                <button @click="resetKeywordConfig" class="bg-gray-500 text-white px-6 py-2 rounded-md hover:bg-gray-600">
                  重置为默认
                </button>
              </div>
            </div>



          </div>
        </div>
      </div>
    </div>

    <!-- 底部免责声明 -->
    <div class="bg-gray-800 text-white p-2 text-xs text-center">
      <span>⚠️ 本软件仅供学习和个人研究使用，请勿用于非法用途 ⚠️</span>
    </div>
    
    <!-- 通知组件 -->
    <Notification ref="notification" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { goApi } from '../api/goApi';
import Notification from '../components/Notification.vue';

defineProps({
  isAuthorized: Boolean
});

// 当前活动选项卡
const activeTab = ref('zodiac');

// 通知组件引用
const notification = ref(null);

// 错误状态管理
const validationErrors = ref({
  zodiac: {},
  colors: {},
  tails: {},
  betTypes: {},
  keywords: {}
});

// 配置选项卡
const configTabs = [
  { key: 'zodiac', name: '12生肖' },
  { key: 'colors', name: '颜色波段' },
  { key: 'tails', name: '尾数配置' },
  { key: 'bet_types', name: '下注类型' },
  { key: 'odds', name: '赔率设置' },
  { key: 'keywords', name: '关键字别名' }
];

// 配置数据
const zodiacConfig = ref([
  { name: '鼠', key: 'rat', numbers: '' },
  { name: '牛', key: 'ox', numbers: '' },
  { name: '虎', key: 'tiger', numbers: '' },
  { name: '兔', key: 'rabbit', numbers: '' },
  { name: '龙', key: 'dragon', numbers: '' },
  { name: '蛇', key: 'snake', numbers: '' },
  { name: '马', key: 'horse', numbers: '' },
  { name: '羊', key: 'goat', numbers: '' },
  { name: '猴', key: 'monkey', numbers: '' },
  { name: '鸡', key: 'rooster', numbers: '' },
  { name: '狗', key: 'dog', numbers: '' },
  { name: '猪', key: 'pig', numbers: '' }
]);

const colorConfig = ref([
  { 
    name: '红波', 
    key: 'red',
    numbers: '', 
    bgClass: 'bg-red-50 border-red-200',
    textClass: 'text-red-700'
  },
  { 
    name: '绿波', 
    key: 'green',
    numbers: '', 
    bgClass: 'bg-green-50 border-green-200',
    textClass: 'text-green-700'
  },
  { 
    name: '蓝波', 
    key: 'blue',
    numbers: '', 
    bgClass: 'bg-blue-50 border-blue-200',
    textClass: 'text-blue-700'
  }
]);

const tailConfig = ref([
  { tail: '0', key: 'tail_0', numbers: '' },
  { tail: '1', key: 'tail_1', numbers: '' },
  { tail: '2', key: 'tail_2', numbers: '' },
  { tail: '3', key: 'tail_3', numbers: '' },
  { tail: '4', key: 'tail_4', numbers: '' },
  { tail: '5', key: 'tail_5', numbers: '' },
  { tail: '6', key: 'tail_6', numbers: '' },
  { tail: '7', key: 'tail_7', numbers: '' },
  { tail: '8', key: 'tail_8', numbers: '' },
  { tail: '9', key: 'tail_9', numbers: '' }
]);

const betTypeConfig = ref([
  { type: 'three_of_three', name: '三中三', aliases: '' },
  { type: 'three_of_two', name: '三中二', aliases: '' },
  { type: 'two_of_two', name: '二中二', aliases: '' },
  { type: 'special', name: '特碰', aliases: '' }
]);

const keywordConfig = ref([
  { type: 'new_macau', name: '新澳', aliases: '' },
  { type: 'old_macau', name: '老澳', aliases: '' },
  { type: 'hong_kong', name: '香港', aliases: '' },
  { type: 'complex', name: '复式', aliases: '' },
  { type: 'drag', name: '拖码', aliases: '' },
  { type: 'each', name: '各', aliases: '' },
  { type: 'per_group', name: '每组', aliases: '' }
]);

const oddsConfig = ref({
  threeOfThree: {
    oddsRatio: 175.0,
    rebate: 0.05
  },
  threeOfTwo: {
    hitTwoOdds: {
      oddsRatio: 7.5,
      rebate: 0.05
    },
    hitThreeOdds: {
      oddsRatio: 175.0,
      rebate: 0.05
    }
  },
  twoOfTwo: {
    oddsRatio: 7.5,
    rebate: 0.05
  },
  special: {
    oddsRatio: 40.0,
    rebate: 0.05
  }
});

// 加载所有配置数据
const loadAllConfigs = async () => {
  try {
    // 加载12生肖配置
    const zodiacData = await goApi.getZodiacConfig();
    zodiacConfig.value.forEach(item => {
      const numbers = zodiacData[item.key] || [];
      item.numbers = numbers.map(n => n.toString().padStart(2, '0')).join(',');
    });

    // 加载颜色配置
    const colorData = await goApi.getColorConfig();
    colorConfig.value.forEach(item => {
      const numbers = colorData[item.key] || [];
      item.numbers = numbers.map(n => n.toString().padStart(2, '0')).join(',');
    });

    // 加载尾数配置
    const tailData = await goApi.getTailConfig();
    tailConfig.value.forEach(item => {
      const numbers = tailData[item.key] || [];
      item.numbers = numbers.map(n => n.toString().padStart(2, '0')).join(',');
    });

    // 加载下注类型别名配置
    const betTypeData = await goApi.getBetTypeAliases();
    betTypeConfig.value.forEach(item => {
      const aliases = betTypeData[item.type] || [];
      item.aliases = aliases.join(',');
    });

    // 加载关键字别名配置
    const keywordData = await goApi.getKeywordAliases();
    keywordConfig.value.forEach(item => {
      const aliases = keywordData[item.type] || [];
      item.aliases = aliases.join(',');
    });

    // 加载赔率配置
    const oddsData = await goApi.getOddsConfig();
    if (oddsData && Object.keys(oddsData).length > 0) {
      oddsConfig.value = {
        threeOfThree: {
          oddsRatio: oddsData.three_of_three?.odds_ratio || 175.0,
          rebate: oddsData.three_of_three?.rebate || 0.05
        },
        threeOfTwo: {
          hitTwoOdds: {
            oddsRatio: oddsData.three_of_two?.hit_two_odds?.odds_ratio || 7.5,
            rebate: oddsData.three_of_two?.hit_two_odds?.rebate || 0.05
          },
          hitThreeOdds: {
            oddsRatio: oddsData.three_of_two?.hit_three_odds?.odds_ratio || 175.0,
            rebate: oddsData.three_of_two?.hit_three_odds?.rebate || 0.05
          }
        },
        twoOfTwo: {
          oddsRatio: oddsData.two_of_two?.odds_ratio || 7.5,
          rebate: oddsData.two_of_two?.rebate || 0.05
        },
        special: {
          oddsRatio: oddsData.special?.odds_ratio || 40.0,
          rebate: oddsData.special?.rebate || 0.05
        }
      };
    }

    console.log('所有配置加载成功');
  } catch (error) {
    console.error('加载配置失败:', error);
    if (notification.value) {
      notification.value.show('加载失败', '加载配置失败: ' + error.message, 'error');
    }
  }
};

// 保存12生肖配置
const saveZodiacConfig = async () => {
  try {
    // 检查是否有错误
    if (Object.keys(validationErrors.value.zodiac).length > 0) {
      if (notification.value) {
        notification.value.show('保存失败', '请先修正输入错误再保存', 'error');
      }
      return;
    }

    // 格式校验
    for (const item of zodiacConfig.value) {
      if (item.numbers.trim()) {
        const validation = validateConfigInput(item.numbers, '12生肖');
        if (!validation.isValid) {
          if (notification.value) {
            notification.value.show('配置格式错误', `${item.name}：${validation.message}`, 'error');
          }
          return;
        }
      }
    }
    
    const config = {};
    zodiacConfig.value.forEach(item => {
      if (item.numbers.trim()) {
        config[item.key] = item.numbers.split(',').map(n => parseInt(n.trim())).filter(n => !isNaN(n));
      }
    });
    
    await goApi.saveZodiacConfig(config);
    if (notification.value) {
      notification.value.show('保存成功', '12生肖配置保存成功！', 'success');
    }
  } catch (error) {
    console.error('保存失败:', error);
    if (notification.value) {
      notification.value.show('保存失败', '保存失败: ' + error.message, 'error');
    }
  }
};

// 重置12生肖配置
const resetZodiacConfig = async () => {
  if (notification.value) {
    const confirmed = await notification.value.show('确认重置', '确定要重置12生肖配置为默认值吗？', 'warning', true);
    if (confirmed) {
      try {
        await goApi.resetSystemConfig();
        await loadAllConfigs();
        notification.value.show('重置成功', '12生肖配置已重置为默认值！', 'success');
      } catch (error) {
        console.error('重置失败:', error);
        notification.value.show('重置失败', '重置失败: ' + error.message, 'error');
      }
    }
  }
};

// 保存颜色配置
const saveColorConfig = async () => {
  try {
    // 检查是否有错误
    if (Object.keys(validationErrors.value.colors).length > 0) {
      if (notification.value) {
        notification.value.show('保存失败', '请先修正输入错误再保存', 'error');
      }
      return;
    }

    // 格式校验
    for (const item of colorConfig.value) {
      if (item.numbers.trim()) {
        const validation = validateConfigInput(item.numbers, '颜色');
        if (!validation.isValid) {
          if (notification.value) {
            notification.value.show('配置格式错误', `${item.name}：${validation.message}`, 'error');
          }
          return;
        }
      }
    }
    
    const config = {};
    colorConfig.value.forEach(item => {
      if (item.numbers.trim()) {
        config[item.key] = item.numbers.split(',').map(n => parseInt(n.trim())).filter(n => !isNaN(n));
      }
    });
    
    await goApi.saveColorConfig(config);
    if (notification.value) {
      notification.value.show('保存成功', '颜色配置保存成功！', 'success');
    }
  } catch (error) {
    console.error('保存失败:', error);
    if (notification.value) {
      notification.value.show('保存失败', '保存失败: ' + error.message, 'error');
    }
  }
};

// 重置颜色配置
const resetColorConfig = async () => {
  if (notification.value) {
    const confirmed = await notification.value.show('确认重置', '确定要重置颜色配置为默认值吗？', 'warning', true);
    if (confirmed) {
      try {
        await goApi.resetSystemConfig();
        await loadAllConfigs();
        notification.value.show('重置成功', '颜色配置已重置为默认值！', 'success');
      } catch (error) {
        console.error('重置失败:', error);
        notification.value.show('重置失败', '重置失败: ' + error.message, 'error');
      }
    }
  }
};

// 保存尾数配置
const saveTailConfig = async () => {
  try {
    // 检查是否有错误
    if (Object.keys(validationErrors.value.tails).length > 0) {
      if (notification.value) {
        notification.value.show('保存失败', '请先修正输入错误再保存', 'error');
      }
      return;
    }

    // 格式校验
    for (const item of tailConfig.value) {
      if (item.numbers.trim()) {
        const validation = validateConfigInput(item.numbers, '尾数');
        if (!validation.isValid) {
          if (notification.value) {
            notification.value.show('配置格式错误', `${item.name}：${validation.message}`, 'error');
          }
          return;
        }
      }
    }
    
    const config = {};
    tailConfig.value.forEach(item => {
      if (item.numbers.trim()) {
        config[item.key] = item.numbers.split(',').map(n => parseInt(n.trim())).filter(n => !isNaN(n));
      }
    });
    
    await goApi.saveTailConfig(config);
    if (notification.value) {
      notification.value.show('保存成功', '尾数配置保存成功！', 'success');
    }
  } catch (error) {
    console.error('保存失败:', error);
    if (notification.value) {
      notification.value.show('保存失败', '保存失败: ' + error.message, 'error');
    }
  }
};

// 重置尾数配置
const resetTailConfig = async () => {
  if (notification.value) {
    const confirmed = await notification.value.show('确认重置', '确定要重置尾数配置为默认值吗？', 'warning', true);
    if (confirmed) {
      try {
        await goApi.resetSystemConfig();
        await loadAllConfigs();
        notification.value.show('重置成功', '尾数配置已重置为默认值！', 'success');
      } catch (error) {
        console.error('重置失败:', error);
        notification.value.show('重置失败', '重置失败: ' + error.message, 'error');
      }
    }
  }
};

// 保存下注类型配置
const saveBetTypeConfig = async () => {
  try {
    // 检查是否有错误
    if (Object.keys(validationErrors.value.betTypes).length > 0) {
      if (notification.value) {
        notification.value.show('保存失败', '请先修正输入错误再保存', 'error');
      }
      return;
    }

    // 格式校验
    for (const item of betTypeConfig.value) {
      if (item.aliases.trim()) {
        const validation = validateConfigInput(item.aliases, '下注类型');
        if (!validation.isValid) {
          if (notification.value) {
            notification.value.show('配置格式错误', `${item.name}：${validation.message}`, 'error');
          }
          return;
        }
      }
    }
    
    const config = {};
    betTypeConfig.value.forEach(item => {
      if (item.aliases.trim()) {
        config[item.type] = item.aliases.split(',').map(alias => alias.trim()).filter(alias => alias);
      }
    });
    
    await goApi.saveBetTypeAliases(config);
    if (notification.value) {
      notification.value.show('保存成功', '下注类型配置保存成功！', 'success');
    }
  } catch (error) {
    console.error('保存失败:', error);
    if (notification.value) {
      notification.value.show('保存失败', '保存失败: ' + error.message, 'error');
    }
  }
};

// 重置下注类型配置
const resetBetTypeConfig = async () => {
  if (notification.value) {
    const confirmed = await notification.value.show('确认重置', '确定要重置下注类型配置为默认值吗？', 'warning', true);
    if (confirmed) {
      try {
        await goApi.resetSystemConfig();
        await loadAllConfigs();
        notification.value.show('重置成功', '下注类型配置已重置为默认值！', 'success');
      } catch (error) {
        console.error('重置失败:', error);
        notification.value.show('重置失败', '重置失败: ' + error.message, 'error');
      }
    }
  }
};

// 保存关键字配置
const saveKeywordConfig = async () => {
  try {
    // 检查是否有错误
    if (Object.keys(validationErrors.value.keywords).length > 0) {
      if (notification.value) {
        notification.value.show('保存失败', '请先修正输入错误再保存', 'error');
      }
      return;
    }

    // 格式校验
    for (const item of keywordConfig.value) {
      if (item.aliases && item.aliases.trim()) {
        const validation = validateConfigInput(item.aliases, '关键字');
        if (!validation.isValid) {
          if (notification.value) {
            notification.value.show('配置格式错误', `${item.name}：${validation.message}`, 'error');
          }
          return;
        }
      }
    }
    
    const config = {};
    keywordConfig.value.forEach(item => {
      const aliases = item.aliases ? item.aliases.split(',').map(s => s.trim()).filter(s => s) : [];
      config[item.type] = aliases;
    });
    
    await goApi.saveKeywordAliases(config);
    if (notification.value) {
      notification.value.show('保存成功', '关键字别名配置已保存！', 'success');
    }
  } catch (error) {
    console.error('保存失败:', error);
    if (notification.value) {
      notification.value.show('保存失败', '保存失败: ' + error.message, 'error');
    }
  }
};

// 重置关键字配置
const resetKeywordConfig = async () => {
  if (notification.value) {
    const confirmed = await notification.value.show('确认重置', '确定要重置关键字别名配置为默认值吗？', 'warning', true);
    if (confirmed) {
      try {
        await goApi.resetSystemConfig();
        await loadAllConfigs();
        notification.value.show('重置成功', '关键字别名配置已重置为默认值！', 'success');
      } catch (error) {
        console.error('重置失败:', error);
        notification.value.show('重置失败', '重置失败: ' + error.message, 'error');
      }
    }
  }
};

// 保存赔率配置
const saveOddsConfig = async () => {
  try {
    const config = {
      three_of_three: {
        odds_ratio: oddsConfig.value.threeOfThree.oddsRatio,
        rebate: oddsConfig.value.threeOfThree.rebate
      },
      three_of_two: {
        hit_two_odds: {
          odds_ratio: oddsConfig.value.threeOfTwo.hitTwoOdds.oddsRatio,
          rebate: oddsConfig.value.threeOfTwo.hitTwoOdds.rebate
        },
        hit_three_odds: {
          odds_ratio: oddsConfig.value.threeOfTwo.hitThreeOdds.oddsRatio,
          rebate: oddsConfig.value.threeOfTwo.hitThreeOdds.rebate
        }
      },
      two_of_two: {
        odds_ratio: oddsConfig.value.twoOfTwo.oddsRatio,
        rebate: oddsConfig.value.twoOfTwo.rebate
      },
      special: {
        odds_ratio: oddsConfig.value.special.oddsRatio,
        rebate: oddsConfig.value.special.rebate
      }
    };
    
    await goApi.saveOddsConfig(config);
    if (notification.value) {
      notification.value.show('保存成功', '赔率配置已保存！', 'success');
    }
  } catch (error) {
    console.error('保存失败:', error);
    if (notification.value) {
      notification.value.show('保存失败', '保存失败: ' + error.message, 'error');
    }
  }
};

// 重置赔率配置
const resetOddsConfig = async () => {
  if (notification.value) {
    const confirmed = await notification.value.show('确认重置', '确定要重置赔率配置为默认值吗？', 'warning', true);
    if (confirmed) {
      try {
        await goApi.resetSystemConfig();
        await loadAllConfigs();
        notification.value.show('重置成功', '赔率配置已重置为默认值！', 'success');
      } catch (error) {
        console.error('重置失败:', error);
        notification.value.show('重置失败', '重置失败: ' + error.message, 'error');
      }
    }
  }
};

// 配置输入格式校验函数
const validateConfigInput = (input, configType) => {
  if (!input || !input.trim()) {
    return { isValid: true, message: '' }; // 空值允许
  }
  
  // 检查是否只包含逗号作为分隔符
  const forbiddenChars = /[^a-zA-Z0-9\u4e00-\u9fff\s,]/;
  if (forbiddenChars.test(input)) {
    const invalidChars = input.match(/[^a-zA-Z0-9\u4e00-\u9fff\s,]/g);
    return {
      isValid: false,
      message: `配置格式错误！只能使用英文逗号分隔，检测到非法字符：${[...new Set(invalidChars)].join(', ')}`
    };
  }
  
  return { isValid: true, message: '' };
};

// 离焦校验函数
const validateOnBlur = (configType, itemKey, value, itemName) => {
  const validation = validateConfigInput(value, configType);
  
  if (!validation.isValid) {
    // 设置错误状态
    validationErrors.value[configType][itemKey] = validation.message;
  } else {
    // 清除错误状态
    delete validationErrors.value[configType][itemKey];
  }
  
  return validation.isValid;
};

// 获取字段错误信息
const getFieldError = (configType, itemKey) => {
  return validationErrors.value[configType] && validationErrors.value[configType][itemKey];
};

// 检查是否有错误
const hasErrors = () => {
  return Object.values(validationErrors.value).some(configErrors => 
    Object.keys(configErrors).length > 0
  );
};

// 页面加载时获取配置
onMounted(() => {
  loadAllConfigs();
});
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