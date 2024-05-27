package clogger

import "log"

// ConsoleLogger is a simple logger that logs output to the console.
var ConsoleLogger = log.New(&consoleWriter{}, "logger: ", log.Ldate|log.Ltime|log.Lshortfile)
