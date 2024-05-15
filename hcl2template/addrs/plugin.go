// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package addrs

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/idna"
)

// Plugin encapsulates a single plugin type.
type Plugin struct {
	Source string
}

// Parts returns the list of components of the source URL, starting with the
// host, and ending with the name of the plugin.
//
// This will correspond more or less to the filesystem hierarchy where
// the plugin is installed.
func (p Plugin) Parts() []string {
	return strings.FieldsFunc(p.Source, func(r rune) bool {
		return r == '/'
	})
}

// Name returns the raw name of the plugin from its source
//
// Exemples:
//   - "github.com/hashicorp/amazon" -> "amazon"
func (p Plugin) Name() string {
	parts := p.Parts()
	return parts[len(parts)-1]
}

func (p Plugin) String() string {
	return p.Source
}

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
//
//	namespace/name
//	hostname/namespace/name
func ParsePluginSourceString(str string) (*Plugin, error) {
	var errs []string

	if strings.HasPrefix(str, "/") {
		errs = append(errs, "A source URL must not start with a '/' character.")
	}

	if strings.HasSuffix(str, "/") {
		errs = append(errs, "A source URL must not end with a '/' character.")
	}

	if strings.Count(str, "/") < 2 {
		errs = append(errs, "A source URL must at least contain a host and a path with 2 components")
	}

	url, err := url.Parse(str)
	if err != nil {
		errs = append(errs, fmt.Sprintf("Failed to parse source URL: %s", err))
	}

	if url != nil && url.Scheme != "" {
		errs = append(errs, "A source URL must not contain a scheme (e.g. https://).")
	}

	if url != nil && url.RawQuery != "" {
		errs = append(errs, "A source URL must not contain a query (e.g. ?var=val)")
	}

	if url != nil && url.Fragment != "" {
		errs = append(errs, "A source URL must not contain a fragment (e.g. #anchor).")
	}

	if errs != nil {
		errsMsg := &strings.Builder{}
		for _, err := range errs {
			fmt.Fprintf(errsMsg, "* %s\n", err)
		}

		return nil, fmt.Errorf("The provided source URL is invalid.\nThe following errors have been discovered:\n%s\nA valid source looks like \"github.com/hashicorp/happycloud\"", errsMsg)
	}

	// check the 'name' portion, which is always the last part
	_, givenName := path.Split(str)
	_, err = ParsePluginPart(givenName)
	if err != nil {
		return nil, fmt.Errorf(`Invalid plugin type %q in source: %s"`, givenName, err)
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
	if strings.HasPrefix(givenName, redundantPrefix) {
		if strings.HasPrefix(givenName, userErrorPrefix) {
			// Likely user error. We only return this specialized error if
			// whatever is after the prefix would otherwise be a
			// syntactically-valid plugin type, so we don't end up advising
			// the user to try something that would be invalid for another
			// reason anyway.
			// (This is mainly just for robustness, because the validation
			// we already did above should've rejected most/all ways for
			// the suggestedType to end up invalid here.)
			suggestedType := strings.Replace(givenName, userErrorPrefix, "", -1)
			if _, err := ParsePluginPart(suggestedType); err == nil {
				return nil, fmt.Errorf("Plugin source has a type with the prefix %q, which isn't valid.\n"+
					"Although that prefix is often used in the names of version control repositories "+
					"for Packer plugins, plugin source strings should not include it.\n"+
					"\nDid you mean %q?", userErrorPrefix, suggestedType)
			}
		}
		// Otherwise, probably instead an incorrectly-named plugin, perhaps
		// arising from a similar instinct to what causes there to be
		// thousands of Python packages on PyPI with "python-"-prefixed
		// names.
		return nil, fmt.Errorf("Plugin source has a type with the %q prefix, which isn't valid.\n"+
			"If you are the author of this plugin, rename it to not include the prefix.\n"+
			"Ex: %q",
			redundantPrefix,
			strings.Replace(givenName, redundantPrefix, "", 1))
	}

	plug := &Plugin{
		Source: str,
	}
	if len(plug.Parts()) > 16 {
		return nil, fmt.Errorf("The source URL must have at most 16 components, and the one provided has %d.\n"+
			"This is unsupported by Packer, please consider using a source that has less components to it.\n"+
			"If this is a blocking issue for you, please open an issue to ask for supporting more "+
			"components to the source URI.",
			len(plug.Parts()))
	}

	return plug, nil
}
