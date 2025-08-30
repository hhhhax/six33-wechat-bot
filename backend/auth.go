package backend

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// 检查是否处于调试状态
func isBeingDebugged() bool {
	// 使用IsDebuggerPresent检测调试器
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	isDebuggerPresent := kernel32.NewProc("IsDebuggerPresent")
	r1, _, _ := isDebuggerPresent.Call()

	// 如果r1不为0，表示正在被调试
	if r1 != 0 {
		return true
	}

	// 尝试其他检测方法，这里使用简单的时间差异检测
	start := time.Now()
	time.Sleep(1 * time.Millisecond)
	elapsed := time.Since(start)

	// 如果时间差异明显大于预期，可能处于调试状态
	// 调试时单步执行会导致时间差异增大
	if elapsed > 10*time.Millisecond {
		return true
	}

	return false
}

// getMachineGuid retrieves the MachineGuid for Windows.
func getMachineGuid() (string, error) {
	// 检查是否被调试
	if isBeingDebugged() {
		// 延迟响应或返回随机数据，而不是直接报错
		time.Sleep(time.Duration(100+time.Now().UnixNano()%1000) * time.Millisecond)
		return "", errors.New("system error: unable to retrieve system information")
	}

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE)
	if err != nil {
		return "", errors.New("access denied: unable to read system information")
	}
	defer key.Close()
	guid, _, err := key.GetStringValue("MachineGuid")
	if err != nil {
		return "", errors.New("access denied: unable to read system identifier")
	}
	return guid, nil
}

// 获取网络时间（百度API，国内可用）
func getNetworkTime() (time.Time, error) {
	// 检查是否被调试
	if isBeingDebugged() {
		// 延迟响应或返回随机数据，而不是直接报错
		time.Sleep(time.Duration(100+time.Now().UnixNano()%1000) * time.Millisecond)
		return time.Time{}, errors.New("network error: unable to connect to time service")
	}

	beijingTime, err := getBeijingTimeFromWorldTimeAPI()
	if err != nil {
		client := &http.Client{
			Timeout: 2 * time.Second,
		}

		resp, err := client.Get("http://quan.suning.com/getSysTime.do")
		if err != nil {
			return time.Time{}, errors.New("unable to retrieve time information")
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return time.Time{}, errors.New("unable to process time data")
		}
		var data struct {
			SysTime2 string `json:"sysTime2"`
		}
		err = json.Unmarshal(body, &data)
		if err != nil {
			return time.Time{}, errors.New("unable to parse time data")
		}
		layout := "2006-01-02 15:04:05"
		// 定义北京时间：UTC+8
		beijing := time.FixedZone("CST", 8*3600)

		parsedTime, err := time.ParseInLocation(layout, data.SysTime2, beijing)
		if err != nil {
			return time.Time{}, errors.New("unable to interpret time data")
		}
		return parsedTime, nil
	}
	return beijingTime, nil
}

// WorldTimeAPIResponse 定义了 WorldTimeAPI 返回的 JSON 结构
type WorldTimeAPIResponse struct {
	Datetime string `json:"datetime"` // 例如: "2025-07-18T01:15:41.123456+08:00"
}

// getBeijingTimeFromWorldTimeAPI 通过 WorldTimeAPI 获取北京时间（带2秒超时）
func getBeijingTimeFromWorldTimeAPI() (time.Time, error) {
	// 创建带超时的 HTTP 客户端
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get("http://worldtimeapi.org/api/timezone/Asia/Shanghai")
	if err != nil {
		return time.Time{}, errors.New("unable to connect to time service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, errors.New("time service error")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return time.Time{}, errors.New("unable to read time data")
	}

	var data WorldTimeAPIResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return time.Time{}, errors.New("unable to parse time data")
	}

	// WorldTimeAPI 返回的日期时间字符串通常是 RFC3339 格式，包含了时区信息
	// Go 的 time.Parse(time.RFC3339, ...) 可以直接解析这种带时区偏移的字符串
	parsedTime, err := time.Parse(time.RFC3339, data.Datetime)
	if err != nil {
		// 如果 RFC3339 解析失败，可以尝试一些常见的变种格式，但 RFC3339 最常用
		// 例如，如果日期字符串是 "2006-01-02T15:04:05+08:00"
		layoutFallback := "2006-01-02T15:04:05-07:00" // 包含毫秒的格式可能需要微调
		parsedTime, err = time.Parse(layoutFallback, data.Datetime)
		if err != nil {
			return time.Time{}, errors.New("unable to interpret time data")
		}
	}

	// 返回的时间 `parsedTime` 已经是 Asia/Shanghai (北京时间) 的时间对象了
	return parsedTime, nil
}

