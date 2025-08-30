package backend

import (
	"time"
	// 移除了decimal包，不再需要
)

// 移除了User结构，六合彩机器人不需要用户管理

// LotteryType 彩种类型
type LotteryType string

const (
	NewMacau LotteryType = "new_macau"
	OldMacau LotteryType = "old_macau"
	HongKong LotteryType = "hongkong"
)

// 移除了ParseConfig和BetTypeConfig，使用简化的配置系统

// LotteryResult 开奖结果
type LotteryResult struct {
	Type          string    `json:"type"`           // 彩种类型: new_macau, old_macau, hongkong
	MainNumbers   []int     `json:"main_numbers"`   // 6个平码
	SpecialNumber int       `json:"special_number"` // 1个特码
	DrawDate      time.Time `json:"draw_date"`      // 开奖日期
	Period        string    `json:"period"`         // 期数
}

// 移除了ParseRecord，使用简化的解析响应结构

// 移除了BetGroup，使用ParsedBet作为下注结构

// 移除了ParseStatistics，不需要复杂的统计功能

// 移除了复杂的智能解析器结构，使用简化的BetParser

// 移除了所有mapao-wechat-bot兼容性结构，六合彩机器人不需要用户系统和结算功能

// SystemConfig 系统配置
type SystemConfig struct {
	ZodiacConfig   ZodiacConfig   `json:"zodiac_config"`    // 12生肖配置
	ColorConfig    ColorConfig    `json:"color_config"`     // 颜色波段配置
	TailConfig     TailConfig     `json:"tail_config"`      // 尾数配置
	BetTypeAliases BetTypeAliases `json:"bet_type_aliases"` // 下注类型别名配置
	KeywordAliases KeywordAliases `json:"keyword_aliases"`  // 关键字别名配置
	OddsConfig     OddsConfig     `json:"odds_config"`      // 赔率配置
}

// ZodiacConfig 12生肖配置
type ZodiacConfig struct {
	Rat     []int `json:"rat"`     // 鼠
	Ox      []int `json:"ox"`      // 牛
	Tiger   []int `json:"tiger"`   // 虎
	Rabbit  []int `json:"rabbit"`  // 兔
	Dragon  []int `json:"dragon"`  // 龙
	Snake   []int `json:"snake"`   // 蛇
	Horse   []int `json:"horse"`   // 马
	Goat    []int `json:"goat"`    // 羊
	Monkey  []int `json:"monkey"`  // 猴
	Rooster []int `json:"rooster"` // 鸡
	Dog     []int `json:"dog"`     // 狗
	Pig     []int `json:"pig"`     // 猪
}

// ColorConfig 颜色波段配置
type ColorConfig struct {
	Red   []int `json:"red"`   // 红波
	Green []int `json:"green"` // 绿波
	Blue  []int `json:"blue"`  // 蓝波
}

// TailConfig 尾数配置 (0-9尾)
type TailConfig struct {
	Tail0 []int `json:"tail_0"` // 0尾
	Tail1 []int `json:"tail_1"` // 1尾
	Tail2 []int `json:"tail_2"` // 2尾
	Tail3 []int `json:"tail_3"` // 3尾
	Tail4 []int `json:"tail_4"` // 4尾
	Tail5 []int `json:"tail_5"` // 5尾
	Tail6 []int `json:"tail_6"` // 6尾
	Tail7 []int `json:"tail_7"` // 7尾
	Tail8 []int `json:"tail_8"` // 8尾
	Tail9 []int `json:"tail_9"` // 9尾
}

// BetTypeAliases 下注类型别名配置
type BetTypeAliases struct {
	ThreeOfThree []string `json:"three_of_three"` // 三中三别名
	ThreeOfTwo   []string `json:"three_of_two"`   // 三中二别名
	TwoOfTwo     []string `json:"two_of_two"`     // 二中二别名
	Special      []string `json:"special"`        // 特碰别名
}

// KeywordAliases 关键字别名配置
type KeywordAliases struct {
	NewMacau []string `json:"new_macau"` // 新澳别名
	OldMacau []string `json:"old_macau"` // 老澳别名
	HongKong []string `json:"hong_kong"` // 香港别名
	Complex  []string `json:"complex"`   // 复式别名
	Each     []string `json:"each"`      // 各别名
	PerGroup []string `json:"per_group"` // 每组别名
}

