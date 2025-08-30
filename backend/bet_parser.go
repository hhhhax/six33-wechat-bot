package backend

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// BetParser 智能下注解析器
type BetParser struct {
	app *App // 引用主应用，用于获取配置
}

// NewBetParser 创建解析器实例
func NewBetParser(app *App) *BetParser {
	return &BetParser{
		app: app,
	}
}

// 获取配置数据的方法
func (p *BetParser) getZodiacMap() map[string][]int {
	config := p.app.GetZodiacConfig()
	return map[string][]int{
		"鼠": config.Rat, "牛": config.Ox, "虎": config.Tiger, "兔": config.Rabbit,
		"龙": config.Dragon, "蛇": config.Snake, "马": config.Horse, "羊": config.Goat,
		"猴": config.Monkey, "鸡": config.Rooster, "狗": config.Dog, "猪": config.Pig,
	}
}

func (p *BetParser) getBetTypeAlias() map[string][]string {
	config := p.app.GetBetTypeAliases()
	return map[string][]string{
		"三中三": config.ThreeOfThree,
		"三中二": config.ThreeOfTwo,
		"二中二": config.TwoOfTwo,
		"特碰":  config.Special,
	}
}

func (p *BetParser) getLotteryAlias() map[string][]string {
	// 这个可以保持固定，或者也从配置读取
	return map[string][]string{
		"新澳": {"新", "新澳", "新澳门", "new_macau"},
		"老澳": {"老", "老澳", "老澳门", "old_macau"},
		"香港": {"港", "香港", "hk", "hongkong"},
	}
}

// ParseBetString 解析下注字符串
func (p *BetParser) ParseBetString(request BetParseRequest) BetParseResponse {
	startTime := time.Now()

	if request.Input == "" {
		return BetParseResponse{
			Success: false,
			Error:   "输入为空",
			Results: []ParsedBet{},
		}
	}

	results := []ParsedBet{}
	lines := strings.Split(strings.TrimSpace(request.Input), "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析单行
		parsed := p.parseLine(line, lines, i)
		results = append(results, parsed...)
	}

	// 生成统计信息
	totalAmount := 0.0
	totalGroups := 0
	for _, bet := range results {
		if !bet.IsError {
			totalAmount += bet.TotalAmount
			totalGroups += bet.Groups
		}
	}

	summary := p.generateSummary(results)
	parseTime := time.Since(startTime)

	return BetParseResponse{
		Success:     len(results) > 0,
		Results:     results,
		TotalBets:   len(results),
		TotalAmount: totalAmount,
		TotalGroups: totalGroups,
		Summary:     summary,
		ParseTime:   parseTime.String(),
	}
}

// parseLine 解析单行
func (p *BetParser) parseLine(line string, allLines []string, lineIndex int) []ParsedBet {
	// 清理输入
	cleanLine := p.cleanLine(line)

	// 尝试不同的解析策略
	strategies := []func(string, []string, int) []ParsedBet{
		p.parseNumberGroupFormat,     // 30，34，45一组30
		p.parseComplexFormat,         // 三中三三中二复式10-20-30-40各25
		p.parseSimpleFormat,          // 16-18-23=20
		p.parseZodiacFormat,          // 龙兔复试三中三，三中二各15
		p.parseDragFormat,            // 三中三21.35拖全场各20
		p.parseLotterySpecificFormat, // 新.三中三3二中二各3
		p.parseBasicFormat,           // 01,02,03/200
		p.parseListFormat,            // 5–13–32，7–23–26、8–22–23=各10元
	}

	for _, strategy := range strategies {
		if results := strategy(cleanLine, allLines, lineIndex); len(results) > 0 {
			return results
		}
	}

	// 如果都无法解析，返回错误
	return []ParsedBet{{
		Type:        "未识别",
		Numbers:     p.extractNumbers(line),
		Amount:      0,
		TotalAmount: 0,
		Groups:      0,
		Description: fmt.Sprintf("未识别格式: %s", line),
		Original:    line,
		IsError:     true,
	}}
}

