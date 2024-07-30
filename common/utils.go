package common

import (
	"math"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// IsContain 判断是否包含
func IsContain(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

// ConvertMetricValue 转化单位
func ConvertMetricValue(v string, u string) float64 {
	switch u {
	case "KB", "kiloBytes":
		value, _ := strconv.ParseInt(v, 10, 64)
		return UnitConversion(float64(value) / float64(1024*1024))
	case "MB":
		value, _ := strconv.ParseInt(v, 10, 64)
		return UnitConversion(float64(value) / float64(1024))
	case "num":
		value, _ := strconv.ParseInt(v, 10, 64)
		return float64(value) * 3
		// MB\%\KBps
	default:
		value, _ := strconv.ParseFloat(v, 64)
		return value
	}
}

func UnitConversion(num float64) float64 {
	var nums float64
	if math.IsNaN(num) || math.IsInf(num, 0) {
		nums = 0
	} else {
		nums, _ = decimal.NewFromFloat(num).Round(2).Float64()
	}
	return nums
}

// TransformMetricAlias 转化 Metric 名称
func TransformMetricAlias(alias string) string {
	if alias == "" {
		return alias
	}
	return strings.Replace(alias, ".", "_", -1)
}
