package common

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

/** low-level parsing */
// strip the comments and extraneous newlines from a byte channel
func uncomment(in <-chan byte) chan byte {
	out := make(chan byte)

	go func(in <-chan byte, out chan byte) {
		var endofline bool

		for {
			by, ok := <-in
			if !ok {
				break
			}

			// If we find a comment, then everything until the end of line
			// needs to be culled. We keep track of that using the `endofline`
			// flag.
			if by == '#' {
				endofline = true

			} else if by == '\n' && endofline {
				endofline = false
			}

			// If we're not in the processing of culling bytes, then write what
			// we've read into our output chan.
			if !endofline {
				out <- by
			}
		}
		close(out)
	}(in, out)
	return out
}

// convert a byte channel into a channel of pseudo-tokens
func tokenizeDhcpConfig(in chan byte) chan string {
	var state string
	var quote bool

	out := make(chan string)
	go func(out chan string) {
		for {
			by, ok := <-in
			if !ok {
				break
			}

			// If we're in a quote, then we continue until we're not in a quote
			// before we start looing for tokens
			if quote {
				if by == '"' {
					out <- state + string(by)
					state, quote = "", false
					continue
				}
				state += string(by)
				continue
			}

			switch by {
			case '"':
				// Otherwise we're outside any quotes and can process bytes normaly
				quote = true
				state += string(by)
				continue

			case '\r':
				fallthrough
			case '\n':
				fallthrough
			case '\t':
				fallthrough
			case ' ':
				// Whitespace is a separator, so we check to see if there's any state.
				// If so, then write our state prior to resetting.

				if len(state) == 0 {
					continue
				}
				out <- state
				state = ""

			case '{':
				fallthrough
			case '}':
				fallthrough
			case ';':
				// If we encounter a brace or a semicolon, then we need to emit our
				// state and then the byte because it can be part of the token.

				if len(state) > 0 {
					out <- state
				}
				out <- string(by)
				state = ""

			default:
				// Just a byte which needs to be aggregated into our state
				state += string(by)
			}
		}

		// If we still have any data left, then make sure to emit that
		if len(state) > 0 {
			out <- state
		}

		// Close our channel since we're responsible for it.
		close(out)
	}(out)
	return out
}

/** mid-level parsing */
type tkParameter struct {
	name    string
	operand []string
}

func (e *tkParameter) String() string {
	var values []string
	for _, val := range e.operand {
		values = append(values, val)
	}
	return fmt.Sprintf("%s [%s]", e.name, strings.Join(values, ","))
}

type tkGroup struct {
	parent *tkGroup
	id     tkParameter

	groups []*tkGroup
	params []tkParameter
}

func (e *tkGroup) String() string {
	var id []string

	id = append(id, e.id.name)
	for _, val := range e.id.operand {
		id = append(id, val)
	}

	var config []string
	for _, val := range e.params {
		config = append(config, val.String())
	}
	return fmt.Sprintf("%s {\n%s\n}", strings.Join(id, " "), strings.Join(config, "\n"))
}

// convert a channel of pseudo-tokens into an tkParameter struct
func parseTokenParameter(in chan string) tkParameter {
	var result tkParameter
	for {
		token, ok := <-in
		if !ok {
			break
		}

		// If there's no name for this parameter yet, then the first token
		// is our name. Snag it into our struct, and grab the next one.
		if result.name == "" {
			result.name = token
			continue
		}

		// If encounter any braces or line-terminators, then we're done parsing.
		// Anything else we find are just operands we need to keep track of.
		if strings.ContainsAny("{};", token) {
			break
		} else {
			result.operand = append(result.operand, token)
		}
	}
	return result
}

// convert a channel of pseudo-tokens into an tkGroup tree */
func parseDhcpConfig(in chan string) (tkGroup, error) {
	var tokens []string
	var result tkGroup

	// This utility function takes a list of tokens and line-terminates them
	// before sending them to parseTokenParameter().
	toParameter := func(tokens []string) tkParameter {
		out := make(chan string)
		go func(out chan string) {
			for _, v := range tokens {
				out <- v
			}
			out <- ";"
			close(out)
		}(out)
		return parseTokenParameter(out)
	}

	// Start building our tree using result as our root node
	node := &result
	for {
		tk, ok := <-in
		if !ok {
			break
		}

		switch tk {
		case "{":
			// If our next token is an opening brace, then we need to collect our
			// current aggregated tokens to parse, push our current node onto the
			// tree, and then pivot into it. Then we can reset our tokens for the child.

			grp := &tkGroup{parent: node}
			grp.id = toParameter(tokens)

			node.groups = append(node.groups, grp)
			node = grp

			tokens = []string{}

		case "}":
			// Otherwise if it's a closing brace, then we need to pop back up to
			// the parent node and resume parsing. If we have any tokens, then
			// that was because they were unterminated. Raise an error in that case.

			if node.parent == nil {
				return tkGroup{}, fmt.Errorf("Refusing to close the global declaration")
			}
			if len(tokens) > 0 {
				return tkGroup{}, fmt.Errorf("List of tokens was left unterminated : %v", tokens)
			}
			node = node.parent

			tokens = []string{}

		case ";":
			// If we encounter a line-terminator, then the list of tokens we've been
			// aggregating are ready to be parsed. Afterwards, we can write them
			// to our current tree node.

			arg := toParameter(tokens)
			node.params = append(node.params, arg)
			tokens = []string{}

		default:
			// Anything else requires us to aggregate our token into our list, and
			// try grabbing the next one.

			tokens = append(tokens, tk)
		}
	}
	return result, nil
}

func tokenizeNetworkMapConfig(in chan byte) chan string {
	var state string
	var quote bool
	var lastnewline bool

	// This logic is very similar to tokenizeDhcpConfig except she needs to handle
	// braces, and we don't. This is the only major difference from us.

	out := make(chan string)
	go func(out chan string) {
		for {
			by, ok := <-in
			if !ok {
				break
			}

			// If we're currently inide a quote, then we need to continue until
			// we encounter the closing quote. We'll keep collecting our state
			// in the meantime.
			if quote {
				if by == '"' {
					out <- state + string(by)
					state, quote = "", false
					continue
				}
				state += string(by)
				continue
			}

			switch by {
			case '"':
				// If we encounter a quote, then we need to transition into our
				// quote-parsing state that keeps collecting data until the closing
				// quote is encountered.

				quote = true
				state += string(by)
				continue

			case '\r':
				fallthrough
			case '\t':
				fallthrough
			case ' ':
				// Whitespace is considered a separator, so if we encounter this
				// then we can write our current state, and then reset.

				if len(state) == 0 {
					continue
				}
				out <- state
				state = ""

			case '\n':
				// Newlines are a somewhat special case because they separate each
				// attribute/line-item, and they can repeat. We need to preserve
				// this token, so we write our current state, then the newline.
				// We also maintain a flag so that we can consolidate multiple
				// newlines together.

				if lastnewline {
					continue
				}
				if len(state) > 0 {
					out <- state
				}
				out <- string(by)
				state = ""
				lastnewline = true
				continue

			case '.':
				fallthrough
			case '=':
				// These characters separate attributes or tokens from one another,
				// so they result in writing the state, the character, and then reset.

				if len(state) > 0 {
					out <- state
				}
				out <- string(by)
				state = ""

			default:
				// Any byte we couldn't parse just gets aggregated for the next pass.
				state += string(by)
			}

			// If we made it here, then we can guarantee that the we didn't just
			// process a newline. Clear this flag for the next one we find.
			lastnewline = false
		}

		// If there's anything left in our state, then the last line was just not
		// newline-terminated. This is a common occurrence, so write our current
		// state before we finish.
		if len(state) > 0 {
			out <- state
		}
		close(out)
	}(out)
	return out
}

func parseNetworkMapConfig(in chan string) (NetworkMap, error) {
	var state []string
	unsorted := make(map[string]map[string]string)

	// A network map has the following syntax "network.attribute = value". This
	// closure is responsible for using the "network" as a key into the `unsorted`
	// mapping, and then assigning the "value" into it keyed by the "attribute".
	addResult := func(network string, attribute string, value string) error {
		_, ok := unsorted[network]
		if !ok {
			unsorted[network] = make(map[string]string)
		}

		val, err := strconv.Unquote(value)
		if err != nil {
			return err
		}

		current := unsorted[network]
		current[attribute] = val
		return nil
	}

	// Loop through all of our tokens making sure to update our unsorted map.
	for {
		tk, ok := <-in
		if !ok {
			// If our token channel is closed, then check to see if we've
			// collected 3 items in our state. If so, then we can add this
			// final attribute/value before we leave.
			if len(state) == 3 {
				err := addResult(state[0], state[1], state[2])
				if err != nil {
					return nil, err
				}
			}
			break
		}

		// This switch makes sure we encounter these tokens in the correct order.
		switch tk {
		case ".":
			if len(state) != 1 {
				return nil, fmt.Errorf("Missing network index")
			}

		case "=":
			if len(state) != 2 {
				return nil, fmt.Errorf("Assignment to empty attribute")
			}

		case "\n":
			if len(state) == 0 {
				continue
			}
			if len(state) != 3 {
				return nil, fmt.Errorf("Invalid attribute assignment : %v", state)
			}
			err := addResult(state[0], state[1], state[2])
			if err != nil {
				return nil, err
			}
			state = make([]string, 0)

		default:
			state = append(state, tk)
		}
	}

	// Go through our unsorted map, and collect all of the keys for "network".
	result := make([]map[string]string, 0)
	var keys []string
	for k := range unsorted {
		keys = append(keys, k)
	}

	// This way we can sort them.
	sort.Strings(keys)

	// And then collect all of them into a list to return to the caller.
	for _, k := range keys {
		result = append(result, unsorted[k])
	}

	return result, nil
}

