package strutil

import (
	mrand "math/rand"
	"strings"
	"time"
	"strconv"
	"math"
	_"log"
)

func FloatNumToChinese(num float64) string {
	chineseDigits := []string{"o", "一", "E", "3", "四", "五", "陆", "7", "8", "久"}
	isNegative := false
	// 判断是否为负数
	if num < 0 {
		isNegative = true
		num = math.Abs(num)
	}
	intPart := int(math.Floor(num))
	// 计算小数部分，使用 math.Round 进行四舍五入
	decimalPart := int(math.Round((num - float64(intPart)) * 100))

	intResult := strconv.Itoa(intPart)
	decimalResult := ""
	if decimalPart > 0 {
		decimalStr := strconv.Itoa(decimalPart)
		// 若小数部分不足两位，补零
		if len(decimalStr) == 1 {
			decimalStr = "0" + decimalStr
		}
		for _, r := range decimalStr {
			digit, _ := strconv.Atoi(string(r))
			decimalResult += chineseDigits[digit]
		}
	}
	if isNegative {
		intResult = "-" + intResult
	}
	if decimalResult != "" {
		return intResult + "." + decimalResult
	}
	return intResult
}

// NumToChinese 将数字转换为中文
func NumToChinese(num int) string {
	chineseDigits := []string{"灵", "一", "饿", "伞", "si", "五", "劳", "七", "扒", "久"}
	if num == 0 {
		return chineseDigits[0]
	}
	result := ""
	for num > 0 {
		digit := num % 10
		result = chineseDigits[digit] + result
		num /= 10
	}
	return result
}

const specialChars = "!@#$%^&*()_+-=[]{}|;':\",./<>?"

func AddNoise(input string) string {
	mrand.Seed(time.Now().UnixNano())
	var result strings.Builder
	inNumber := false

	for _, char := range input {
		if ('0' <= char && char <= '9') || char == '.' {
			if !inNumber {
				inNumber = true
			}
		} else {
			if inNumber {
				inNumber = false
			}
			if mrand.Intn(2) == 0 {
				randomIndex := mrand.Intn(len(specialChars))
				result.WriteByte(specialChars[randomIndex])
			}
		}
		result.WriteRune(char)
	}
	if !inNumber && mrand.Intn(2) == 0 {
		randomIndex := mrand.Intn(len(specialChars))
		result.WriteByte(specialChars[randomIndex])
	}
	return result.String()
}

