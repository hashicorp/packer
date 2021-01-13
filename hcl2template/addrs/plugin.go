package addrs

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"golang.org/x/net/idna"
)

// Plugin encapsulates a single plugin type.
type Plugin struct {
	Type      string
	Namespace string
	Hostname  string
}

func (p Plugin) RealRelativePath() string {
	ns := DefaultPluginNamespace
	if p.Namespace != "" {
		ns = p.Namespace
	}
	return ns + "/packer-plugin-" + p.Type
}

func (p Plugin) Parts() []string {
	return []string{p.Hostname, p.Namespace, p.Type}
}

func (p Plugin) String() string {
	return strings.Join(p.Parts(), "/")
}

// ForDisplay returns a user-friendly FQN string, simplified for readability. If
// the plugin is using the default hostname, the hostname is omitted.
func (p *Plugin) ForDisplay() string {
	parts := []string{}
	if p.Hostname != DefaultPluginHost {
		parts = append(parts, p.Hostname)
	}
	if p.Namespace != DefaultPluginNamespace {
		parts = append(parts, p.Namespace)
	}
	parts = append(parts, p.Type)
	return strings.Join(parts, "/")
}

const DefaultPluginHost = "github.com"
const DefaultPluginNamespace = "hashicorp"

// ParsePluginPart processes an addrs.Plugin namespace or type string
// provided by an end-user, producing a normalized version if possible or
// an error if the string contains invalid characters.
//
// A plugin part is processed in the same way as an individual label in a DNS
// domain name: it is transformed to lowercase per the usual DNS case mapping
// and normalization rules and may contain only letters, digits, and dashes.
// Additionally, dashes may not appear at the start or end of the string.
//
// These restrictions are intended to allow these names to appear in fussy
// contexts such as directory/file names on case-insensitive filesystems,
// repository names on GitHub, etc. We're using the DNS rules in particular,
// rather than some similar rules defined locally, because the hostname part
// of an addrs.Plugin is already a hostname and it's ideal to use exactly
// the same case folding and normalization rules for all of the parts.
//
// It's valid to pass the result of this function as the argument to a
// subsequent call, in which case the result will be identical.
func ParsePluginPart(given string) (string, error) {
	if len(given) == 0 {
		return "", fmt.Errorf("must have at least one character")
	}

	// We're going to process the given name using the same "IDNA" library we
	// use for the hostname portion, since it already implements the case
	// folding rules we want.
	//
	// The idna library doesn't expose individual label parsing directly, but
	// once we've verified it doesn't contain any dots we can just treat it
	// like a top-level domain for this library's purposes.
	if strings.ContainsRune(given, '.') {
		return "", fmt.Errorf("dots are not allowed")
	}

	// We don't allow names containing multiple consecutive dashes, just as
	// a matter of preference: they look confusing, or incorrect.
	// This also, as a side-effect, prevents the use of the "punycode"
	// indicator prefix "xn--" that would cause the IDNA library to interpret
	// the given name as punycode, because that would be weird and unexpected.
	if strings.Contains(given, "--") {
		return "", fmt.Errorf("cannot use multiple consecutive dashes")
	}

	result, err := idna.Lookup.ToUnicode(given)
	if err != nil {
		return "", fmt.Errorf("must contain only letters, digits, and dashes, and may not use leading or trailing dashes: %w", err)
	}

	return result, nil
}

// IsPluginPartNormalized compares a given string to the result of ParsePluginPart(string)
func IsPluginPartNormalized(str string) (bool, error) {
	normalized, err := ParsePluginPart(str)
	if err != nil {
		return false, err
	}
	if str == normalized {
		return true, nil
	}
	return false, nil
}