/** higher-level parsing */
/// parameters
type pParameter interface {
	repr() string
}

type pParameterInclude struct {
	filename string
}

func (e pParameterInclude) repr() string { return fmt.Sprintf("include-file:filename=%s", e.filename) }

type pParameterOption struct {
	name  string
	value string
}

func (e pParameterOption) repr() string { return fmt.Sprintf("option:%s=%s", e.name, e.value) }

// allow some-kind-of-something
type pParameterGrant struct {
	verb      string // allow,deny,ignore
	attribute string
}

func (e pParameterGrant) repr() string { return fmt.Sprintf("grant:%s,%s", e.verb, e.attribute) }

type pParameterAddress4 []string

func (e pParameterAddress4) repr() string {
	return fmt.Sprintf("fixed-address4:%s", strings.Join(e, ","))
}

type pParameterAddress6 []string

func (e pParameterAddress6) repr() string {
	return fmt.Sprintf("fixed-address6:%s", strings.Join(e, ","))
}

// hardware address 00:00:00:00:00:00
type pParameterHardware struct {
	class   string
	address []byte
}

func (e pParameterHardware) repr() string {
	res := make([]string, 0)

	for _, v := range e.address {
		res = append(res, fmt.Sprintf("%02x", v))
	}

	return fmt.Sprintf("hardware-address:%s[%s]", e.class, strings.Join(res, ":"))
}

type pParameterBoolean struct {
	parameter string
	truancy   bool
}

func (e pParameterBoolean) repr() string { return fmt.Sprintf("boolean:%s=%v", e.parameter, e.truancy) }

type pParameterClientMatch struct {
	name string
	data string
}

func (e pParameterClientMatch) repr() string {
	return fmt.Sprintf("match-client:%s=%s", e.name, e.data)
}

// range 127.0.0.1 127.0.0.255
type pParameterRange4 struct {
	min net.IP
	max net.IP
}

func (e pParameterRange4) repr() string {
	return fmt.Sprintf("range4:%s-%s", e.min.String(), e.max.String())
}

type pParameterRange6 struct {
	min net.IP
	max net.IP
}

func (e pParameterRange6) repr() string {
	return fmt.Sprintf("range6:%s-%s", e.min.String(), e.max.String())
}

type pParameterPrefix6 struct {
	min  net.IP
	max  net.IP
	bits int
}

func (e pParameterPrefix6) repr() string {
	return fmt.Sprintf("prefix6:/%d:%s-%s", e.bits, e.min.String(), e.max.String())
}

// some-kind-of-parameter 1024
type pParameterOther struct {
	parameter string
	value     string
}

func (e pParameterOther) repr() string { return fmt.Sprintf("parameter:%s=%s", e.parameter, e.value) }

type pParameterExpression struct {
	parameter  string
	expression string
}

func (e pParameterExpression) repr() string {
	return fmt.Sprintf("parameter-expression:%s=\"%s\"", e.parameter, e.expression)
}

type pDeclarationIdentifier interface {
	repr() string
}

type pDeclaration struct {
	id           pDeclarationIdentifier
	parent       *pDeclaration
	parameters   []pParameter
	declarations []pDeclaration
}

func (e *pDeclaration) short() string {
	return e.id.repr()
}

func (e *pDeclaration) repr() string {
	res := e.short()

	var parameters []string
	for _, v := range e.parameters {
		parameters = append(parameters, v.repr())
	}

	var groups []string
	for _, v := range e.declarations {
		groups = append(groups, fmt.Sprintf("-> %s", v.short()))
	}

	if e.parent != nil {
		res = fmt.Sprintf("%s parent:%s", res, e.parent.short())
	}

	return fmt.Sprintf("%s\n%s\n%s\n", res, strings.Join(parameters, "\n"), strings.Join(groups, "\n"))
}

type pDeclarationGlobal struct{}

func (e pDeclarationGlobal) repr() string { return fmt.Sprintf("{global}") }

type pDeclarationShared struct{ name string }

func (e pDeclarationShared) repr() string { return fmt.Sprintf("{shared-network %s}", e.name) }

type pDeclarationSubnet4 struct{ net.IPNet }

func (e pDeclarationSubnet4) repr() string { return fmt.Sprintf("{subnet4 %s}", e.String()) }

type pDeclarationSubnet6 struct{ net.IPNet }

func (e pDeclarationSubnet6) repr() string { return fmt.Sprintf("{subnet6 %s}", e.String()) }

type pDeclarationHost struct{ name string }

func (e pDeclarationHost) repr() string { return fmt.Sprintf("{host name:%s}", e.name) }

type pDeclarationPool struct{}

func (e pDeclarationPool) repr() string { return fmt.Sprintf("{pool}") }

type pDeclarationGroup struct{}

func (e pDeclarationGroup) repr() string { return fmt.Sprintf("{group}") }

type pDeclarationClass struct{ name string }

func (e pDeclarationClass) repr() string { return fmt.Sprintf("{class}") }

/** parsers */
func parseParameter(val tkParameter) (pParameter, error) {
	switch val.name {
	case "include":
		if len(val.operand) != 2 {
			return nil, fmt.Errorf("Invalid number of parameters for pParameterInclude : %v", val.operand)
		}

		name := val.operand[0]
		return pParameterInclude{filename: name}, nil

	case "option":
		if len(val.operand) != 2 {
			return nil, fmt.Errorf("Invalid number of parameters for pParameterOption : %v", val.operand)
		}

		name, value := val.operand[0], val.operand[1]
		return pParameterOption{name: name, value: value}, nil

	case "allow":
		fallthrough
	case "deny":
		fallthrough
	case "ignore":
		if len(val.operand) < 1 {
			return nil, fmt.Errorf("Invalid number of parameters for pParameterGrant : %v", val.operand)
		}

		attribute := strings.Join(val.operand, " ")
		return pParameterGrant{verb: strings.ToLower(val.name), attribute: attribute}, nil

	case "range":
		if len(val.operand) < 1 {
			return nil, fmt.Errorf("Invalid number of parameters for pParameterRange4 : %v", val.operand)
		}

		idxAddress := map[bool]int{true: 1, false: 0}[strings.ToLower(val.operand[0]) == "bootp"]
		if len(val.operand) > 2+idxAddress {
			return nil, fmt.Errorf("Invalid number of parameters for pParameterRange : %v", val.operand)
		}

		if idxAddress+1 > len(val.operand) {
			res := net.ParseIP(val.operand[idxAddress])
			return pParameterRange4{min: res, max: res}, nil
		}

		addr1 := net.ParseIP(val.operand[idxAddress])
		addr2 := net.ParseIP(val.operand[idxAddress+1])
		return pParameterRange4{min: addr1, max: addr2}, nil

	case "range6":
		if len(val.operand) == 1 {
			address := val.operand[0]
			if strings.Contains(address, "/") {
				cidr := strings.SplitN(address, "/", 2)
				if len(cidr) != 2 {
					return nil, fmt.Errorf("Unknown ipv6 format : %v", address)
				}

				address := net.ParseIP(cidr[0])

				bits, err := strconv.Atoi(cidr[1])
				if err != nil {
					return nil, err
				}
				mask := net.CIDRMask(bits, net.IPv6len*8)

				// figure out the network address
				network := address.Mask(mask)

				// make a broadcast address
				broadcast := network
				networkSize, totalSize := mask.Size()
				hostSize := totalSize - networkSize
				for i := networkSize / 8; i < totalSize/8; i++ {
					broadcast[i] = byte(0xff)
				}

				octetIndex := network[networkSize/8]
				bitsLeft := (uint32)(hostSize % 8)
				broadcast[octetIndex] = network[octetIndex] | ((1 << bitsLeft) - 1)

				// FIXME: check that the broadcast address was made correctly
				return pParameterRange6{min: network, max: broadcast}, nil
			}
			res := net.ParseIP(address)
			return pParameterRange6{min: res, max: res}, nil
		}

		if len(val.operand) == 2 {
			addr := net.ParseIP(val.operand[0])
			if strings.ToLower(val.operand[1]) == "temporary" {
				return pParameterRange6{min: addr, max: addr}, nil
			}

			other := net.ParseIP(val.operand[1])
			return pParameterRange6{min: addr, max: other}, nil
		}
		return nil, fmt.Errorf("Invalid number of parameters for pParameterRange6 : %v", val.operand)

	case "prefix6":
		if len(val.operand) != 3 {
			return nil, fmt.Errorf("Invalid number of parameters for pParameterRange6 : %v", val.operand)
		}

		bits, err := strconv.Atoi(val.operand[2])
		if err != nil {
			return nil, fmt.Errorf("Invalid bits for pParameterPrefix6 : %v", val.operand[2])
		}

		minaddr := net.ParseIP(val.operand[0])
		maxaddr := net.ParseIP(val.operand[1])
		return pParameterPrefix6{min: minaddr, max: maxaddr, bits: bits}, nil

	case "hardware":
		if len(val.operand) != 2 {
			return nil, fmt.Errorf("Invalid number of parameters for pParameterHardware : %v", val.operand)
		}

		class := val.operand[0]
		octets := strings.Split(val.operand[1], ":")
		address := make([]byte, 0)
		for _, v := range octets {
			b, err := strconv.ParseInt(v, 16, 0)
			if err != nil {
				return nil, err
			}
			address = append(address, byte(b))
		}

		return pParameterHardware{class: class, address: address}, nil

	case "fixed-address":
		ip4addrs := make(pParameterAddress4, len(val.operand))
		copy(ip4addrs, val.operand)
		return ip4addrs, nil

	case "fixed-address6":
		ip6addrs := make(pParameterAddress6, len(val.operand))
		copy(ip6addrs, val.operand)
		return ip6addrs, nil

	case "host-identifier":
		if len(val.operand) != 3 {
			return nil, fmt.Errorf("Invalid number of parameters for pParameterClientMatch : %v", val.operand)
		}

		if val.operand[0] != "option" {
			return nil, fmt.Errorf("Invalid match parameter : %v", val.operand[0])
		}

		optionName := val.operand[1]
		optionData := val.operand[2]
		return pParameterClientMatch{name: optionName, data: optionData}, nil

	default:
		length := len(val.operand)

		if length < 1 {
			return pParameterBoolean{parameter: val.name, truancy: true}, nil

		} else if length > 1 {
			if val.operand[0] == "=" {
				return pParameterExpression{parameter: val.name, expression: strings.Join(val.operand[1:], "")}, nil
			}
		}

		if length != 1 {
			return nil, fmt.Errorf("Invalid number of parameters for pParameterOther : %v", val.operand)
		}

		if strings.ToLower(val.name) == "not" {
			return pParameterBoolean{parameter: val.operand[0], truancy: false}, nil
		}

		return pParameterOther{parameter: val.name, value: val.operand[0]}, nil
	}
}

