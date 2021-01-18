package main

import (
	"os"

	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/sirupsen/logrus"

	"ydtb/bot"
	"ydtb/config"
)

func main() {
	logrus.SetFormatter(&runtime.Formatter{
		ChildFormatter: &logrus.TextFormatter{},
		File:           true,
		BaseNameOnly:   true,
		Line:           true},
	)
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	conf, err := config.Get()
	if err != nil {
		logrus.Fatalln("error while retrieve config:", err)
		return
	}

	b, err := bot.New(conf)
	if err != nil {
		logrus.Fatalln("error while create new bot instance:", err)
	}

	b.Start()
}
