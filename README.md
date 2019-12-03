# go-transformer
converter between yaml/json text and golang struct text
```
json/yaml文本与golang struct文本之间的转换:
  读取json/yaml文本，生成struct文本
  读取struct文本，生成json/yaml文本
```

## example
[example.go](https://github.com/berix9/go-transformer/blob/master/examples/examples.go)
### json to struct
* input
```json
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
```
* output 
```go
type Mega struct {
    Root Root `json:"Root"`
}

type Root struct {
    Users []Users `json:"users"`
}

type Users struct {
    Age int `json:"age"`
    Friends []Friends `json:"friends"`
    Id int `json:"id"`
    NameLoc string `json:"name_loc"`
    Rate []int `json:"rate"`
    Vip bool `json:"vip"`
}

type Friends struct {
    Id int `json:"id"`
    Name string `json:"name"`
}

```
### struct to yaml
* input
```go
type Model struct {
	Name     string            `json:"name" bson:"name"` //一般注释，指定值->  #eg#my name
	Creator  string            `json:"creator" bson:"creator"`
	Labels   []string          `json:"labels" bson:"labels"`
	Versions []ModelSubVersion `json:"versions" bson:"versions"`
}

//some comment
type ModelSubVersion struct {
	Version          int         `json:"version" bson:"version"`
	OriginalFilePath string      `json:"originalFilePath" bson:"originalFilePath"`
	Volume           VolumeClaim `json:"volume" bson:"volume"`
	FilePath         string      `json:"filePath" bson:"filePath"`
	Summary          SummarySpec `json:"summary" bson:"summary"`
	Desc             string      `json:"desc" bson:"desc"`
}
```
* output
```yaml
Creator: string-a
Labels:
- str-a
- str-b
- string-c
Name: my name
Versions:
- Desc: string-a
  FilePath: string-a
  OriginalFilePath: string-a
  Version: 1
```
