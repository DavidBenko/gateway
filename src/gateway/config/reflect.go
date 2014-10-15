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
		func(flagName, defaultValue, flagUsage string) bool {
			return flagName != ""
		},
		func(fieldValue reflect.Value, flagName, flagValue, flagUsage string) {
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
		},
	)
}

func setDefaults(value reflect.Value) {
	reflectConfiguration(
		value,
		func(flagName, defaultValue, flagUsage string) bool {
			return defaultValue != ""
		},
		func(fieldValue reflect.Value, flagName, defaultValue, flagUsage string) {
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
		func(flagName, flagValue, flagUsage string) bool {
			_, ok := setFlags[flagName]
			return ok
		},
		func(fieldValue reflect.Value, flagName, flagValue, flagUsage string) {
			setFlagValue, _ := setFlags[flagName]
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
