package backend

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/shopspring/decimal"
)

var roundIDCounter int64

// IntelligentBetParser 智能下注解析器
type IntelligentBetParser struct {
	config IntelligentBetParserConfig
}

// NumbersAndAmount 号码和金额结构
type NumbersAndAmount struct {
	Numbers []int
	Amount  float64
}

// NumberGroupsAndAmount 多组号码和金额结构
type NumberGroupsAndAmount struct {
	NumberGroups [][]int
	Amount       float64
}

// BetContext 下注上下文
type BetContext struct {
	inheritedLotteries []string // 继承的体彩类型
}

func NewIntelligentBetParser(config IntelligentBetParserConfig) *IntelligentBetParser {
	return &IntelligentBetParser{config: config}
}

// ParseBetString 智能解析下注字符串
func (p *IntelligentBetParser) ParseBetString(request BetParseRequest) BetParsingResult {
	startTime := time.Now()
	roundID := strconv.FormatInt(atomic.AddInt64(&roundIDCounter, 1), 10)

	result := BetParsingResult{
		RoundID:       roundID,
		OriginalText:  request.Input,
		ParseTime:     startTime,
		ErrorMessages: make([]string, 0),
	}

	if strings.TrimSpace(request.Input) == "" {
		result.HasError = true
		result.ErrorMessages = append(result.ErrorMessages, "输入为空")
		return result
	}

	safeLogger.AppendLog(fmt.Sprintf("开始智能解析: %s", request.Input))
		// 1. 替换关键词（生肖、颜色、尾数）
	processedText := p.replaceKeywords(request.Input)
	safeLogger.AppendLog(fmt.Sprintf("替换关键词: %s", processedText))
	// 2. 字符串预处理
	processedText = p.preprocessText(processedText)

	safeLogger.AppendLog(fmt.Sprintf("字符串预处理: %s", processedText))
	

	// 3. 移除非关键字的所有中文记清理关键词后面的空格
	processedText = p.removeChineseChars(processedText)
	safeLogger.AppendLog(fmt.Sprintf("移除非关键字的所有中文及替换莫名的空格: %s", processedText))

	// 4. 分割为多笔下注
	betSegments := p.segmentBets(processedText)
	safeLogger.AppendLog(fmt.Sprintf("分割为多笔下注: %s", betSegments))

	// 5. 解析每笔下注
	parsedBets := make([]SingleBetParsing, 0)
	context := &BetContext{inheritedLotteries: make([]string, 0)}

	for i, segment := range betSegments {
		betID := fmt.Sprintf("%s_bet_%d", roundID, i+1)
		parsed := p.parseSingleBet(betID, segment, context)
		parseJson, _ := json.Marshal(parsed)
		safeLogger.AppendLog(fmt.Sprintf("解析每笔下注: %s", string(parseJson)))
		parsedBets = append(parsedBets, parsed)

		// 更新体彩继承状态
		if len(parsed.LotteryBets) > 0 {
			context.inheritedLotteries = make([]string, 0)
			for lotteryType := range parsed.LotteryBets {
				context.inheritedLotteries = append(context.inheritedLotteries, lotteryType)
			}
		}
	}

	result.ParsedBets = parsedBets

	// 6. 生成统计信息
	result.RoundStatistics = p.generateRoundStatistics(parsedBets)

	// 7. 检查错误
	for _, bet := range parsedBets {
		if bet.HasError {
			result.HasError = true
			result.ErrorMessages = append(result.ErrorMessages, bet.ErrorMessage)
		}
	}

	result.ParseTime = time.Now()
	return result
}

