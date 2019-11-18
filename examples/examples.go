package main

import (
	"fmt"
	"go-transformer/megatron"
	"go-transformer/optimus"
	"io/ioutil"
)

func main() {
	//json2struct()
	//yaml2struct()
	struct2ToJson()
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
	s, e := megatron.YamlToStruct(str, &megatron.Option{Recursive: true})
	if e != nil {
		fmt.Println("error:", e)
		return
	}
	fmt.Printf(s)
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
	s, e := megatron.Json2Struct(str, &megatron.Option{"", true})

	if e != nil {
		fmt.Println("error:", e)
		return
	}
	fmt.Printf(s)
}

func struct2ToJson() {
	f, _ := ioutil.ReadFile("/Users/yangsen/Documents/go-ys/src/go-transformer/examples/struct2json-3")
	str := string(f)

	fmt.Println(optimus.StructsToJson(str))

}