// ParsePluginSourceString parses the source attribute and returns a plugin.
// This is intended primarily to parse the FQN-like strings
//
// The following are valid source string formats:
// 		name
// 		namespace/name
// 		hostname/namespace/name
func ParsePluginSourceString(str string) (*Plugin, hcl.Diagnostics) {
	ret := &Plugin{
		Hostname:  DefaultPluginHost,
		Namespace: DefaultPluginNamespace,
	}
	var diags hcl.Diagnostics

	// split the source string into individual components
	parts := strings.Split(str, "/")
	if len(parts) == 0 || len(parts) > 3 {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid plugin source string",
			Detail:   `The "source" attribute must be in the format "[hostname/][namespace/]name"`,
		})
		return nil, diags
	}

	// check for an invalid empty string in any part
	for i := range parts {
		if parts[i] == "" {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid plugin source string",
				Detail:   `The "source" attribute must be in the format "[hostname/][namespace/]name"`,
			})
			return nil, diags
		}
	}

	// check the 'name' portion, which is always the last part
	givenName := parts[len(parts)-1]
	name, err := ParsePluginPart(givenName)
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid plugin type",
			Detail:   fmt.Sprintf(`Invalid plugin type %q in source %q: %s"`, givenName, str, err),
		})
		return nil, diags
	}
	ret.Type = name

	if len(parts) == 1 {
		return ret, diags
	}

	if len(parts) >= 2 {
		// the namespace is always the second-to-last part
		givenNamespace := parts[len(parts)-2]
		namespace, err := ParsePluginPart(givenNamespace)
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid plugin namespace",
				Detail:   fmt.Sprintf(`Invalid plugin namespace %q in source %q: %s"`, namespace, str, err),
			})
			return nil, diags
		}
		ret.Namespace = namespace
	}

	// Final Case: 3 parts
	if len(parts) == 3 {
		// the hostname is always the first part in a three-part source string
		hostname := parts[0]
		// TODO(azr): validate host ? Can this be something else than a
		// github.com host for now?
		ret.Hostname = hostname
	}

	// Due to how plugin executables are named and plugin git repositories
	// are conventionally named, it's a reasonable and
	// apparently-somewhat-common user error to incorrectly use the
	// "packer-plugin-" prefix in a plugin source address. There is
	// no good reason for a plugin to have the prefix "packer-" anyway,
	// so we've made that invalid from the start both so we can give feedback
	// to plugin developers about the packer- prefix being redundant
	// and give specialized feedback to folks who incorrectly use the full
	// packer-plugin- prefix to help them self-correct.
	const redundantPrefix = "packer-"
	const userErrorPrefix = "packer-plugin-"
	if strings.HasPrefix(ret.Type, redundantPrefix) {
		if strings.HasPrefix(ret.Type, userErrorPrefix) {
			// Likely user error. We only return this specialized error if
			// whatever is after the prefix would otherwise be a
			// syntactically-valid plugin type, so we don't end up advising
			// the user to try something that would be invalid for another
			// reason anyway.
			// (This is mainly just for robustness, because the validation
			// we already did above should've rejected most/all ways for
			// the suggestedType to end up invalid here.)
			suggestedType := ret.Type[len(userErrorPrefix):]
			if _, err := ParsePluginPart(suggestedType); err == nil {
				suggestedAddr := ret
				suggestedAddr.Type = suggestedType
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid plugin type",
					Detail:   fmt.Sprintf("Plugin source %q has a type with the prefix %q, which isn't valid. Although that prefix is often used in the names of version control repositories for Packer plugins, plugin source strings should not include it.\n\nDid you mean %q?", ret.ForDisplay(), userErrorPrefix, suggestedAddr.ForDisplay()),
				})
				return nil, diags
			}
		}
		// Otherwise, probably instead an incorrectly-named plugin, perhaps
		// arising from a similar instinct to what causes there to be
		// thousands of Python packages on PyPI with "python-"-prefixed
		// names.
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid plugin type",
			Detail:   fmt.Sprintf("Plugin source %q has a type with the prefix %q, which isn't allowed because it would be redundant to name a Packer plugin with that prefix. If you are the author of this plugin, rename it to not include the prefix.", ret, redundantPrefix),
		})
		return nil, diags
	}

	return ret, diags
}