// preprocessText 字符串预处理
func (p *IntelligentBetParser) preprocessText(text string) string {
	// 1. 换行符替换为空格
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	// 2. 压缩连续分隔符并统一替换为逗号，同时保留第一个分隔符
	consecutiveRe := regexp.MustCompile(`([./\\\-=:,、+。*])([./\\\-=:,、+。*])+`)
	text = consecutiveRe.ReplaceAllString(text, "$1")

	separatorRe := regexp.MustCompile(`[./\\\-=:,、+。*]+`)
	// `ReplaceAllString` 一次性完成所有替换和压缩
	text = separatorRe.ReplaceAllString(text, "-")

	// 定义你想要替换的符号字符串
	symbolsToReplace := "【】[]{}“‘”’" // 可以在这里添加你需要的任何符号
	// 将字符串中的所有元字符转义，然后作为正则表达式的模式
	pattern := "[" + regexp.QuoteMeta(symbolsToReplace) + "]+"

	invalidCharsRe := regexp.MustCompile(pattern)
	text = invalidCharsRe.ReplaceAllString(text, " ")

	// 3. 处理分隔符后的空格：保留分隔符，移除空格
	spaceAfterSeparatorRe := regexp.MustCompile(`([./\\\-=:,、+。*])\s+`)
	text = spaceAfterSeparatorRe.ReplaceAllString(text, "$1")

	// 4. 清理多余空格
	spaceRe := regexp.MustCompile(`\s+`)
	text = spaceRe.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// replaceKeywords 替换关键词（生肖、颜色、尾数）
func (p *IntelligentBetParser) replaceKeywords(text string) string {
	// 合并所有关键词映射
	allKeywords := make(map[string][]int)
	for keyword, numbers := range p.config.ZodiacMap {
		allKeywords[keyword] = numbers
	}
	for keyword, numbers := range p.config.ColorMap {
		allKeywords[keyword] = numbers
	}
	for keyword, numbers := range p.config.TailMap {
		allKeywords[keyword] = numbers
	}

	// 按关键词长度降序排序（从最多字开始替换）
	keywords := make([]string, 0, len(allKeywords))
	for keyword := range allKeywords {
		keywords = append(keywords, keyword)
	}

	// 排序：按长度降序
	for i := 0; i < len(keywords)-1; i++ {
		for j := i + 1; j < len(keywords); j++ {
			if len(keywords[i]) < len(keywords[j]) {
				keywords[i], keywords[j] = keywords[j], keywords[i]
			}
		}
	}

	// 替换关键词
	for _, keyword := range keywords {
		if numbers, exists := allKeywords[keyword]; exists {
			numbersStr := make([]string, len(numbers))
			for i, num := range numbers {
				numbersStr[i] = fmt.Sprintf("%02d", num)
			}
			replacement := strings.Join(numbersStr, ",")
			replacement = "," + replacement
			text = strings.ReplaceAll(text, keyword, replacement)
		}
	}

	return text
}

// removeChineseChars 移除非关键字的所有中文字符
func (p *IntelligentBetParser) removeChineseChars(text string) string {

	// 清理结束关键词后的空格和符号
	text = p.cleanEndKeywords(text)

	// 保留的关键词
	preservedKeywords := make([]string, 0)

	// 添加下注类型关键词
	for _, aliases := range p.config.BetTypeAliases {
		preservedKeywords = append(preservedKeywords, aliases...)
	}

	// 添加体彩关键词
	for _, aliases := range p.config.LotteryAliases {
		preservedKeywords = append(preservedKeywords, aliases...)
	}
	//  添加结束关键词
	for _, aliases := range p.config.EndKeywords {
		preservedKeywords = append(preservedKeywords, aliases...)
	}
	//  添加关键字关键词
	for _, aliases := range p.config.KeywordAliases {
		preservedKeywords = append(preservedKeywords, aliases...)
	}

	// 按长度降序排序
	for i := 0; i < len(preservedKeywords)-1; i++ {
		for j := i + 1; j < len(preservedKeywords); j++ {
			if len(preservedKeywords[i]) < len(preservedKeywords[j]) {
				preservedKeywords[i], preservedKeywords[j] = preservedKeywords[j], preservedKeywords[i]
			}
		}
	}

	// 先用占位符替换要保留的关键词
	placeholders := make(map[string]string)
	for i, keyword := range preservedKeywords {
		placeholder := fmt.Sprintf("__PRESERVE_%d__", i)
		if strings.Contains(text, keyword) {
			text = strings.ReplaceAll(text, keyword, placeholder)
			placeholders[placeholder] = keyword
		}
	}

	// 移除所有中文字符 - 使用简单方法逐字符检查
	var result strings.Builder
	for _, char := range text {
		// 检查是否为中文字符（简化版本）
		if char >= 0x4e00 && char <= 0x9fff {
			// 跳过中文字符
			continue
		}
		result.WriteRune(char)
	}
	text = result.String()

	// 恢复保留的关键词
	for placeholder, keyword := range placeholders {
		safeLogger.AppendLog(fmt.Sprintf("恢复保留的关键词: %s,%s", placeholder, keyword))

		// 标记是否已找到并替换
		replaced := false

		// 尝试在 BetTypeAliases 中查找并替换
		for k, alias := range p.config.BetTypeAliases {
			if slices.Contains(alias, keyword) {
				text = strings.ReplaceAll(text, placeholder, k)
				replaced = true // 找到并替换，退出内层 alias 循环
				break           // 退出 k 循环
			}
		}
		if replaced {
			continue // 已经替换，处理下一个 placeholder
		}

		// 尝试在 LotteryAliases 中查找并替换
		for k, alias := range p.config.LotteryAliases {
			if slices.Contains(alias, keyword) {
				text = strings.ReplaceAll(text, placeholder, k)
				replaced = true
				break
			}
		}
		if replaced {
			continue
		}

		// 尝试在 KeywordAliases 中查找并替换
		for k, alias := range p.config.KeywordAliases {
			if slices.Contains(alias, keyword) {
				text = strings.ReplaceAll(text, placeholder, k)
				replaced = true
				break
			}
		}
		if replaced {
			continue
		}

		// 尝试在 EndKeywords 中查找并替换
		for k, alias := range p.config.EndKeywords {
			if slices.Contains(alias, keyword) {
				text = strings.ReplaceAll(text, placeholder, k)
				replaced = true
				break
			}
		}
		if replaced {
			continue
		}

		// 如果在所有配置中都没找到，则保持原关键词
		text = strings.ReplaceAll(text, placeholder, keyword)
	}

	return text
}

// cleanEndKeywords 清理结束关键词后的空格和符号
func (p *IntelligentBetParser) cleanEndKeywords(text string) string {
	// 获取所有结束关键词
	endKeywords := p.config.EndKeywords

	// 过滤出实际的结束关键词（去重）
	uniqueEndKeywords := make([]string, 0)
	for k, _ := range endKeywords {
		uniqueEndKeywords = append(uniqueEndKeywords, k)
	}

	// 按长度降序排序，确保长关键词优先处理
	for i := 0; i < len(uniqueEndKeywords)-1; i++ {
		for j := i + 1; j < len(uniqueEndKeywords); j++ {
			if len(uniqueEndKeywords[i]) < len(uniqueEndKeywords[j]) {
				uniqueEndKeywords[i], uniqueEndKeywords[j] = uniqueEndKeywords[j], uniqueEndKeywords[i]
			}
		}
	}

	// 定义一个正则表达式，用于匹配“十”到“万”的中文数字
	chineseNumRe := `(一|二|三|四|五|六|七|八|九|十|百|千|万)`

	for _, keyword := range uniqueEndKeywords {
		// 步骤一：移除结束关键词后的空格和标点符号
		// 匹配：关键词 + 任意空格 + 任意标点符号
		pattern := regexp.QuoteMeta(keyword) + `[\s\p{P}]*`
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		text = re.ReplaceAllString(text, keyword+"")

		// 步骤二：处理该关键词后面紧跟的中文数字
		// 匹配：关键词 + 紧跟的中文数字
		chineseNumPattern := regexp.QuoteMeta(keyword) + chineseNumRe + `+`
		reNum, err := regexp.Compile(chineseNumPattern)
		if err != nil {
			continue
		}

		// 查找所有匹配该模式的字符串
		matches := reNum.FindAllString(text, -1)
		for _, match := range matches {
			// 提取中文数字部分
			numPart := strings.TrimPrefix(match, keyword)

			// 将中文数字转换为阿拉伯数字
			if val, ok := chineseToNumber(numPart); ok {
				// 替换原始字符串中的“关键词+中文数字”为“关键词+阿拉伯数字”
				text = strings.ReplaceAll(text, match, keyword+strconv.Itoa(val))
			}
		}
	}

	return text
}

// chineseToNumber 辅助函数：将中文数字（十到万）转换为阿拉伯数字
// 这是一个简化版本，只处理整十、整百、整千、整万的情况
// 比如 "十" -> 10, "二十" -> 20, "一百" -> 100, "一万" -> 10000
func chineseToNumber(chinese string) (int, bool) {
	numMap := map[rune]int{'零': 0, '一': 1, '二': 2, '三': 3, '四': 4, '五': 5, '六': 6, '七': 7, '八': 8, '九': 9}
	unitMap := map[rune]int{'十': 10, '百': 100, '千': 1000, '万': 10000}

	var result int
	var tempNum int = 1 // 用于处理“十”开头的数字，如“十”或“十三”

	runes := []rune(chinese)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if val, ok := numMap[r]; ok {
			tempNum = val
			// 如果这是最后一个字符，直接加到结果中
			if i == len(runes)-1 {
				result += tempNum
			}
		} else if unit, ok := unitMap[r]; ok {
			if tempNum == 1 && unit == 10 && i == 0 {
				// 特殊处理“十”，表示10
				result += 10
			} else {
				result += tempNum * unit
			}
			tempNum = 0 // 重置临时数字
		}
	}

	// 检查是否处理了至少一个数字或单位
	if result > 0 {
		return result, true
	}

	// 特殊处理纯粹的单位，如“十”
	if val, ok := unitMap[runes[0]]; ok && len(runes) == 1 {
		return val, true
	}

	return 0, false // 无法转换
}

