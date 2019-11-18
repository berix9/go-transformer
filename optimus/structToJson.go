package optimus

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-transformer/utils"
	"io"
	"strconv"
	"strings"
	"unicode"
)

const (
	unknown            = "unknown"
	specific_value_tag = "#eg#"
)

type object struct {
	Name         string
	JsonKey      string
	ParentKey    string
	CommonFileds []*field
	SubObj       []*object
	Tp           string
}

type field struct {
	JsonKey string
	Tp      string
	Value   interface{}
}
type RealStruct struct {
	Fields map[string]interface{}
}

func StructsToJson(str string) string {
	bufReader := bufio.NewReader(strings.NewReader(str))
	obj := new(object)
	objs, err := splitStructs(bufReader, obj, "root", false)
	if err != nil {
		panic(err)
	}
	structMaps := convertStructsToMap(objs)
	return mergeStructMaps(structMaps)
}

func mergeStructMaps(maps map[string]map[string]interface{}) string {
	var subed []string
	result := make(map[string]interface{})
	for structName, m := range maps {
		for filedKey, v := range m {
			switch v.(type) {
			case string:
				if v == unknown {
					//"k" is JsonKey.Tp
					k := strings.Split(filedKey, ".")
					//fmt.Println(filedKey, v)
					if subObj, ok := maps[strings.TrimPrefix(k[1], "[]")]; ok {
						m[k[0]] = subObj
						if strings.HasPrefix(k[1], "[]") {
							m[k[0]] = []interface{}{subObj}
						}
						//result[structName] = m
						subed = append(subed, strings.TrimPrefix(k[1], "[]"))
						//delete(result, strings.TrimPrefix(k[1], "[]"))
					} else {
						m[k[0]] = v
					}
					delete(m, filedKey)
				}
			}
		}
		result[structName] = m
	}
	//fmt.Println(subed)

	for k, _ := range result {
		for _, excluded := range subed {
			if strings.TrimPrefix(k, "[]") == excluded {
				delete(result, k)
			}
		}
	}
	buf := bytes.Buffer{}
	moreThanOne := false
	for _, v := range result {
		if moreThanOne {
			buf.WriteString("\n\n********************************\n\n")
		}
		by, _ := json.MarshalIndent(v, "", "    ")
		buf.Write(by)
		moreThanOne = true
	}
	return buf.String()
}

func convertStructsToMap(objs []*object) map[string]map[string]interface{} {
	structMaps := make(map[string]map[string]interface{})
	for _, obj := range objs {
		if obj.Name == "" {
			continue
		}
		m := make(map[string]interface{})
		for _, f := range obj.CommonFileds {
			//fmt.Println("fields:", obj.Name, "|", f.Tp, f.Value, "|", reflect.TypeOf(f.Value))
			v := f.fillValue().Value
			switch v.(type) {
			case string:
				if v == unknown {
					m[fmt.Sprintf("%s.%s", f.JsonKey, f.Tp)] = v
				} else {
					m[f.JsonKey] = v
				}
			default:
				m[f.JsonKey] = v
			}
		}
		structMaps[obj.Name] = m
	}

	/*by, _ := json.Marshal(structMaps)
	fmt.Println("maps:\n", string(by))*/
	return structMaps

}
func splitStructs(reader *bufio.Reader, obj *object, parentKey string, nested bool) ([]*object, error) {
	obj.ParentKey = parentKey
	objects := make([]*object, 0, 1)
	multiComment := false
	for {
		lineBys, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				//fmt.Println("read done")
				break
			}
			return nil, err
		}
		line := string(lineBys)
		//fmt.Println("line:", line, line == "")

		//handle multi-lines comments and blank line
		if strings.HasPrefix(line, "/*") {
			multiComment = true
			continue
		}
		if strings.HasSuffix(line, "*/") {
			multiComment = false
			continue
		}
		if len(line) < 1 || strings.HasPrefix(line, "//") || multiComment {
			continue
		}

		//new struct: especially for nested struct
		if !strings.Contains(line, "type ") && strings.Contains(line, "{") {
			newObj := new(object)
			newObj.parseLine(line)
			obj.SubObj = append(obj.SubObj, newObj)
			obj.CommonFileds = append(obj.CommonFileds, &field{JsonKey: newObj.JsonKey, Tp: newObj.Tp})
			objRes, err := splitStructs(reader, newObj, obj.Name, true)
			if err != nil {
				fmt.Println("newObj err:", err)
				return nil, err
			}
			objects = append(objects, objRes...)
			continue
		}
		if strings.Contains(line, "}") {
			obj.parseLine(line)
			o := *obj
			objects = append(objects, &o)
			if nested {
				return objects, nil
			}
			obj = new(object)
			continue
		}
		obj.parseLine(line)
	}
	return objects, nil
}

