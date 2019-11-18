package megatron

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-conv/utils"
	"reflect"
	"strings"
)

type Option struct {
	StructName string //specify struct Name
	Recursive  bool   //if true, parse the json recursively

}

func Json2Struct(js string, opts ...*Option) (string, error) {
	opt := mergeOptions(opts)
	return jsonToStruct(strings.TrimSpace(js), opt)
}

//convert json string to struct string
func jsonToStruct(js string, opt *Option) (res string, err error) {
	//json string is array/object
	m := make(map[string]interface{})
	if strings.HasPrefix(js, "[") {
		//if array, unmarshal into map-array, then get the first object
		maps := make([]map[string]interface{}, 0, 1)
		if err = json.Unmarshal([]byte(js), &maps); err != nil {
			return
		}
		if len(maps) > 0 {
			m = maps[0]
		} else {
			return "", errors.New("the json string given has no valid content")
		}
	} else {
		//unmarshal into map
		if err = json.Unmarshal([]byte(js), &m); err != nil {
			return
		}
	}

	buffer := bytes.NewBufferString("")
	if err = jsonMapToStruct(buffer, m, opt); err != nil {
		fmt.Println("map2struct:", err)
		return
	}

	return buffer.String(), nil
}

/*
convert map to struct string
	m: the map after json.Unmarshal
*/
func jsonMapToStruct(buffer *bytes.Buffer, m map[string]interface{}, opt *Option) error {
	opt.StructName = opt.StructName[strings.LastIndex(opt.StructName, "]")+1:]
	buffer.WriteString(fmt.Sprintf("type %s struct {\n", opt.StructName))
	objs := make(map[string]interface{})
	typeValue := ""
	for k, v := range m {
		//fmt.Println("-----", k, reflect.TypeOf(k), "---|---", v, reflect.TypeOf(v))
		switch v.(type) {
		case float64:
			vs := fmt.Sprintf("%v", v)
			if strings.Contains(vs, ".") {
				typeValue = "float64"
			} else {
				typeValue = "int"
			}
		case string:
			typeValue = "string"
		case bool:
			typeValue = "bool"
		case map[string]interface{}:
			//if value is object and recursive-option is true,
			// 	use key as nested-struct name, and save value into the map waiting to be process recursively.
			if opt.Recursive {
				typeValue = utils.UpperWords(k)
				objs[typeValue] = v
			} else {
				typeValue = "interface{}"
			}
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
					typeValue = fmt.Sprintf("[]%s", utils.UpperWords(k))
					objs[typeValue] = v.([]interface{})[0]
				}
			} else {
				typeValue = "[]interface{}"
			}
		default:
			typeValue = "interface{}"
		}
		upKey := utils.UpperWords(k)
		descText := fmt.Sprintf("`json:\"%s\"`", k)
		buffer.WriteString(fmt.Sprintf("    %s %s %s\n", upKey, typeValue, descText))
	}
	buffer.WriteString("}\n\n")

	//process nested-struct
	if len(objs) > 0 {
		for k, v := range objs {
			opt.StructName = k
			fmt.Println("-----", k, reflect.TypeOf(k), "---|---", v, reflect.TypeOf(v))
			jsonMapToStruct(buffer, v.(map[string]interface{}), opt)
		}
	}

	return nil
}

func mergeOptions(opts []*Option) *Option {
	opt := &Option{}
	for _, op := range opts {
		if op.Recursive == true {
			opt.Recursive = true
		}
		if op.StructName != "" {
			name := utils.UpperWords(op.StructName)
			opt.StructName = name
		}
	}
	if opt.StructName == "" {
		opt.StructName = default_struct_name
	}
	return opt
}
