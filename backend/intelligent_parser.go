package backend

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
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

// BetContext 下注上下文
type BetContext struct {
	inheritedLotteries []string // 继承的体彩类型
	currentLotteries   []string // 当前识别的体彩类型
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

	// 1. 字符串预处理
	processedText := p.preprocessText(request.Input)

	// 2. 替换关键词（生肖、颜色、尾数）
	expandedText := p.replaceKeywords(processedText)

	// 3. 移除非关键字的所有中文
	cleanedText := p.removeChineseChars(expandedText)

	// 4. 分割为多笔下注
	betSegments := p.segmentBets(cleanedText)

	// 5. 解析每笔下注
	parsedBets := make([]SingleBetParsing, 0)
	context := &BetContext{inheritedLotteries: make([]string, 0)}

	for i, segment := range betSegments {
		betID := fmt.Sprintf("%s_bet_%d", roundID, i+1)
		parsed := p.parseSingleBet(betID, segment, context)
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
			text = strings.ReplaceAll(text, keyword, replacement)
		}
	}

	return text
}

// removeChineseChars 移除非关键字的所有中文字符
func (p *IntelligentBetParser) removeChineseChars(text string) string {
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

	// 添加结束关键词
	preservedKeywords = append(preservedKeywords, p.config.EndKeywords...)

	// 添加其他关键词
	preservedKeywords = append(preservedKeywords, "复式", "拖", "各", "每组", "元", "块", "死活", "硬软", "特串", "特碰")

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
		text = strings.ReplaceAll(text, placeholder, keyword)
	}

	return text
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
		for _, endKeyword := range endKeywords {
			if strings.Contains(word, endKeyword) {
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

// parseSingleBet 解析单笔下注（完整版）
func (p *IntelligentBetParser) parseSingleBet(betID string, segment string, context *BetContext) SingleBetParsing {
	result := SingleBetParsing{
		BetID:        betID,
		OriginalText: segment,
		LotteryBets:  make(map[string]LotteryBetInfo),
	}

	// 1. 识别体彩类型
	lotteries := p.identifyLotteries(segment)
	if len(lotteries) == 0 && len(context.inheritedLotteries) > 0 {
		lotteries = context.inheritedLotteries // 继承上一笔的体彩
	}
	if len(lotteries) == 0 {
		lotteries = []string{"新澳"} // 默认新澳
	}

	// 2. 识别下注类型
	betTypes := p.identifyBetTypes(segment)

	// 3. 检查复式和拖码
	isComplex := p.isComplexBet(segment)
	isDrag := p.isDragBet(segment)

	// 4. 根据下注类型智能提取号码和金额
	numbersAndAmount := p.smartExtractNumbersAndAmount(segment, betTypes, isComplex, isDrag)

	if len(numbersAndAmount.Numbers) == 0 || numbersAndAmount.Amount <= 0 {
		result.HasError = true
		result.ErrorMessage = "无法提取有效的号码或金额"
		return result
	}

	// 5. 为每个体彩创建下注信息
	for _, lottery := range lotteries {
		lotteryInfo := LotteryBetInfo{
			LotteryType: lottery,
			BetTypes:    make(map[string]BetTypeInfo),
		}

		// 如果没有明确的下注类型，根据号码数量推断
		if len(betTypes) == 0 {
			betTypes = p.inferBetTypes(numbersAndAmount.Numbers, segment)
		}

		for _, betType := range betTypes {
			groups := p.calculateGroups(numbersAndAmount.Numbers, betType, isComplex, isDrag)
			totalAmount := float64(groups) * numbersAndAmount.Amount

			betTypeInfo := BetTypeInfo{
				BetType:     betType,
				TotalGroups: groups,
				TotalAmount: totalAmount,
				IsComplex:   isComplex,
				IsDrag:      isDrag,
				HasNumbers:  true,
				BetDetails: []BetDetail{{
					Numbers:     numbersAndAmount.Numbers,
					Amount:      numbersAndAmount.Amount,
					Groups:      groups,
					Description: fmt.Sprintf("%s %s", lottery, betType),
				}},
			}

			lotteryInfo.BetTypes[betType] = betTypeInfo
			lotteryInfo.TotalAmount += totalAmount
			lotteryInfo.TotalGroups += groups
		}

		result.LotteryBets[lottery] = lotteryInfo
	}

	// 生成统计信息
	result.BetStatistics = p.generateBetStatistics(result.LotteryBets)

	return result
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
		`(\d+)元`,  // 20元
		`(\d+)块`,  // 20块
		`=(\d+)`,  // =20
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
	return strings.Contains(text, "复式") || strings.Contains(text, "复试")
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

// generateBetStatistics 生成单笔下注统计
func (p *IntelligentBetParser) generateBetStatistics(lotteryBets map[string]LotteryBetInfo) BetStatistics {
	stats := BetStatistics{
		LotteryBetTypeStats: make(map[string]map[string]BetTypeStat),
	}

	for lottery, lotteryInfo := range lotteryBets {
		stats.TotalAmount += lotteryInfo.TotalAmount
		stats.TotalGroups += lotteryInfo.TotalGroups
		stats.LotteryCount++

		betTypeStats := make(map[string]BetTypeStat)
		for betType, betTypeInfo := range lotteryInfo.BetTypes {
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

// generateRoundStatistics 生成整轮统计
func (p *IntelligentBetParser) generateRoundStatistics(parsedBets []SingleBetParsing) RoundBetStatistics {
	stats := RoundBetStatistics{
		TotalBets:           len(parsedBets),
		LotteryBetTypeStats: make(map[string]map[string]BetTypeStat),
		LotteryTotals:       make(map[string]BetTypeStat),
		BetTypeTotals:       make(map[string]BetTypeStat),
	}

	for _, bet := range parsedBets {
		if bet.HasError {
			continue
		}

		stats.TotalAmount += bet.BetStatistics.TotalAmount
		stats.TotalGroups += bet.BetStatistics.TotalGroups

		// 累计各项统计
		for lottery, lotteryStats := range bet.BetStatistics.LotteryBetTypeStats {
			if _, exists := stats.LotteryBetTypeStats[lottery]; !exists {
				stats.LotteryBetTypeStats[lottery] = make(map[string]BetTypeStat)
			}

			lotteryTotal := stats.LotteryTotals[lottery]

			for betType, betStat := range lotteryStats {
				existing := stats.LotteryBetTypeStats[lottery][betType]
				existing.Amount += betStat.Amount
				existing.Groups += betStat.Groups
				existing.Count += betStat.Count
				stats.LotteryBetTypeStats[lottery][betType] = existing

				lotteryTotal.Amount += betStat.Amount
				lotteryTotal.Groups += betStat.Groups
				lotteryTotal.Count += betStat.Count

				typeTotal := stats.BetTypeTotals[betType]
				typeTotal.Amount += betStat.Amount
				typeTotal.Groups += betStat.Groups
				typeTotal.Count += betStat.Count
				stats.BetTypeTotals[betType] = typeTotal
			}

			stats.LotteryTotals[lottery] = lotteryTotal
		}
	}

	return stats
}
