package printer

import (
	"fmt"
	"reflect"
)

func PrintMap[V any](m map[string]V) {
	for key, _ := range m {
		fmt.Println(key)
	}
}

func PrintConfiguration(configuration interface{}) {
	fmt.Println()
	reflectObj := reflect.ValueOf(configuration)
	for i := 0; i < reflectObj.NumField(); i++ {
		field := reflectObj.Type().Field(i)
		value := reflectObj.Field(i)
		fmt.Printf("%s: %+v", field.Name, value)
	}
	fmt.Println()
}