// cleanLine 清理输入行
func (p *BetParser) cleanLine(line string) string {
	// 统一标点符号
	replacements := map[string]string{
		"，": ",", "。": ".", "–": "-", "—": "-", "=": "=",
		"（": "(", "）": ")", "【": "[", "】": "]",
	}

	for old, new := range replacements {
		line = strings.ReplaceAll(line, old, new)
	}

	// 规范化空格
	line = regexp.MustCompile(`\s+`).ReplaceAllString(line, " ")
	return strings.TrimSpace(line)
}

// parseNumberGroupFormat 格式1: 数字组合格式 (30，34，45一组30)
func (p *BetParser) parseNumberGroupFormat(line string, allLines []string, lineIndex int) []ParsedBet {
	patterns := []string{
		`(\d+[,.\-]+\d+[,.\-]+\d+)[^0-9]*一组[^0-9]*(\d+)`,
		`(\d+[,.\-]+\d+[,.\-]+\d+)[^0-9]*(\d+)$`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(line); len(matches) >= 3 {
			numbers := p.extractNumbers(matches[1])
			amount, err := strconv.ParseFloat(matches[2], 64)

			if len(numbers) == 3 && err == nil && amount > 0 {
				return []ParsedBet{{
					Type:        "三中三",
					Numbers:     numbers,
					Amount:      amount,
					TotalAmount: amount,
					Groups:      1,
					Description: fmt.Sprintf("%v 三中三 %.0f元", numbers, amount),
					Original:    line,
				}}
			}
		}
	}
	return []ParsedBet{}
}

// parseComplexFormat 格式2: 复式格式 (三中三三中二复式10-20-30-40各25)
func (p *BetParser) parseComplexFormat(line string, allLines []string, lineIndex int) []ParsedBet {
	re := regexp.MustCompile(`(.+)复式([0-9\-,.\s]+)各\s*(\d+)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) >= 4 {
		betTypesStr := matches[1]
		numbers := p.extractNumbers(matches[2])
		amountEach, err := strconv.ParseFloat(matches[3], 64)

		if len(numbers) >= 3 && err == nil && amountEach > 0 {
			betTypes := p.extractBetTypes(betTypesStr)
			results := []ParsedBet{}

			for _, betType := range betTypes {
				requiredNums := p.getRequiredNumbers(betType)
				groups := p.calculateCombinations(len(numbers), requiredNums)
				totalAmount := float64(groups) * amountEach

				results = append(results, ParsedBet{
					Type:        betType,
					Numbers:     numbers,
					Amount:      amountEach,
					TotalAmount: totalAmount,
					Groups:      groups,
					Description: fmt.Sprintf("%v %s %d组×%.0f元=%.0f元", numbers, betType, groups, amountEach, totalAmount),
					Original:    line,
				})
			}
			return results
		}
	}
	return []ParsedBet{}
}

// parseSimpleFormat 格式3: 简单等号格式 (16-18-23=20)
func (p *BetParser) parseSimpleFormat(line string, allLines []string, lineIndex int) []ParsedBet {
	patterns := []string{
		`(\d+[\-,.]+\d+[\-,.]+\d+)=(\d+)`,
		`(\d+[\-,.]+\d+)=(\d+)`, // 二中二格式
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(line); len(matches) >= 3 {
			numbers := p.extractNumbers(matches[1])
			amount, err := strconv.ParseFloat(matches[2], 64)

			if len(numbers) >= 2 && err == nil && amount > 0 {
				betType := "三中三"
				if len(numbers) == 2 {
					betType = "二中二"
				}

				return []ParsedBet{{
					Type:        betType,
					Numbers:     numbers,
					Amount:      amount,
					TotalAmount: amount,
					Groups:      1,
					Description: fmt.Sprintf("%v %s %.0f元", numbers, betType, amount),
					Original:    line,
				}}
			}
		}
	}
	return []ParsedBet{}
}

// parseZodiacFormat 格式4: 生肖格式 (龙兔复试三中三，三中二各15)
func (p *BetParser) parseZodiacFormat(line string, allLines []string, lineIndex int) []ParsedBet {
	re := regexp.MustCompile(`([鼠牛虎兔龙蛇马羊猴鸡狗猪]+)复试?(.+)各\s*(\d+)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) >= 4 {
		zodiacNames := strings.Split(matches[1], "")
		betTypesStr := matches[2]
		amountEach, err := strconv.ParseFloat(matches[3], 64)

		// 获取生肖对应数字
		allNumbers := []int{}
		zodiacMap := p.getZodiacMap()
		for _, zodiac := range zodiacNames {
			if numbers, exists := zodiacMap[zodiac]; exists {
				allNumbers = append(allNumbers, numbers...)
			}
		}

		if len(allNumbers) > 0 && err == nil && amountEach > 0 {
			betTypes := p.extractBetTypes(betTypesStr)
			results := []ParsedBet{}

			for _, betType := range betTypes {
				requiredNums := p.getRequiredNumbers(betType)
				groups := p.calculateCombinations(len(allNumbers), requiredNums)
				totalAmount := float64(groups) * amountEach

				results = append(results, ParsedBet{
					Type:        betType,
					Numbers:     allNumbers,
					Amount:      amountEach,
					TotalAmount: totalAmount,
					Groups:      groups,
					Description: fmt.Sprintf("%s生肖 %s %d组×%.0f元=%.0f元", strings.Join(zodiacNames, ""), betType, groups, amountEach, totalAmount),
					Original:    line,
				})
			}
			return results
		}
	}
	return []ParsedBet{}
}

