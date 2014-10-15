package config

import (
	"flag"
	"log"
	"reflect"
	"strconv"
)

// These could be DRY'd up, I'm sure.

func setupFlags(value reflect.Value) {
	if value.Kind() == reflect.Struct {
		t := reflect.TypeOf(value.Interface())
		n := t.NumField()
		for i := 0; i < n; i++ {
			field := t.Field(i)

			flagName := field.Tag.Get("flag")
			flagValue := field.Tag.Get("default")
			flagUsage := field.Tag.Get("usage")

			fieldValue := value.FieldByName(field.Name)

			if flagName != "" {
				switch fieldValue.Kind() {
				case reflect.Int64:
					intValue, err := strconv.ParseInt(flagValue, 10, 64)
					if err != nil {
						log.Fatal(err)
					}
					flag.Int64(flagName, intValue, flagUsage)
				case reflect.String:
					flag.String(flagName, flagValue, flagUsage)
				}
			} else if fieldValue.Kind() == reflect.Struct {
				setupFlags(fieldValue)
			}
		}
	}
}

func setDefaults(value reflect.Value) {
	if value.Kind() == reflect.Struct {
		t := reflect.TypeOf(value.Interface())
		n := t.NumField()
		for i := 0; i < n; i++ {
			field := t.Field(i)
			defaultValue := field.Tag.Get("default")
			fieldValue := value.FieldByName(field.Name)
			if defaultValue != "" {
				switch fieldValue.Kind() {
				case reflect.Int64:
					flagInt, err := strconv.ParseInt(defaultValue, 10, 64)
					if err != nil {
						log.Fatal(err)
					}
					fieldValue.SetInt(flagInt)
				case reflect.String:
					fieldValue.SetString(defaultValue)
				}
			} else if fieldValue.Kind() == reflect.Struct {
				setDefaults(fieldValue)
			}
		}
	}
}

func setFromFlags(value reflect.Value) {
	setFlags := make(map[string]flag.Value)
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = f.Value
	})

	if value.Kind() == reflect.Struct {
		t := reflect.TypeOf(value.Interface())
		n := t.NumField()
		for i := 0; i < n; i++ {
			field := t.Field(i)
			flag := field.Tag.Get("flag")
			fieldValue := value.FieldByName(field.Name)
			if setFlagValue, ok := setFlags[flag]; ok {
				switch fieldValue.Kind() {
				case reflect.Int64:
					flagInt, err := strconv.ParseInt(setFlagValue.String(), 10, 64)
					if err != nil {
						log.Fatal(err)
					}
					fieldValue.SetInt(flagInt)
				case reflect.String:
					fieldValue.SetString(setFlagValue.String())
				}
			} else if fieldValue.Kind() == reflect.Struct {
				setFromFlags(fieldValue)
			}
		}
	}
}