// segmentBets 分割为多笔下注
func (p *IntelligentBetParser) segmentBets(text string) []string {
	// 识别结束标识符
	endKeywords := p.config.EndKeywords

	segments := make([]string, 0)
	currentSegment := ""

	// 简单实现：按结束关键词分割
	words := strings.Fields(text)
	for _, word := range words {
		currentSegment += word + " "

		// 检查是否包含结束关键词
		hasEndKeyword := false
		for k := range endKeywords {
			if strings.Contains(word, k) {
				segments = append(segments, strings.TrimSpace(currentSegment))
				currentSegment = ""
				hasEndKeyword = true
				break
			}
		}

		if hasEndKeyword {
			continue
		}
	}

	// 处理最后一段
	if strings.TrimSpace(currentSegment) != "" {
		segments = append(segments, strings.TrimSpace(currentSegment))
	}

	return segments
}

// identifyLotteries 识别体彩类型（从最多字开始替换）
func (p *IntelligentBetParser) identifyLotteries(text string) []string {
	result := make([]string, 0)

	// 按别名长度降序排序
	for lotteryType, aliases := range p.config.LotteryAliases {
		// 对当前体彩的别名按长度排序
		sortedAliases := make([]string, len(aliases))
		copy(sortedAliases, aliases)

		for i := 0; i < len(sortedAliases)-1; i++ {
			for j := i + 1; j < len(sortedAliases); j++ {
				if len(sortedAliases[i]) < len(sortedAliases[j]) {
					sortedAliases[i], sortedAliases[j] = sortedAliases[j], sortedAliases[i]
				}
			}
		}

		// 从最长别名开始匹配
		for _, alias := range sortedAliases {
			if strings.Contains(text, alias) {
				result = append(result, lotteryType)
				break
			}
		}
	}
	return result
}

// identifyBetTypes 识别下注类型
func (p *IntelligentBetParser) identifyBetTypes(text string) []string {
	result := make([]string, 0)
	for betType, aliases := range p.config.BetTypeAliases {
		for _, alias := range aliases {
			if strings.Contains(text, alias) {
				result = append(result, betType)
				break
			}
		}
	}
	return result
}