func parseTokenGroup(val tkGroup) (*pDeclaration, error) {
	params := val.id.operand

	switch val.id.name {
	case "group":
		return &pDeclaration{id: pDeclarationGroup{}}, nil

	case "pool":
		return &pDeclaration{id: pDeclarationPool{}}, nil

	case "host":
		if len(params) == 1 {
			return &pDeclaration{id: pDeclarationHost{name: params[0]}}, nil
		}

	case "subnet":
		if len(params) == 3 && strings.ToLower(params[1]) == "netmask" {
			addr := make([]byte, 4)
			for i, v := range strings.SplitN(params[2], ".", 4) {
				res, err := strconv.ParseInt(v, 10, 0)
				if err != nil {
					return nil, err
				}
				addr[i] = byte(res)
			}
			oc1, oc2, oc3, oc4 := addr[0], addr[1], addr[2], addr[3]
			if subnet, mask := net.ParseIP(params[0]), net.IPv4Mask(oc1, oc2, oc3, oc4); subnet != nil && mask != nil {
				return &pDeclaration{id: pDeclarationSubnet4{net.IPNet{IP: subnet, Mask: mask}}}, nil
			}
		}

	case "subnet6":
		if len(params) == 1 {
			ip6 := strings.SplitN(params[0], "/", 2)
			if len(ip6) == 2 && strings.Contains(ip6[0], ":") {
				address := net.ParseIP(ip6[0])
				prefix, err := strconv.Atoi(ip6[1])
				if err != nil {
					return nil, err
				}
				return &pDeclaration{id: pDeclarationSubnet6{net.IPNet{IP: address, Mask: net.CIDRMask(prefix, net.IPv6len*8)}}}, nil
			}
		}

	case "shared-network":
		if len(params) == 1 {
			return &pDeclaration{id: pDeclarationShared{name: params[0]}}, nil
		}

	case "":
		return &pDeclaration{id: pDeclarationGlobal{}}, nil
	}
	return nil, fmt.Errorf("Invalid pDeclaration : %v : %v", val.id.name, params)
}

func flattenDhcpConfig(root tkGroup) (*pDeclaration, error) {
	result, err := parseTokenGroup(root)

	if err != nil {
		return nil, err
	}

	for _, p := range root.params {
		param, err := parseParameter(p)
		if err != nil {
			return nil, err
		}
		result.parameters = append(result.parameters, param)
	}

	for _, p := range root.groups {
		group, err := flattenDhcpConfig(*p)
		if err != nil {
			return nil, err
		}
		group.parent = result
		result.declarations = append(result.declarations, *group)
	}

	return result, nil
}

/** reduce the tree into the things that we care about */
type grant uint

const (
	ALLOW  grant = iota
	IGNORE grant = iota
	DENY   grant = iota
)

type configDeclaration struct {
	id         []pDeclarationIdentifier
	composites []pDeclaration

	address []pParameter

	options     map[string]string
	grants      map[string]grant
	attributes  map[string]bool
	parameters  map[string]string
	expressions map[string]string

	hostid []pParameterClientMatch
}

func createDeclaration(node pDeclaration) configDeclaration {
	var hierarchy []pDeclaration

	for n := &node; n != nil; n = n.parent {
		hierarchy = append(hierarchy, *n)
	}

	var result configDeclaration
	result.address = make([]pParameter, 0)

	result.options = make(map[string]string)
	result.grants = make(map[string]grant)
	result.attributes = make(map[string]bool)
	result.parameters = make(map[string]string)
	result.expressions = make(map[string]string)

	result.hostid = make([]pParameterClientMatch, 0)

	// walk from globals to pDeclaration collecting all parameters
	for i := len(hierarchy) - 1; i >= 0; i-- {
		result.composites = append(result.composites, hierarchy[(len(hierarchy)-1)-i])
		result.id = append(result.id, hierarchy[(len(hierarchy)-1)-i].id)

		// update configDeclaration parameters
		for _, p := range hierarchy[i].parameters {
			switch p := p.(type) {
			case pParameterOption:
				result.options[p.name] = p.value
			case pParameterGrant:
				Grant := map[string]grant{"ignore": IGNORE, "allow": ALLOW, "deny": DENY}
				result.grants[p.attribute] = Grant[p.verb]
			case pParameterBoolean:
				result.attributes[p.parameter] = p.truancy
			case pParameterClientMatch:
				result.hostid = append(result.hostid, p)
			case pParameterExpression:
				result.expressions[p.parameter] = p.expression
			case pParameterOther:
				result.parameters[p.parameter] = p.value
			default:
				result.address = append(result.address, p)
			}
		}
	}
	return result
}

func (e *configDeclaration) repr() string {
	var result []string

	res := make([]string, 0)
	for _, v := range e.id {
		res = append(res, v.repr())
	}
	result = append(result, strings.Join(res, ","))

	if len(e.address) > 0 {
		res := make([]string, 0)
		for _, v := range e.address {
			res = append(res, v.repr())
		}
		result = append(result, fmt.Sprintf("address : %v", strings.Join(res, ",")))
	}

	if len(e.options) > 0 {
		result = append(result, fmt.Sprintf("options : %v", e.options))
	}
	if len(e.grants) > 0 {
		result = append(result, fmt.Sprintf("grants : %v", e.grants))
	}
	if len(e.attributes) > 0 {
		result = append(result, fmt.Sprintf("attributes : %v", e.attributes))
	}
	if len(e.parameters) > 0 {
		result = append(result, fmt.Sprintf("parameters : %v", e.parameters))
	}
	if len(e.expressions) > 0 {
		result = append(result, fmt.Sprintf("parameter-expressions : %v", e.expressions))
	}

	if len(e.hostid) > 0 {
		res := make([]string, 0)
		for _, v := range e.hostid {
			res = append(res, v.repr())
		}
		result = append(result, fmt.Sprintf("hostid : %v", strings.Join(res, " ")))
	}

	return strings.Join(result, "\n") + "\n"
}

func (e *configDeclaration) IP4() (net.IP, error) {
	var result []string

	for _, entry := range e.address {
		switch entry.(type) {
		case pParameterAddress4:
			for _, s := range entry.(pParameterAddress4) {
				result = append(result, s)
			}
		}
	}

	if len(result) > 1 {
		return nil, fmt.Errorf("More than one address4 returned : %v", result)

	} else if len(result) == 0 {
		return nil, fmt.Errorf("No IP4 addresses found")
	}

	// Try and parse it as an IP4. If so, then it's good to return it as-is.
	if res := net.ParseIP(result[0]); res != nil {
		return res, nil
	}

	// Otherwise make an attempt to resolve it to an address.
	res, err := net.ResolveIPAddr("ip4", result[0])
	if err != nil {
		return nil, err
	}

	return res.IP, nil
}

func (e *configDeclaration) IP6() (net.IP, error) {
	var result []string

	for _, entry := range e.address {
		switch entry.(type) {
		case pParameterAddress6:
			for _, s := range entry.(pParameterAddress6) {
				result = append(result, s)
			}
		}
	}

	if len(result) > 1 {
		return nil, fmt.Errorf("More than one address6 returned : %v", result)

	} else if len(result) == 0 {
		return nil, fmt.Errorf("No IP6 addresses found")
	}

	// If we were able to parse it into an IP, then we can just return it.
	if res := net.ParseIP(result[0]); res != nil {
		return res, nil
	}

	// Otherwise, try to resolve it into an address.
	res, err := net.ResolveIPAddr("ip6", result[0])
	if err != nil {
		return nil, err
	}
	return res.IP, nil
}

func (e *configDeclaration) Hardware() (net.HardwareAddr, error) {
	var result []pParameterHardware

	for _, addr := range e.address {
		switch addr.(type) {
		case pParameterHardware:
			result = append(result, addr.(pParameterHardware))
		}
	}

	if len(result) > 0 {
		return nil, fmt.Errorf("More than one hardware address returned : %v", result)
	}

	res := make(net.HardwareAddr, 0)
	for _, by := range result[0].address {
		res = append(res, by)
	}
	return res, nil
}

/*** Dhcp Configuration */
type DhcpConfiguration []configDeclaration

