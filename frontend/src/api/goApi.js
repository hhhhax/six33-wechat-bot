// six33-wechat-bot/frontend/src/api/goApi.js
// 封装与 Go 后端通信的 API 调用
// `window.go.backend.App` 是 Wails 自动注入的 Go 后端对象

const goApp = window.go.backend.App;

export const goApi = {
    /**
     * 授权登录
     * @param {string} code 授权码
     * @returns {Promise<boolean>} 是否授权成功
     */
    authorize: async (code) => {
        try {
            const result = await goApp.Authorize(code);
            console.log('goApi.authorize result:', result, 'type:', typeof result);
            return Boolean(result);
        } catch (error) {
            console.error("授权失败:", error);
            return false;
        }
    },

    /**
     * 检查是否已授权且未过期
     * @returns {Promise<boolean>} 是否已授权
     */
    isAuthorized: async () => {
        try {
            const result = await goApp.IsAuthorized();
            console.log('goApi.isAuthorized result:', result, 'type:', typeof result);
            return Boolean(result);
        } catch (error) {
            console.error("检查授权状态失败:", error);
            return false;
        }
    },

    // 移除了不需要的授权过期时间和文件对话框API

    // 移除了旧的parseBetInput方法，使用下面的新版本

    // 移除了解析历史和统计相关方法，不再需要

    // 移除了开奖结果相关方法，目前不需要

    // 移除了旧的解析配置和训练相关方法

    // ================================
    // 系统配置相关API
    // ================================

    /**
     * 获取系统配置
     * @returns {Promise<Object>} 系统配置对象
     */
    getSystemConfig: async () => {
        try {
            const result = await goApp.GetSystemConfig();
            return result;
        } catch (error) {
            console.error("获取系统配置失败:", error);
            return {};
        }
    },

    /**
     * 获取12生肖配置
     * @returns {Promise<Object>} 12生肖配置对象
     */
    getZodiacConfig: async () => {
        try {
            const result = await goApp.GetZodiacConfig();
            return result;
        } catch (error) {
            console.error("获取12生肖配置失败:", error);
            return {};
        }
    },

    /**
     * 保存12生肖配置
     * @param {Object} config 12生肖配置对象
     * @returns {Promise<boolean>} 是否成功
     */
    saveZodiacConfig: async (config) => {
        try {
            await goApp.SaveZodiacConfig(config);
            return true;
        } catch (error) {
            console.error("保存12生肖配置失败:", error);
            throw error;
        }
    },

    /**
     * 获取颜色波段配置
     * @returns {Promise<Object>} 颜色配置对象
     */
    getColorConfig: async () => {
        try {
            const result = await goApp.GetColorConfig();
            return result;
        } catch (error) {
            console.error("获取颜色配置失败:", error);
            return {};
        }
    },

    /**
     * 保存颜色波段配置
     * @param {Object} config 颜色配置对象
     * @returns {Promise<boolean>} 是否成功
     */
    saveColorConfig: async (config) => {
        try {
            await goApp.SaveColorConfig(config);
            return true;
        } catch (error) {
            console.error("保存颜色配置失败:", error);
            throw error;
        }
    },

    /**
     * 获取尾数配置
     * @returns {Promise<Object>} 尾数配置对象
     */
    getTailConfig: async () => {
        try {
            const result = await goApp.GetTailConfig();
            return result;
        } catch (error) {
            console.error("获取尾数配置失败:", error);
            return {};
        }
    },

    /**
     * 保存尾数配置
     * @param {Object} config 尾数配置对象
     * @returns {Promise<boolean>} 是否成功
     */
    saveTailConfig: async (config) => {
        try {
            await goApp.SaveTailConfig(config);
            return true;
        } catch (error) {
            console.error("保存尾数配置失败:", error);
            throw error;
        }
    },

    /**
     * 获取下注类型别名配置
     * @returns {Promise<Object>} 下注类型别名配置对象
     */
    getBetTypeAliases: async () => {
        try {
            const result = await goApp.GetBetTypeAliases();
            return result;
        } catch (error) {
            console.error("获取下注类型别名配置失败:", error);
            return {};
        }
    },

    /**
     * 保存下注类型别名配置
     * @param {Object} config 下注类型别名配置对象
     * @returns {Promise<boolean>} 是否成功
     */
    saveBetTypeAliases: async (config) => {
        try {
            await goApp.SaveBetTypeAliases(config);
            return true;
        } catch (error) {
            console.error("保存下注类型别名配置失败:", error);
            throw error;
        }
    },

    /**
     * 重置系统配置为默认值
     * @returns {Promise<boolean>} 是否成功
     */
    resetSystemConfig: async () => {
        try {
            await goApp.ResetSystemConfig();
            return true;
        } catch (error) {
            console.error("重置系统配置失败:", error);
            throw error;
        }
    },

    /**
     * 获取关键字别名配置
     * @returns {Promise<Object>} 关键字别名配置对象
     */
    getKeywordAliases: async () => {
        try {
            const result = await goApp.GetKeywordAliases();
            return result;
        } catch (error) {
            console.error("获取关键字别名配置失败:", error);
            return {};
        }
    },

    /**
     * 保存关键字别名配置
     * @param {Object} config 关键字别名配置对象
     * @returns {Promise<boolean>} 是否成功
     */
    saveKeywordAliases: async (config) => {
        try {
            await goApp.SaveKeywordAliases(config);
            return true;
        } catch (error) {
            console.error("保存关键字别名配置失败:", error);
            throw error;
        }
    },

    /**
     * 获取赔率配置
     * @returns {Promise<Object>} 赔率配置对象
     */
    getOddsConfig: async () => {
        try {
            const result = await goApp.GetOddsConfig();
            return result;
        } catch (error) {
            console.error("获取赔率配置失败:", error);
            return {};
        }
    },

    /**
     * 保存赔率配置
     * @param {Object} config 赔率配置对象
     * @returns {Promise<boolean>} 是否成功
     */
    saveOddsConfig: async (config) => {
        try {
            await goApp.SaveOddsConfig(config);
            return true;
        } catch (error) {
            console.error("保存赔率配置失败:", error);
            throw error;
        }
    },

    // 移除了结算和导出Excel方法，目前不需要

    // ================================
    // 智能解析相关API
    // ================================

    /**
     * 解析下注输入 (旧版本 - 保持兼容性)
     * @param {string} input 输入的下注字符串
     * @param {Array<string>} enabledTypes 启用的彩种类型
     * @returns {Promise<Object>} 解析结果对象
     */
    parseBetInput: async (input, enabledTypes) => {
        try {
            const result = await goApp.ParseBetInput(input, enabledTypes || []);
            return result;
        } catch (error) {
            console.error("解析下注输入失败:", error);
            throw error;
        }
    },

    /**
     * 智能解析下注输入 (新版本 - 智能解析器)
     * @param {string} input 输入的下注字符串
     * @param {Array<string>} enabledTypes 启用的彩种类型
     * @returns {Promise<Object>} 智能解析结果对象
     */
    parseBetInputIntelligent: async (input, enabledTypes) => {
        try {
            const result = await goApp.ParseBetInputIntelligent(input, enabledTypes || []);
            return result;
        } catch (error) {
            console.error("智能解析下注输入失败:", error);
            throw new Error(`智能解析失败: ${error.message || error}`);
        }
    }
};

// 默认导出
export default goApi;
