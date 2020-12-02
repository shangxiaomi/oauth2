package mylog

import (
	"log"
	"os"
)

var (
	Debug *log.Logger
	Info  *log.Logger
	Error *log.Logger
	Warn  *log.Logger
)
var LogFile *os.File

func init() {
	log.Println("init ...")
	Debug = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	Warn = log.New(os.Stdout, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	file := "./" + "oauth2" + ".log"
	// TODO : 这里如果直接对 LogFile赋值，会变成局部变量？有待考证
	myFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	LogFile = myFile
	if err != nil {
		panic(err)
	}
	Debug.SetOutput(LogFile) // 将文件设置为log输出的文件
	Info.SetOutput(LogFile)  // 将文件设置为log输出的文件
	Error.SetOutput(LogFile) // 将文件设置为log输出的文件
	Warn.SetOutput(LogFile)  // 将文件设置为log输出的文件
	log.SetOutput(LogFile)   // 将文件设置为log输出的文件
}

func GetLogFile() *os.File {
	return LogFile
}