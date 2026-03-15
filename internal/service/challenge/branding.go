package challenge

import (
	"sync"
)

// BrandConfig allows customers to customize the challenge UI appearance.
type BrandConfig struct {
	AppID           string `json:"app_id"`
	LogoURL         string `json:"logo_url,omitempty"`
	PrimaryColor    string `json:"primary_color,omitempty"`    // hex e.g. "#4A90D9"
	BackgroundColor string `json:"background_color,omitempty"` // hex
	TextColor       string `json:"text_color,omitempty"`       // hex
	BorderRadius    int    `json:"border_radius,omitempty"`    // px
	FontFamily      string `json:"font_family,omitempty"`
	PoweredByText   string `json:"powered_by_text,omitempty"`  // override "Protected by EngageLab"
	DarkMode        bool   `json:"dark_mode,omitempty"`
	CustomCSS       string `json:"custom_css,omitempty"`       // enterprise only
}

// BrandManager manages per-app branding configurations.
type BrandManager struct {
	mu      sync.RWMutex
	configs map[string]*BrandConfig // appID -> config
}

func NewBrandManager() *BrandManager {
	return &BrandManager{
		configs: make(map[string]*BrandConfig),
	}
}

func (m *BrandManager) Set(config *BrandConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.configs[config.AppID] = config
}

func (m *BrandManager) Get(appID string) *BrandConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if cfg, ok := m.configs[appID]; ok {
		return cfg
	}
	return defaultBrand()
}

func (m *BrandManager) Delete(appID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.configs, appID)
}

func (m *BrandManager) List() []*BrandConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*BrandConfig
	for _, c := range m.configs {
		result = append(result, c)
	}
	return result
}

func defaultBrand() *BrandConfig {
	return &BrandConfig{
		PrimaryColor:    "#4A90D9",
		BackgroundColor: "#FFFFFF",
		TextColor:       "#333333",
		BorderRadius:    12,
		FontFamily:      "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
		PoweredByText:   "Protected by EngageLab CAPTCHA",
	}
}

// GenerateCSS produces a CSS string from the brand config for the challenge widget.
func (c *BrandConfig) GenerateCSS() string {
	css := `.ec-captcha-box {`
	if c.BackgroundColor != "" {
		css += `background:` + c.BackgroundColor + `;`
	}
	if c.BorderRadius > 0 {
		css += `border-radius:` + itoa(c.BorderRadius) + `px;`
	}
	if c.FontFamily != "" {
		css += `font-family:` + c.FontFamily + `;`
	}
	css += `}`

	if c.PrimaryColor != "" {
		css += `.ec-slider-thumb,.ec-target{background:` + c.PrimaryColor + `;}`
	}
	if c.TextColor != "" {
		css += `.ec-captcha-box *{color:` + c.TextColor + `;}`
	}

	if c.DarkMode {
		css += `.ec-captcha-box{background:#1a1a2e;color:#e0e0e0;}`
		css += `.ec-slider-track{background:#2a2a4a;border-color:#3a3a5a;}`
	}

	if c.CustomCSS != "" {
		css += c.CustomCSS
	}

	return css
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
