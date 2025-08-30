package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	ConfigDirName  = "config"
	ConfigFileName = "system_config.json"
)

// 全局配置文件访问锁
var configMutex sync.RWMutex

// getConfigFilePath 获取配置文件路径
func getConfigFilePath() (string, error) {
	// 获取可执行文件目录
	exeDir, err := getExecutableDir()
	if err != nil {
		return "", fmt.Errorf("获取可执行文件目录失败: %v", err)
	}

	// 构建配置目录路径
	configDir := filepath.Join(exeDir, ConfigDirName)

	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 返回配置文件完整路径
	return filepath.Join(configDir, ConfigFileName), nil
}

// getExecutableDir 获取可执行文件所在目录
func getExecutableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

// loadSystemConfigFromFile 从文件加载系统配置（只在初始化时调用）
func loadSystemConfigFromFile() (*SystemConfig, error) {
	configMutex.RLock()
	defer configMutex.RUnlock()
	configPath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	// 如果配置文件不存在，返回默认配置并创建文件
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		safeLogger.AppendLog("配置文件不存在，使用默认配置并创建文件: " + configPath)
		defaultConfig := getDefaultSystemConfig()
		// 直接写入文件，不使用saveSystemConfigToFile避免递归锁
		data, marshalErr := json.MarshalIndent(defaultConfig, "", "  ")
		if marshalErr == nil {
			os.WriteFile(configPath, data, 0644)
		}
		return defaultConfig, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析JSON
	var config SystemConfig
	if err := json.Unmarshal(data, &config); err != nil {
		safeLogger.AppendLog("配置文件格式错误，使用默认配置: " + err.Error())
		// 配置文件损坏时，备份原文件并使用默认配置
		backupPath := configPath + ".backup"
		os.Rename(configPath, backupPath)
		defaultConfig := getDefaultSystemConfig()
		// 直接写入文件，避免递归锁
		data, marshalErr := json.MarshalIndent(defaultConfig, "", "  ")
		if marshalErr == nil {
			os.WriteFile(configPath, data, 0644)
		}
		return defaultConfig, nil
	}

	safeLogger.AppendLog("成功从文件加载系统配置: " + configPath)
	return &config, nil
}

// saveSystemConfigToFile 保存系统配置到文件（线程安全）
func saveSystemConfigToFile(config *SystemConfig) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	safeLogger.AppendLog("成功保存系统配置到文件: " + configPath)
	return nil
}

// getDefaultSystemConfig 获取默认系统配置
func getDefaultSystemConfig() *SystemConfig {
	return &SystemConfig{
		ZodiacConfig: ZodiacConfig{
			Rat:     []int{1, 13, 25, 37, 49},
			Ox:      []int{2, 14, 26, 38},
			Tiger:   []int{3, 15, 27, 39},
			Rabbit:  []int{4, 16, 28, 40},
			Dragon:  []int{5, 17, 29, 41},
			Snake:   []int{6, 18, 30, 42},
			Horse:   []int{7, 19, 31, 43},
			Goat:    []int{8, 20, 32, 44},
			Monkey:  []int{9, 21, 33, 45},
			Rooster: []int{10, 22, 34, 46},
			Dog:     []int{11, 23, 35, 47},
			Pig:     []int{12, 24, 36, 48},
		},
		ColorConfig: ColorConfig{
			Red:   []int{1, 2, 7, 8, 12, 13, 18, 19, 23, 24, 29, 30, 34, 35, 40, 45, 46},
			Green: []int{5, 6, 11, 16, 17, 21, 22, 27, 28, 32, 33, 38, 39, 43, 44, 49},
			Blue:  []int{3, 4, 9, 10, 14, 15, 20, 25, 26, 31, 36, 37, 41, 42, 47, 48},
		},
		TailConfig: TailConfig{
			Tail0: []int{10, 20, 30, 40},
			Tail1: []int{1, 11, 21, 31, 41},
			Tail2: []int{2, 12, 22, 32, 42},
			Tail3: []int{3, 13, 23, 33, 43},
			Tail4: []int{4, 14, 24, 34, 44},
			Tail5: []int{5, 15, 25, 35, 45},
			Tail6: []int{6, 16, 26, 36, 46},
			Tail7: []int{7, 17, 27, 37, 47},
			Tail8: []int{8, 18, 28, 38, 48},
			Tail9: []int{9, 19, 29, 39, 49},
		},
		BetTypeAliases: BetTypeAliases{
			ThreeOfThree: []string{"死", "三中三", "三全中", "3中3"},
			ThreeOfTwo:   []string{"活", "三中二", "三种二", "3中2"},
			TwoOfTwo:     []string{"二全中", "二中二", "2中2"},
			Special:      []string{"特碰"},
		},
		KeywordAliases: KeywordAliases{
			NewMacau: []string{"新", "新澳", "新澳门"},
			OldMacau: []string{"老", "老澳", "老澳门", "旧"},
			HongKong: []string{"香", "香港", "港"},
			Complex:  []string{"复式", "复试", "组合"},
			Each:     []string{"各", "每个", "分别", "都"},
			PerGroup: []string{"每组", "一组"},
		},
		OddsConfig: OddsConfig{
			ThreeOfThree: ThreeOfThreeOdds{
				OddsRatio: 175.0, // 三中三默认赔率 1:175
				Rebate:    0.05,  // 默认回水 5%
			},
			ThreeOfTwo: ThreeOfTwoOdds{
				HitTwoOdds: HitTwoOdds{
					OddsRatio: 7.5,  // 三中二中二个赔率 1:7.5
					Rebate:    0.05, // 默认回水 5%
				},
				HitThreeOdds: HitThreeOdds{
					OddsRatio: 175.0, // 三中二中三个赔率 1:175
					Rebate:    0.05,  // 默认回水 5%
				},
			},
			TwoOfTwo: TwoOfTwoOdds{
				OddsRatio: 7.5,  // 二中二默认赔率 1:7.5
				Rebate:    0.05, // 默认回水 5%
			},
			Special: SpecialOdds{
				OddsRatio: 40.0, // 特碰默认赔率 1:40
				Rebate:    0.05, // 默认回水 5%
			},
		},
	}
}

// InitializeConfig 初始化配置系统（只在启动时调用一次）
func InitializeConfig() error {
	configMutex.Lock()
	defer configMutex.Unlock()

	// 确保配置目录存在
	configPath, err := getConfigFilePath()
	if err != nil {
		safeLogger.AppendLog("初始化配置目录失败: " + err.Error())
		return err
	}

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		safeLogger.AppendLog("配置文件不存在，创建默认配置: " + configPath)
		defaultConfig := getDefaultSystemConfig()
		// 直接写入文件，避免递归调用造成死锁
		data, marshalErr := json.MarshalIndent(defaultConfig, "", "  ")
		if marshalErr != nil {
			safeLogger.AppendLog("序列化默认配置失败: " + marshalErr.Error())
			return marshalErr
		}
		if writeErr := os.WriteFile(configPath, data, 0644); writeErr != nil {
			safeLogger.AppendLog("创建默认配置文件失败: " + writeErr.Error())
			return writeErr
		}
	}

	safeLogger.AppendLog("配置系统初始化成功")
	return nil
}