// identifyBetTypeFlags 识别下注类型标识（新版本）
func (p *IntelligentBetParser) identifyBetTypeFlags(text string) BetTypeFlags {
	flags := BetTypeFlags{}

	// 检查三中三关键词
	if p.containsBetTypeKeywords(text, "三中三") {
		flags.HasThreeOfThree = true
	}

	// 检查三中二关键词
	if p.containsBetTypeKeywords(text, "三中二") {
		flags.HasThreeOfTwo = true
	}

	// 检查二中二关键词
	if p.containsBetTypeKeywords(text, "二中二") {
		flags.HasTwoOfTwo = true
	}

	// 检查特碰关键词
	if p.containsBetTypeKeywords(text, "特碰") {
		flags.HasSpecial = true
	}

	return flags
}

// containsBetTypeKeywords 检查是否包含指定下注类型的关键词
func (p *IntelligentBetParser) containsBetTypeKeywords(text, betType string) bool {
	if aliases, exists := p.config.BetTypeAliases[betType]; exists {
		for _, alias := range aliases {
			if strings.Contains(text, alias) {
				return true
			}
		}
	}
	return false
}

// processBetType 处理单个下注类型，返回该类型的所有模式信息
func (p *IntelligentBetParser) processBetType(
	betType string,
	sourceNumbers []int,
	unitAmount decimal.Decimal,
	text string,
) BetTypeDetail {

	detail := BetTypeDetail{
		BetType:     betType,
		Modes:       make(map[string]BetModeInfo),
		TotalGroups: 0,
		TotalAmount: decimal.NewFromInt(0),
		HasNumbers:  len(sourceNumbers) > 0,
	}

	// 检查并处理复式模式
	if p.isComplexBet(text) {
		modeInfo := p.processComplexMode(betType, sourceNumbers, unitAmount)
		detail.Modes["complex"] = modeInfo
		detail.TotalGroups += modeInfo.Groups
		detail.TotalAmount = detail.TotalAmount.Add(modeInfo.Amount)
	}

	// 检查并处理拖码模式
	if p.isDragBet(text) {
		modeInfo := p.processDragMode(betType, sourceNumbers, unitAmount, text)
		detail.Modes["drag"] = modeInfo
		detail.TotalGroups += modeInfo.Groups
		detail.TotalAmount = detail.TotalAmount.Add(modeInfo.Amount)
	}

	// 检查并处理多组独立模式
	multiGroups := p.extractMultipleNumberGroups(text)
	if len(multiGroups.NumberGroups) > 1 {
		modeInfo := p.processMultipleMode(betType, multiGroups, unitAmount)
		detail.Modes["multiple"] = modeInfo
		detail.TotalGroups += modeInfo.Groups
		detail.TotalAmount = detail.TotalAmount.Add(modeInfo.Amount)
	} else if !p.isComplexBet(text) && !p.isDragBet(text) {
		// 处理单组模式（当没有其他模式时）
		modeInfo := p.processSingleMode(betType, sourceNumbers, unitAmount)
		detail.Modes["single"] = modeInfo
		detail.TotalGroups += modeInfo.Groups
		detail.TotalAmount = detail.TotalAmount.Add(modeInfo.Amount)
	}

	return detail
}

// processComplexMode 处理复式模式
func (p *IntelligentBetParser) processComplexMode(
	betType string,
	sourceNumbers []int,
	unitAmount decimal.Decimal,
) BetModeInfo {

	modeInfo := BetModeInfo{
		ModeName:   "complex",
		BetDetails: make([]BetDetail, 0),
		UnitAmount: unitAmount,
	}

	var combinations [][]int

	switch betType {
	case "三中三":
		combinations = p.generateNCombinations(sourceNumbers, 3)
	case "二中二":
		combinations = p.generateNCombinations(sourceNumbers, 2)
	case "三中二":
		combinations = p.generateNCombinations(sourceNumbers, 3) // 三中二也是选3个号码
	case "特碰":
		combinations = p.generateNCombinations(sourceNumbers, 2)
	}

	for _, combo := range combinations {
		modeInfo.BetDetails = append(modeInfo.BetDetails, BetDetail{
			Numbers:     combo,
			Amount:      unitAmount,
			Description: fmt.Sprintf("%s 复式", betType),
		})
	}

	modeInfo.Groups = len(combinations)
	modeInfo.Amount = unitAmount.Mul(decimal.NewFromInt(int64(len(combinations))))

	return modeInfo
}

// processDragMode 处理拖码模式
func (p *IntelligentBetParser) processDragMode(
	betType string,
	sourceNumbers []int,
	unitAmount decimal.Decimal,
	text string,
) BetModeInfo {

	modeInfo := BetModeInfo{
		ModeName:   "drag",
		BetDetails: make([]BetDetail, 0),
		UnitAmount: unitAmount,
	}

	// 解析拖码：如 "1,2,3拖10,11,12拖18,20,22"
	dragGroups := p.parseDragGroups(text)

	// 生成拖码组合（笛卡尔积）
	dragCombinations := p.generateDragCombinations(dragGroups, betType)

	for _, combo := range dragCombinations {
		modeInfo.BetDetails = append(modeInfo.BetDetails, BetDetail{
			Numbers:     combo,
			Amount:      unitAmount,
			Description: fmt.Sprintf("%s 拖码", betType),
		})
	}

	modeInfo.Groups = len(dragCombinations)
	modeInfo.Amount = unitAmount.Mul(decimal.NewFromInt(int64(len(dragCombinations))))

	return modeInfo
}