func (obj *object) parseLine(line string) {
	//the line declares struct, get struct name
	// `type Friends struct {`  or  anonymous nested struct `FieldName struct {`
	if strings.Contains(line, "struct") && strings.Contains(line, "{") {
		obj.getStructName(line)
		return
	} else if strings.Contains(line, "}") {
		obj.getJsonKey(line)
		return
	}
	// the field of struct
	//		City          string `json:"city"`          // NewYork
	arr := splitFields(line)
	var f field
	if len(arr) < 2 {
		panic(errors.New(fmt.Sprintf("invalid struct field:%s", line)))
	}
	f.JsonKey = arr[0]
	f.Tp = arr[1]
	if len(arr) > 2 && (strings.Contains(arr[2], "json:") || strings.Contains(arr[2], "json :")) {
		f.getJsonKey(arr[2])
	} else {
		f.JsonKey = utils.CamelToSnake(f.JsonKey)
	}
	if len(arr) > 3 {
		f.Value = arr[3]
	}
	obj.CommonFileds = append(obj.CommonFileds, &f)

}

/*
*************object***************
 */
//get struct name
// when get the struct name, assume it to be  TP and JsonKey.
func (obj *object) getStructName(line string) {
	/*nested := false
	if !strings.Contains(line, "type ") {
		nested = true
	}*/

	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(strings.TrimSuffix(line, "{"), "type")
	line = strings.TrimSpace(line)
	obj.Name = strings.TrimSpace(strings.TrimSuffix(line, "struct"))
	if strings.Contains(obj.Name, "[]") {
		obj.Name = strings.TrimSpace(strings.TrimSuffix(obj.Name, "[]"))
	}
	obj.Tp = obj.Name
	obj.JsonKey = obj.Name
	/*if nested {
		obj.CommonFileds = append(obj.CommonFileds, &field{JsonKey: obj.Name, Tp: obj.Name})
	}*/
}

func (obj *object) getJsonKey(line string) {
	k := doGetJsonTag(line)
	if len(k) > 0 {
		obj.JsonKey = k
	} else {
		obj.JsonKey = utils.CamelToSnake(obj.Name)
	}
}

/*
*************field***************
 */
func (f *field) getJsonKey(js string) {
	k := doGetJsonTag(js)
	if len(k) > 0 {
		f.JsonKey = k
	}
}

