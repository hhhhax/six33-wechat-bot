package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/shopspring/decimal"
)

var roundIDCounter int64

// IntelligentBetParser 智能下注解析器
type IntelligentBetParser struct {
	config IntelligentBetParserConfig
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
	processedText := request.Input
	safeLogger.AppendLog(fmt.Sprintf("开始智能解析: %s", processedText))
	// 1. 替换关键词（生肖、颜色、尾数）
	processedText = p.replaceKeywords(processedText)
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
			result.ErrorMessages = append(result.ErrorMessages, bet.ErrorMessage...)
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

	// 2. 智能处理分隔符
	text = p.smartReplaceSeparators(text)


	// 定义你想要替换的符号字符串
	symbolsToReplace := "【】[]{}“‘”’" // 可以在这里添加你需要的任何符号
	// 将字符串中的所有元字符转义，然后作为正则表达式的模式
	pattern := "[" + regexp.QuoteMeta(symbolsToReplace) + "]+"

	invalidCharsRe := regexp.MustCompile(pattern)
	text = invalidCharsRe.ReplaceAllString(text, " ")

	// 3. 处理分隔符后的空格：保留分隔符，移除空格
	spaceAfterSeparatorRe := regexp.MustCompile(`([./\\\-=:,，、+。*])\s+`)
	text = spaceAfterSeparatorRe.ReplaceAllString(text, "$1")

	// 4. 清理多余空格
	spaceRe := regexp.MustCompile(`\s+`)
	text = spaceRe.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// smartReplaceSeparators 智能替换分隔符
func (p *IntelligentBetParser) smartReplaceSeparators(text string) string {
	// 定义所有分隔符（同一级别）
	separators := map[rune]bool{
		'.': true, '/': true, '\\': true, '-': true, '=': true, ':': true,
		',': true, '，': true, '、': true, '+': true, '。': true, '*': true,
	}
	
	result := make([]rune, 0, len(text))
	runes := []rune(text)
	
	for i := 0; i < len(runes); i++ {
		char := runes[i]
		
		// 如果是分隔符
		if separators[char] {
			// 处理连续分隔符
			separatorGroup := []rune{char}
			j := i + 1
			
			// 收集连续的分隔符
			for j < len(runes) && separators[runes[j]] {
				separatorGroup = append(separatorGroup, runes[j])
				j++
			}
			
			// 判断用什么替换
			replacement := p.decideSeparatorReplacement(runes, i, separatorGroup)
			result = append(result, []rune(replacement)...)
			
			// 跳过已处理的分隔符
			i = j - 1
		} else {
			result = append(result, char)
		}
	}
	
	return string(result)
}

// decideSeparatorReplacement 决定分隔符的替换方式
func (p *IntelligentBetParser) decideSeparatorReplacement(runes []rune, startPos int, separatorGroup []rune) string {
	// 1. 检查前后是否为数字
	prevIsDigit := false
	if startPos > 0 && unicode.IsDigit(runes[startPos-1]) {
		prevIsDigit = true
	}
	
	nextIsDigit := false
	nextPos := startPos + len(separatorGroup)
	if nextPos < len(runes) && unicode.IsDigit(runes[nextPos]) {
		nextIsDigit = true
	}
	
	// 2. 如果前后都是数字，且分隔符组是同一类型，替换为"-"
	if prevIsDigit && nextIsDigit && p.isSameSeparatorType(separatorGroup) {
		return "-"
	}
	
	// 3. 其他情况替换为空格
	return " "
}

// isSameSeparatorType 检查分隔符组是否为同一类型
func (p *IntelligentBetParser) isSameSeparatorType(separatorGroup []rune) bool {
	if len(separatorGroup) == 0 {
		return false
	}
	
	firstSep := separatorGroup[0]
	for _, sep := range separatorGroup {
		if sep != firstSep {
			return false
		}
	}
	return true
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

	// 替换关键词,且默认不加
	for _, keyword := range keywords {
		if numbers, exists := allKeywords[keyword]; exists {
			numbersStr := make([]string, len(numbers))
			for i, num := range numbers {
				numbersStr[i] = fmt.Sprintf("%02d", num)
			}
			replacement := strings.Join(numbersStr, "-")

			var newText strings.Builder
			lastIndex := 0

			for {
				index := strings.Index(text[lastIndex:], keyword)
				if index == -1 {
					break
				}

				realIndex := lastIndex + index

				newText.WriteString(text[lastIndex:realIndex])

				if realIndex > 0 && unicode.IsDigit(rune(text[realIndex-1])) {
					newText.WriteString("-" + replacement)
				} else {
					newText.WriteString(replacement)
				}

				lastIndex = realIndex + len(keyword)
			}

			newText.WriteString(text[lastIndex:])
			text = newText.String()
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
		if strings.Contains(text, "-"+keyword+"-") {
			text = strings.ReplaceAll(text, "-"+keyword+"-", keyword)
		}
		if strings.Contains(text, "-"+keyword) {
			text = strings.ReplaceAll(text, "-"+keyword, keyword)
		}
		if strings.Contains(text, keyword+"-") {
			text = strings.ReplaceAll(text, keyword+"-", keyword)
		}
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

// segmentBets 通过金额分割为多笔下注
func (p *IntelligentBetParser) segmentBets(text string) []string {
	// 金额模式（按优先级排序）
	patterns := []string{
		`各(\d+)`,  // 各20
		`每组(\d+)`, // 每组20
	}

	segments := make([]string, 0)

	// 编译所有正则表达式
	compiledPatterns := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		compiledPatterns[i] = regexp.MustCompile(pattern)
	}

	// 查找所有金额位置
	amountPositions := make([]AmountMatch, 0)

	for _, re := range compiledPatterns {
		matches := re.FindAllStringSubmatchIndex(text, -1)
		for _, match := range matches {
			if len(match) >= 4 { // 确保有捕获组
				amountPositions = append(amountPositions, AmountMatch{
					Start: match[0],
					End:   match[1],
				})
			}
		}
	}

	// 如果没有找到金额，返回整个文本作为一段
	if len(amountPositions) == 0 {
		return []string{strings.TrimSpace(text)}
	}

	// 按位置排序
	sort.Slice(amountPositions, func(i, j int) bool {
		return amountPositions[i].Start < amountPositions[j].Start
	})

	// 去重相同位置的匹配
	uniquePositions := make([]AmountMatch, 0)
	for i, pos := range amountPositions {
		if i == 0 || pos.Start != amountPositions[i-1].Start {
			uniquePositions = append(uniquePositions, pos)
		}
	}

	// 根据金额位置分割文本
	lastEnd := 0
	for i, pos := range uniquePositions {
		// 分割点：当前金额表达式的结束位置
		splitPoint := pos.End

		// 如果不是第一个金额，从上一个分割点开始
		if i > 0 {
			segmentText := text[lastEnd:splitPoint]
			segmentText = p.cleanSegment(segmentText)
			if segmentText != "" {
				segments = append(segments, segmentText)
			}
		} else {
			// 第一个金额，从文本开头到金额结束
			segmentText := text[0:splitPoint]
			segmentText = p.cleanSegment(segmentText)
			if segmentText != "" {
				segments = append(segments, segmentText)
			}
		}

		lastEnd = splitPoint
	}

	// 处理最后一段（如果有剩余文本）
	if lastEnd < len(text) {
		segmentText := text[lastEnd:]
		segmentText = p.cleanSegment(segmentText)
		if segmentText != "" {
			segments = append(segments, segmentText)
		}
	}

	return segments
}

// cleanSegment 清理分段文本，移除前后的"-"符号
func (p *IntelligentBetParser) cleanSegment(segment string) string {
	// 去除首尾空白
	segment = strings.TrimSpace(segment)

	// 移除开头的"-"符号
	for strings.HasPrefix(segment, "-") {
		segment = strings.TrimPrefix(segment, "-")
		segment = strings.TrimSpace(segment)
	}

	// 移除结尾的"-"符号
	for strings.HasSuffix(segment, "-") {
		segment = strings.TrimSuffix(segment, "-")
		segment = strings.TrimSpace(segment)
	}

	return segment
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
				text = strings.ReplaceAll(text, alias, "")
				result = append(result, lotteryType)
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
	if strings.Contains(text, "三中三") {
		flags.HasThreeOfThree = true
	}

	// 检查三中二关键词
	if strings.Contains(text, "三中二") {
		flags.HasThreeOfTwo = true
	}

	// 检查二中二关键词
	if strings.Contains(text, "二中二") {
		flags.HasTwoOfTwo = true
	}

	// 检查特碰关键词
	if strings.Contains(text, "特碰") {
		flags.HasSpecial = true
	}

	return flags
}

// processBetType 处理单个下注类型，返回该类型的所有模式信息
func (p *IntelligentBetParser) processBetType(
	betType string,
	text string,
) (*BetTypeDetail, error) {

	detail := &BetTypeDetail{
		BetType:     betType,
		Modes:       make(map[string]BetModeInfo),
		TotalGroups: 0,
		TotalAmount: decimal.NewFromInt(0),
	}
	// 检查并处理拖码模式
	if p.isDragBet(text) {
		modeInfo, err := p.processDragMode(betType, text)
		if err != nil {
			return nil, err
		}
		detail.Modes["drag"] = *modeInfo
		detail.TotalGroups += modeInfo.Groups
		detail.TotalAmount = detail.TotalAmount.Add(modeInfo.Amount)
	}

	// 检查并处理复式模式
	if p.isComplexBet(betType, text) {
		modeInfo, err := p.processComplexMode(betType, text)
		if err != nil {
			return nil, err
		}
		detail.Modes["complex"] = *modeInfo
		detail.TotalGroups += modeInfo.Groups
		detail.TotalAmount = detail.TotalAmount.Add(modeInfo.Amount)
	}

	return detail, nil
}

// processComplexMode 处理复式
func (p *IntelligentBetParser) processComplexMode(
	betType string,
	text string,
) (*BetModeInfo, error) {
	unitAmount := p.extractAmountSmart(text)
	if unitAmount.IsZero() {
		return nil, errors.New("存在复式下注，但不存在下注金额,请在下注金额前手动添加'各'或'每组'表示每组金额")
	}

	modeInfo := &BetModeInfo{
		ModeName:   "complex",
		BetDetails: make([]BetDetail, 0),
		UnitAmount: unitAmount,
	}

	var combinations []BetCombination

	switch betType {
	case "三中三":
		combinations, _ = p.generateNCombinations(text, 3)
	case "二中二":
		combinations, _ = p.generateNCombinations(text, 2)
	case "三中二":
		combinations, _ = p.generateNCombinations(text, 3)
	case "特碰":
		combinations, _ = p.generateNCombinations(text, 2)
	default:
		return nil, fmt.Errorf("不支持的下注类型: %s", betType)
	}

	// 如果没有找到组合，返回错误
	if combinations == nil {
		return nil, errors.New("未找到有效的下注组合")
	}

	for _, combo := range combinations {
		modeInfo.BetDetails = append(modeInfo.BetDetails, BetDetail{
			Numbers:     combo.Numbers,
			Amount:      unitAmount,
			Description: fmt.Sprintf("复式%s: %s", betType, combo.OriginalText),
		})
	}

	modeInfo.Groups = len(combinations)
	modeInfo.Amount = unitAmount.Mul(decimal.NewFromInt(int64(len(combinations))))

	return modeInfo, nil
}

// processDragMode 处理拖码模式，支持多组输入
func (p *IntelligentBetParser) processDragMode(
	betType string,
	text string,
) (*BetModeInfo, error) {
	unitAmount := p.extractAmountSmart(text)
	if unitAmount.IsZero() {
		return nil, errors.New("存在拖类型下注，但不存在下注金额,请在下注金额前手动添加'各'或'每组'表示每组金额")
	}

	modeInfo := &BetModeInfo{
		ModeName:   "drag",
		BetDetails: make([]BetDetail, 0),
		UnitAmount: unitAmount,
	}

	// 使用正则表达式按空格或逗号分割整个文本，以处理多组拖码
	reDragGroups := regexp.MustCompile(`(\d{1,2}(?:-\d{1,2})*拖\d{1,2}(?:-\d{1,2})*(?:拖\d{1,2}(?:-\d{1,2})*)?)`)
	dragStrings := reDragGroups.FindAllString(text, -1)

	if len(dragStrings) == 0 {
		return nil, errors.New("未找到有效的拖码组合")
	}

	var totalCombinations int64 = 0

	for _, dragString := range dragStrings {
		// 解析单个拖码组，如 "1-2-3拖10-11-12"
		dragGroups, err := p.parseDragGroups(dragString)
		if err != nil {
			return nil, err
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

		// 生成拖码组合（笛卡尔积）
		combinations, _ := p.generateCartesianProduct(dragGroups, requiredCount)

		for _, combo := range combinations {
			modeInfo.BetDetails = append(modeInfo.BetDetails, BetDetail{
				Numbers:     combo,
				Amount:      unitAmount,
				Description: fmt.Sprintf("%s拖码: %s", betType, dragString),
			})
		}
		totalCombinations += int64(len(combinations))
	}

	modeInfo.Groups = int(totalCombinations)
	modeInfo.Amount = unitAmount.Mul(decimal.NewFromInt(totalCombinations))

	return modeInfo, nil
}

// generateNCombinations 完整优化版本
func (p *IntelligentBetParser) generateNCombinations(text string, n int) ([]BetCombination, error) {
	// 使用最简单的正则，完全兼容Go
	re := regexp.MustCompile(`\d{1,2}(?:-\d{1,2})*`)

	// 找到所有匹配及其位置
	matchIndices := re.FindAllStringIndex(text, -1)

	var allCombinations []BetCombination
	for _, indices := range matchIndices {
		start, end := indices[0], indices[1]
		match := text[start:end]

		// 检查前后字符，过滤掉"拖"字相邻或数字相邻的情况
		shouldSkip := false

		// 检查前一个字符
		if start > 0 {
			// 安全地获取前一个字符（处理UTF-8）
			beforeText := text[:start]
			if len(beforeText) > 0 {
				runes := []rune(beforeText)
				if len(runes) > 0 {
					prevRune := runes[len(runes)-1]
					if prevRune == '拖' || unicode.IsDigit(prevRune) {
						shouldSkip = true
					}
				}
			}
		}

		// 检查后一个字符
		if end < len(text) && !shouldSkip {
			// 安全地获取后一个字符（处理UTF-8）
			afterText := text[end:]
			if len(afterText) > 0 {
				runes := []rune(afterText)
				if len(runes) > 0 {
					nextRune := runes[0]
					if nextRune == '拖' || unicode.IsDigit(nextRune) {
						shouldSkip = true
					}
				}
			}
		}

		if shouldSkip {
			continue
		}

		// 分割字符串得到单个数字
		numStrings := strings.Split(match, "-")
		if len(numStrings) < n {
			continue
		}

		// 将字符串转换为整数
		var numbers []int
		for _, s := range numStrings {
			num, err := strconv.Atoi(s)
			if err != nil {
				continue
			}
			numbers = append(numbers, num)
		}

		// 确保组合的数字个数正好是n
		if len(numbers) == n {
			allCombinations = append(allCombinations, BetCombination{
				Numbers:      numbers,
				OriginalText: match,
			})
		} else if len(numbers) > n {
			// 如果组合数字多于n个，则生成所有n个数字的组合
			combinations, err := p.getCombinations(numbers, n)
			if err != nil {
				return nil, err
			}
			for _, combo := range combinations {
				allCombinations = append(allCombinations, BetCombination{
					Numbers:      combo,
					OriginalText: match,
				})
			}
		}
	}

	if len(allCombinations) == 0 {
		return nil, errors.New("未找到有效的下注组合")
	}

	return allCombinations, nil
}

// getCombinations 从一组数字中生成所有n个数字的组合
func (p *IntelligentBetParser) getCombinations(numbers []int, n int) ([][]int, error) {
	if n < 1 || n > len(numbers) {
		return nil, fmt.Errorf("无效的组合数n: %d", n)
	}

	var result [][]int
	var f func(start int, combo []int)

	f = func(start int, combo []int) {
		if len(combo) == n {
			temp := make([]int, n)
			copy(temp, combo)
			result = append(result, temp)
			return
		}

		for i := start; i < len(numbers); i++ {
			f(i+1, append(combo, numbers[i]))
		}
	}

	f(0, []int{})
	return result, nil
}

// parseDragGroups 解析单个拖码组，如"1-2-3拖10-11-12"
func (p *IntelligentBetParser) parseDragGroups(text string) ([][]int, error) {
	var groups [][]int

	// 使用正则表达式按 "拖" 分割字符串，并提取数字
	re := regexp.MustCompile(`\b\d{1,2}(?:-\d{1,2})*\b`)
	parts := strings.Split(text, "拖")

	for _, part := range parts {
		matches := re.FindAllString(part, -1)
		var groupNumbers []int
		for _, match := range matches {
			numStrings := strings.Split(match, "-")
			for _, s := range numStrings {
				if num, err := strconv.Atoi(s); err == nil {
					groupNumbers = append(groupNumbers, num)
				}
			}
		}
		if len(groupNumbers) > 0 {
			groups = append(groups, groupNumbers)
		}
	}

	if len(groups) == 0 {
		return nil, errors.New("未找到有效的拖码组")
	}

	return groups, nil
}

// generateCartesianProduct 生成笛卡尔积
func (p *IntelligentBetParser) generateCartesianProduct(sets [][]int, requiredSize int) ([][]int, error) {
	var result [][]int
	if len(sets) == 0 {
		return result, nil
	}

	// 递归生成笛卡尔积
	var generate func(index int, currentCombo []int)
	generate = func(index int, currentCombo []int) {
		if index == len(sets) {
			if len(currentCombo) == requiredSize {
				temp := make([]int, requiredSize)
				copy(temp, currentCombo)
				result = append(result, temp)
			}
			return
		}

		for _, num := range sets[index] {
			// 确保组合中没有重复数字
			isDuplicate := false
			for _, cNum := range currentCombo {
				if cNum == num {
					isDuplicate = true
					break
				}
			}

			if !isDuplicate {
				generate(index+1, append(currentCombo, num))
			}
		}
	}

	generate(0, []int{})
	return result, nil
}

// parseSingleBet 解析单笔下注（新优化版本）
func (p *IntelligentBetParser) parseSingleBet(betID string, segment string, context *BetContext) SingleBetParsing {
	result := SingleBetParsing{
		BetID:        betID,
		OriginalText: segment,
		LotteryBets:  make(map[string]LotteryBetInfo),
		ErrorMessage: make([]string, 0),
	}

	// 1. 识别体彩类型,并移除相关字符串
	lotteries := p.identifyLotteries(segment)
	if len(lotteries) == 0 && len(context.inheritedLotteries) > 0 {
		lotteries = context.inheritedLotteries
	}
	if len(lotteries) == 0 {
		lotteries = []string{"新澳"}
	}

	// 2. 识别下注类型标识
	betTypeFlags := p.identifyBetTypeFlags(segment)

	// 3. 为每个体彩处理
	for _, lottery := range lotteries {
		lotteryInfo := LotteryBetInfo{
			LotteryType:    lottery,
			BetTypeFlags:   betTypeFlags,
			BetTypeDetails: make(map[string]BetTypeDetail),
			TotalAmount:    decimal.NewFromInt(0),
			TotalGroups:    0,
		}

		// 4. 处理每种存在的下注类型
		if betTypeFlags.HasThreeOfThree {
			detail, err := p.processBetType("三中三", segment)
			if err != nil {
				result.HasError = true
				result.ErrorMessage = append(result.ErrorMessage, err.Error())
				return result
			} else {
				lotteryInfo.BetTypeDetails["三中三"] = *detail
				lotteryInfo.TotalAmount = lotteryInfo.TotalAmount.Add(detail.TotalAmount)
				lotteryInfo.TotalGroups += detail.TotalGroups
			}
		}

		if betTypeFlags.HasTwoOfTwo {
			detail, err := p.processBetType("二中二", segment)
			if err != nil {
				result.HasError = true
				result.ErrorMessage = append(result.ErrorMessage, err.Error())
				return result
			} else {
				lotteryInfo.BetTypeDetails["二中二"] = *detail
				lotteryInfo.TotalAmount = lotteryInfo.TotalAmount.Add(detail.TotalAmount)
				lotteryInfo.TotalGroups += detail.TotalGroups
			}
		}

		if betTypeFlags.HasThreeOfTwo {
			detail, err := p.processBetType("三中二", segment)
			if err != nil {
				result.HasError = true
				result.ErrorMessage = append(result.ErrorMessage, err.Error())
				return result
			} else {
				lotteryInfo.BetTypeDetails["三中二"] = *detail
				lotteryInfo.TotalAmount = lotteryInfo.TotalAmount.Add(detail.TotalAmount)
				lotteryInfo.TotalGroups += detail.TotalGroups
			}
		}

		if betTypeFlags.HasSpecial {
			detail, err := p.processBetType("特碰", segment)
			if err != nil {
				result.HasError = true
				result.ErrorMessage = append(result.ErrorMessage, err.Error())
				return result
			} else {
				lotteryInfo.BetTypeDetails["特碰"] = *detail
				lotteryInfo.TotalAmount = lotteryInfo.TotalAmount.Add(detail.TotalAmount)
				lotteryInfo.TotalGroups += detail.TotalGroups
			}
		}

		// 检查是否有下注类型但没有具体号码
		hasAnyBetType := betTypeFlags.HasThreeOfThree || betTypeFlags.HasTwoOfTwo ||
			betTypeFlags.HasThreeOfTwo || betTypeFlags.HasSpecial

		if !hasAnyBetType {
			result.HasError = true
			result.ErrorMessage = append(result.ErrorMessage, "没有识别到任何的下注类型,请检查是否有下注包含：三中三、二中二、三中二、特碰，一种或多种下注类型")
			return result
		}

		result.LotteryBets[lottery] = lotteryInfo
	}

	// 生成统计信息
	result.BetStatistics = p.generateNewBetStatistics(result.LotteryBets)

	return result
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

// extractAmountSmart 智能提取金额
func (p *IntelligentBetParser) extractAmountSmart(text string) decimal.Decimal {
	// 金额模式（按优先级排序）
	patterns := []string{
		`各(\d+)`,  // 各20
		`每组(\d+)`, // 每组20
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			if amount, err := strconv.ParseFloat(matches[1], 64); err == nil {
				return decimal.NewFromFloat(amount)
			}
		}
	}
	return decimal.NewFromInt(0)
}

// isComplexBet 检查是否为复式下注
func (p *IntelligentBetParser) isComplexBet(betType string, text string) bool {
	// 规则1: 文本中包含"复式"关键词
	// if strings.Contains(text, "复式") {
	// 	return true
	// }

	switch betType {
	case "三中三", "三中二":
		// 三中三或三中二: 连续4个或以上数字
		return p.hasConsecutiveNumbers(text, 4)
	case "二中二", "特碰":
		// 二中二或特碰: 连续3个或以上数字
		return p.hasConsecutiveNumbers(text, 3)
	}
	return false
}

// hasConsecutiveNumbers 检查文本中是否存在连续的n个数字,需要排除"拖"字
func (p *IntelligentBetParser) hasConsecutiveNumbers(text string, n int) bool {
	if n <= 0 {
		return false
	}

	// 匹配所有数字组合
	re := regexp.MustCompile(`\d{1,2}(?:-\d{1,2})*`)
	matchIndices := re.FindAllStringIndex(text, -1)

	for _, indices := range matchIndices {
		start, end := indices[0], indices[1]
		match := text[start:end]

		// 检查前后字符
		shouldSkip := false

		// 检查前一个字符
		if start > 0 {
			beforeText := text[:start]
			if len(beforeText) > 0 {
				runes := []rune(beforeText)
				if len(runes) > 0 {
					prevRune := runes[len(runes)-1]
					if prevRune == '拖' || unicode.IsDigit(prevRune) {
						shouldSkip = true
					}
				}
			}
		}

		// 检查后一个字符
		if end < len(text) && !shouldSkip {
			afterText := text[end:]
			if len(afterText) > 0 {
				runes := []rune(afterText)
				if len(runes) > 0 {
					nextRune := runes[0]
					if nextRune == '拖' || unicode.IsDigit(nextRune) {
						shouldSkip = true
					}
				}
			}
		}

		if shouldSkip {
			continue
		}

		// 验证这个匹配确实包含至少n个数字
		numStrings := strings.Split(match, "-")
		if len(numStrings) >= n {
			return true
		}
	}

	return false
}

// isDragBet 检查是否拖码下注
func (p *IntelligentBetParser) isDragBet(text string) bool {
	if strings.Contains(text, "拖") {
		return true
	}
	return false
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