func ReadDhcpConfiguration(fd *os.File) (DhcpConfiguration, error) {
	fromfile := consumeFile(fd)
	uncommented := uncomment(fromfile)
	tokenized := tokenizeDhcpConfig(uncommented)

	// Parse the tokenized DHCP configuration into a tree. We need it as a tree
	// because DHCP declarations can inherit options from their parent
	parsetree, err := parseDhcpConfig(tokenized)
	if err != nil {
		return nil, err
	}

	// Flatten the tree into a list of pDeclaration objects. This is responsible
	// for actually propagating options from the parent pDeclaration into all of
	// its children.
	global, err := flattenDhcpConfig(parsetree)
	if err != nil {
		return nil, err
	}

	// This closure is just to the goro that follows it in recursively walking
	// through all of the declarations and writing them individually to a chan.
	var walkDeclarations func(root pDeclaration, out chan *configDeclaration)

	walkDeclarations = func(root pDeclaration, out chan *configDeclaration) {
		res := createDeclaration(root)
		out <- &res
		for _, p := range root.declarations {
			walkDeclarations(p, out)
		}
	}

	// That way this goro can take each individual declaration and write it to
	// a channel.
	each := make(chan *configDeclaration)
	go func(out chan *configDeclaration) {
		walkDeclarations(*global, out)
		out <- nil
	}(each)

	// For this loop to convert it into a itemized list.
	var result DhcpConfiguration
	for decl := <-each; decl != nil; decl = <-each {
		result = append(result, *decl)
	}
	return result, nil
}

func (e *DhcpConfiguration) Global() configDeclaration {
	result := (*e)[0]
	if len(result.id) != 1 {
		panic(fmt.Errorf("Something that can't happen happened"))
	}
	return result
}

func (e *DhcpConfiguration) SubnetByAddress(address net.IP) (configDeclaration, error) {
	var result []configDeclaration
	for _, entry := range *e {
		switch entry.id[0].(type) {
		case pDeclarationSubnet4:
			id := entry.id[0].(pDeclarationSubnet4)
			if id.Contains(address) {
				result = append(result, entry)
			}
		case pDeclarationSubnet6:
			id := entry.id[0].(pDeclarationSubnet6)
			if id.Contains(address) {
				result = append(result, entry)
			}
		}
	}
	if len(result) == 0 {
		return configDeclaration{}, fmt.Errorf("No network declarations containing %s found", address.String())
	}
	if len(result) > 1 {
		return configDeclaration{}, fmt.Errorf("More than 1 network declaration found : %v", result)
	}
	return result[0], nil
}

func (e *DhcpConfiguration) HostByName(host string) (configDeclaration, error) {
	var result []configDeclaration
	for _, entry := range *e {
		switch entry.id[0].(type) {
		case pDeclarationHost:
			id := entry.id[0].(pDeclarationHost)
			if strings.EqualFold(id.name, host) {
				result = append(result, entry)
			}
		}
	}
	if len(result) == 0 {
		return configDeclaration{}, fmt.Errorf("No host declarations containing %s found", host)
	}
	if len(result) > 1 {
		return configDeclaration{}, fmt.Errorf("More than 1 host declaration found : %v", result)
	}
	return result[0], nil
}

/*** Network Map */
type NetworkMap []map[string]string

type NetworkNameMapper interface {
	NameIntoDevices(string) ([]string, error)
	DeviceIntoName(string) (string, error)
}

