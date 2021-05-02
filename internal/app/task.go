package app

import (
	"log"
	"sync"
	"time"

	"github.com/wangyingbo/proxypool-1/internal/cache"
	"github.com/wangyingbo/proxypool-1/internal/database"
	"github.com/wangyingbo/proxypool-1/pkg/provider"
	"github.com/wangyingbo/proxypool-1/pkg/proxy"
)

var location, _ = time.LoadLocation("PRC")

func CrawlGo() {
	wg := &sync.WaitGroup{}
	var pc = make(chan proxy.Proxy)
	for _, g := range Getters {
		wg.Add(1)
		go g.Get2Chan(pc, wg)
	}
	proxies := cache.GetProxies("allproxies")
	proxies = append(proxies, database.GetAllProxies()...)
	go func() {
		wg.Wait()
		close(pc)
	}()
	for node := range pc {
		if node != nil {
			proxies = append(proxies, node)
		}
	}
	// 节点衍生并去重
	proxies = proxies.Deduplication().Derive()
	log.Println("CrawlGo node count:", len(proxies))
	proxies = provider.Clash{
		provider.Base{
			Proxies: &proxies,
		},
	}.CleanProxies()
	log.Println("CrawlGo cleaned node count:", len(proxies))
	proxies.NameSetCounrty().Sort().NameAddIndex().NameAddTG()
	log.Println("Proxy rename DONE!")

	// 全节点存储到数据库
	database.SaveProxyList(proxies)

	cache.SetProxies("allproxies", proxies)
	cache.AllProxiesCount = proxies.Len()
	log.Println("AllProxiesCount:", cache.AllProxiesCount)
	cache.SSProxiesCount = proxies.TypeLen("ss")
	log.Println("SSProxiesCount:", cache.SSProxiesCount)
	cache.SSRProxiesCount = proxies.TypeLen("ssr")
	log.Println("SSRProxiesCount:", cache.SSRProxiesCount)
	cache.VmessProxiesCount = proxies.TypeLen("vmess")
	log.Println("VmessProxiesCount:", cache.VmessProxiesCount)
	cache.TrojanProxiesCount = proxies.TypeLen("trojan")
	log.Println("TrojanProxiesCount:", cache.TrojanProxiesCount)
	cache.LastCrawlTime = time.Now().In(location).Format("2006-01-02 15:04:05")

	// 可用性检测
	log.Println("Now proceed proxy health check...")
	proxies = proxy.CleanBadProxiesWithGrpool(proxies)
	log.Println("CrawlGo clash usable node count:", len(proxies))
	proxies.NameReIndex().NameAddTG()
	cache.SetProxies("proxies", proxies)
	cache.UsefullProxiesCount = proxies.Len()

	cache.SetString("clashproxies", provider.Clash{
		provider.Base{
			Proxies: &proxies,
		},
	}.Provide())
	cache.SetString("surgeproxies", provider.Surge{
		provider.Base{
			Proxies: &proxies,
		},
	}.Provide())
}