// parseDragFormat 格式5: 拖码格式 (三中三21.35拖全场各20)
func (p *BetParser) parseDragFormat(line string, allLines []string, lineIndex int) []ParsedBet {
	re := regexp.MustCompile(`三中三(\d+[,.]\d+)拖全场各(\d+)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) >= 3 {
		baseNumbers := p.extractNumbers(matches[1])
		amountEach, err := strconv.ParseFloat(matches[2], 64)

		if len(baseNumbers) >= 2 && err == nil && amountEach > 0 {
			// 拖全场计算：与1-49中其他数字组合
			groups := 49 - len(baseNumbers) // 简化计算
			totalAmount := float64(groups) * amountEach

			return []ParsedBet{{
				Type:        "三中三",
				Numbers:     baseNumbers,
				Amount:      amountEach,
				TotalAmount: totalAmount,
				Groups:      groups,
				Description: fmt.Sprintf("%v拖全场 三中三 %d组×%.0f元=%.0f元", baseNumbers, groups, amountEach, totalAmount),
				Original:    line,
			}}
		}
	}
	return []ParsedBet{}
}

// parseLotterySpecificFormat 格式6: 彩种指定格式 (新.三中三3二中二各3)
func (p *BetParser) parseLotterySpecificFormat(line string, allLines []string, lineIndex int) []ParsedBet {
	re := regexp.MustCompile(`([新老港澳]+)\.(.+)各\s*(\d+)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) >= 4 {
		lottery := p.normalizeLottery(matches[1])
		betTypes := p.extractBetTypes(matches[2])
		amountEach, err := strconv.ParseFloat(matches[3], 64)

		// 查找下一行的数字
		numbers := []int{}
		if lineIndex+1 < len(allLines) {
			numbers = p.extractNumbers(allLines[lineIndex+1])
		}

		if lottery != "" && len(betTypes) > 0 && len(numbers) >= 2 && err == nil && amountEach > 0 {
			results := []ParsedBet{}

			for _, betType := range betTypes {
				requiredNums := p.getRequiredNumbers(betType)
				groups := p.calculateCombinations(len(numbers), requiredNums)
				totalAmount := float64(groups) * amountEach

				results = append(results, ParsedBet{
					Type:        betType,
					Lottery:     lottery,
					Numbers:     numbers,
					Amount:      amountEach,
					TotalAmount: totalAmount,
					Groups:      groups,
					Description: fmt.Sprintf("%s %v %s %d组×%.0f元=%.0f元", lottery, numbers, betType, groups, amountEach, totalAmount),
					Original:    line,
				})
			}
			return results
		}
	}
	return []ParsedBet{}
}

