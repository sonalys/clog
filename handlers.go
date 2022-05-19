package main

import "github.com/valyala/fastjson"

var jsonParser = fastjson.Parser{}

func jsonHandler(line []byte) (fields []field) {
	v, _ := jsonParser.ParseBytes(line)
	obj := v.GetObject()
	obj.Visit(func(key []byte, v *fastjson.Value) {
		fields = append(fields, field{
			key:   key,
			value: v.GetStringBytes(),
		})
	})
	return
}

func textHandler(line []byte) (fields []field) {
	lineLen := len(line)
	cur := 0
	isKey := true
	field := field{}

	for i := 0; i < lineLen; i++ {
		switch line[i] {
		case ' ':
			if isKey {
				field.key = line[cur:i]
			} else {
				field.value = line[cur:i]
			}
			isKey = true
			cur = i + 1
			fields = append(fields, field)
		case '=':
			if isKey {
				isKey = false
				field.key = line[cur:i]
				cur = i + 1
			}
		}
	}
	return
}

func defineParseStrategy(line []byte) parserHandler {
	switch line[0] {
	case '{':
		return jsonHandler
	default:
		return textHandler
	}
}
