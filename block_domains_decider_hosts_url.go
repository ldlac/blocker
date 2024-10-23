package blocker

import (
	"bufio"
	"net/http"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type BlockDomainsDeciderHostsUrl struct {
	blocklist    map[string]bool
	blocklistUrl string
	lastUpdated  time.Time
	log          Logger
}

// Name ...
func NewBlockDomainsDeciderHostsUrl(url string, logger Logger) BlockDomainsDecider {
	return &BlockDomainsDeciderHostsUrl{
		blocklistUrl: url,
		log:          logger,
		blocklist:    map[string]bool{},
	}
}

// IsDomainBlocked ...
func (d *BlockDomainsDeciderHostsUrl) IsDomainBlocked(domain string) bool {
	return d.blocklist[domain]
}

// StartBlocklistUpdater ...
func (d *BlockDomainsDeciderHostsUrl) StartBlocklistUpdater(ticker *time.Ticker) {
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
func (d *BlockDomainsDeciderHostsUrl) UpdateBlocklist() error {
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
		comps := strings.Split(hostLine, " ")
		if len(comps) < 2 {
			// Bad line in the input file
			d.log.Warningf("unformatted line present in the input file: %s", hostLine)
			continue
		}

		domain := comps[1]
		d.blocklist[dns.Fqdn(domain)] = true
	}

	d.lastUpdated = time.Now()

	d.log.Infof("updated blocklist; blocked domains: before: %d, after: %d; last updated: before: %v, after: %v",
		numBlockedDomainsBefore, len(d.blocklist), lastUpdatedBefore, d.lastUpdated)

	return nil
}

// IsBlocklistUpdateRequired ...
func (d *BlockDomainsDeciderHostsUrl) IsBlocklistUpdateRequired() bool {
	return false
}