func ReadNetworkMap(fd *os.File) (NetworkMap, error) {
	fromfile := consumeFile(fd)
	uncommented := uncomment(fromfile)
	tokenized := tokenizeNetworkMapConfig(uncommented)

	// Now that we've tokenized the network map, we just need to parse it into
	// a list of maps.
	result, err := parseNetworkMapConfig(tokenized)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (e NetworkMap) NameIntoDevices(name string) ([]string, error) {
	var devices []string

	for _, val := range e {
		if strings.EqualFold(val["name"], name) {
			devices = append(devices, val["device"])
		}
	}

	if len(devices) > 0 {
		return devices, nil
	}

	return make([]string, 0), fmt.Errorf("Network name not found : %v", name)
}
func (e NetworkMap) DeviceIntoName(device string) (string, error) {
	for _, val := range e {
		if strings.EqualFold(val["device"], device) {
			return val["name"], nil
		}
	}

	return "", fmt.Errorf("Device name not found : %v", device)
}
func (e *NetworkMap) repr() string {
	var result []string

	for idx, val := range *e {
		result = append(result, fmt.Sprintf("network%d.name = \"%s\"", idx, val["name"]))
		result = append(result, fmt.Sprintf("network%d.device = \"%s\"", idx, val["device"]))
	}

	return strings.Join(result, "\n")
}

/*** parser for VMware Fusion's networking file */
func tokenizeNetworkingConfig(in chan byte) chan string {
	var state string
	var repeat_newline bool

	out := make(chan string)
	go func(out chan string) {
		for {
			by, ok := <-in
			if !ok {
				break
			}

			switch by {
			case '\t':
				fallthrough
			case ' ':
				// If we receive whitespace, then this is just a write to our
				// state and then we reset.
				if len(state) == 0 {
					continue
				}
				out <- state
				state = ""

			case '\r':
				// If windows has tampered with our newlines, then we normalize
				// things by converting its value from 0x0d to 0x0a.
				fallthrough
			case '\n':
				// Newlines can repeat, so this case is responsible for writing
				// to the chan, and consolidating multiple newlines into a single.
				if repeat_newline {
					continue
				}
				if len(state) > 0 {
					out <- state
				}
				out <- "\n"
				state = ""
				repeat_newline = true
				continue

			default:
				// Anything other bytes just need to be aggregated into a string.
				state += string(by)
			}
			repeat_newline = false
		}

		// If there's anything left in our state after the chan has been closed,
		// then the input just wasn't terminated properly. It's still valid, so
		// write we have to the channel.
		if len(state) > 0 {
			out <- state
		}
		close(out)
	}(out)
	return out
}

func splitNetworkingConfig(in chan string) chan []string {
	out := make(chan []string)

	// This goroutine is simple in that it takes a chan of tokens, and splits
	// them across the newlines.

	go func(out chan []string) {
		row := make([]string, 0)
		for {
			tk, ok := <-in
			if !ok {
				break
			}

			if tk == "\n" {
				// If we received a newline token, then we need to write our
				// aggregated list of tokens and reset our "splitting" state.

				if len(row) > 0 {
					out <- row
				}

				row = make([]string, 0)

			} else {
				// Anything else just requires us to aggregate the token into
				// our list.
				row = append(row, tk)
			}
		}

		if len(row) > 0 {
			out <- row
		}
		close(out)
	}(out)
	return out
}

/// All token types in networking file.
// VERSION token
type networkingVERSION struct {
	value string
}

func networkingReadVersion(row []string) (*networkingVERSION, error) {
	if len(row) != 1 {
		return nil, fmt.Errorf("Unexpected format for VERSION entry : %v", row)
	}
	res := &networkingVERSION{value: row[0]}
	if !res.Valid() {
		return nil, fmt.Errorf("Unexpected format for VERSION entry : %v", row)
	}
	return res, nil
}

func (s networkingVERSION) Repr() string {
	if !s.Valid() {
		return fmt.Sprintf("VERSION{INVALID=\"%v\"}", s.value)
	}
	return fmt.Sprintf("VERSION{%f}", s.Number())
}

func (s networkingVERSION) Valid() bool {
	tokens := strings.SplitN(s.value, "=", 2)
	if len(tokens) != 2 || tokens[0] != "VERSION" {
		return false
	}

	tokens = strings.Split(tokens[1], ",")
	if len(tokens) != 2 {
		return false
	}

	for _, t := range tokens {
		_, err := strconv.ParseUint(t, 10, 64)
		if err != nil {
			return false
		}
	}
	return true
}

func (s networkingVERSION) Number() float64 {
	var result float64
	tokens := strings.SplitN(s.value, "=", 2)
	tokens = strings.Split(tokens[1], ",")

	integer, err := strconv.ParseUint(tokens[0], 10, 64)
	if err != nil {
		integer = 0
	}
	result = float64(integer)

	mantissa, err := strconv.ParseUint(tokens[1], 10, 64)
	if err != nil {
		return result
	}
	denomination := math.Pow(10.0, float64(len(tokens[1])))
	return result + (float64(mantissa) / denomination)
}

// VNET_X token
type networkingVNET struct {
	value string
}

func (s networkingVNET) Valid() bool {
	if strings.ToUpper(s.value) != s.value {
		return false
	}
	tokens := strings.SplitN(s.value, "_", 3)
	if len(tokens) != 3 || tokens[0] != "VNET" {
		return false
	}
	_, err := strconv.ParseUint(tokens[1], 10, 64)
	if err != nil {
		return false
	}
	return true
}

func (s networkingVNET) Number() int {
	tokens := strings.SplitN(s.value, "_", 3)
	res, err := strconv.Atoi(tokens[1])
	if err != nil {
		return ^int(0)
	}
	return res
}

func (s networkingVNET) Option() string {
	tokens := strings.SplitN(s.value, "_", 3)
	if len(tokens) == 3 {
		return tokens[2]
	}
	return ""
}

func (s networkingVNET) Repr() string {
	if !s.Valid() {
		tokens := strings.SplitN(s.value, "_", 3)
		return fmt.Sprintf("VNET{INVALID=%v}", tokens)
	}
	return fmt.Sprintf("VNET{%d} %s", s.Number(), s.Option())
}

// Interface name
type networkingInterface struct {
	name string
}

func (s networkingInterface) Interface() (*net.Interface, error) {
	return net.InterfaceByName(s.name)
}

// networking command entry types
type networkingCommandEntry_answer struct {
	vnet  networkingVNET
	value string
}
type networkingCommandEntry_remove_answer struct {
	vnet networkingVNET
}
type networkingCommandEntry_add_nat_portfwd struct {
	vnet        int
	protocol    string
	port        int
	target_host net.IP
	target_port int
}
type networkingCommandEntry_remove_nat_portfwd struct {
	vnet     int
	protocol string
	port     int
}
type networkingCommandEntry_add_dhcp_mac_to_ip struct {
	vnet int
	mac  net.HardwareAddr
	ip   net.IP
}
type networkingCommandEntry_remove_dhcp_mac_to_ip struct {
	vnet int
	mac  net.HardwareAddr
}
type networkingCommandEntry_add_bridge_mapping struct {
	intf networkingInterface
	vnet int
}
type networkingCommandEntry_remove_bridge_mapping struct {
	intf networkingInterface
}
type networkingCommandEntry_add_nat_prefix struct {
	vnet   int
	prefix int
}
type networkingCommandEntry_remove_nat_prefix struct {
	vnet   int
	prefix int
}

type networkingCommandEntry struct {
	entry                 interface{}
	answer                *networkingCommandEntry_answer
	remove_answer         *networkingCommandEntry_remove_answer
	add_nat_portfwd       *networkingCommandEntry_add_nat_portfwd
	remove_nat_portfwd    *networkingCommandEntry_remove_nat_portfwd
	add_dhcp_mac_to_ip    *networkingCommandEntry_add_dhcp_mac_to_ip
	remove_dhcp_mac_to_ip *networkingCommandEntry_remove_dhcp_mac_to_ip
	add_bridge_mapping    *networkingCommandEntry_add_bridge_mapping
	remove_bridge_mapping *networkingCommandEntry_remove_bridge_mapping
	add_nat_prefix        *networkingCommandEntry_add_nat_prefix
	remove_nat_prefix     *networkingCommandEntry_remove_nat_prefix
}

func (e networkingCommandEntry) Name() string {
	switch e.entry.(type) {
	case networkingCommandEntry_answer:
		return "answer"
	case networkingCommandEntry_remove_answer:
		return "remove_answer"
	case networkingCommandEntry_add_nat_portfwd:
		return "add_nat_portfwd"
	case networkingCommandEntry_remove_nat_portfwd:
		return "remove_nat_portfwd"
	case networkingCommandEntry_add_dhcp_mac_to_ip:
		return "add_dhcp_mac_to_ip"
	case networkingCommandEntry_remove_dhcp_mac_to_ip:
		return "remove_dhcp_mac_to_ip"
	case networkingCommandEntry_add_bridge_mapping:
		return "add_bridge_mapping"
	case networkingCommandEntry_remove_bridge_mapping:
		return "remove_bridge_mapping"
	case networkingCommandEntry_add_nat_prefix:
		return "add_nat_prefix"
	case networkingCommandEntry_remove_nat_prefix:
		return "remove_nat_prefix"
	}
	return ""
}

func (e networkingCommandEntry) Entry() reflect.Value {
	this := reflect.ValueOf(e)
	switch e.entry.(type) {
	case networkingCommandEntry_answer:
		return reflect.Indirect(this.FieldByName("answer"))
	case networkingCommandEntry_remove_answer:
		return reflect.Indirect(this.FieldByName("remove_answer"))
	case networkingCommandEntry_add_nat_portfwd:
		return reflect.Indirect(this.FieldByName("add_nat_portfwd"))
	case networkingCommandEntry_remove_nat_portfwd:
		return reflect.Indirect(this.FieldByName("remove_nat_portfwd"))
	case networkingCommandEntry_add_dhcp_mac_to_ip:
		return reflect.Indirect(this.FieldByName("add_dhcp_mac_to_ip"))
	case networkingCommandEntry_remove_dhcp_mac_to_ip:
		return reflect.Indirect(this.FieldByName("remove_dhcp_mac_to_ip"))
	case networkingCommandEntry_add_bridge_mapping:
		return reflect.Indirect(this.FieldByName("add_bridge_mapping"))
	case networkingCommandEntry_remove_bridge_mapping:
		return reflect.Indirect(this.FieldByName("remove_bridge_mapping"))
	case networkingCommandEntry_add_nat_prefix:
		return reflect.Indirect(this.FieldByName("add_nat_prefix"))
	case networkingCommandEntry_remove_nat_prefix:
		return reflect.Indirect(this.FieldByName("remove_nat_prefix"))
	}
	return reflect.Value{}
}

func (e networkingCommandEntry) Repr() string {
	result := make(map[string]interface{})

	entryN, entry := e.Name(), e.Entry()
	entryT := entry.Type()
	for i := 0; i < entry.NumField(); i++ {
		fld, fldT := entry.Field(i), entryT.Field(i)
		result[fldT.Name] = fld
	}
	return fmt.Sprintf("%s -> %v", entryN, result)
}

// networking command entry parsers
func parseNetworkingCommand_answer(row []string) (*networkingCommandEntry, error) {
	if len(row) != 2 {
		return nil, fmt.Errorf("Expected %d arguments. Received only %d.", 2, len(row))
	}

	vnet := networkingVNET{value: row[0]}
	if !vnet.Valid() {
		return nil, fmt.Errorf("Invalid format for VNET.")
	}

	result := networkingCommandEntry_answer{vnet: vnet, value: row[1]}
	return &networkingCommandEntry{entry: result, answer: &result}, nil
}
func parseNetworkingCommand_remove_answer(row []string) (*networkingCommandEntry, error) {
	if len(row) != 1 {
		return nil, fmt.Errorf("Expected %d argument. Received %d.", 1, len(row))
	}
	vnet := networkingVNET{value: row[0]}
	if !vnet.Valid() {
		return nil, fmt.Errorf("Invalid format for VNET.")
	}

	result := networkingCommandEntry_remove_answer{vnet: vnet}
	return &networkingCommandEntry{entry: result, remove_answer: &result}, nil
}
func parseNetworkingCommand_add_nat_portfwd(row []string) (*networkingCommandEntry, error) {
	if len(row) != 5 {
		return nil, fmt.Errorf("Expected %d arguments. Received only %d.", 5, len(row))
	}

	vnet, err := strconv.Atoi(row[0])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse first argument as an integer. : %v", row[0])
	}

	protocol := strings.ToLower(row[1])
	if !(protocol == "tcp" || protocol == "udp") {
		return nil, fmt.Errorf("Expected \"tcp\" or \"udp\" for second argument. : %v", row[1])
	}

	sport, err := strconv.Atoi(row[2])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse third argument as an integer. : %v", row[2])
	}

	dest := net.ParseIP(row[3])
	if dest == nil {
		return nil, fmt.Errorf("Unable to parse fourth argument as an IPv4 address. : %v", row[2])
	}

	dport, err := strconv.Atoi(row[4])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse fifth argument as an integer. : %v", row[4])
	}

	result := networkingCommandEntry_add_nat_portfwd{vnet: vnet - 1, protocol: protocol, port: sport, target_host: dest, target_port: dport}
	return &networkingCommandEntry{entry: result, add_nat_portfwd: &result}, nil
}
func parseNetworkingCommand_remove_nat_portfwd(row []string) (*networkingCommandEntry, error) {
	if len(row) != 3 {
		return nil, fmt.Errorf("Expected %d arguments. Received only %d.", 3, len(row))
	}

	vnet, err := strconv.Atoi(row[0])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse first argument as an integer. : %v", row[0])
	}

	protocol := strings.ToLower(row[1])
	if !(protocol == "tcp" || protocol == "udp") {
		return nil, fmt.Errorf("Expected \"tcp\" or \"udp\" for second argument. : %v", row[1])
	}

	sport, err := strconv.Atoi(row[2])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse third argument as an integer. : %v", row[2])
	}

	result := networkingCommandEntry_remove_nat_portfwd{vnet: vnet - 1, protocol: protocol, port: sport}
	return &networkingCommandEntry{entry: result, remove_nat_portfwd: &result}, nil
}
func parseNetworkingCommand_add_dhcp_mac_to_ip(row []string) (*networkingCommandEntry, error) {
	if len(row) != 3 {
		return nil, fmt.Errorf("Expected %d arguments. Received only %d.", 3, len(row))
	}

	vnet, err := strconv.Atoi(row[0])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse first argument as an integer. : %v", row[0])
	}

	mac, err := net.ParseMAC(row[1])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse second argument as hwaddr. : %v", row[1])
	}

	ip := net.ParseIP(row[2])
	if ip == nil {
		return nil, fmt.Errorf("Unable to parse third argument as ipv4. : %v", row[2])
	}

	result := networkingCommandEntry_add_dhcp_mac_to_ip{vnet: vnet - 1, mac: mac, ip: ip}
	return &networkingCommandEntry{entry: result, add_dhcp_mac_to_ip: &result}, nil
}
func parseNetworkingCommand_remove_dhcp_mac_to_ip(row []string) (*networkingCommandEntry, error) {
	if len(row) != 2 {
		return nil, fmt.Errorf("Expected %d arguments. Received only %d.", 2, len(row))
	}

	vnet, err := strconv.Atoi(row[0])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse first argument as an integer. : %v", row[0])
	}

	mac, err := net.ParseMAC(row[1])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse second argument as hwaddr. : %v", row[1])
	}

	result := networkingCommandEntry_remove_dhcp_mac_to_ip{vnet: vnet - 1, mac: mac}
	return &networkingCommandEntry{entry: result, remove_dhcp_mac_to_ip: &result}, nil
}
func parseNetworkingCommand_add_bridge_mapping(row []string) (*networkingCommandEntry, error) {
	if len(row) != 2 {
		return nil, fmt.Errorf("Expected %d arguments. Received only %d.", 2, len(row))
	}
	intf := networkingInterface{name: row[0]}

	vnet, err := strconv.Atoi(row[1])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse second argument as an integer. : %v", row[2])
	}

	result := networkingCommandEntry_add_bridge_mapping{intf: intf, vnet: vnet - 1}
	return &networkingCommandEntry{entry: result, add_bridge_mapping: &result}, nil
}
func parseNetworkingCommand_remove_bridge_mapping(row []string) (*networkingCommandEntry, error) {
	if len(row) != 1 {
		return nil, fmt.Errorf("Expected %d argument. Received %d.", 1, len(row))
	}
	intf := networkingInterface{name: row[0]}
	/*
		number, err := strconv.Atoi(row[0])
		if err != nil {
			return nil, fmt.Errorf("Unable to parse first argument as an integer. : %v", row[0])
		}
	*/
	result := networkingCommandEntry_remove_bridge_mapping{intf: intf}
	return &networkingCommandEntry{entry: result, remove_bridge_mapping: &result}, nil
}
func parseNetworkingCommand_add_nat_prefix(row []string) (*networkingCommandEntry, error) {
	if len(row) != 2 {
		return nil, fmt.Errorf("Expected %d arguments. Received only %d.", 2, len(row))
	}

	vnet, err := strconv.Atoi(row[0])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse first argument as an integer. : %v", row[0])
	}

	if !strings.HasPrefix(row[1], "/") {
		return nil, fmt.Errorf("Expected second argument to begin with \"/\". : %v", row[1])
	}

	prefix, err := strconv.Atoi(row[1][1:])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse prefix out of second argument. : %v", row[1])
	}

	result := networkingCommandEntry_add_nat_prefix{vnet: vnet - 1, prefix: prefix}
	return &networkingCommandEntry{entry: result, add_nat_prefix: &result}, nil
}
func parseNetworkingCommand_remove_nat_prefix(row []string) (*networkingCommandEntry, error) {
	if len(row) != 2 {
		return nil, fmt.Errorf("Expected %d arguments. Received only %d.", 2, len(row))
	}

	vnet, err := strconv.Atoi(row[0])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse first argument as an integer. : %v", row[0])
	}

	if !strings.HasPrefix(row[1], "/") {
		return nil, fmt.Errorf("Expected second argument to begin with \"/\". : %v", row[1])
	}

	prefix, err := strconv.Atoi(row[1][1:])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse prefix out of second argument. : %v", row[1])
	}

	result := networkingCommandEntry_remove_nat_prefix{vnet: vnet - 1, prefix: prefix}
	return &networkingCommandEntry{entry: result, remove_nat_prefix: &result}, nil
}