// processMultipleMode 处理多组独立模式
func (p *IntelligentBetParser) processMultipleMode(
	betType string,
	multiGroups NumberGroupsAndAmount,
	unitAmount decimal.Decimal,
) BetModeInfo {

	modeInfo := BetModeInfo{
		ModeName:   "multiple",
		BetDetails: make([]BetDetail, 0),
		UnitAmount: unitAmount,
	}

	for i, numberGroup := range multiGroups.NumberGroups {
		modeInfo.BetDetails = append(modeInfo.BetDetails, BetDetail{
			Numbers:     numberGroup,
			Amount:      unitAmount,
			Description: fmt.Sprintf("%s 第%d组", betType, i+1),
		})
	}

	modeInfo.Groups = len(multiGroups.NumberGroups)
	modeInfo.Amount = unitAmount.Mul(decimal.NewFromInt(int64(len(multiGroups.NumberGroups))))

	return modeInfo
}

// processSingleMode 处理单组模式
func (p *IntelligentBetParser) processSingleMode(
	betType string,
	sourceNumbers []int,
	unitAmount decimal.Decimal,
) BetModeInfo {

	modeInfo := BetModeInfo{
		ModeName:   "single",
		BetDetails: make([]BetDetail, 0),
		UnitAmount: unitAmount,
		Groups:     1,
		Amount:     unitAmount,
	}

	modeInfo.BetDetails = append(modeInfo.BetDetails, BetDetail{
		Numbers:     sourceNumbers,
		Amount:      unitAmount,
		Description: fmt.Sprintf("%s 单组", betType),
	})

	return modeInfo
}

// generateNCombinations 生成N个数的组合（C(n,r)）
func (p *IntelligentBetParser) generateNCombinations(numbers []int, r int) [][]int {
	if len(numbers) < r || r <= 0 {
		return [][]int{}
	}

	var result [][]int

	// 递归生成组合
	var generate func(start int, current []int)
	generate = func(start int, current []int) {
		if len(current) == r {
			// 复制当前组合
			combination := make([]int, len(current))
			copy(combination, current)
			result = append(result, combination)
			return
		}

		for i := start; i <= len(numbers)-(r-len(current)); i++ {
			current = append(current, numbers[i])
			generate(i+1, current)
			current = current[:len(current)-1]
		}
	}

	generate(0, make([]int, 0, r))
	return result
}

// parseDragGroups 解析拖码组（如"1,2,3拖10,11,12拖18,20,22"）
func (p *IntelligentBetParser) parseDragGroups(text string) [][]int {
	var groups [][]int

	// 查找拖码标识符
	dragKeywords := []string{"拖", "tuo"}

	parts := strings.Split(text, ",")
	currentGroup := make([]int, 0)

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// 检查是否包含拖码关键词
		hasDragKeyword := false
		for _, keyword := range dragKeywords {
			if strings.Contains(part, keyword) {
				hasDragKeyword = true
				// 移除拖码关键词，提取数字
				part = strings.ReplaceAll(part, keyword, "")
				break
			}
		}

		// 提取数字
		re := regexp.MustCompile(`\d+`)
		matches := re.FindAllString(part, -1)

		for _, match := range matches {
			if num, err := strconv.Atoi(match); err == nil {
				currentGroup = append(currentGroup, num)
			}
		}

		// 如果发现拖码关键词，则当前组结束
		if hasDragKeyword && len(currentGroup) > 0 {
			groups = append(groups, currentGroup)
			currentGroup = make([]int, 0)
		}
	}

	// 处理最后一组
	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}

	return groups
}

// generateDragCombinations 生成拖码组合（笛卡尔积）
func (p *IntelligentBetParser) generateDragCombinations(dragGroups [][]int, betType string) [][]int {
	if len(dragGroups) == 0 {
		return [][]int{}
	}

	// 根据下注类型确定需要的数字个数
	var requiredCount int
	switch betType {
	case "三中三", "三中二":
		requiredCount = 3
	case "二中二", "特碰":
		requiredCount = 2
	default:
		requiredCount = 2
	}

	var result [][]int

	// 递归生成笛卡尔积
	var generate func(groupIndex int, current []int)
	generate = func(groupIndex int, current []int) {
		if len(current) == requiredCount {
			// 复制当前组合
			combination := make([]int, len(current))
			copy(combination, current)
			result = append(result, combination)
			return
		}

		if groupIndex >= len(dragGroups) {
			return
		}

		// 从当前组选择一个数字
		for _, num := range dragGroups[groupIndex] {
			// 检查是否重复
			duplicate := false
			for _, existing := range current {
				if existing == num {
					duplicate = true
					break
				}
			}

			if !duplicate {
				current = append(current, num)
				generate(groupIndex+1, current)
				current = current[:len(current)-1]
			}
		}
	}

	generate(0, make([]int, 0, requiredCount))
	return result
}

// extractSourceNumbersAndAmount 提取源号码和单组金额
func (p *IntelligentBetParser) extractSourceNumbersAndAmount(text string, context *BetContext) ([]int, decimal.Decimal) {
	// 使用现有的智能提取逻辑
	numbersAndAmount := p.smartExtractNumbersAndAmount(text, []string{}, false, false)

	// 转换为decimal类型
	unitAmount := decimal.NewFromFloat(numbersAndAmount.Amount)

	return numbersAndAmount.Numbers, unitAmount
}

