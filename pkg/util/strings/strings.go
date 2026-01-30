// Copyright 2024 许铭杰 (1044011439@qq.com). All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package strings

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/asaskevich/govalidator"
)

type frequencyInfo struct {
	s         string
	frequency int
}

type frequencyInfoSlice []frequencyInfo

func (fi frequencyInfoSlice) Len() int {
	return len(fi)
}

func (fi frequencyInfoSlice) Swap(i, j int) {
	fi[i], fi[j] = fi[j], fi[i]
}

func (fi frequencyInfoSlice) Less(i, j int) bool {
	return fi[i].frequency < fi[j].frequency
}

// Diff 剔除 base 中 包含的 exclude 的字符串.
func Diff(base, exclude []string) (result []string) {
	excludeMap := make(map[string]bool)
	for _, s := range exclude {
		excludeMap[s] = true
	}

	for _, s := range base {
		if !excludeMap[s] {
			result = append(result, s)
		}
	}

	return result
}

// Include 得到 include 中 包含的 base 的字符串.
func Include(base, include []string) (result []string) {
	baseMap := make(map[string]bool)
	for _, s := range base {
		baseMap[s] = true
	}

	for _, s := range include {
		if baseMap[s] {
			result = append(result, s)
		}
	}

	return result
}

// Unique 去除重复字符串.
func Unique(ss []string) (result []string) {
	sMap := make(map[string]bool)
	for _, s := range ss {
		sMap[s] = true
	}

	for s := range sMap {
		result = append(result, s)
	}

	return result
}

// CamelCaseToUnderscore 将大/小驼峰命名法装为 _ 命名法.
func CamelCaseToUnderscore(str string) string {
	return govalidator.CamelCaseToUnderscore(str)
}

// UnderscoreToCamelCase 将 _ 命名法转为大/小驼峰命名法.
func UnderscoreToCamelCase(str string) string {
	return govalidator.UnderscoreToCamelCase(str)
}

// FindString 找到 array 中 str 的索引下标.
func FindString(array []string, str string) int {
	for index, s := range array {
		if s == str {
			return index
		}
	}

	return -1
}

// StringIn 检查 str 是否在 array 中存在.
func StringIn(str string, array []string) bool {
	return FindString(array, str) > -1
}

// Reverse 反转字符串.
func Reverse(s string) string {
	size := len(s)
	buf := make([]byte, size)
	for start := 0; start < size; {
		r, n := utf8.DecodeRuneInString(s[start:])
		start += n
		utf8.EncodeRune(buf[size-start:], r)
	}

	return string(buf)
}

// Filter 从 list 中 过滤掉 strToFilter.
func Filter(list []string, strToFilter string) (newList []string) {
	for _, s := range list {
		if s != strToFilter {
			newList = append(newList, s)
		}
	}
	return
}

// Add 向 list 中添加 str.
func Add(list []string, str string) []string {
	for _, s := range list {
		if s == str {
			return list
		}
	}
	list = append(list, str)
	return list
}

// Contains list 中 是否包含 strToSearch.
func Contains(list []string, strToSearch string) bool {
	for _, item := range list {
		if item == strToSearch {
			return true
		}
	}

	return false
}

// FrequencySort 按照 list 中字符串出现的频率进行排序.
func FrequencySort(list []string) []string {
	cnt := map[string]int{}

	for _, s := range list {
		cnt[s]++
	}

	infos := make([]frequencyInfo, 0, len(cnt))
	for s, c := range cnt {
		infos = append(infos, frequencyInfo{
			s:         s,
			frequency: c,
		})
	}

	sort.Sort(frequencyInfoSlice(infos))

	ret := make([]string, 0, len(infos))
	for _, info := range infos {
		ret = append(ret, info.s)
	}

	return ret
}

// ContainseQualFold 如果给定的 slice 包含 unicode 大小写折叠下的字符串 s ，则返回 true.
func ContainsEqualFold(slice []string, s string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, s) {
			return true
		}
	}

	return false
}
