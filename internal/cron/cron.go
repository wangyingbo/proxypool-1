package cron

import (
	"runtime"

	"github.com/jasonlvhit/gocron"
	"github.com/wangyingbo/proxypool-1/internal/app"
)

func Cron() {
	_ = gocron.Every(480).Minutes().Do(crawlTask)
	<-gocron.Start()
}

func crawlTask() {
	_ = app.InitConfigAndGetters("")
	app.CrawlGo()
	app.Getters = nil
	runtime.GC()
}
