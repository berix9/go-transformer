package main

import (
	"fmt"
	"go-transformer/megatron"
	"go-transformer/optimus"
	"go-transformer/utils"
	"io/ioutil"
)

func main() {
	//json2struct()
	//yaml2struct()
	struct2ToJson()
	//genJsonKey("IDUserNameIDStop")
}

func yaml2struct() {
	str := `   
- name_loc: ss
  age: 18
  vip: true
  friends:
  - name: ys
    gender: 男
    age: 33
    sub_sub_struct:
      name: ys
      gender: 男
      age: 33
`

	if e := megatron.NewMegatron(str, megatron.Yaml_to_struct).
		SetOptionRecursive().SetStructName("Mega").ToStruct().Error(); e != nil {
		fmt.Println("error:", e)
		return
	}
}

func json2struct() {
	str := ` 
{
    "Root": {
        "users": [
            {
                "age": 1,
                "friends": [
                    {
                        "id": 1,
                        "name": "berix"
                    }
                ],
                "id": 3,
                "name_loc": "string-a",
                "rate": [
                    1,
                    1,
                    1
                ],
                "vip": false
            }
        ]
    }
}
`

	//设置输出到文件[SetXxxx方法均为可选]
	if e := megatron.NewMegatron(str, megatron.Json_to_struct).
		SetOutputFile("/Users/yangsen/Documents/go-ys/src/go-transformer/examples/a/b", true).
		SetOptionRecursive().SetStructName("Mega").ToStruct().Error(); e != nil {
		fmt.Println("error:", e)
		return
	}
	//默认输出到标准输出[SetXxxx方法均为可选]
	if e := megatron.NewMegatron(str, megatron.Json_to_struct).SetOptionRecursive().SetStructName("Mega").
		ToStruct().Error(); e != nil {
		fmt.Println("error:", e)
		return
	}
}

func struct2ToJson() {
	f, e := ioutil.ReadFile("/Users/yangsen/Documents/go-ys/src/go-transformer/examples/struct2json-3")
	if e != nil {
		fmt.Println("err:", e)
		return
	}
	str := string(f)
	//fmt.Println(optimus.StructsToJson(str, optimus.ConvertTypeJson))
	fmt.Println(optimus.StructsCube(str, optimus.ConvertTypeYaml))
}

func genJsonKey(s string) {
	fmt.Println(utils.CamelToSnake(s))
}
