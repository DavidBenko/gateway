package logger

import "log"

func Fatal(v ...interface{}) {
	log.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

func Fatalln(v ...interface{}) {
	log.Fatalln(v...)
}

func Panic(v ...interface{}) {
	log.Panic(v...)
}

func Panicf(format string, v ...interface{}) {
	log.Panicf(format, v...)
}

func Panicln(v ...interface{}) {
	log.Panicln(v...)
}

func Print(v ...interface{}) {
	log.Print(v...)
}

func Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func Println(v ...interface{}) {
	log.Println(v...)
}