type networkingCommandParser struct {
	command  string
	callback func([]string) (*networkingCommandEntry, error)
}

var NetworkingCommandParsers = []networkingCommandParser{
	/* DictRecordParseFunct */ {command: "answer", callback: parseNetworkingCommand_answer},
	/* DictRecordParseFunct */ {command: "remove_answer", callback: parseNetworkingCommand_remove_answer},
	/* NatFwdRecordParseFunct */ {command: "add_nat_portfwd", callback: parseNetworkingCommand_add_nat_portfwd},
	/* NatFwdRecordParseFunct */ {command: "remove_nat_portfwd", callback: parseNetworkingCommand_remove_nat_portfwd},
	/* DhcpMacRecordParseFunct */ {command: "add_dhcp_mac_to_ip", callback: parseNetworkingCommand_add_dhcp_mac_to_ip},
	/* DhcpMacRecordParseFunct */ {command: "remove_dhcp_mac_to_ip", callback: parseNetworkingCommand_remove_dhcp_mac_to_ip},
	/* BridgeMappingRecordParseFunct */ {command: "add_bridge_mapping", callback: parseNetworkingCommand_add_bridge_mapping},
	/* BridgeMappingRecordParseFunct */ {command: "remove_bridge_mapping", callback: parseNetworkingCommand_remove_bridge_mapping},
	/* NatPrefixRecordParseFunct */ {command: "add_nat_prefix", callback: parseNetworkingCommand_add_nat_prefix},
	/* NatPrefixRecordParseFunct */ {command: "remove_nat_prefix", callback: parseNetworkingCommand_remove_nat_prefix},
}

func NetworkingParserByCommand(command string) *func([]string) (*networkingCommandEntry, error) {
	for _, p := range NetworkingCommandParsers {
		if p.command == command {
			return &p.callback
		}
	}
	return nil
}

func parseNetworkingConfig(rows chan []string) chan networkingCommandEntry {
	out := make(chan networkingCommandEntry)

	go func(in chan []string, out chan networkingCommandEntry) {
		for {
			row, ok := <-in
			if !ok {
				break
			}

			if len(row) >= 1 {
				parser := NetworkingParserByCommand(row[0])
				if parser == nil {
					log.Printf("Invalid command : %v", row)
					continue
				}

				callback := *parser

				entry, err := callback(row[1:])
				if err != nil {
					log.Printf("Unable to parse command : %v %v", err, row)
					continue
				}
				out <- *entry
			}
		}
		close(out)
	}(rows, out)
	return out
}

type NetworkingConfig struct {
	answer         map[int]map[string]string
	nat_portfwd    map[int]map[string]string
	dhcp_mac_to_ip map[int]map[string]net.IP
	//bridge_mapping map[net.Interface]uint64	// XXX: we don't need the actual interface for anything but informing the user.
	bridge_mapping map[string]int
	nat_prefix     map[int][]int
}

func (c NetworkingConfig) repr() string {
	return fmt.Sprintf("answer -> %v\nnat_portfwd -> %v\ndhcp_mac_to_ip -> %v\nbridge_mapping -> %v\nnat_prefix -> %v", c.answer, c.nat_portfwd, c.dhcp_mac_to_ip, c.bridge_mapping, c.nat_prefix)
}

func flattenNetworkingConfig(in chan networkingCommandEntry) NetworkingConfig {
	var result NetworkingConfig
	var vmnet int

	result.answer = make(map[int]map[string]string)
	result.nat_portfwd = make(map[int]map[string]string)
	result.dhcp_mac_to_ip = make(map[int]map[string]net.IP)
	result.bridge_mapping = make(map[string]int)
	result.nat_prefix = make(map[int][]int)

	for {
		e, ok := <-in
		if !ok {
			break
		}

		switch e.entry.(type) {
		case networkingCommandEntry_answer:
			vnet := e.answer.vnet
			answers, exists := result.answer[vnet.Number()]
			if !exists {
				answers = make(map[string]string)
				result.answer[vnet.Number()] = answers
			}
			answers[vnet.Option()] = e.answer.value

		case networkingCommandEntry_remove_answer:
			vnet := e.remove_answer.vnet
			answers, exists := result.answer[vnet.Number()]
			if exists {
				delete(answers, vnet.Option())
			} else {
				log.Printf("Unable to remove answer %s as specified by `remove_answer`.\n", vnet.Repr())
			}

		case networkingCommandEntry_add_nat_portfwd:
			vmnet = e.add_nat_portfwd.vnet
			protoport := fmt.Sprintf("%s/%d", e.add_nat_portfwd.protocol, e.add_nat_portfwd.port)
			target := fmt.Sprintf("%s:%d", e.add_nat_portfwd.target_host, e.add_nat_portfwd.target_port)
			portfwds, exists := result.nat_portfwd[vmnet]
			if !exists {
				portfwds = make(map[string]string)
				result.nat_portfwd[vmnet] = portfwds
			}
			portfwds[protoport] = target

		case networkingCommandEntry_remove_nat_portfwd:
			vmnet = e.remove_nat_portfwd.vnet
			protoport := fmt.Sprintf("%s/%d", e.remove_nat_portfwd.protocol, e.remove_nat_portfwd.port)
			portfwds, exists := result.nat_portfwd[vmnet]
			if exists {
				delete(portfwds, protoport)
			} else {
				log.Printf("Unable to remove nat port-forward %s from interface %s%d as requested by `remove_nat_portfwd`.\n", protoport, NetworkingInterfacePrefix, vmnet)
			}

		case networkingCommandEntry_add_dhcp_mac_to_ip:
			vmnet = e.add_dhcp_mac_to_ip.vnet
			dhcpmacs, exists := result.dhcp_mac_to_ip[vmnet]
			if !exists {
				dhcpmacs = make(map[string]net.IP)
				result.dhcp_mac_to_ip[vmnet] = dhcpmacs
			}
			dhcpmacs[e.add_dhcp_mac_to_ip.mac.String()] = e.add_dhcp_mac_to_ip.ip

		case networkingCommandEntry_remove_dhcp_mac_to_ip:
			vmnet = e.remove_dhcp_mac_to_ip.vnet
			dhcpmacs, exists := result.dhcp_mac_to_ip[vmnet]
			if exists {
				delete(dhcpmacs, e.remove_dhcp_mac_to_ip.mac.String())
			} else {
				log.Printf("Unable to remove dhcp_mac_to_ip entry %v from interface %s%d as specified by `remove_dhcp_mac_to_ip`.\n", e.remove_dhcp_mac_to_ip, NetworkingInterfacePrefix, vmnet)
			}

		case networkingCommandEntry_add_bridge_mapping:
			intf := e.add_bridge_mapping.intf
			if _, err := intf.Interface(); err != nil {
				log.Printf("Interface \"%s\" as specified by `add_bridge_mapping` was not found on the current platform. This is a non-critical error. Ignoring.", intf.name)
			}
			result.bridge_mapping[intf.name] = e.add_bridge_mapping.vnet

		case networkingCommandEntry_remove_bridge_mapping:
			intf := e.remove_bridge_mapping.intf
			if _, err := intf.Interface(); err != nil {
				log.Printf("Interface \"%s\" as specified by `remove_bridge_mapping` was not found on the current platform. This is a non-critical error. Ignoring.", intf.name)
			}
			delete(result.bridge_mapping, intf.name)

		case networkingCommandEntry_add_nat_prefix:
			vmnet = e.add_nat_prefix.vnet
			_, exists := result.nat_prefix[vmnet]
			if exists {
				result.nat_prefix[vmnet] = append(result.nat_prefix[vmnet], e.add_nat_prefix.prefix)
			} else {
				result.nat_prefix[vmnet] = []int{e.add_nat_prefix.prefix}
			}

		case networkingCommandEntry_remove_nat_prefix:
			vmnet = e.remove_nat_prefix.vnet
			prefixes, exists := result.nat_prefix[vmnet]
			if exists {
				for index := 0; index < len(prefixes); index++ {
					if prefixes[index] == e.remove_nat_prefix.prefix {
						result.nat_prefix[vmnet] = append(prefixes[:index], prefixes[index+1:]...)
						break
					}
				}

			} else {
				log.Printf("Unable to remove nat prefix /%d from interface %s%d as specified by `remove_nat_prefix`.\n", e.remove_nat_prefix.prefix, NetworkingInterfacePrefix, vmnet)
			}
		}
	}
	return result
}