// OddsConfig 赔率配置
type OddsConfig struct {
	ThreeOfThree ThreeOfThreeOdds `json:"three_of_three"` // 三中三赔率
	ThreeOfTwo   ThreeOfTwoOdds   `json:"three_of_two"`   // 三中二赔率
	TwoOfTwo     TwoOfTwoOdds     `json:"two_of_two"`     // 二中二赔率
	Special      SpecialOdds      `json:"special"`        // 特碰赔率
}

// ThreeOfThreeOdds 三中三赔率配置
type ThreeOfThreeOdds struct {
	OddsRatio float64 `json:"odds_ratio"` // 赔率
	Rebate    float64 `json:"rebate"`     // 回水率
}

// ThreeOfTwoOdds 三中二赔率配置（有两种情况）
type ThreeOfTwoOdds struct {
	HitTwoOdds   HitTwoOdds   `json:"hit_two_odds"`   // 中二个时的赔率
	HitThreeOdds HitThreeOdds `json:"hit_three_odds"` // 中三个时的赔率
}

// HitTwoOdds 中二个时的赔率配置
type HitTwoOdds struct {
	OddsRatio float64 `json:"odds_ratio"` // 赔率
	Rebate    float64 `json:"rebate"`     // 回水率
}

// HitThreeOdds 中三个时的赔率配置
type HitThreeOdds struct {
	OddsRatio float64 `json:"odds_ratio"` // 赔率
	Rebate    float64 `json:"rebate"`     // 回水率
}

// TwoOfTwoOdds 二中二赔率配置
type TwoOfTwoOdds struct {
	OddsRatio float64 `json:"odds_ratio"` // 赔率
	Rebate    float64 `json:"rebate"`     // 回水率
}

// SpecialOdds 特碰赔率配置
type SpecialOdds struct {
	OddsRatio float64 `json:"odds_ratio"` // 赔率
	Rebate    float64 `json:"rebate"`     // 回水率
}

// ================================
// 解析引擎相关模型
// ================================

// BetParseRequest 解析请求
type BetParseRequest struct {
	Input        string                 `json:"input"`         // 输入的下注字符串
	EnabledTypes []string               `json:"enabled_types"` // 启用的彩种类型
	UserSettings map[string]interface{} `json:"user_settings"` // 用户设置
}

// BetParseResponse 解析响应
type BetParseResponse struct {
	Success     bool             `json:"success"`      // 是否解析成功
	Error       string           `json:"error"`        // 错误信息
	Results     []ParsedBet      `json:"results"`      // 解析结果
	TotalBets   int              `json:"total_bets"`   // 总下注数
	TotalAmount float64          `json:"total_amount"` // 总金额
	TotalGroups int              `json:"total_groups"` // 总组数
	Summary     []BetTypeSummary `json:"summary"`      // 分类统计
	ParseTime   string           `json:"parse_time"`   // 解析耗时
}

// ParsedBet 解析后的下注
type ParsedBet struct {
	Type        string  `json:"type"`         // 下注类型（三中三、三中二等）
	Lottery     string  `json:"lottery"`      // 彩种（新澳、老澳、香港）
	Numbers     []int   `json:"numbers"`      // 号码列表
	Amount      float64 `json:"amount"`       // 单组金额
	TotalAmount float64 `json:"total_amount"` // 总金额
	Groups      int     `json:"groups"`       // 组数
	Description string  `json:"description"`  // 描述
	Original    string  `json:"original"`     // 原始输入行
	IsError     bool    `json:"is_error"`     // 是否解析错误
}

// BetTypeSummary 下注类型统计
type BetTypeSummary struct {
	Type        string  `json:"type"`         // 下注类型
	Lottery     string  `json:"lottery"`      // 彩种
	TotalGroups int     `json:"total_groups"` // 总组数
	TotalAmount float64 `json:"total_amount"` // 总金额
}

// ================================
// 智能解析器数据结构
// ================================

// BetParsingResult 整轮下注解析结果
type BetParsingResult struct {
	RoundID         string             `json:"roundId"`         // 轮次ID (递增数字)
	OriginalText    string             `json:"originalText"`    // 原始下注文本
	ParsedBets      []SingleBetParsing `json:"parsedBets"`      // 每笔下注解析结果
	RoundStatistics RoundBetStatistics `json:"roundStatistics"` // 整轮统计信息
	ParseTime       time.Time          `json:"parseTime"`       // 解析时间
	HasError        bool               `json:"hasError"`        // 是否有错误
	ErrorMessages   []string           `json:"errorMessages"`   // 错误信息列表
}

