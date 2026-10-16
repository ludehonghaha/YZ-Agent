package xray

import (
	"testing"

	"github.com/P0me1oo/YZ-Agent/internal/model"
	"github.com/P0me1oo/YZ-Agent/internal/panel"
)

func TestSmartRelayRules(t *testing.T) {
	node := &model.NodeSpec{
		Relay: &model.RelayConfig{Mode: "entry", RouteID: 11, Children: []model.RelayChild{
			{NodeID: 7, Tag: "relay-7", RouteID: 12},
			{NodeID: 8, Tag: "relay-8", RouteID: 13},
		}},
		SmartRelay: &panel.SmartRelayConfig{Enabled: true, AINodeID: 7, DefaultNodeID: 8, Domains: []string{"openai.com", "*.claude.ai"}},
	}
	rules := buildSmartRelayRoutingRules(node)
	if len(rules) != 2 {
		t.Fatalf("want 2 rules got %d", len(rules))
	}
	if rules[0]["outboundTag"] != "relay-7" || rules[1]["outboundTag"] != "relay-8" {
		t.Fatalf("unexpected outbounds: %#v", rules)
	}
	if rules[0]["vlessRoute"] != "11" || rules[1]["vlessRoute"] != "11" {
		t.Fatalf("smart rules should target entry only: %#v", rules)
	}
	node.SmartRelay.DefaultNodeID = 99
	if got := buildSmartRelayRoutingRules(node); len(got) != 0 {
		t.Fatalf("missing landing must disable rules: %#v", got)
	}
}
