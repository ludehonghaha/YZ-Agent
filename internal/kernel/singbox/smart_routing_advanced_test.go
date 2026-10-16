package singbox

import (
	"testing"

	"github.com/P0me1oo/YZ-Agent/internal/model"
	"github.com/P0me1oo/YZ-Agent/internal/panel"
)

func TestAdvancedRoutingNormalInbound(t *testing.T) {
	node := &model.NodeSpec{
		Protocol: "mieru",
		CustomOutbounds: []model.OutboundConfig{
			{Tag: "sr-t12345", Protocol: "socks", Settings: map[string]any{
				"server": "192.0.2.19", "server_port": 1080,
			}},
		},
		SmartRelay: &panel.SmartRelayConfig{
			Mode: "advanced", Enabled: true, DefaultTarget: "sr-t12345",
			Rules: []panel.SmartRouteRule{
				{Name: "AI", Kind: "domain_suffix", Values: []string{"openai.com", "claude.ai"}, Target: "direct"},
				{Name: "CIDR", Kind: "ip_cidr", Values: []string{"198.51.100.0/24"}, Target: "sr-t12345"},
			},
		},
	}
	rules := buildSmartRelayRoutingRules(node, nil)
	if len(rules) != 3 {
		t.Fatalf("want 3 rules got %d: %#v", len(rules), rules)
	}
	if rules[0]["outbound"] != "direct" || rules[0]["domain_suffix"] == nil {
		t.Fatal(rules[0])
	}
	if rules[1]["ip_cidr"] == nil {
		t.Fatal(rules[1])
	}
	if rules[2]["outbound"] != "sr-t12345" {
		t.Fatal(rules[2])
	}
	if rules[2]["inbound"] == nil {
		t.Fatal("missing inbound filter")
	}
	node.SmartRelay.Rules[0].Target = "sr-unknown"
	if got := buildSmartRelayRoutingRules(node, nil); len(got) != 0 {
		t.Fatal("unknown target must fail closed")
	}
}

func TestAdvancedRoutingDirectAsDefault(t *testing.T) {
	node := &model.NodeSpec{Protocol: "vless", SmartRelay: &panel.SmartRelayConfig{
		Mode: "advanced", Enabled: true, DefaultTarget: "direct",
	}}
	rules := buildSmartRelayRoutingRules(node, nil)
	// Direct default is handled by the existing post-resolve guard/final route.
	if len(rules) != 0 {
		t.Fatalf("direct fallback must preserve the private-IP guard: %#v", rules)
	}
	node.SmartRelay.Enabled = false
	if len(buildSmartRelayRoutingRules(node, nil)) != 0 {
		t.Fatal("disabled must emit no rules")
	}
}