// extractMultipleNumberGroups 提取多组号码（临时实现）
func (p *IntelligentBetParser) extractMultipleNumberGroups(text string) NumberGroupsAndAmount {
	// 简单实现：如果包含复式关键词则返回空，否则使用单组
	if p.isComplexBet(text) || p.isDragBet(text) {
		return NumberGroupsAndAmount{NumberGroups: [][]int{}, Amount: 0}
	}

	// 使用现有的号码提取逻辑
	numbersAndAmount := p.smartExtractNumbersAndAmount(text, []string{}, false, false)
	if len(numbersAndAmount.Numbers) > 0 {
		return NumberGroupsAndAmount{
			NumberGroups: [][]int{numbersAndAmount.Numbers},
			Amount:       numbersAndAmount.Amount,
		}
	}

	return NumberGroupsAndAmount{NumberGroups: [][]int{}, Amount: 0}
}

// parseSingleBet 解析单笔下注（新优化版本）
func (p *IntelligentBetParser) parseSingleBet(betID string, segment string, context *BetContext) SingleBetParsing {
	result := SingleBetParsing{
		BetID:        betID,
		OriginalText: segment,
		LotteryBets:  make(map[string]LotteryBetInfo),
	}

	// 1. 识别体彩类型
	lotteries := p.identifyLotteries(segment)
	if len(lotteries) == 0 && len(context.inheritedLotteries) > 0 {
		lotteries = context.inheritedLotteries
	}
	if len(lotteries) == 0 {
		lotteries = []string{"新澳"}
	}

	// 2. 提取源号码和金额
	sourceNumbers, unitAmount := p.extractSourceNumbersAndAmount(segment, context)

	// 3. 识别下注类型标识
	betTypeFlags := p.identifyBetTypeFlags(segment)

	// 4. 为每个体彩处理
	for _, lottery := range lotteries {
		lotteryInfo := LotteryBetInfo{
			LotteryType:    lottery,
			BetTypeFlags:   betTypeFlags,
			SourceNumbers:  sourceNumbers,
			UnitAmount:     unitAmount,
			BetTypeDetails: make(map[string]BetTypeDetail),
			TotalAmount:    decimal.NewFromInt(0),
			TotalGroups:    0,
		}

		// 5. 处理每种存在的下注类型
		if betTypeFlags.HasThreeOfThree {
			detail := p.processBetType("三中三", sourceNumbers, unitAmount, segment)
			lotteryInfo.BetTypeDetails["三中三"] = detail
			lotteryInfo.TotalAmount = lotteryInfo.TotalAmount.Add(detail.TotalAmount)
			lotteryInfo.TotalGroups += detail.TotalGroups
		}

		if betTypeFlags.HasTwoOfTwo {
			detail := p.processBetType("二中二", sourceNumbers, unitAmount, segment)
			lotteryInfo.BetTypeDetails["二中二"] = detail
			lotteryInfo.TotalAmount = lotteryInfo.TotalAmount.Add(detail.TotalAmount)
			lotteryInfo.TotalGroups += detail.TotalGroups
		}

		if betTypeFlags.HasThreeOfTwo {
			detail := p.processBetType("三中二", sourceNumbers, unitAmount, segment)
			lotteryInfo.BetTypeDetails["三中二"] = detail
			lotteryInfo.TotalAmount = lotteryInfo.TotalAmount.Add(detail.TotalAmount)
			lotteryInfo.TotalGroups += detail.TotalGroups
		}

		if betTypeFlags.HasSpecial {
			detail := p.processBetType("特碰", sourceNumbers, unitAmount, segment)
			lotteryInfo.BetTypeDetails["特碰"] = detail
			lotteryInfo.TotalAmount = lotteryInfo.TotalAmount.Add(detail.TotalAmount)
			lotteryInfo.TotalGroups += detail.TotalGroups
		}

		// 检查是否有下注类型但没有具体号码
		hasAnyBetType := betTypeFlags.HasThreeOfThree || betTypeFlags.HasTwoOfTwo ||
			betTypeFlags.HasThreeOfTwo || betTypeFlags.HasSpecial

		if hasAnyBetType && len(sourceNumbers) == 0 {
			result.HasError = true
			result.ErrorMessage = "识别到下注类型但没有找到具体号码"
			return result
		}

		// 如果没有任何下注类型，尝试推断
		if !hasAnyBetType && len(sourceNumbers) > 0 {
			inferredBetType := p.inferBetTypeFromNumbers(sourceNumbers)
			if inferredBetType != "" {
				detail := p.processBetType(inferredBetType, sourceNumbers, unitAmount, segment)
				lotteryInfo.BetTypeDetails[inferredBetType] = detail
				lotteryInfo.TotalAmount = lotteryInfo.TotalAmount.Add(detail.TotalAmount)
				lotteryInfo.TotalGroups += detail.TotalGroups
			}
		}

		result.LotteryBets[lottery] = lotteryInfo
	}

	// 生成统计信息
	result.BetStatistics = p.generateNewBetStatistics(result.LotteryBets)

	return result
}

// inferBetTypeFromNumbers 根据号码数量推断下注类型
func (p *IntelligentBetParser) inferBetTypeFromNumbers(numbers []int) string {
	count := len(numbers)
	switch count {
	case 2:
		return "二中二"
	case 3:
		return "三中三"
	default:
		if count > 3 {
			return "三中三" // 默认复式三中三
		}
		return ""
	}
}

