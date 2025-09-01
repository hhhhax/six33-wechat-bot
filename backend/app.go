package backend

import (
	"context"
	"crypto/rsa"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// SafeLogger 安全的日志记录器
type SafeLogger struct {
	mutex sync.Mutex
}

// 全局日志记录器
var safeLogger = &SafeLogger{}

// WriteLog 安全地写入日志文件
func (sl *SafeLogger) WriteLog(message string) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)

	// 尝试写入日志文件，如果失败也不要panic
	if err := os.WriteFile("start.log", []byte(logMessage), 0644); err != nil {
		// 如果无法写入文件，至少输出到标准错误
		fmt.Fprintf(os.Stderr, "Failed to write log: %v, original message: %s\n", err, logMessage)
	}
}

// AppendLog 追加日志到文件
func (sl *SafeLogger) AppendLog(message string) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)

	// 尝试追加到日志文件
	file, err := os.OpenFile("start.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v, message: %s\n", err, logMessage)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(logMessage); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to log file: %v, message: %s\n", err, logMessage)
	}
}

// recoverWithLog 统一的panic恢复和日志记录
func recoverWithLog(functionName string) {
	if r := recover(); r != nil {
		stack := debug.Stack()
		errorMsg := fmt.Sprintf("PANIC in %s: %v\nStack trace:\n%s", functionName, r, string(stack))
		safeLogger.AppendLog(errorMsg)

		// 也输出到标准错误以便调试
		fmt.Fprintf(os.Stderr, "%s\n", errorMsg)
	}
}

