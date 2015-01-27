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
		func(name, _ string) bool {
			return name != ""
		},
		func(field reflect.Value, name, value string) {
			usage, ok := usageStrings[name]
			if !ok {
				log.Fatalf("No usage string for flag %s", name)
			}
			switch field.Kind() {
			case reflect.Int64:
				flag.Int64(name, parseInt(value), usage)
			case reflect.String:
				flag.String(name, value, usage)
			case reflect.Bool:
				flag.Bool(name, parseBool(value), usage)
			}
		},
	)
}

func setDefaults(value reflect.Value) {
	reflectConfiguration(
		value,
		func(_, defaultValue string) bool {
			return defaultValue != ""
		},
		func(field reflect.Value, _, value string) {
			switch field.Kind() {
			case reflect.Int64:
				field.SetInt(parseInt(value))
			case reflect.String:
				field.SetString(value)
			case reflect.Bool:
				field.SetBool(parseBool(value))
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
		func(flagName, _ string) bool {
			_, ok := setFlags[flagName]
			return ok
		},
		func(field reflect.Value, name, _ string) {
			setFlagValue, _ := setFlags[name]
			switch field.Kind() {
			case reflect.Int64:
				field.SetInt(parseInt(setFlagValue.String()))
			case reflect.String:
				field.SetString(setFlagValue.String())
			case reflect.Bool:
				field.SetBool(parseBool(setFlagValue.String()))
			}
		},
	)
}

func reflectConfiguration(
	value reflect.Value,
	shouldHandle func(flagName, flagValue string) bool,
	handle func(fieldValue reflect.Value, flagName, flagValue string),
) {
	if value.Kind() == reflect.Struct {
		t := reflect.TypeOf(value.Interface())
		n := t.NumField()
		for i := 0; i < n; i++ {
			field := t.Field(i)

			flagName := field.Tag.Get("flag")
			flagValue := field.Tag.Get("default")

			fieldValue := value.FieldByName(field.Name)

			if shouldHandle(flagName, flagValue) {
				handle(fieldValue, flagName, flagValue)
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

func parseBool(value string) bool {
	b, err := strconv.ParseBool(value)
	if err != nil {
		log.Fatal(err)
	}
	return b
}
