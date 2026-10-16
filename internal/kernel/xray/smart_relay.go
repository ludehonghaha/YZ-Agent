package xray

import (
	"strconv"
	"strings"

	"github.com/P0me1oo/YZ-Agent/internal/model"
)

// buildSmartRelayRoutingRules applies only to the entry identity. Explicit
// logical-node subscriptions still use their original route identifiers.
// Callers position these AFTER private-network guards and BEFORE relay routes.
func buildSmartRelayRoutingRules(node *model.NodeSpec) []M {
	if !node.IsRelayEntry() || node.SmartRelay == nil || !node.SmartRelay.Enabled || node.Relay.RouteID <= 0 {
		return nil
	}
	var aiTag, defaultTag string
	for _, child := range node.Relay.Children {
		switch child.NodeID {
		case node.SmartRelay.AINodeID:
			aiTag = child.Tag
		case node.SmartRelay.DefaultNodeID:
			defaultTag = child.Tag
		}
	}
	if aiTag == "" || defaultTag == "" || aiTag == defaultTag {
		return nil
	}
	domains := make([]string, 0, len(node.SmartRelay.Domains))
	for _, domain := range node.SmartRelay.Domains {
		domain = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(domain)), "*.")
		if domain != "" && !strings.ContainsAny(domain, "/: ") {
			domains = append(domains, "domain:"+domain)
		}
	}
	if len(domains) == 0 {
		return nil
	}
	id := strconv.Itoa(node.Relay.RouteID)
	return []M{
		{"type": "field", "vlessRoute": id, "domain": domains, "outboundTag": aiTag},
		{"type": "field", "vlessRoute": id, "outboundTag": defaultTag},
	}
}
