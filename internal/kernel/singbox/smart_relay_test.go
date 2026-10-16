package singbox

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
		SmartRelay: &panel.SmartRelayConfig{Enabled: true, AINodeID: 7, DefaultNodeID: 8, Domains: []string{"openai.com"}},
	}
	users := []model.UserSpec{{ID: 1, UUID: "4b8d1832-6839-45cb-882b-2e996d4b137e"}}
	rules := buildSmartRelayRoutingRules(node, users)
	if len(rules) != 2 {
		t.Fatalf("got %d rules", len(rules))
	}
	if rules[0]["outbound"] != "relay-7" || rules[1]["outbound"] != "relay-8" {
		t.Fatalf("unexpected rules: %#v", rules)
	}
	if len(rules[0]["auth_user"].([]string)) != 1 || len(rules[1]["auth_user"].([]string)) != 1 {
		t.Fatal("entry identity missing")
	}
	node.SmartRelay.AINodeID = 99
	if got := buildSmartRelayRoutingRules(node, users); len(got) != 0 {
		t.Fatalf("missing landing must disable rules: %#v", got)
	}
}
