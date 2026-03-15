package risk

import (
	"net"
	"strings"
	"sync"
)

// ProxyDetector identifies proxy, VPN, Tor, and datacenter traffic.
type ProxyDetector struct {
	mu sync.RWMutex
	// Known Tor exit node IPs (in production, periodically updated from Tor project)
	torExitNodes map[string]bool
	// Known residential proxy providers
	proxyASNs map[string]bool
	// VPN provider IP ranges
	vpnCIDRs []*net.IPNet
	// Additional datacenter ranges beyond the base engine list
	extraDCCIDRs []*net.IPNet
}

func NewProxyDetector() *ProxyDetector {
	d := &ProxyDetector{
		torExitNodes: make(map[string]bool),
		proxyASNs:    make(map[string]bool),
	}

	// Seed some known Tor exit nodes (in production, pull from https://check.torproject.org/torbulkexitlist)
	torIPs := []string{"185.220.100.252", "185.220.100.253", "185.220.101.1", "185.220.101.2"}
	for _, ip := range torIPs {
		d.torExitNodes[ip] = true
	}

	// Known residential proxy ASNs
	proxyASNs := []string{"AS9009", "AS202425", "AS44477", "AS62904"}
	for _, asn := range proxyASNs {
		d.proxyASNs[asn] = true
	}

	// Common VPN provider CIDRs (simplified)
	vpnCIDRs := []string{
		"198.18.0.0/15",    // Benchmark testing
		"100.64.0.0/10",    // CGNAT (often used by VPN)
	}
	for _, cidr := range vpnCIDRs {
		if _, ipnet, err := net.ParseCIDR(cidr); err == nil {
			d.vpnCIDRs = append(d.vpnCIDRs, ipnet)
		}
	}

	// Extra datacenter ranges
	extraDC := []string{
		"159.89.0.0/16",    // DigitalOcean
		"167.99.0.0/16",    // DigitalOcean
		"139.59.0.0/16",    // DigitalOcean
		"128.199.0.0/16",   // DigitalOcean
		"45.55.0.0/16",     // DigitalOcean
		"157.245.0.0/16",   // DigitalOcean
		"5.101.0.0/16",     // Hetzner
		"95.216.0.0/15",    // Hetzner
		"65.108.0.0/15",    // Hetzner
		"141.94.0.0/15",    // OVH
		"51.77.0.0/16",     // OVH
		"164.90.0.0/16",    // DigitalOcean
	}
	for _, cidr := range extraDC {
		if _, ipnet, err := net.ParseCIDR(cidr); err == nil {
			d.extraDCCIDRs = append(d.extraDCCIDRs, ipnet)
		}
	}

	return d
}

type ProxyResult struct {
	Score    float64
	Labels   []string
	IsTor    bool
	IsVPN    bool
	IsProxy  bool
	IsDC     bool
}

func (d *ProxyDetector) Check(ipStr string) ProxyResult {
	result := ProxyResult{}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return result
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	// Tor exit node
	if d.torExitNodes[ipStr] {
		result.IsTor = true
		result.Score += 25
		result.Labels = append(result.Labels, "tor_exit_node")
	}

	// VPN CIDR
	for _, cidr := range d.vpnCIDRs {
		if cidr.Contains(ip) {
			result.IsVPN = true
			result.Score += 15
			result.Labels = append(result.Labels, "vpn_range")
			break
		}
	}

	// Extra datacenter
	for _, cidr := range d.extraDCCIDRs {
		if cidr.Contains(ip) {
			result.IsDC = true
			result.Score += 12
			result.Labels = append(result.Labels, "datacenter_ip_extended")
			break
		}
	}

	return result
}

// BotPatternDetector identifies advanced bot patterns.
type BotPatternDetector struct{}

func NewBotPatternDetector() *BotPatternDetector {
	return &BotPatternDetector{}
}

type BotPatternResult struct {
	Score  float64
	Labels []string
	IsBot  bool
}

func (d *BotPatternDetector) Analyze(ua string, behaviorData map[string]interface{}) BotPatternResult {
	result := BotPatternResult{}

	// 1. Headless browser detection
	lower := strings.ToLower(ua)
	headlessIndicators := []string{"headlesschrome", "headless", "phantomjs", "slimerjs"}
	for _, h := range headlessIndicators {
		if strings.Contains(lower, h) {
			result.Score += 20
			result.Labels = append(result.Labels, "headless_browser")
			result.IsBot = true
			break
		}
	}

	// 2. WebDriver detection (from behavior data)
	if webdriver, ok := behaviorData["webdriver"].(bool); ok && webdriver {
		result.Score += 25
		result.Labels = append(result.Labels, "webdriver_detected")
		result.IsBot = true
	}

	// 3. Automation framework signatures
	automationKeys := []string{"__selenium_unwrapped", "__webdriver_evaluate", "__fxdriver_evaluate", "_phantom", "callPhantom", "__nightmare"}
	for _, key := range automationKeys {
		if _, ok := behaviorData[key]; ok {
			result.Score += 20
			result.Labels = append(result.Labels, "automation_framework")
			result.IsBot = true
			break
		}
	}

	// 4. Impossible timing patterns
	if duration, ok := getFloat(behaviorData, "duration_ms"); ok {
		if mouseCount, ok2 := getFloat(behaviorData, "mouse_moves"); ok2 {
			if duration > 0 && mouseCount > 0 {
				// Mouse events per second
				eventsPerSec := mouseCount / (duration / 1000)
				if eventsPerSec > 100 {
					// Humanly impossible mouse event rate
					result.Score += 15
					result.Labels = append(result.Labels, "impossible_mouse_rate")
					result.IsBot = true
				}
			}
		}
	}

	// 5. Perfect linear mouse movement (zero jitter)
	if jitter, ok := getFloat(behaviorData, "mouse_jitter"); ok {
		if jitter == 0 {
			result.Score += 10
			result.Labels = append(result.Labels, "zero_jitter")
		}
	}

	// 6. Missing browser APIs that real browsers have
	if plugins, ok := getFloat(behaviorData, "plugin_count"); ok {
		if plugins == 0 {
			result.Score += 5
			result.Labels = append(result.Labels, "no_plugins")
		}
	}

	return result
}
