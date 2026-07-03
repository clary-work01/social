package main

import (
	"fmt"
	"reflect"
	"strings"
)

func main() {
	type Post struct {
		Title string `json:"title" validate:"required,max=10"`
	}

	var post Post
	t := reflect.TypeOf(post) // t: main.Post

	field, _ := t.FieldByName("Title") // field: {Title  string json:"title" validate:"required,max=10" 0 [0] false}

	validateTag := field.Tag.Get("validate") // validateTag: required,max=10

	validateRules := strings.Split(validateTag, ",") // validateRules: [required max=10]
	fmt.Println(validateRules)
}