// Constructor for networking file
func ReadNetworkingConfig(fd *os.File) (NetworkingConfig, error) {

	// start piecing together all of the differents parts of the file and split
	// it into its individual rows.
	fromfile := consumeFile(fd)
	tokenized := tokenizeNetworkingConfig(fromfile)
	rows := splitNetworkingConfig(tokenized)

	// consume the version _first_. this is important because if the version is
	// wrong, then there's likely tokens that we won't know how to interpret.
	parsed_version, err := networkingReadVersion(<-rows)
	if err != nil {
		return NetworkingConfig{}, err
	}

	// verify that it's 1.0 since that's all we support for now.
	if version := parsed_version.Number(); version != 1.0 {
		return NetworkingConfig{}, fmt.Errorf("Expected version %f of networking file. Received version %f.", 1.0, version)
	}

	// now that our version has been confirmed, we can proceed to parse the
	// rest of the file and parseNetworkingConfig is free to consume rows as
	// much as it wants to.
	entries := parseNetworkingConfig(rows)

	// convert what we've parsed into a configuration that's easy to interpret
	return flattenNetworkingConfig(entries), nil
}

// netmapper interface
type NetworkingType int

const (
	NetworkingType_HOSTONLY = iota + 1
	NetworkingType_NAT
	NetworkingType_BRIDGED
)

func networkingConfig_InterfaceTypes(config NetworkingConfig) map[int]NetworkingType {
	result := make(map[int]NetworkingType)

	// defaults
	result[0] = NetworkingType_BRIDGED
	result[1] = NetworkingType_HOSTONLY
	result[8] = NetworkingType_NAT

	// walk through config collecting bridged interfaces
	for _, vmnet := range config.bridge_mapping {
		result[vmnet] = NetworkingType_BRIDGED
	}

	// walk through answers finding out which ones are nat versus hostonly
	for vmnet, table := range config.answer {

		// everything should be defined as a virtual adapter...
		if table["VIRTUAL_ADAPTER"] == "yes" {

			// validate that the VNET entry contains everything we expect it to
			_, subnetQ := table["HOSTONLY_SUBNET"]
			_, netmaskQ := table["HOSTONLY_NETMASK"]
			if !(subnetQ && netmaskQ) {
				log.Printf("Interface %s%d is missing some expected keys (HOSTONLY_SUBNET, HOSTONLY_NETMASK). This is non-critical. Ignoring..", NetworkingInterfacePrefix, vmnet)
			}

			// distinguish between nat or hostonly
			if table["NAT"] == "yes" {
				result[vmnet] = NetworkingType_NAT

			} else {
				result[vmnet] = NetworkingType_HOSTONLY
			}

		} else {
			// if it's not a virtual_adapter, then it must be an alias (really a bridge).
			result[vmnet] = NetworkingType_BRIDGED
		}
	}
	return result
}

func networkingConfig_NamesToVmnet(config NetworkingConfig) map[NetworkingType][]int {
	types := networkingConfig_InterfaceTypes(config)

	// now sort the keys
	var keys []int
	for vmnet := range types {
		keys = append(keys, vmnet)
	}
	sort.Ints(keys)

	// build result dictionary
	result := make(map[NetworkingType][]int)

	for i := 0; i < len(keys); i++ {
		t := types[keys[i]]
		result[t] = append(result[t], keys[i])
	}
	return result
}

const NetworkingInterfacePrefix = "vmnet"

func (e NetworkingConfig) NameIntoDevices(name string) ([]string, error) {
	netmapper := networkingConfig_NamesToVmnet(e)
	name = strings.ToLower(name)

	var vmnets []string
	var networkingType NetworkingType
	if name == "hostonly" && len(netmapper[NetworkingType_HOSTONLY]) > 0 {
		networkingType = NetworkingType_HOSTONLY

	} else if name == "nat" && len(netmapper[NetworkingType_NAT]) > 0 {
		networkingType = NetworkingType_NAT

	} else if name == "bridged" && len(netmapper[NetworkingType_BRIDGED]) > 0 {
		networkingType = NetworkingType_BRIDGED

	} else {
		return make([]string, 0), fmt.Errorf("Network name not found: %v", name)
	}

	for i := 0; i < len(netmapper[networkingType]); i++ {
		vmnets = append(vmnets, fmt.Sprintf("%s%d", NetworkingInterfacePrefix, netmapper[networkingType][i]))
	}
	return vmnets, nil
}

func (e NetworkingConfig) DeviceIntoName(device string) (string, error) {
	types := networkingConfig_InterfaceTypes(e)

	lowerdevice := strings.ToLower(device)
	if !strings.HasPrefix(lowerdevice, NetworkingInterfacePrefix) {
		return device, nil
	}
	vmnet, err := strconv.Atoi(lowerdevice[len(NetworkingInterfacePrefix):])
	if err != nil {
		return "", err
	}
	network := types[vmnet]
	switch network {
	case NetworkingType_HOSTONLY:
		return "hostonly", nil

	case NetworkingType_NAT:
		return "nat", nil

	case NetworkingType_BRIDGED:
		return "bridged", nil
	}
	return "", fmt.Errorf("Unable to determine network type for device %s%d.", NetworkingInterfacePrefix, vmnet)
}

/** generic async file reader */
func consumeFile(fd *os.File) chan byte {
	fromFile := make(chan byte)
	go func() {
		b := make([]byte, 1)
		for {
			_, err := fd.Read(b)
			if err != nil {
				// In case of any error we must stop
				// ErrClosed may appear since file is closed and this goroutine still left running
				break
			}
			fromFile <- b[0]
		}
		close(fromFile)
	}()
	return fromFile
}

/** Consume a byte channel until a terminal byte is reached, and write each list of bytes to a channel */
func consumeUntilSentinel(sentinel byte, in chan byte) (result []byte, ok bool) {

	// This is a simple utility that will consume from a channel until a sentinel
	// byte has been reached. Consumed data is returned in `result, and if
	// there's no more data to read, then `ok` will be false.
	for ok = true; ; {
		if by, success := <-in; !success {
			ok = false
			break

		} else if by == sentinel {
			break

		} else {
			result = append(result, by)

		}
	}
	return
}

/** Simple utility to ignore chars when consuming a channel */
func filterOutCharacters(ignore []byte, in chan byte) chan byte {
	out := make(chan byte)

	go func(ignore_s string) {
		for {
			if by, ok := <-in; !ok {
				break

			} else if !strings.ContainsAny(ignore_s, string(by)) {
				out <- by
			}
		}
		close(out)
	}(string(ignore))

	return out
}

/**
This consumes bytes within a pair of some bytes, like parentheses, brackets, braces...

We start by reading bytes until we encounter openByte. These will be returned as
the first parameter. Then we can enter a goro and consume bytes until we get to
closeByte. At that point we're done, and we can exit.
**/
func consumeOpenClosePair(openByte, closeByte byte, in chan byte) ([]byte, chan byte) {
	result := make([]byte, 0)

	// Consume until we get to openByte. We'll return what we consumed because
	// it isn't actually relevant to what we're trying to accomplish.
	for by := range in {
		if by == openByte {
			break

		} else {
			result = append(result, by)
		}
	}

	// Now we can feed input to our goro and a consumer can see what's contained
	// between their requested pairs
	out := make(chan byte)
	go func(out chan byte) {
		by := openByte

		// We only made it here because we received an openByte, so let's make
		// sure we send it down the channel.
		out <- by

		// Now just spin in a loop shipping bytes down the channel until we hit
		// closeByte, or we're at the very end...whichever comes first.
		var ok bool
		for by != closeByte {
			by, ok = <-in
			if !ok {
				by = closeByte
			}

			out <- by
		}
		close(out)
	}(out)

	// Return what we consumed, and a channel that yields everything in between
	// the openByte and closeByte pair.
	return result, out
}

// Basic decoding of a dhcpd lease address
func decodeDhcpdLeaseBytes(input string) ([]byte, error) {
	processed := &bytes.Buffer{}

	// Split the string into pieces as we'll need to validate it.
	for _, item := range strings.Split(input, ":") {
		if len(item) != 2 {
			return []byte{}, fmt.Errorf("bytes are not well-formed (%v)", input)
		}
		processed.WriteString(item)
	}

	length := hex.DecodedLen(processed.Len())

	// Decode the processed data into the result...
	result := make([]byte, length)
	if n, err := hex.Decode(result, processed.Bytes()); err != nil {
		return []byte{}, err

		// Check that our decode length corresponds to what was intended
	} else if n != length {
		return []byte{}, fmt.Errorf("expected to decode %d bytes, got %d instead", length, n)
	}

	// ...and then return it.
	return result, nil
}

/*** Dhcp Leases */
type dhcpLeaseEntry struct {
	address                      string
	starts, ends                 time.Time
	starts_weekday, ends_weekday int
	ether, uid                   []byte
	extra                        []string
}

