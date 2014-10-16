package config

import (
	"flag"
	"log"
	"reflect"
	"strconv"
)

func setupFlags(value reflect.Value) {
	reflectConfiguration(
		value,
		func(name, _, _ string) bool {
			return name != ""
		},
		func(field reflect.Value, name, value, usage string) {
			switch field.Kind() {
			case reflect.Int64:
				flag.Int64(name, parseInt(value), usage)
			case reflect.String:
				flag.String(name, value, usage)
			}
		},
	)
}

func setDefaults(value reflect.Value) {
	reflectConfiguration(
		value,
		func(_, defaultValue, _ string) bool {
			return defaultValue != ""
		},
		func(field reflect.Value, _, value, _ string) {
			switch field.Kind() {
			case reflect.Int64:
				field.SetInt(parseInt(value))
			case reflect.String:
				field.SetString(value)
			}
		},
	)
}

func setFromFlags(value reflect.Value) {
	setFlags := make(map[string]flag.Value)
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = f.Value
	})

	reflectConfiguration(
		value,
		func(flagName, _, _ string) bool {
			_, ok := setFlags[flagName]
			return ok
		},
		func(field reflect.Value, name, _, _ string) {
			setFlagValue, _ := setFlags[name]
			switch field.Kind() {
			case reflect.Int64:
				field.SetInt(parseInt(setFlagValue.String()))
			case reflect.String:
				field.SetString(setFlagValue.String())
			}
		},
	)
}

func reflectConfiguration(
	value reflect.Value,
	shouldHandle func(flagName, flagValue, flagUsage string) bool,
	handle func(fieldValue reflect.Value, flagName, flagValue, flagUsage string),
) {
	if value.Kind() == reflect.Struct {
		t := reflect.TypeOf(value.Interface())
		n := t.NumField()
		for i := 0; i < n; i++ {
			field := t.Field(i)

			flagName := field.Tag.Get("flag")
			flagValue := field.Tag.Get("default")
			flagUsage := field.Tag.Get("usage")

			fieldValue := value.FieldByName(field.Name)

			if shouldHandle(flagName, flagValue, flagUsage) {
				handle(fieldValue, flagName, flagValue, flagUsage)
			} else if fieldValue.Kind() == reflect.Struct {
				reflectConfiguration(fieldValue, shouldHandle, handle)
			}
		}
	}
}

func parseInt(value string) int64 {
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return i
}