// parseBasicFormat 格式7: 基础格式 (01,02,03/200)
func (p *BetParser) parseBasicFormat(line string, allLines []string, lineIndex int) []ParsedBet {
	re := regexp.MustCompile(`([0-9,.\-\s]+)/(\d+)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) >= 3 {
		numbers := p.extractNumbers(matches[1])
		amount, err := strconv.ParseFloat(matches[2], 64)

		if len(numbers) >= 2 && err == nil && amount > 0 {
			betType := "三中三"
			if len(numbers) == 2 {
				betType = "二中二"
			}

			return []ParsedBet{{
				Type:        betType,
				Numbers:     numbers,
				Amount:      amount,
				TotalAmount: amount,
				Groups:      1,
				Description: fmt.Sprintf("%v %s %.0f元", numbers, betType, amount),
				Original:    line,
			}}
		}
	}
	return []ParsedBet{}
}

// parseListFormat 格式8: 列表格式 (5–13–32，7–23–26、8–22–23=各10元)
func (p *BetParser) parseListFormat(line string, allLines []string, lineIndex int) []ParsedBet {
	re := regexp.MustCompile(`(.+)=?各(\d+)元?`)
	matches := re.FindStringSubmatch(line)

	if len(matches) >= 3 {
		groupsStr := matches[1]
		amountEach, err := strconv.ParseFloat(matches[2], 64)

		if err == nil && amountEach > 0 {
			// 分割多个组合
			groups := regexp.MustCompile(`[,，、]`).Split(groupsStr, -1)
			results := []ParsedBet{}

			for _, group := range groups {
				numbers := p.extractNumbers(group)
				if len(numbers) >= 2 {
					betType := "三中三"
					if len(numbers) == 2 {
						betType = "二中二"
					}

					results = append(results, ParsedBet{
						Type:        betType,
						Numbers:     numbers,
						Amount:      amountEach,
						TotalAmount: amountEach,
						Groups:      1,
						Description: fmt.Sprintf("%v %s %.0f元", numbers, betType, amountEach),
						Original:    line,
					})
				}
			}
			return results
		}
	}
	return []ParsedBet{}
}

// 辅助方法
func (p *BetParser) extractNumbers(str string) []int {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(str, -1)

	numbers := []int{}
	for _, match := range matches {
		if num, err := strconv.Atoi(match); err == nil && num >= 1 && num <= 49 {
			numbers = append(numbers, num)
		}
	}
	return numbers
}

func (p *BetParser) extractBetTypes(str string) []string {
	types := []string{}
	betTypeAlias := p.getBetTypeAlias()
	for standard, aliases := range betTypeAlias {
		for _, alias := range aliases {
			if strings.Contains(str, alias) {
				types = append(types, standard)
				break
			}
		}
	}

	// 去重
	uniqueTypes := []string{}
	seen := make(map[string]bool)
	for _, t := range types {
		if !seen[t] {
			uniqueTypes = append(uniqueTypes, t)
			seen[t] = true
		}
	}

	return uniqueTypes
}

func (p *BetParser) normalizeLottery(str string) string {
	lotteryAlias := p.getLotteryAlias()
	for standard, aliases := range lotteryAlias {
		for _, alias := range aliases {
			if strings.Contains(str, alias) {
				return standard
			}
		}
	}
	return ""
}

func (p *BetParser) getRequiredNumbers(betType string) int {
	switch betType {
	case "二中二":
		return 2
	case "三中三", "三中二":
		return 3
	case "特碰":
		return 2
	default:
		return 3
	}
}

func (p *BetParser) calculateCombinations(n, r int) int {
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

func (p *BetParser) generateSummary(results []ParsedBet) []BetTypeSummary {
	summary := make(map[string]*BetTypeSummary)

	for _, bet := range results {
		if bet.IsError {
			continue
		}

		key := bet.Lottery + "_" + bet.Type
		if _, exists := summary[key]; !exists {
			summary[key] = &BetTypeSummary{
				Type:    bet.Type,
				Lottery: bet.Lottery,
			}
		}

		summary[key].TotalGroups += bet.Groups
		summary[key].TotalAmount += bet.TotalAmount
	}

	result := []BetTypeSummary{}
	for _, s := range summary {
		result = append(result, *s)
	}

	return result
}