// generateNewBetStatistics 生成新版本的下注统计
func (p *IntelligentBetParser) generateNewBetStatistics(lotteryBets map[string]LotteryBetInfo) BetStatistics {
	stats := BetStatistics{
		TotalAmount:         decimal.NewFromInt(0),
		TotalGroups:         0,
		LotteryCount:        len(lotteryBets),
		LotteryBetTypeStats: make(map[string]map[string]BetTypeStat),
	}

	for lotteryType, lotteryInfo := range lotteryBets {
		stats.TotalAmount = stats.TotalAmount.Add(lotteryInfo.TotalAmount)
		stats.TotalGroups += lotteryInfo.TotalGroups

		// 初始化该体彩的统计
		if stats.LotteryBetTypeStats[lotteryType] == nil {
			stats.LotteryBetTypeStats[lotteryType] = make(map[string]BetTypeStat)
		}

		// 统计各下注类型
		for betType, detail := range lotteryInfo.BetTypeDetails {
			stats.LotteryBetTypeStats[lotteryType][betType] = BetTypeStat{
				Amount: detail.TotalAmount,
				Groups: detail.TotalGroups,
				Count:  1,
			}
		}
	}

	return stats
}

// smartExtractNumbersAndAmount 智能提取号码和金额
func (p *IntelligentBetParser) smartExtractNumbersAndAmount(text string, betTypes []string, isComplex, isDrag bool) NumbersAndAmount {
	result := NumbersAndAmount{}

	// 1. 先尝试识别明确的金额模式
	amount := p.extractAmountSmart(text)
	if amount > 0 {
		// 找到明确金额，提取号码
		result.Amount = amount
		result.Numbers = p.extractNumbersExcludingAmount(text, amount)
		return result
	}
	safeLogger.AppendLog(fmt.Sprintf("智能提取金额: %f", amount))

	// 2. 没有明确金额标识，需要智能分析
	allNumbers := p.extractAllNumbers(text)
	if len(allNumbers) < 2 {
		return result // 至少需要1个号码+1个金额
	}

	// 3. 分析模式，寻找最后一个重复数字作为金额
	result = p.analyzeNumberPattern(allNumbers, text)

	return result
}

// extractAmountSmart 智能提取金额
func (p *IntelligentBetParser) extractAmountSmart(text string) float64 {
	// 金额模式（按优先级排序）
	patterns := []string{
		`各(\d+)`,  // 各20
		`每组(\d+)`, // 每组20
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			if amount, err := strconv.ParseFloat(matches[1], 64); err == nil {
				return amount
			}
		}
	}
	return 0
}

// analyzeNumberPattern 分析数字模式
func (p *IntelligentBetParser) analyzeNumberPattern(numbers []int, text string) NumbersAndAmount {
	result := NumbersAndAmount{}

	if len(numbers) < 2 {
		return result
	}

	// 对于12-38=20这种情况，根据分隔符数量推断下注类型
	separatorCount := strings.Count(text, "-") + strings.Count(text, ".") + strings.Count(text, ",")

	if separatorCount == 2 {
		// 两个分隔符，默认二中二，第三个是金额
		if len(numbers) >= 3 {
			result.Numbers = numbers[:2]
			result.Amount = float64(numbers[2])
			return result
		}
	} else if separatorCount == 3 {
		// 三个分隔符，默认三中三，第四个是金额
		if len(numbers) >= 4 {
			result.Numbers = numbers[:3]
			result.Amount = float64(numbers[3])
			return result
		}
	}

	// 寻找最后出现的重复数字，很可能是金额
	lastNum := numbers[len(numbers)-1]

	// 检查是否有多个相同的最后数字
	sameAsLastCount := 0
	for _, num := range numbers {
		if num == lastNum {
			sameAsLastCount++
		}
	}

	if sameAsLastCount > 1 {
		// 最后的数字重复出现，很可能是金额
		result.Amount = float64(lastNum)
		// 提取不等于金额的号码
		for _, num := range numbers {
			if num != lastNum && num >= 1 && num <= 49 {
				result.Numbers = append(result.Numbers, num)
			}
		}
	} else {
		// 分析特殊模式：如 12,30,50,30 或 8,20,50,30
		result = p.analyzeSpecialPatterns(numbers, text)
	}

	return result
}

// analyzeSpecialPatterns 分析特殊模式
func (p *IntelligentBetParser) analyzeSpecialPatterns(numbers []int, text string) NumbersAndAmount {
	result := NumbersAndAmount{}

	if len(numbers) < 3 {
		return result
	}

	// 检查最后一个数字是否可能是金额（通常金额会比较小或者重复）
	lastNum := numbers[len(numbers)-1]

	// 如果最后数字小于等于1000且前面有至少2个数字，认为是金额
	if lastNum <= 1000 && len(numbers) >= 3 {
		result.Amount = float64(lastNum)
		result.Numbers = numbers[:len(numbers)-1]

		// 过滤无效号码
		validNumbers := make([]int, 0)
		for _, num := range result.Numbers {
			if num >= 1 && num <= 49 {
				validNumbers = append(validNumbers, num)
			}
		}
		result.Numbers = validNumbers
	} else {
		// 默认处理：最后一个数字是金额
		if len(numbers) > 1 {
			result.Amount = float64(numbers[len(numbers)-1])
			result.Numbers = numbers[:len(numbers)-1]
		}
	}

	return result
}

// extractNumbersExcludingAmount 提取号码，排除金额
func (p *IntelligentBetParser) extractNumbersExcludingAmount(text string, amount float64) []int {
	allNumbers := p.extractAllNumbers(text)
	amountInt := int(amount)

	numbers := make([]int, 0)
	for _, num := range allNumbers {
		if num != amountInt && num >= 1 && num <= 49 {
			numbers = append(numbers, num)
		}
	}
	return numbers
}

