package megatron

import (
	"bytes"
	"errors"
	"fmt"
	"go-conv/utils"
	"gopkg.in/yaml.v2"
	"reflect"
	"strings"
)

func YamlToStruct(yamlStr string, opts ...*Option) (string, error) {
	opt := mergeOptions(opts)
	yamlStr = strings.TrimSpace(yamlStr)
	return yamlToStruct(yamlStr, opt)
}

//convert yaml string to struct string
func yamlToStruct(yamlStr string, opt *Option) (res string, err error) {
	//yaml string is arry/object
	m := make(map[interface{}]interface{})
	if strings.HasPrefix(yamlStr, "-") {
		//if array, unmarshal into map-array, then use the first object
		maps := make([]map[interface{}]interface{}, 0, 1)
		if err = yaml.Unmarshal([]byte(yamlStr), &maps); err != nil {
			fmt.Println(yamlStr)
			fmt.Println(1, err)
			return
		}
		if len(maps) > 0 {
			m = maps[0]
		} else {
			return "", errors.New("the json string given has no valid content")
		}
	} else {
		//unmarshal into map
		if err = yaml.Unmarshal([]byte(yamlStr), &m); err != nil {
			return
		}
	}

	buffer := bytes.NewBufferString("")
	if err = yamlMapToStruct(buffer, m, opt); err != nil {
		fmt.Println("map2struct:", err)
		return
	}

	return buffer.String(), nil
}

/*
convert map to struct string
	m: the map after unmarshal
*/
func yamlMapToStruct(buffer *bytes.Buffer, m map[interface{}]interface{}, opt *Option) error {
	opt.StructName = opt.StructName[strings.LastIndex(opt.StructName, "]")+1:]
	buffer.WriteString(fmt.Sprintf("type %s struct {\n", opt.StructName))
	objs := make(map[interface{}]interface{})
	typeValue := ""
	for k, v := range m {
		//fmt.Println("**********", k, reflect.TypeOf(k), "*****|*****", v, reflect.TypeOf(v))
		switch v.(type) {
		case float64:
			vs := fmt.Sprintf("%v", v)
			if strings.Contains(vs, ".") {
				typeValue = "float64"
			} else {
				typeValue = "int"
			}
		case int:
			typeValue = "int"
		case int64:
			typeValue = "int64"
		case string:
			typeValue = "string"
		case bool:
			typeValue = "bool"
		case []interface{}:
			//if value is object-array and recursive-option is true,
			// 	use key as nested-struct name, and save value into the map waiting to be process recursively.
			if opt.Recursive && len(v.([]interface{})) > 0 {
				nv := v.([]interface{})[0]
				switch nv.(type) {
				case float64:
					vs := fmt.Sprintf("%v", nv)
					if strings.Contains(vs, ".") {
						typeValue = "[]float64"
					} else {
						typeValue = "[]int"
					}
				case bool:
					typeValue = "[]bool"
				case string:
					typeValue = "[]string"
				default:
					typeValue = fmt.Sprintf("[]%s", utils.UpperWords(k.(string)))
					objs[typeValue] = v.([]interface{})[0]
				}
			} else {
				typeValue = "[]interface{}"
			}
		case map[interface{}]interface{}:
			//if value is object and recursive-option is true,
			// 	use key as nested-struct name, and save value into the map waiting to be process recursively.
			if opt.Recursive {
				typeValue = utils.UpperWords(k.(string))
				objs[typeValue] = v
			} else {
				typeValue = "interface{}"
			}
		default:

		}
		upKey := utils.UpperWords(k.(string))
		descText := fmt.Sprintf("`json:\"%s\"`", k)
		buffer.WriteString(fmt.Sprintf("    %s %s %s\n", upKey, typeValue, descText))
	}
	buffer.WriteString("}\n\n")

	//process nested-struct
	if len(objs) > 0 {
		for k, v := range objs {
			fmt.Println("-----", k, reflect.TypeOf(k), "---|---", v, reflect.TypeOf(v))
			opt.StructName = k.(string)
			yamlMapToStruct(buffer, v.(map[interface{}]interface{}), opt)
		}
	}

	return nil
}
