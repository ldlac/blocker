package blocker

import (
	"bufio"
	"net/http"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type BlockDomainsDeciderABPUrl struct {
	blocklist    map[string]bool
	blocklistUrl string
	lastUpdated  time.Time
	log          Logger
}

// Name ...
func NewBlockDomainsDeciderABPUrl(url string, logger Logger) BlockDomainsDecider {
	return &BlockDomainsDeciderABPUrl{
		blocklistUrl: url,
		log:          logger,
		blocklist:    map[string]bool{},
	}
}

// IsDomainBlocked ...
func (d *BlockDomainsDeciderABPUrl) IsDomainBlocked(domain string) bool {
	// We will check every subdomain of the given domain against the blocklist. i.e. if example.com
	// is blocked, then every subdomain of that (subdomain.example.com, sub1.sub2.example.com) is
	// blocked. However, example.com.org is not blocked.
	comps := strings.Split(domain, ".")
	current := comps[len(comps)-1]
	for i := len(comps) - 2; i >= 0; i-- {
		newCurrent := strings.Join([]string{
			comps[i],
			current,
		}, ".")

		if d.blocklist[newCurrent] {
			return true
		}

		current = newCurrent
	}

	return false
}

// StartBlocklistUpdater ...
func (d *BlockDomainsDeciderABPUrl) StartBlocklistUpdater(ticker *time.Ticker) {
	go func() {
		for {
			tick := <-ticker.C
			d.log.Debugf("Ticker arrived at time: %v", tick)

			if d.IsBlocklistUpdateRequired() {
				d.log.Debug("update required")
				d.UpdateBlocklist()
			} else {
				d.log.Debug("update not required")
			}
		}
	}()
}

// UpdateBlocklist ...
func (d *BlockDomainsDeciderABPUrl) UpdateBlocklist() error {
	// Update process
	response, err := http.Get(d.blocklistUrl)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	blocklistContent := response.Body

	numBlockedDomainsBefore := len(d.blocklist)
	lastUpdatedBefore := d.lastUpdated

	scanner := bufio.NewScanner(blocklistContent)
	for scanner.Scan() {
		hostLine := scanner.Text()
		if !strings.HasPrefix(hostLine, "||") || !strings.HasSuffix(hostLine, "^") {
			d.log.Warningf("line \"%s\" does not match parseable ABP syntax subset", hostLine)
			continue
		}

		hostLine = strings.TrimPrefix(hostLine, "||")
		hostLine = strings.TrimSuffix(hostLine, "^")
		d.blocklist[dns.Fqdn(hostLine)] = true
	}

	d.lastUpdated = time.Now()

	d.log.Infof("updated blocklist; blocked domains: before: %d, after: %d; last updated: before: %v, after: %v",
		numBlockedDomainsBefore, len(d.blocklist), lastUpdatedBefore, d.lastUpdated)

	return nil
}

// IsBlocklistUpdateRequired ...
func (d *BlockDomainsDeciderABPUrl) IsBlocklistUpdateRequired() bool {
	return false
}
