package egress

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/ottstack/mechaproxy/pkg/innerip"
	"github.com/ottstask/configapi/pkg/meta"
	"github.com/ottstask/configapi/pkg/watchclient"
)

const defaultNamespace = "default"

var egressCfg = atomic.Value{}

var domainCfg = sync.Map{}

var domainAccessCh = make(chan string, 10000)

var listenConfig = &serveConfig{
	Addr:          ":19000",
	GetConfigAddr: "http://localhost:8080/api/config/domain",
}

type serveConfig struct {
	Addr          string
	GetConfigAddr string
}

func init() {
	err := envconfig.Process("egress", listenConfig)
	if err != nil {
		log.Fatal(err)
	}
}

func getNamespace(ip string) string {
	val := egressCfg.Load()
	if val == nil {
		return defaultNamespace
	}
	cfg := val.(*meta.EgressConfig)
	if v, ok := cfg.HostNamespace[ip]; ok {
		return v
	}
	return defaultNamespace
}

func getDomainDownstream(domain string) *meta.DownStreamInfo {
	select {
	case domainAccessCh <- domain:
	default:
	}

	val, ok := domainCfg.Load(domain)
	if !ok {
		return getDownstreamFromRemote(domain)
	}
	cfg := val.(*meta.DomainConfig)
	num := len(cfg.DownStreams)
	if num == 0 {
		return nil
	}
	idx := atomic.AddInt32(&cfg.Index, 1)
	if idx >= int32(num) {
		atomic.StoreInt32(&cfg.Index, 0)
	}
	ds := cfg.DownStreams[int(idx)%num]
	return ds
}

func getDownstreamFromRemote(domain string) *meta.DownStreamInfo {
	rsp, err := http.Get(listenConfig.GetConfigAddr + "?domain=" + domain)
	if err != nil {
		log.Println("getDownstreamFromRemote error", err)
		return nil
	}
	defer rsp.Body.Close()
	bs, _ := io.ReadAll(rsp.Body)
	if rsp.StatusCode != 200 {
		log.Println("getDownstreamFromRemote bad status code", rsp.StatusCode, string(bs))
		return nil
	}
	domainInfo := &meta.DomainConfig{}
	if err := json.Unmarshal(bs, domainInfo); err != nil {
		log.Println("getDownstreamFromRemote Unmarshal error", err)
		return nil
	}
	if domainInfo.IsZero {
		// TODO: is zero
		return nil
	}

	// TODO: robin select
	for _, ds := range domainInfo.DownStreams {
		return ds
	}
	return nil
}

func WatchDomain() {
	gcInterval := time.NewTimer(time.Minute)
	curDomainAccess := make(map[string]bool)
	allDomainAccess := make(map[string]chan struct{})
	for {
		select {
		case domain := <-domainAccessCh:
			isNewDomain := false
			if _, ok := allDomainAccess[domain]; !ok {
				isNewDomain = true
			}
			curDomainAccess[domain] = true

			if isNewDomain {
				ch := make(chan struct{})
				allDomainAccess[domain] = ch
				go func(domain string) {
					domainConfigKey := meta.DomainConfigKeyPrefix + domain
					for {
						select {
						case val := <-watchclient.Watch(domainConfigKey, &meta.DomainConfig{}):
							domainCfg.Store(domain, val)
						case <-ch:
							fmt.Println("close ch")
							return
						}
					}
				}(domain)
			}
		case <-gcInterval.C:
			for domain, ch := range allDomainAccess {
				if !curDomainAccess[domain] {
					watchclient.Delete(meta.DomainConfigKeyPrefix + domain)
					domainCfg.Delete(domain)
					delete(allDomainAccess, domain)
					close(ch)
				}
			}
			curDomainAccess = make(map[string]bool)
		}
	}
}

func WatchConfig() {
	egressConfigKey := meta.EgressConfigKeyPrefix + innerip.Get()
	for val := range watchclient.Watch(egressConfigKey, &meta.EgressConfig{}) {
		egressCfg.Store(val)
	}
}