// extractAllNumbers 提取所有数字
func (p *IntelligentBetParser) extractAllNumbers(text string) []int {
	numberRe := regexp.MustCompile(`\d+`)
	matches := numberRe.FindAllString(text, -1)

	numbers := make([]int, 0)
	for _, match := range matches {
		if num, err := strconv.Atoi(match); err == nil {
			numbers = append(numbers, num)
		}
	}
	return numbers
}

// inferBetTypes 根据号码数量推断下注类型
func (p *IntelligentBetParser) inferBetTypes(numbers []int, text string) []string {
	numCount := len(numbers)

	// 检查是否有特殊下注类型关键词
	if strings.Contains(text, "特串") || strings.Contains(text, "特碰") {
		return []string{"特碰"}
	}

	switch numCount {
	case 1:
		return []string{"特碰"}
	case 2:
		return []string{"二中二"}
	case 3:
		return []string{"三中三"}
	default:
		if numCount > 3 {
			return []string{"三中三"} // 默认三中三
		}
		return []string{"特碰"} // 默认特碰
	}
}

// isComplexBet 检查是否复式下注
func (p *IntelligentBetParser) isComplexBet(text string) bool {
	return strings.Contains(text, "复式")
}

// isDragBet 检查是否拖码下注
func (p *IntelligentBetParser) isDragBet(text string) bool {
	return strings.Contains(text, "拖")
}

// calculateGroups 计算组数
func (p *IntelligentBetParser) calculateGroups(numbers []int, betType string, isComplex, isDrag bool) int {
	if isDrag {
		// 拖码计算：需要更复杂的逻辑 - 简化实现
		return 1
	}

	if isComplex {
		// 复式计算
		switch betType {
		case "三中三":
			return p.combination(len(numbers), 3)
		case "三中二":
			return p.combination(len(numbers), 3)
		case "二中二":
			return p.combination(len(numbers), 2)
		case "特碰":
			return p.combination(len(numbers), 2)
		}
	}

	return 1 // 默认1组
}

// combination 计算组合数C(n,r)
func (p *IntelligentBetParser) combination(n, r int) int {
	if r > n || r < 0 {
		return 0
	}
	if r == 0 || r == n {
		return 1
	}

	result := 1
	for i := 0; i < r; i++ {
		result = result * (n - i) / (i + 1)
	}
	return result
}

// generateBetStatistics 生成单笔下注统计（保留旧版本兼容）
func (p *IntelligentBetParser) generateBetStatistics(lotteryBets map[string]LotteryBetInfo) BetStatistics {
	stats := BetStatistics{
		TotalAmount:         decimal.NewFromInt(0),
		LotteryBetTypeStats: make(map[string]map[string]BetTypeStat),
	}

	for lottery, lotteryInfo := range lotteryBets {
		stats.TotalAmount = stats.TotalAmount.Add(lotteryInfo.TotalAmount)
		stats.TotalGroups += lotteryInfo.TotalGroups
		stats.LotteryCount++

		betTypeStats := make(map[string]BetTypeStat)
		for betType, betTypeInfo := range lotteryInfo.BetTypeDetails {
			betTypeStats[betType] = BetTypeStat{
				Amount: betTypeInfo.TotalAmount,
				Groups: betTypeInfo.TotalGroups,
				Count:  1,
			}
		}
		stats.LotteryBetTypeStats[lottery] = betTypeStats
	}

	return stats
}

// generateRoundStatistics 生成整轮统计（新版本）
func (p *IntelligentBetParser) generateRoundStatistics(parsedBets []SingleBetParsing) RoundBetStatistics {
	stats := RoundBetStatistics{
		TotalAmount:         decimal.NewFromInt(0),
		TotalBets:           len(parsedBets),
		LotteryBetTypeStats: make(map[string]map[string]BetTypeStat),
		LotteryTotals:       make(map[string]BetTypeStat),
		BetTypeTotals:       make(map[string]BetTypeStat),
	}

	for _, bet := range parsedBets {
		if bet.HasError {
			continue
		}

		stats.TotalAmount = stats.TotalAmount.Add(bet.BetStatistics.TotalAmount)
		stats.TotalGroups += bet.BetStatistics.TotalGroups

		// 累计各项统计
		for lottery, lotteryStats := range bet.BetStatistics.LotteryBetTypeStats {
			if _, exists := stats.LotteryBetTypeStats[lottery]; !exists {
				stats.LotteryBetTypeStats[lottery] = make(map[string]BetTypeStat)
			}

			lotteryTotal := stats.LotteryTotals[lottery]

			for betType, betStat := range lotteryStats {
				existing := stats.LotteryBetTypeStats[lottery][betType]
				existing.Amount = existing.Amount.Add(betStat.Amount)
				existing.Groups += betStat.Groups
				existing.Count += betStat.Count
				stats.LotteryBetTypeStats[lottery][betType] = existing

				lotteryTotal.Amount = lotteryTotal.Amount.Add(betStat.Amount)
				lotteryTotal.Groups += betStat.Groups
				lotteryTotal.Count += betStat.Count

				typeTotal := stats.BetTypeTotals[betType]
				typeTotal.Amount = typeTotal.Amount.Add(betStat.Amount)
				typeTotal.Groups += betStat.Groups
				typeTotal.Count += betStat.Count
				stats.BetTypeTotals[betType] = typeTotal
			}

			stats.LotteryTotals[lottery] = lotteryTotal
		}
	}

	return stats
}