// 公钥内容，直接嵌入代码中
const publicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA6DXjS1hnuphH8AGHtwM4
NJH+plk+e52eT9JIMwBnh+qvEGCSg9eQ8xl5KVWR5NING5L5GODUjayzRXDCYRat
Feeva5oBYEi+5nLTSuS/5aL/6vmea3f57zreleqc42JB1XpIeMuFuc/TLA8LcQPj
favw/kzch/oaIU0PfNc9BmrOEvmX7hfxZhlwdB2rk89m1sarwsm6brhZxg0oHieM
c+Ftqh8WM0yT46VbZZsZPjHW7aiL1WBZ3Bt9tG7cdKMp6Ojbli2bSftCUB4qFrSP
WKb14Y/b1VXk9zu8MhCziokp0OjROpUz4zVW7uSTYnLAk6eC6wjz4QCYA3yrd+jb
1QIDAQAB
-----END PUBLIC KEY-----`

var pubKey *rsa.PublicKey

// App 结构体定义了 Go 后端应用
type App struct {
	ctx          context.Context
	mutex        sync.RWMutex  // 读写锁
	shutdownChan chan struct{} // 优雅关闭通道
	shutdownOnce sync.Once     // 确保只关闭一次
	authCode     string        // 授权码
	authExpiry   time.Time     // 授权过期时间

	// 六合彩相关数据
	lotteryResults map[string]*LotteryResult // 开奖结果 (new_macau, old_macau, hongkong)
	systemConfig   *SystemConfig             // 系统配置（内存缓存）
}

// NewApp 创建并返回一个新的 App 实例
func NewApp() *App {
	defer recoverWithLog("NewApp")

	safeLogger.WriteLog("六合彩智能解析机器人开始启动")

	// 设置全局panic处理
	defer func() {
		if r := recover(); r != nil {
			errorMsg := fmt.Sprintf("NewApp PANIC: %v\nStack: %s", r, string(debug.Stack()))
			safeLogger.AppendLog(errorMsg)
			fmt.Fprintf(os.Stderr, "%s\n", errorMsg)
		}
	}()

	// 加载公钥
	var err error
	pubKey, err = loadPublicKey(publicKeyPEM)
	if err != nil {
		safeLogger.WriteLog(fmt.Sprintf("加载公钥失败: %v", err))
		return nil
	}

	// 初始化配置系统
	if initErr := InitializeConfig(); initErr != nil {
		safeLogger.WriteLog(fmt.Sprintf("初始化配置系统失败: %v", initErr))
		// 配置初始化失败不应该阻止应用启动，继续使用默认配置
	}

	// 加载系统配置到内存
	systemConfig, configErr := loadSystemConfigFromFile()
	if configErr != nil {
		safeLogger.WriteLog(fmt.Sprintf("加载系统配置失败，使用默认配置: %v", configErr))
		systemConfig = getDefaultSystemConfig()
	}

	app := &App{
		shutdownChan:   make(chan struct{}),
		authExpiry:     time.Time{},
		lotteryResults: make(map[string]*LotteryResult),
		systemConfig:   systemConfig,
	}

	safeLogger.WriteLog("六合彩智能解析机器人实例创建成功")
	return app
}

// Startup 应用启动时调用
func (a *App) Startup(ctx context.Context) {
	defer recoverWithLog("startup")
	a.ctx = ctx
	safeLogger.AppendLog("六合彩智能解析机器人已启动")
	if ctx != nil {
		wailsRuntime.LogInfo(ctx, "App has started up.")
	}
}

// DomReady 当DOM准备就绪时调用
func (a *App) DomReady(ctx context.Context) {
	defer recoverWithLog("domReady")
	safeLogger.AppendLog("DOM已准备就绪")
	if ctx != nil {
		wailsRuntime.LogInfo(ctx, "DOM is ready.")
	}
}

// Shutdown 应用关闭时调用
func (a *App) Shutdown(ctx context.Context) {
	defer recoverWithLog("shutdown")

	safeLogger.AppendLog("六合彩智能解析机器人开始关闭")
	if a.ctx != nil {
		wailsRuntime.LogInfo(a.ctx, "App is shutting down.")
	}

	// 触发优雅关闭
	a.gracefulShutdown()
}

func (a *App) gracefulShutdown() {
	a.shutdownOnce.Do(func() {
		close(a.shutdownChan)
		time.Sleep(100 * time.Millisecond)
		safeLogger.AppendLog("程序即将退出")
		os.Exit(1)
	})
}

// ================================
// 授权验证相关方法
// ================================

// Authorize 处理授权登录
func (a *App) Authorize(code string) bool {
	defer recoverWithLog("Authorize")

	if a == nil {
		safeLogger.AppendLog("App实例为nil，无法授权")
		return false
	}

	// 检查是否处于调试状态
	if isBeingDebugged() {
		safeLogger.AppendLog("检测到调试工具，拒绝授权")
		return false
	}

	if success, _, _, _ := verifyAuthCode(pubKey, code); success {
		a.mutex.Lock()
		a.authCode = code
		a.authExpiry = time.Now().Add(24 * time.Hour) // 24小时过期
		a.mutex.Unlock()

		safeLogger.AppendLog(fmt.Sprintf("用户授权成功，授权码: %s", code))
		return true
	}

	safeLogger.AppendLog(fmt.Sprintf("用户授权失败，无效授权码: %s", code))
	return false
}

// IsAuthorized 检查是否已授权
func (a *App) IsAuthorized() bool {
	defer recoverWithLog("IsAuthorized")

	if a == nil {
		return false
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	// 检查授权码是否为空
	if a.authCode == "" {
		return false
	}

	// 检查是否过期
	if time.Now().After(a.authExpiry) {
		return false
	}

	return true
}

// GetAuthStatus 获取授权状态
func (a *App) GetAuthStatus() map[string]interface{} {
	defer recoverWithLog("GetAuthStatus")

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	status := map[string]interface{}{
		"authorized": a.IsAuthorized(),
		"expiry":     a.authExpiry.Format("2006-01-02 15:04:05"),
	}

	if a.authCode != "" {
		// 只显示授权码的前几位和后几位
		maskedCode := fmt.Sprintf("%s****%s", a.authCode[:4], a.authCode[len(a.authCode)-4:])
		status["authCode"] = maskedCode
	}

	return status
}

// ================================
// 智能解析相关方法
// ================================

// ParseBetInput 解析下注输入
func (a *App) ParseBetInput(input string, enabledTypes []string) (*BetParseResponse, error) {
	defer recoverWithLog("ParseBetInput")

	if strings.TrimSpace(input) == "" {
		response := &BetParseResponse{
			Success: false,
			Error:   "输入内容为空",
			Results: []ParsedBet{},
		}
		return response, nil
	}

	// 创建解析器，传入app实例
	parser := NewBetParser(a)

	// 构建解析请求
	request := BetParseRequest{
		Input:        input,
		EnabledTypes: enabledTypes,
		UserSettings: make(map[string]interface{}),
	}

	// 执行解析
	response := parser.ParseBetString(request)

	// 记录解析日志
	if response.Success {
		safeLogger.AppendLog(fmt.Sprintf("下注解析成功: %d项下注, 总金额%.2f元, 耗时%s", response.TotalBets, response.TotalAmount, response.ParseTime))
	} else {
		safeLogger.AppendLog(fmt.Sprintf("下注解析失败: %s", response.Error))
	}

	return &response, nil
}

// ================================
// 系统配置相关方法
// ================================

// GetSystemConfig 获取系统配置
func (a *App) GetSystemConfig() SystemConfig {
	defer recoverWithLog("GetSystemConfig")
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return *a.systemConfig
}

// GetZodiacConfig 获取生肖配置
func (a *App) GetZodiacConfig() ZodiacConfig {
	defer recoverWithLog("GetZodiacConfig")
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.systemConfig.ZodiacConfig
}

// SaveZodiacConfig 保存生肖配置
func (a *App) SaveZodiacConfig(config ZodiacConfig) error {
	defer recoverWithLog("SaveZodiacConfig")

	// 先更新内存中的配置
	a.mutex.Lock()
	a.systemConfig.ZodiacConfig = config
	a.mutex.Unlock()

	// 再保存到文件
	if err := saveSystemConfigToFile(a.systemConfig); err != nil {
		safeLogger.AppendLog(fmt.Sprintf("保存生肖配置失败: %v", err))
		return err
	}

	safeLogger.AppendLog("生肖配置已更新")
	return nil
}

// GetColorConfig 获取颜色配置
func (a *App) GetColorConfig() ColorConfig {
	defer recoverWithLog("GetColorConfig")
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.systemConfig.ColorConfig
}

// SaveColorConfig 保存颜色配置
func (a *App) SaveColorConfig(config ColorConfig) error {
	defer recoverWithLog("SaveColorConfig")

	// 先更新内存中的配置
	a.mutex.Lock()
	a.systemConfig.ColorConfig = config
	a.mutex.Unlock()

	// 再保存到文件
	if err := saveSystemConfigToFile(a.systemConfig); err != nil {
		safeLogger.AppendLog(fmt.Sprintf("保存颜色配置失败: %v", err))
		return err
	}

	safeLogger.AppendLog("颜色配置已更新")
	return nil
}

// GetTailConfig 获取尾数配置
func (a *App) GetTailConfig() TailConfig {
	defer recoverWithLog("GetTailConfig")
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.systemConfig.TailConfig
}

// SaveTailConfig 保存尾数配置
func (a *App) SaveTailConfig(config TailConfig) error {
	defer recoverWithLog("SaveTailConfig")

	// 先更新内存中的配置
	a.mutex.Lock()
	a.systemConfig.TailConfig = config
	a.mutex.Unlock()

	// 再保存到文件
	if err := saveSystemConfigToFile(a.systemConfig); err != nil {
		safeLogger.AppendLog(fmt.Sprintf("保存尾数配置失败: %v", err))
		return err
	}

	safeLogger.AppendLog("尾数配置已更新")
	return nil
}

// GetBetTypeAliases 获取下注类型别名配置
func (a *App) GetBetTypeAliases() BetTypeAliases {
	defer recoverWithLog("GetBetTypeAliases")
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.systemConfig.BetTypeAliases
}

// SaveBetTypeAliases 保存下注类型别名配置
func (a *App) SaveBetTypeAliases(config BetTypeAliases) error {
	defer recoverWithLog("SaveBetTypeAliases")

	// 先更新内存中的配置
	a.mutex.Lock()
	a.systemConfig.BetTypeAliases = config
	a.mutex.Unlock()

	// 再保存到文件
	if err := saveSystemConfigToFile(a.systemConfig); err != nil {
		safeLogger.AppendLog(fmt.Sprintf("保存下注类型别名配置失败: %v", err))
		return err
	}

	safeLogger.AppendLog("下注类型别名配置已更新")
	return nil
}

// GetKeywordAliases 获取关键字别名配置
func (a *App) GetKeywordAliases() KeywordAliases {
	defer recoverWithLog("GetKeywordAliases")
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.systemConfig.KeywordAliases
}

// SaveKeywordAliases 保存关键字别名配置
func (a *App) SaveKeywordAliases(config KeywordAliases) error {
	defer recoverWithLog("SaveKeywordAliases")

	// 先更新内存中的配置
	a.mutex.Lock()
	a.systemConfig.KeywordAliases = config
	a.mutex.Unlock()

	// 再保存到文件
	if err := saveSystemConfigToFile(a.systemConfig); err != nil {
		safeLogger.AppendLog(fmt.Sprintf("保存关键字别名配置失败: %v", err))
		return err
	}

	safeLogger.AppendLog("关键字别名配置已更新")
	return nil
}

// GetOddsConfig 获取赔率配置
func (a *App) GetOddsConfig() OddsConfig {
	defer recoverWithLog("GetOddsConfig")
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.systemConfig.OddsConfig
}

// SaveOddsConfig 保存赔率配置
func (a *App) SaveOddsConfig(config OddsConfig) error {
	defer recoverWithLog("SaveOddsConfig")

	// 先更新内存中的配置
	a.mutex.Lock()
	a.systemConfig.OddsConfig = config
	a.mutex.Unlock()

	// 再保存到文件
	if err := saveSystemConfigToFile(a.systemConfig); err != nil {
		safeLogger.AppendLog(fmt.Sprintf("保存赔率配置失败: %v", err))
		return err
	}

	safeLogger.AppendLog("赔率配置已更新")
	return nil
}

// ResetSystemConfig 重置系统配置
func (a *App) ResetSystemConfig() error {
	defer recoverWithLog("ResetSystemConfig")

	// 获取默认配置
	defaultConfig := getDefaultSystemConfig()

	// 先更新内存中的配置
	a.mutex.Lock()
	a.systemConfig = defaultConfig
	a.mutex.Unlock()

	// 再保存到文件
	if err := saveSystemConfigToFile(defaultConfig); err != nil {
		return err
	}

	safeLogger.AppendLog("系统配置已重置为默认值")
	return nil
}

// ================================
// 智能解析器API方法
// ================================

// ParseBetInputIntelligent 智能解析下注输入
func (a *App) ParseBetInputIntelligent(input string, enabledTypes []string) (*BetParsingResult, error) {
	defer recoverWithLog("ParseBetInputIntelligent")

	if strings.TrimSpace(input) == "" {
		result := &BetParsingResult{
			HasError:      true,
			ErrorMessages: []string{"输入内容为空"},
			ParsedBets:    []SingleBetParsing{},
		}
		return result, nil
	}

	// 创建解析器配置
	config := a.createParserConfig()

	// 创建智能解析器
	parser := NewIntelligentBetParser(config)

	// 构建解析请求
	request := BetParseRequest{
		Input:        input,
		EnabledTypes: enabledTypes,
		UserSettings: make(map[string]interface{}),
	}

	// 执行智能解析
	result := parser.ParseBetString(request)

	// 记录解析日志
	if !result.HasError {
		safeLogger.AppendLog(fmt.Sprintf("智能解析成功: %d笔下注, 总金额%s元",
			result.RoundStatistics.TotalBets, result.RoundStatistics.TotalAmount.String()))
	} else {
		safeLogger.AppendLog(fmt.Sprintf("智能解析失败: %v", result.ErrorMessages))
	}

	return &result, nil
}

// createParserConfig 创建解析器配置
func (a *App) createParserConfig() IntelligentBetParserConfig {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	zodiacConfig := a.systemConfig.ZodiacConfig
	colorConfig := a.systemConfig.ColorConfig
	tailConfig := a.systemConfig.TailConfig
	betTypeAliases := a.systemConfig.BetTypeAliases
	keywordAliases := a.systemConfig.KeywordAliases

	return IntelligentBetParserConfig{
		ZodiacMap: map[string][]int{
			"鼠": zodiacConfig.Rat, "牛": zodiacConfig.Ox, "虎": zodiacConfig.Tiger, "兔": zodiacConfig.Rabbit,
			"龙": zodiacConfig.Dragon, "蛇": zodiacConfig.Snake, "马": zodiacConfig.Horse, "羊": zodiacConfig.Goat,
			"猴": zodiacConfig.Monkey, "鸡": zodiacConfig.Rooster, "狗": zodiacConfig.Dog, "猪": zodiacConfig.Pig,
		},
		ColorMap: map[string][]int{
			"红": colorConfig.Red, "蓝": colorConfig.Blue, "绿": colorConfig.Green,
		},
		TailMap: map[string][]int{
			"0尾": tailConfig.Tail0, "1尾": tailConfig.Tail1, "2尾": tailConfig.Tail2, "3尾": tailConfig.Tail3,
			"4尾": tailConfig.Tail4, "5尾": tailConfig.Tail5, "6尾": tailConfig.Tail6, "7尾": tailConfig.Tail7,
			"8尾": tailConfig.Tail8, "9尾": tailConfig.Tail9,
		},
		BetTypeAliases: map[string][]string{
			"三中三": betTypeAliases.ThreeOfThree,
			"三中二": betTypeAliases.ThreeOfTwo,
			"二中二": betTypeAliases.TwoOfTwo,
			"特碰":  betTypeAliases.Special,
		},
		LotteryAliases: map[string][]string{
			"新澳": keywordAliases.NewMacau,
			"老澳": keywordAliases.OldMacau,
			"香港": keywordAliases.HongKong,
		},
		KeywordAliases: map[string][]string{
			"复式": keywordAliases.Complex,
			"拖":  keywordAliases.Drag,
		},
		EndKeywords: map[string][]string{
			"各":  keywordAliases.Each,
			"每组": keywordAliases.PerGroup,
		},
	}
}
