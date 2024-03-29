package utils

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

func SnakeToCamel(name string) string {
	pers := strings.Split(name, "_")
	res := ""
	for _, p := range pers {
		res += strings.ToUpper(p[0:1]) + p[1:]
	}
	return res
}

func CamelToSnake(name string) string {
	var words []string
	lastCaped := false

	buf := bytes.Buffer{}
	firstWord := true
	for _, r := range name {
		if r >= 65 && r <= 90 { //caped letter
			//two continuous capital letters will be regarded as one word
			if lastCaped {
				buf.WriteRune(r + 32)
				words = append(words, buf.String())
				buf.Reset()
				lastCaped = false
			} else {
				if !firstWord && buf.Len() > 0 {
					fmt.Println(string(r + 32))
					words = append(words, buf.String())
					buf.Reset()
				}
				buf.WriteRune(r + 32)
				lastCaped = true
			}
			firstWord = false
		} else {
			buf.WriteRune(r)
			lastCaped = false
		}
	}
	words = append(words, buf.String())
	return strings.Join(words, "_")

}

//删除字符串中的多余空格，有多个空格时，仅保留一个空格
func DeleteExtraSpace(s string) string {
	s1 := strings.Replace(s, "	", " ", -1)       //替换tab为空格
	regstr := "\\s{2,}"                          //两个及两个以上空格的正则表达式
	reg, _ := regexp.Compile(regstr)             //编译正则表达式
	s2 := make([]byte, len(s1))                  //定义字符数组切片
	copy(s2, s1)                                 //将字符串复制到切片
	spc_index := reg.FindStringIndex(string(s2)) //在字符串中搜索
	for len(spc_index) > 0 {                     //找到适配项
		s2 = append(s2[:spc_index[0]+1], s2[spc_index[1]:]...) //删除多余空格
		spc_index = reg.FindStringIndex(string(s2))            //继续在字符串中搜索
	}
	return string(s2)
}
