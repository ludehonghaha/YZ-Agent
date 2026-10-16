package singbox

import (
	"strings"

	"github.com/P0me1oo/YZ-Agent/internal/model"
)

// buildSmartRelayRoutingRules matches only credentials for the entry itself.
// A dedicated logical-node credential is never redirected by this feature.
func buildSmartRelayRoutingRules(node *model.NodeSpec, users []model.UserSpec) []M {
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
			domains = append(domains, domain)
		}
	}
	if len(domains) == 0 {
		return nil
	}
	names := make([]string, 0, len(users))
	for _, user := range users {
		if user.AllowsRelayRoute(node.Relay.RouteID) {
			names = append(names, relayUserName(user, node.Relay.RouteID))
		}
	}
	if len(names) == 0 {
		return nil
	}
	return []M{
		{"auth_user": names, "domain_suffix": domains, "action": "route", "outbound": aiTag},
		{"auth_user": names, "action": "route", "outbound": defaultTag},
	}
}