// SingleBetParsing 单笔下注解析结果
type SingleBetParsing struct {
	BetID         string                    `json:"betId"`         // 下注ID
	OriginalText  string                    `json:"originalText"`  // 原始下注文本
	LotteryBets   map[string]LotteryBetInfo `json:"lotteryBets"`   // 各体彩下注信息 key: 体彩类型("新澳"/"老澳"/"香港")
	BetStatistics BetStatistics             `json:"betStatistics"` // 本笔下注统计
	HasError      bool                      `json:"hasError"`      // 是否有错误
	ErrorMessage  string                    `json:"errorMessage"`  // 错误信息
}

// LotteryBetInfo 单个体彩的下注信息
type LotteryBetInfo struct {
	LotteryType string                 `json:"lotteryType"` // 体彩类型("新澳"/"老澳"/"香港")
	BetTypes    map[string]BetTypeInfo `json:"betTypes"`    // 各下注类型信息 key: 下注类型("三中三"/"二中二"/"特碰")
	TotalAmount float64                `json:"totalAmount"` // 该体彩总下注金额
	TotalGroups int                    `json:"totalGroups"` // 该体彩总下注组数
}

// BetTypeInfo 单个下注类型的信息
type BetTypeInfo struct {
	BetType     string      `json:"betType"`     // 下注类型("三中三"/"二中二"/"特碰")
	BetDetails  []BetDetail `json:"betDetails"`  // 具体下注明细列表
	TotalGroups int         `json:"totalGroups"` // 该类型总组数
	TotalAmount float64     `json:"totalAmount"` // 该类型总金额
	IsComplex   bool        `json:"isComplex"`   // 是否复式下注（从N个号码选组合）
	IsDrag      bool        `json:"isDrag"`      // 是否拖码下注（多组号码笛卡尔积）
	HasNumbers  bool        `json:"hasNumbers"`  // 是否有具体号码（用于错误检查）
}

// BetDetail 具体下注明细
type BetDetail struct {
	Numbers     []int   `json:"numbers"`     // 具体号码组合
	Amount      float64 `json:"amount"`      // 该组合的下注金额
	Groups      int     `json:"groups"`      // 该明细的组数（复式/拖码时可能>1）
	Description string  `json:"description"` // 描述信息
}

// BetStatistics 单笔下注统计
type BetStatistics struct {
	TotalAmount  float64 `json:"totalAmount"`  // 总金额
	TotalGroups  int     `json:"totalGroups"`  // 总组数
	LotteryCount int     `json:"lotteryCount"` // 体彩数量
	// 按体彩分类的下注类型统计
	LotteryBetTypeStats map[string]map[string]BetTypeStat `json:"lotteryBetTypeStats"` // [体彩][下注类型] -> 统计
}

// RoundBetStatistics 整轮下注统计
type RoundBetStatistics struct {
	TotalAmount float64 `json:"totalAmount"` // 总金额
	TotalGroups int     `json:"totalGroups"` // 总组数
	TotalBets   int     `json:"totalBets"`   // 总笔数
	// 各体彩各下注类型的统计
	LotteryBetTypeStats map[string]map[string]BetTypeStat `json:"lotteryBetTypeStats"` // [体彩][下注类型] -> 统计
	// 体彩总计（所有下注类型合计）
	LotteryTotals map[string]BetTypeStat `json:"lotteryTotals"` // [体彩] -> 总计
	// 下注类型总计（跨所有体彩）
	BetTypeTotals map[string]BetTypeStat `json:"betTypeTotals"` // [下注类型] -> 总计
}

// BetTypeStat 下注类型统计信息
type BetTypeStat struct {
	Amount float64 `json:"amount"` // 金额
	Groups int     `json:"groups"` // 组数
	Count  int     `json:"count"`  // 笔数
}

// IntelligentBetParserConfig 智能解析器配置
type IntelligentBetParserConfig struct {
	ZodiacMap      map[string][]int    `json:"zodiacMap"`      // 生肖映射
	ColorMap       map[string][]int    `json:"colorMap"`       // 颜色映射
	TailMap        map[string][]int    `json:"tailMap"`        // 尾数映射
	BetTypeAliases map[string][]string `json:"betTypeAliases"` // 下注类型别名
	LotteryAliases map[string][]string `json:"lotteryAliases"` // 体彩别名
	EndKeywords    []string            `json:"endKeywords"`    // 结束关键词
}