func (f *field) fillValue() *field {
	switch f.Tp {
	case "int":
		if f.Value != nil {
			v, e := strconv.Atoi(f.Value.(string))
			if e != nil {
				panic(e)
			}
			f.Value = v
		} else {
			f.Value = 1
		}
	case "[]int":
		if f.Value != nil {
			var v []int
			if e := json.Unmarshal([]byte(f.Value.(string)), &v); e != nil {
				panic(e)
			}
			f.Value = v
		} else {
			f.Value = []int{1, 1, 1}
		}
	case "int32":
		if f.Value != nil {
			v, e := strconv.ParseInt(f.Value.(string), 10, 64)
			if e != nil {
				panic(e)
			}
			f.Value = int32(v)
		} else {
			f.Value = 2
		}
	case "[]int32":
		if f.Value != nil {
			var v []int32
			if e := json.Unmarshal([]byte(f.Value.(string)), &v); e != nil {
				panic(e)
			}
			f.Value = v
		} else {
			f.Value = []int32{2, 2, 2}
		}
	case "int64":
		if f.Value != nil {
			v, e := strconv.ParseInt(f.Value.(string), 10, 64)
			if e != nil {
				panic(e)
			}
			f.Value = v
		} else {
			f.Value = 3
		}
	case "[]int64":
		if f.Value != nil {
			var v []int64
			if e := json.Unmarshal([]byte(f.Value.(string)), &v); e != nil {
				panic(e)
			}
			f.Value = v
		} else {
			f.Value = []int64{3, 3, 3}
		}
	case "float64":
		if f.Value != nil {
			v, e := strconv.ParseFloat(f.Value.(string), 64)
			if e != nil {
				panic(e)
			}
			f.Value = v
		} else {
			f.Value = 1.1
		}
	case "[]float64":
		if f.Value != nil {
			var v []float64
			if e := json.Unmarshal([]byte(f.Value.(string)), &v); e != nil {
				panic(e)
			}
			f.Value = v
		} else {
			f.Value = []float64{1.1, 1.2, 1.3}
		}

	case "bool":
		if f.Value != nil {
			v, e := strconv.ParseBool(f.Value.(string))
			if e != nil {
				panic(e)
			}
			f.Value = v
		} else {
			f.Value = false
		}
	case "[]bool":
		if f.Value != nil {
			var v []bool
			if e := json.Unmarshal([]byte(f.Value.(string)), &v); e != nil {
				panic(e)
			}
			f.Value = v
		} else {
			f.Value = []bool{true, false, true}
		}

	case "string":
		if f.Value != nil {

			f.Value = f.Value.(string)
		} else {
			f.Value = "string-a"
		}
	case "[]string":
		if f.Value != nil {
			var v []string
			if e := json.Unmarshal([]byte(f.Value.(string)), &v); e != nil {
				panic(e)
			}
			f.Value = v
		} else {
			f.Value = []string{"str-a", "str-b", "string-c"}
		}
	default:
		f.Value = unknown

	}
	return f
}

/*
*************common***************
 */
func doGetJsonTag(js string) string {
	valid := false
	startInd := strings.Index(js, "json:")
	if startInd < 0 {
		startInd = strings.Index(js, "json :")
	}
	if startInd < 0 {
		return ""
	}
	words := js[startInd+4:]
	var name []rune
	for _, r := range words {
		if unicode.IsLetter(r) || r == '-' || r == '_' {
			name = append(name, r)
			valid = true
		} else if valid {
			break
		}
	}
	return string(name)
}

func splitFields(line string) []string {
	var name, tp, tag, value string
	line = strings.TrimSpace(utils.DeleteExtraSpace(line))
	arr := make([]string, 0, 4)
	ta := strings.Split(line, "//")
	if len(ta) > 1 && strings.Contains(ta[len(ta)-1], specific_value_tag) {
		value = strings.TrimSpace(ta[len(ta)-1])
		value = strings.TrimSpace(value[strings.LastIndex(value, specific_value_tag)+4:])
	}
	line = strings.TrimSpace(ta[0])

	ta = strings.Split(line, "`")
	if len(ta) > 2 {
		tag = ta[1]
	}
	line = strings.TrimSpace(ta[0])

	ta = strings.Split(line, " ")
	if len(ta) > 1 {
		name = ta[0]
		tp = ta[1]
	}
	if name != "" {
		arr = append(arr, name)
	}
	if tp != "" {
		arr = append(arr, tp)
	}
	if tag != "" {
		arr = append(arr, tag)
	}
	if value != "" {
		arr = append(arr, value)
	}

	return arr
}
