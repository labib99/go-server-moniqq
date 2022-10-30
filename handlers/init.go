package handlers

import (
	"log"
	"time"
)

var timeLoc *time.Location
var dtLayout = [6]string{
	"02/01/2006 15:04:05",
	"02/01/2006 15:04:05 − 15:04:05",
	"02/01/2006 15:04:05 - 15:04:05",
	"Jan/02/2006 15:04",
	"02/01/2006 15.04",
	"2006-01-02",
}
var parameters = [4]string{"throughput", "packet loss", "delay", "jitter"}

func init() {
	var err error

	loadEnv()
	createDBInstance()
	sessionAndAuth()

	timeLoc, err = time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Fatalln(err)
	}
}
