package config

import "reflect"

// cleanConfigStrings recursively walks through the config and removes quotes from string fields
func cleanConfigStrings(v reflect.Value) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		switch field.Kind() {
		case reflect.String:
			if field.CanSet() {
				field.SetString(removeStringSign(field.String()))
			}
		case reflect.Struct:
			cleanConfigStrings(field)
		case reflect.Pointer:
			if !field.IsNil() && field.Elem().Kind() == reflect.Struct {
				cleanConfigStrings(field.Elem())
			}
		}
	}
}

// removeStringSign removes quotes from a string
func removeStringSign(str string) string {
	if len(str) >= 2 && (str[0] == '"' || str[0] == '\'') && (str[len(str)-1] == '"' || str[len(str)-1] == '\'') {
		return str[1 : len(str)-1]
	}
	return str
}
