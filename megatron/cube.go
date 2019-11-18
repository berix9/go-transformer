package megatron

import (
	"bytes"
	"fmt"
	"go-transformer/utils"
	"io"
	"os"
	"path"
	"strings"
)

const (
	default_struct_name = "AutoStructName"
	Json_to_struct      = 1
	Yaml_to_struct      = 2
)

type megatron struct {
	Content string
	Tp      int
	Err     error
	Options []*option
}

type option struct {
	StructName string //specify struct Name
	Recursive  bool   //if true, parse the json recursively
	Writers    []io.Writer
}

func NewMegatron(content string, tp int) *megatron {
	return &megatron{Content: content, Tp: tp}
}

func (m *megatron) SetStructName(structName string) *megatron {
	m.Options = append(m.Options, &option{StructName: structName})
	return m
}

//
func (m *megatron) SetOptionRecursive() *megatron {
	m.Options = append(m.Options, &option{Recursive: true})
	return m
}

//
func (m *megatron) Error() error {
	return m.Err
}

//parse
func (m *megatron) SetOutputFile(f string, alsoStdout bool) *megatron {
	if m.occurError() {
		return m
	}
	var writers []io.Writer
	var file *os.File
	_, err := os.Stat(f)
	if err != nil {
		_, e := os.Stat(path.Dir(f))
		if e != nil {
			if os.IsNotExist(e) {
				if e := os.MkdirAll(path.Dir(f), os.ModePerm); e != nil {
					m.setError(err)
					return m
				}
			} else {
				m.setError(err)
				return m
			}
		}
	}
	file, err = os.OpenFile(f, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		m.setError(err)
		return m
	}

	writers = append(writers, file)
	if alsoStdout {
		writers = append(writers, os.Stdout)
	}

	m.Options = append(m.Options, &option{Writers: writers})
	return m
}

//convert json string to struct string
func (meg *megatron) ToStruct() *megatron {
	if meg.occurError() {
		return meg
	}
	opt := meg.mergeOptions()
	var err error
	m := make(map[string]interface{})
	if meg.Tp == Json_to_struct {
		m, err = jsonUnmarshal(strings.TrimSpace(meg.Content))
		if err != nil {
			meg.setError(err)
			return meg
		}
	} else if meg.Tp == Yaml_to_struct {
		m, err = yamlUnmarshal(strings.TrimSpace(meg.Content))
		if err != nil {
			meg.setError(err)
			return meg
		}
	}

	buffer := bytes.NewBufferString("")
	if err := mapToStruct(buffer, m, opt); err != nil {
		fmt.Println("map2struct:", err)
		meg.setError(err)
		return meg
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

	return meg
}

/*
*******************private*********************
 */

func (m *megatron) mergeOptions() *option {
	opt := &option{}
	for _, op := range m.Options {
		if op.Recursive == true {
			opt.Recursive = true
		}
		if op.StructName != "" {
			name := utils.UpperWords(op.StructName)
			opt.StructName = name
		}
		if len(op.Writers) > 0 {
			opt.Writers = op.Writers
		}
	}
	if opt.StructName == "" {
		opt.StructName = default_struct_name
	}
	if len(opt.Writers) < 1 {
		opt.Writers = append(opt.Writers, os.Stdout)
	}

	return opt
}

/*
convert map to struct string
	m: the map after unmarshal
*/
func mapToStruct(buffer *bytes.Buffer, m map[string]interface{}, opt *option) error {
	opt.StructName = opt.StructName[strings.LastIndex(opt.StructName, "]")+1:]
	buffer.WriteString(fmt.Sprintf("type %s struct {\n", opt.StructName))
	objs := make(map[string]interface{})
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
				case map[interface{}]interface{}:
					if opt.Recursive {
						typeValue = utils.UpperWords(k)
						m := make(map[string]interface{})
						for tk, tv := range nv.(map[interface{}]interface{}) {
							m[fmt.Sprintf("%v", tk)] = tv
						}
						objs[typeValue] = m
					} else {
						typeValue = "[]interface{}"
					}
				default:
					typeValue = fmt.Sprintf("[]%s", utils.UpperWords(k))
					objs[typeValue] = v.([]interface{})[0]
				}
			} else {
				typeValue = "[]interface{}"
			}
		case map[string]interface{}:
			//if value is object and recursive-option is true,
			// 	use key as nested-struct name, and save value into the map waiting to be process recursively.
			if opt.Recursive {
				typeValue = utils.UpperWords(k)
				objs[typeValue] = v
			} else {
				typeValue = "interface{}"
			}
		case map[interface{}]interface{}:
			//if value is object and recursive-option is true,
			// 	use key as nested-struct name, and save value into the map waiting to be process recursively.
			if opt.Recursive {
				typeValue = utils.UpperWords(k)
				m := make(map[string]interface{})
				for tk, tv := range v.(map[interface{}]interface{}) {
					m[fmt.Sprintf("%v", tk)] = tv
				}
				objs[typeValue] = m
			} else {
				typeValue = "interface{}"
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
			//fmt.Println("-----", k, reflect.TypeOf(k), "---|---", v, reflect.TypeOf(v))
			opt.StructName = k
			mapToStruct(buffer, v.(map[string]interface{}), opt)
		}
	}

	return nil
}

func (m *megatron) setError(e error) {
	m.Err = e
}

func (m *megatron) occurError() bool {
	return m.Err != nil
}