// loadPublicKey loads the public key from a string.
func loadPublicKey(pemContent string) (*rsa.PublicKey, error) {
	publicKeyPEM := []byte(pemContent)

	block, _ := pem.Decode(publicKeyPEM)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("invalid key format")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("invalid key data")
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("unsupported key type")
	}
	return rsaPub, nil
}

// 防止栈检查操作
func obfuscatedStackCheck() {
	// 检查调用栈是否被篡改
	stack := string(debug.Stack())
	// 使用栈信息执行某些检查，但不存储结果到未使用的变量
	if len(stack) > 0 && strings.Contains(stack, "debugger") {
		// 如果发现调试器相关的栈信息，引入随机延迟
		time.Sleep(time.Duration(time.Now().UnixNano()%100) * time.Millisecond)
	}
}

// verifyAuthCode verifies the validity of the authorization code.
// publicKey: RSA public key used for signature verification.
// signedAuthCode: The authorization code string containing raw data and digital signature.
func verifyAuthCode(publicKey *rsa.PublicKey, signedAuthCode string) (bool, string, time.Time, error) {
	// 检查是否被调试
	if isBeingDebugged() {
		// 引入随机延迟
		time.Sleep(time.Duration(100+time.Now().UnixNano()%1000) * time.Millisecond)
		return false, "", time.Time{}, errors.New("verification error: security check failed")
	}

	// 栈检查
	obfuscatedStackCheck()

	if publicKey == nil {
		return false, "", time.Time{}, errors.New("authorization error: missing verification key")
	}

	// Authorization code format: base64(data)|base64(signature)
	parts := strings.Split(signedAuthCode, "|")
	if len(parts) != 2 {
		return false, "", time.Time{}, errors.New("authorization error: invalid format")
	}

	encodedData := parts[0]
	encodedSignature := parts[1]

	// Decode raw data
	data, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return false, "", time.Time{}, errors.New("authorization error: invalid data encoding")
	}
	// Decode digital signature
	signature, err := base64.StdEncoding.DecodeString(encodedSignature)
	if err != nil {
		return false, "", time.Time{}, errors.New("authorization error: invalid signature encoding")
	}

	// Verify digital signature
	hashed := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		return false, "", time.Time{}, errors.New("authorization error: signature verification failed")
	}

	// Parse raw data: MachineGuid|ExpireTime
	dataStr := string(data)
	dataParts := strings.Split(dataStr, "|")
	if len(dataParts) != 2 {
		return false, "", time.Time{}, errors.New("authorization error: invalid data format")
	}

	authorizedMachineGuid := dataParts[0]
	expiryTimeString := dataParts[1]

	// Get local MachineGuid
	localMachineGuid, err := getMachineGuid()
	if err != nil {
		return false, "", time.Time{}, errors.New("authorization error: unable to verify system identity")
	}

	// Compare machine GUIDs
	if localMachineGuid != authorizedMachineGuid {
		// 简化错误消息，不暴露具体GUID
		return false, "", time.Time{}, errors.New("authorization error: system identity mismatch")
	}

	// Parse authorization expiry time
	beijing := time.FixedZone("CST", 8*3600)
	parsedExpireTime, err := time.ParseInLocation("2006-01-02 15:04:05", expiryTimeString, beijing)
	if err != nil {
		return false, "", time.Time{}, errors.New("authorization error: invalid expiration data")
	}

	// Get network time
	netTime, err := getNetworkTime()
	if err != nil {
		return false, "", time.Time{}, errors.New("authorization error: unable to verify time")
	}

	// Check if expired
	if netTime.After(parsedExpireTime) {
		return false, "", time.Time{}, errors.New("authorization error: expired")
	}

	// 对验证过程增加随机延迟，防止时间分析攻击
	time.Sleep(time.Duration(50+time.Now().UnixNano()%100) * time.Millisecond)

	return true, authorizedMachineGuid, parsedExpireTime, nil
}