func readDhcpdLeaseEntry(in chan byte) (entry *dhcpLeaseEntry, err error) {

	// Build the regexes we'll use to legitimately parse each item
	ipLineRe := regexp.MustCompile(`lease\s+(.+?)\s*$`)
	startTimeLineRe := regexp.MustCompile(`starts\s+(\d+)\s+(.+?)\s*$`)
	endTimeLineRe := regexp.MustCompile(`ends\s+(\d+)\s+(.+?)\s*$`)
	macLineRe := regexp.MustCompile(`hardware\s+ethernet\s+(.+?)\s*$`)
	uidLineRe := regexp.MustCompile(`uid\s+(.+?)\s*$`)

	/// Read up to the lease item and validate that it actually matches
	lease, ch := consumeOpenClosePair('{', '}', in)

	// If we couldn't read the lease, then this item is busted and we're prolly
	// done reading the channel.
	if len(lease) == 0 {
		return nil, nil
	}

	matches := ipLineRe.FindStringSubmatch(string(lease))
	if matches == nil {
		res := strings.TrimSpace(string(lease))
		return &dhcpLeaseEntry{extra: []string{res}}, fmt.Errorf("Unable to parse lease entry (%#v)", string(lease))
	}

	if by, ok := <-ch; ok && by == '{' {
		// If we found a lease match and we're definitely beginning a lease
		// entry, then create our storage.
		entry = &dhcpLeaseEntry{address: matches[1]}

	} else if ok {
		// If we didn't see a begin brace, then this entry is mangled which
		// means that we should probably bail
		return &dhcpLeaseEntry{address: matches[1]}, fmt.Errorf("Missing parameters for lease entry %v", matches[1])

	} else if !ok {
		// If our channel is closed, so we bail "cleanly".
		return nil, nil
	}

	/// Now we can parse the inside of the block.
	for insideBraces := true; insideBraces; {
		item, ok := consumeUntilSentinel(';', ch)
		item_s := string(item)

		if !ok {
			insideBraces = false
		}

		// Parse out the start time
		matches = startTimeLineRe.FindStringSubmatch(item_s)
		if matches != nil {
			if entry.starts, err = time.Parse("2006/01/02 15:04:05", matches[2]); err != nil {
				log.Printf("Error trying to parse start time (%v) for entry %v", matches[2], entry.address)
			}
			if entry.starts_weekday, err = strconv.Atoi(matches[1]); err != nil {
				log.Printf("Error trying to parse start weekday (%v) for entry %v", matches[1], entry.address)
			}
			continue
		}

		// Parse out the end time
		matches = endTimeLineRe.FindStringSubmatch(item_s)
		if matches != nil {
			if entry.ends, err = time.Parse("2006/01/02 15:04:05", matches[2]); err != nil {
				log.Printf("Error trying to parse end time (%v) for entry %v", matches[2], entry.address)
			}
			if entry.ends_weekday, err = strconv.Atoi(matches[1]); err != nil {
				log.Printf("Error trying to parse end weekday (%v) for entry %v", matches[1], entry.address)
			}
			continue
		}

		// Parse out the hardware ethernet
		matches = macLineRe.FindStringSubmatch(item_s)
		if matches != nil {
			if entry.ether, err = decodeDhcpdLeaseBytes(matches[1]); err != nil {
				log.Printf("Error trying to parse hardware ethernet address (%v) for entry %v", matches[1], entry.address)
			}
			continue
		}

		// Parse out the uid
		matches = uidLineRe.FindStringSubmatch(item_s)
		if matches != nil {
			if entry.uid, err = decodeDhcpdLeaseBytes(matches[1]); err != nil {
				log.Printf("Error trying to parse uid (%v) for entry %v", matches[1], entry.address)
			}
			continue
		}

		// Check to see if we're terminating the brace, so we can skip
		// to the next iteration.
		if strings.HasSuffix(item_s, "}") {
			continue
		}

		// Just stash it for now because we have no idea what it is.
		entry.extra = append(entry.extra, strings.TrimSpace(item_s))
	}

	return entry, nil
}

func ReadDhcpdLeaseEntries(fd *os.File) ([]dhcpLeaseEntry, error) {
	fch := consumeFile(fd)
	uncommentedch := uncomment(fch)
	wch := filterOutCharacters([]byte{'\n', '\r', '\v'}, uncommentedch)

	result := make([]dhcpLeaseEntry, 0)
	errors := make([]error, 0)

	// Consume dhcpd lease entries from the channel until we just plain run out.
	for i := 0; ; i += 1 {
		if entry, err := readDhcpdLeaseEntry(wch); entry == nil {
			// If our entry is nil, then we've run out of input and finished
			// parsing the file to completion.
			break

		} else if err != nil {
			// If we received an error, then log it and keep track of it. This
			// way we can warn the user later which entries we had issues with.
			log.Printf("Error parsing dhcpd lease entry #%d: %s", 1+i, err)
			errors = append(errors, err)

		} else {
			// If we've parsed an entry successfully, then aggregate it to
			// our slice of results.
			result = append(result, *entry)
		}
	}

	// If we received any errors then include alongside our results.
	if len(errors) > 0 {
		return result, fmt.Errorf("Errors found while parsing dhcpd lease entries: %v", errors)
	}
	return result, nil
}

/*** Apple Dhcp Leases */

// Here is what an Apple DHCPD lease entry looks like:
// {
// 	ip_address=192.168.111.2
// 	hw_address=1,0:50:56:20:ac:33
// 	identifier=1,0:50:56:20:ac:33
// 	lease=0x5fd72edc
// 	name=vagrant-2019
// }

type appleDhcpLeaseEntry struct {
	ipAddress     string
	hwAddress, id []byte
	lease         string
	name          string
	extra         map[string]string
}

func readAppleDhcpdLeaseEntry(in chan byte) (entry *appleDhcpLeaseEntry, err error) {
	entry = &appleDhcpLeaseEntry{extra: map[string]string{}}
	mandatoryFieldCount := 0
	// Read up to the lease item and validate that it actually matches
	_, ch := consumeOpenClosePair('{', '}', in)
	for insideBraces := true; insideBraces; {
		item, ok := consumeUntilSentinel('\n', ch)
		item_s := strings.TrimSpace(string(item))

		if !ok {
			insideBraces = false
		}
		if strings.Contains(item_s, "{") || strings.Contains(item_s, "}") {
			continue
		}
		splittedLine := strings.Split(item_s, "=")
		var key, val string
		switch len(splittedLine) {
		case 0:
			// should never happens as Split always returns at least 1 item
			fallthrough
		case 1:
			log.Printf("Error parsing invalid line: `%s`", item_s)
			continue
		case 2:
			key = strings.TrimSpace(splittedLine[0])
			val = strings.TrimSpace(splittedLine[1])
		default:
			// There were more than one '=' on this line, we'll keep the part before the first '=' as the key and
			// the rest will be the value
			key = strings.TrimSpace(splittedLine[0])
			val = strings.TrimSpace(strings.Join(splittedLine[1:], "="))
		}
		switch key {
		case "ip_address":
			entry.ipAddress = val
			mandatoryFieldCount++
		case "identifier":
			fallthrough
		case "hw_address":
			if strings.Count(val, ",") != 1 {
				log.Printf("Error %s `%s` is not properly formatted for entry %s", key, val, entry.name)
				break
			}
			splittedVal := strings.Split(val, ",")
			mac := splittedVal[1]
			splittedMac := strings.Split(mac, ":")
			// Pad the retrieved hw address with '0' when necessary
			for idx := range splittedMac {
				if len(splittedMac[idx]) == 1 {
					splittedMac[idx] = "0" + splittedMac[idx]
				}
			}
			mac = strings.Join(splittedMac, ":")
			decodedLease, err := decodeDhcpdLeaseBytes(mac)
			if err != nil {
				log.Printf("Error trying to parse %s (%v) for entry %s - %v", key, val, entry.name, mac)
				break
			}
			if key == "identifier" {
				entry.id = decodedLease
			} else {
				entry.hwAddress = decodedLease
			}
			mandatoryFieldCount++
		case "lease":
			entry.lease = val
		case "name":
			entry.name = val
		default:
			// Just stash it for now because we have no idea what it is.
			entry.extra[key] = val
		}
	}
	// we have most likely parsed the whole file
	if mandatoryFieldCount == 0 {
		return nil, nil
	}
	// an entry is composed of 3 mandatory fields, we'll check that they all have been set during the parsing
	if mandatoryFieldCount < 3 {
		return entry, fmt.Errorf("Error entry `%v` is missing mandatory information", entry)
	}
	return entry, nil
}

func ReadAppleDhcpdLeaseEntries(fd *os.File) ([]appleDhcpLeaseEntry, error) {
	fch := consumeFile(fd)
	uncommentedch := uncomment(fch)
	wch := filterOutCharacters([]byte{'\r', '\v'}, uncommentedch)

	result := make([]appleDhcpLeaseEntry, 0)
	errors := make([]error, 0)

	// Consume apple dhcpd lease entries from the channel until we just plain run out.
	for i := 0; ; i++ {
		if entry, err := readAppleDhcpdLeaseEntry(wch); entry == nil {
			// If our entry is nil, then we've run out of input and finished
			// parsing the file to completion.
			break

		} else if err != nil {
			// If we received an error, then log it and keep track of it. This
			// way we can warn the user later which entries we had issues with.
			log.Printf("Error parsing apple dhcpd lease entry #%d: %s", 1+i, err)
			errors = append(errors, err)

		} else {
			// If we've parsed an entry successfully, then aggregate it to
			// our slice of results.
			result = append(result, *entry)
		}
	}

	// If we received any errors then include alongside our results.
	if len(errors) > 0 {
		return result, fmt.Errorf("Errors found while parsing apple dhcpd lease entries: %v", errors)
	}
	return result, nil
}
