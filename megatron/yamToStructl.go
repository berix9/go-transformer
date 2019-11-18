package megatron

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"strings"
)

func yamlUnmarshal(content string) (map[string]interface{}, error) {
	restMap := make(map[string]interface{})
	m := make(map[interface{}]interface{})
	if strings.HasPrefix(content, "-") {
		//if array, unmarshal into map-array, then use the first object
		maps := make([]map[interface{}]interface{}, 0, 1)
		if err := yaml.Unmarshal([]byte(content), &maps); err != nil {
			return restMap, err
		}
		if len(maps) > 0 {
			m = maps[0]
		} else {
			return restMap, errors.New("the json string given has no valid content")
		}
	} else {
		//unmarshal into map
		if err := yaml.Unmarshal([]byte(content), &m); err != nil {
			return restMap, err
		}
	}
	for k, v := range m {
		restMap[fmt.Sprintf("%v", k)] = v
	}
	return restMap, nil
}

/*func (m *megatron) YamlToStruct() (string, error) {
	opt := m.mergeOptions()
	return yamlToStruct(strings.TrimSpace(m.Content), opt)
}

//convert yaml string to struct string
func yamlToStruct(yamlStr string, opt *option) (res string, err error) {
	//yaml string is arry/object
	m := make(map[interface{}]interface{})
	if strings.HasPrefix(yamlStr, "-") {
		//if array, unmarshal into map-array, then use the first object
		maps := make([]map[interface{}]interface{}, 0, 1)
		if err = yaml.Unmarshal([]byte(yamlStr), &maps); err != nil {
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

	for _, w := range opt.Writers {
		io.Copy(w, buffer)
		switch w.(type) {
		case *os.File:
			w.(*os.File).Seek(0, 0)
			info, _ := w.(*os.File).Stat()
			if info.Name() == "stdout" {
				continue
			}
			io.Copy(buffer, w.(*os.File))
		}
	}


	return buffer.String(), nil
}
*/

/*
convert map to struct string
	m: the map after unmarshal
*/
/*func yamlMapToStruct(buffer *bytes.Buffer, m map[interface{}]interface{}, opt *option) error {
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
			typeValue = "interface{}"
		}
		upKey := utils.UpperWords(k.(string))
		descText := fmt.Sprintf("`json:\"%s\"`", k)
		buffer.WriteString(fmt.Sprintf("    %s %s %s\n", upKey, typeValue, descText))
	}
	buffer.WriteString("}\n\n")

	//process nested-struct
	if len(objs) > 0 {
		for k, v := range objs {
			//fmt.Println("-----", k, reflect.TypeOf(k), "---|---", v, reflect.TypeOf(v))
			opt.StructName = k.(string)
			yamlMapToStruct(buffer, v.(map[interface{}]interface{}), opt)
		}
	}

	return nil
}*/
