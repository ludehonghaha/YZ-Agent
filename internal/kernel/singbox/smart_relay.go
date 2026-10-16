package singbox

import (
	"strings"

	"github.com/P0me1oo/YZ-Agent/internal/model"
)

// buildSmartRelayRoutingRules matches only credentials for the entry itself.
// A dedicated logical-node credential is never redirected by this feature.
func buildSmartRelayRoutingRules(node *model.NodeSpec, users []model.UserSpec) []M {
	if node.SmartRelay == nil || !node.SmartRelay.Enabled {
		return nil
	}
	if node.SmartRelay.Mode == "advanced" {
		return buildAdvancedSmartRules(node, users)
	}
	if !node.IsRelayEntry() || node.Relay.RouteID <= 0 {
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

// buildAdvancedSmartRules applies an ordered selection to this entry only.
// In a relay entry, child logical identities retain their own independent exits.
// These rules sit after private address blocking and before relay identity rules.
func buildAdvancedSmartRules(node *model.NodeSpec, users []model.UserSpec) []M {
	profile := node.SmartRelay
	if profile == nil || profile.DefaultTarget == "" || node.IsRelayLanding() {
		return nil
	}
	allowed := map[string]bool{"direct": true}
	for _, outbound := range node.CustomOutbounds {
		if strings.HasPrefix(outbound.Tag, "sr-") {
			allowed[outbound.Tag] = true
		}
	}
	if node.IsRelayEntry() {
		for _, child := range node.Relay.Children {
			allowed[child.Tag] = true
		}
	}
	if !allowed[profile.DefaultTarget] {
		return nil
	}
	for _, rule := range profile.Rules {
		if !allowed[rule.Target] || len(rule.Values) == 0 {
			return nil
		}
		if rule.Kind != "domain_suffix" && rule.Kind != "ip_cidr" {
			return nil
		}
	}
	// No auth_user filter is necessary for ordinary inbound nodes, which run
	// in a dedicated core. For relay entries match *only* the entry identity.
	base := M{}
	if node.IsRelayEntry() {
		names := make([]string, 0, len(users))
		for _, user := range users {
			if user.AllowsRelayRoute(node.Relay.RouteID) {
				names = append(names, relayUserName(user, node.Relay.RouteID))
			}
		}
		if len(names) == 0 {
			return nil
		}
		base["auth_user"] = names
	} else {
		base["inbound"] = []string{node.Protocol + "-in"}
	}

	rules := make([]M, 0, len(profile.Rules)+1)
	for _, item := range profile.Rules {
		route := M{"action": "route", "outbound": item.Target}
		for key, value := range base {
			route[key] = value
		}
		values := make([]string, 0, len(item.Values))
		for _, v := range item.Values {
			v = strings.TrimSpace(strings.ToLower(v))
			if v == "" {
				continue
			}
			if item.Kind == "domain_suffix" {
				v = strings.TrimPrefix(v, "*.")
			}
			values = append(values, v)
		}
		if len(values) == 0 {
			return nil
		}
		if item.Kind == "domain_suffix" {
			route["domain_suffix"] = values
		} else {
			route["ip_cidr"] = values
		}
		rules = append(rules, route)
	}
	// A direct default must fall through to the existing DNS-resolution and
	// private-IP guard. A catch-all direct route here would bypass that guard.
	if profile.DefaultTarget == "direct" {
		return rules
	}
	final := M{"action": "route", "outbound": profile.DefaultTarget}
	for key, value := range base {
		final[key] = value
	}
	return append(rules, final)
}
