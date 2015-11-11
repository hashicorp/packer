package common
import (
	"fmt"
	"os"
	"io"
	"strings"
	"strconv"
	"net"
	"sort"
)

type sentinelSignaller chan struct{}

/** low-level parsing */
// strip the comments and extraneous newlines from a byte channel
func uncomment(eof sentinelSignaller, in <-chan byte) chan byte {
	out := make(chan byte)

	go func(in <-chan byte, out chan byte) {
		var endofline bool
		for stillReading := true; stillReading; {
			select {
				case <-eof:
					stillReading = false
				case ch := <-in:
					switch ch {
						case '#':
							endofline = true
						case '\n':
							if endofline { endofline = false }
					}
					if !endofline { out <- ch }
			}
		}
	}(in, out)
	return out
}

// convert a byte channel into a channel of pseudo-tokens
func tokenizeDhcpConfig(eof sentinelSignaller, in chan byte) chan string {
	var ch byte
	var state string
	var quote bool

	out := make(chan string)
	go func(out chan string) {
		for stillReading := true; stillReading; {
			select {
				case <-eof:
					stillReading = false

				case ch = <-in:
					if quote {
						if ch == '"' {
							out <- state + string(ch)
							state,quote = "",false
							continue
						}
						state += string(ch)
						continue
					}

					switch ch {
						case '"':
							quote = true
							state += string(ch)
							continue

						case '\r':
							fallthrough
						case '\n':
							fallthrough
						case '\t':
							fallthrough
						case ' ':
							if len(state) == 0 { continue }
							out <- state
							state = ""

						case '{': fallthrough
						case '}': fallthrough
						case ';':
							if len(state) > 0 { out <- state }
							out <- string(ch)
							state = ""

						default:
							state += string(ch)
					}
			}
		}
		if len(state) > 0 { out <- state }
	}(out)
	return out
}

/** mid-level parsing */
type tkParameter struct {
	name string
	operand []string
}
func (e *tkParameter) String() string {
	var values []string
	for _,val := range e.operand {
		values = append(values, val)
	}
	return fmt.Sprintf("%s [%s]", e.name, strings.Join(values, ","))
}

type tkGroup struct {
	parent *tkGroup
	id tkParameter

	groups []*tkGroup
	params []tkParameter
}
func (e *tkGroup) String() string {
	var id []string

	id = append(id, e.id.name)
	for _,val := range e.id.operand {
		id = append(id, val)
	}

	var config []string
	for _,val := range e.params {
		config = append(config, val.String())
	}
	return fmt.Sprintf("%s {\n%s\n}", strings.Join(id, " "), strings.Join(config, "\n"))
}

// convert a channel of pseudo-tokens into an tkParameter struct
func parseTokenParameter(in chan string) tkParameter {
	var result tkParameter
	for {
		token := <-in
		if result.name == "" {
			result.name = token
			continue
		}
		switch token {
			case "{": fallthrough
			case "}": fallthrough
			case ";": goto leave
			default:
				result.operand = append(result.operand, token)
		}
	}
leave:
	return result
}

// convert a channel of pseudo-tokens into an tkGroup tree */
func parseDhcpConfig(eof sentinelSignaller, in chan string) (tkGroup,error) {
	var tokens []string
	var result tkGroup

	toParameter := func(tokens []string) tkParameter {
		out := make(chan string)
		go func(out chan string){
			for _,v := range tokens { out <- v }
			out <- ";"
		}(out)
		return parseTokenParameter(out)
	}

	for stillReading,currentgroup := true,&result; stillReading; {
		select {
			case <-eof:
				stillReading = false

			case tk := <-in:
				switch tk {
					case "{":
						grp := &tkGroup{parent:currentgroup}
						grp.id = toParameter(tokens)
						currentgroup.groups = append(currentgroup.groups, grp)
						currentgroup = grp
					case "}":
						if currentgroup.parent == nil {
							return tkGroup{}, fmt.Errorf("Unable to close the global declaration")
						}
						if len(tokens) > 0 {
							return tkGroup{}, fmt.Errorf("List of tokens was left unterminated : %v", tokens)
						}
						currentgroup = currentgroup.parent
					case ";":
						arg := toParameter(tokens)
						currentgroup.params = append(currentgroup.params, arg)
					default:
						tokens = append(tokens, tk)
						continue
				}
				tokens = []string{}
		}
	}
	return result,nil
}

func tokenizeNetworkMapConfig(eof sentinelSignaller, in chan byte) chan string {
	var ch byte
	var state string
	var quote bool
	var lastnewline bool

	out := make(chan string)
	go func(out chan string) {
		for stillReading := true; stillReading; {
			select {
				case <-eof:
					stillReading = false

				case ch = <-in:
					if quote {
						if ch == '"' {
							out <- state + string(ch)
							state,quote = "",false
							continue
						}
						state += string(ch)
						continue
					}

					switch ch {
						case '"':
							quote = true
							state += string(ch)
							continue

						case '\r':
							fallthrough
						case '\t':
							fallthrough
						case ' ':
							if len(state) == 0 { continue }
							out <- state
							state = ""

						case '\n':
							if lastnewline { continue }
							if len(state) > 0 { out <- state }
							out <- string(ch)
							state = ""
							lastnewline = true
							continue

						case '.': fallthrough
						case '=':
							if len(state) > 0 { out <- state }
							out <- string(ch)
							state = ""

						default:
							state += string(ch)
					}
					lastnewline = false
			}
		}
		if len(state) > 0 { out <- state }
	}(out)
	return out
}

func parseNetworkMapConfig(eof sentinelSignaller, in chan string) (NetworkMap,error) {
	var unsorted map[string]map[string]string
	var state []string

	addResult := func(network string, attribute string, value string) error {
		_,ok := unsorted[network]
		if !ok { unsorted[network] = make(map[string]string) }

		val,err := strconv.Unquote(value)
		if err != nil { return err }

		current := unsorted[network]
		current[attribute] = val
		return nil
	}

	stillReading := true
	for unsorted = make(map[string]map[string]string); stillReading; {
		select {
			case <-eof:
				if len(state) == 3 {
					err := addResult(state[0], state[1], state[2])
					if err != nil { return nil,err }
				}
				stillReading = false
			case tk := <-in:
				switch tk {
					case ".":
						if len(state) != 1 { return nil,fmt.Errorf("Missing network index") }
					case "=":
						if len(state) != 2 { return nil,fmt.Errorf("Assignment to empty attribute") }
					case "\n":
						if len(state) == 0 { continue }
						if len(state) != 3 { return nil,fmt.Errorf("Invalid attribute assignment : %v", state) }
						err := addResult(state[0], state[1], state[2])
						if err != nil { return nil,err }
						state = make([]string, 0)
					default:
						state = append(state, tk)
				}
		}
	}
	result := make([]map[string]string, 0)
	var keys []string
	for k := range unsorted { keys = append(keys, k) }
	sort.Strings(keys)
	for _,k := range keys {
		result = append(result, unsorted[k])
	}
	return result,nil
}

/** higher-level parsing */
/// parameters
type pParameter interface { repr() string }

type pParameterInclude struct {
	filename string
}
func (e pParameterInclude) repr() string { return fmt.Sprintf("include-file:filename=%s",e.filename) }

type pParameterOption struct {
	name string
	value string
}
func (e pParameterOption) repr() string { return fmt.Sprintf("option:%s=%s",e.name,e.value) }

// allow some-kind-of-something
type pParameterGrant struct {
	verb string 		// allow,deny,ignore
	attribute string
}
func (e pParameterGrant) repr() string { return fmt.Sprintf("grant:%s,%s",e.verb,e.attribute) }

type pParameterAddress4 []string
func (e pParameterAddress4) repr() string {
	return fmt.Sprintf("fixed-address4:%s",strings.Join(e,","))
}

type pParameterAddress6 []string
func (e pParameterAddress6) repr() string {
	return fmt.Sprintf("fixed-address6:%s",strings.Join(e,","))
}

// hardware address 00:00:00:00:00:00
type pParameterHardware struct {
	class string
	address []byte
}
func (e pParameterHardware) repr() string {
	res := make([]string, 0)
	for _,v := range e.address {
		res = append(res, fmt.Sprintf("%02x",v))
	}
	return fmt.Sprintf("hardware-address:%s[%s]",e.class,strings.Join(res,":"))
}

type pParameterBoolean struct {
	parameter string
	truancy bool
}
func (e pParameterBoolean) repr() string { return fmt.Sprintf("boolean:%s=%s",e.parameter,e.truancy) }

type pParameterClientMatch struct {
	name string
	data string
}
func (e pParameterClientMatch) repr() string { return fmt.Sprintf("match-client:%s=%s",e.name,e.data) }

// range 127.0.0.1 127.0.0.255
type pParameterRange4 struct {
	min net.IP
	max net.IP
}
func (e pParameterRange4) repr() string { return fmt.Sprintf("range4:%s-%s",e.min.String(),e.max.String()) }

type pParameterRange6 struct {
	min net.IP
	max net.IP
}
func (e pParameterRange6) repr() string { return fmt.Sprintf("range6:%s-%s",e.min.String(),e.max.String()) }

type pParameterPrefix6 struct {
	min net.IP
	max net.IP
	bits int
}
func (e pParameterPrefix6) repr() string { return fmt.Sprintf("prefix6:/%d:%s-%s",e.bits,e.min.String(),e.max.String()) }

// some-kind-of-parameter 1024
type pParameterOther struct {
	parameter string
	value string
}
func (e pParameterOther) repr() string { return fmt.Sprintf("parameter:%s=%s",e.parameter,e.value) }

type pParameterExpression struct {
	parameter string
	expression string
}
func (e pParameterExpression) repr() string { return fmt.Sprintf("parameter-expression:%s=\"%s\"",e.parameter,e.expression) }

type pDeclarationIdentifier interface { repr() string }

type pDeclaration struct {
	id pDeclarationIdentifier
	parent *pDeclaration
	parameters []pParameter
	declarations []pDeclaration
}

func (e *pDeclaration) short() string {
	return e.id.repr()
}

func (e *pDeclaration) repr() string {
	res := e.short()

	var parameters []string
	for _,v := range e.parameters {
		parameters = append(parameters, v.repr())
	}

	var groups []string
	for _,v := range e.declarations {
		groups = append(groups, fmt.Sprintf("-> %s",v.short()))
	}

	if e.parent != nil {
		res = fmt.Sprintf("%s parent:%s",res,e.parent.short())
	}
	return fmt.Sprintf("%s\n%s\n%s\n", res, strings.Join(parameters,"\n"), strings.Join(groups,"\n"))
}

type pDeclarationGlobal struct {}
func (e pDeclarationGlobal) repr() string { return fmt.Sprintf("{global}") }

type pDeclarationShared struct { name string }
func (e pDeclarationShared) repr() string { return fmt.Sprintf("{shared-network %s}", e.name) }

type pDeclarationSubnet4 struct { net.IPNet }
func (e pDeclarationSubnet4) repr() string { return fmt.Sprintf("{subnet4 %s}", e.String()) }

type pDeclarationSubnet6 struct { net.IPNet }
func (e pDeclarationSubnet6) repr() string { return fmt.Sprintf("{subnet6 %s}", e.String()) }

type pDeclarationHost struct { name string }
func (e pDeclarationHost) repr() string { return fmt.Sprintf("{host name:%s}", e.name) }

type pDeclarationPool struct {}
func (e pDeclarationPool) repr() string { return fmt.Sprintf("{pool}") }

type pDeclarationGroup struct {}
func (e pDeclarationGroup) repr() string { return fmt.Sprintf("{group}") }

type pDeclarationClass struct { name string }
func (e pDeclarationClass) repr() string { return fmt.Sprintf("{class}") }

/** parsers */
func parseParameter(val tkParameter) (pParameter,error) {
	switch val.name {
		case "include":
			if len(val.operand) != 2 {
				return nil,fmt.Errorf("Invalid number of parameters for pParameterInclude : %v",val.operand)
			}
			name := val.operand[0]
			return pParameterInclude{filename: name},nil

		case "option":
			if len(val.operand) != 2 {
				return nil,fmt.Errorf("Invalid number of parameters for pParameterOption : %v",val.operand)
			}
			name, value := val.operand[0], val.operand[1]
			return pParameterOption{name: name, value: value},nil

		case "allow": fallthrough
		case "deny": fallthrough
		case "ignore":
			if len(val.operand) < 1 {
				return nil,fmt.Errorf("Invalid number of parameters for pParameterGrant : %v",val.operand)
			}
			attribute := strings.Join(val.operand," ")
			return pParameterGrant{verb: strings.ToLower(val.name), attribute: attribute},nil

		case "range":
			if len(val.operand) < 1 {
				return nil,fmt.Errorf("Invalid number of parameters for pParameterRange4 : %v",val.operand)
			}
			idxAddress := map[bool]int{true:1,false:0}[strings.ToLower(val.operand[0]) == "bootp"]
			if len(val.operand) > 2 + idxAddress {
				return nil,fmt.Errorf("Invalid number of parameters for pParameterRange : %v",val.operand)
			}
			if idxAddress + 1 > len(val.operand) {
				res := net.ParseIP(val.operand[idxAddress])
				return pParameterRange4{min: res, max: res},nil
			}
			addr1 := net.ParseIP(val.operand[idxAddress])
			addr2 := net.ParseIP(val.operand[idxAddress+1])
			return pParameterRange4{min: addr1, max: addr2},nil

		case "range6":
			if len(val.operand) == 1 {
				address := val.operand[0]
				if (strings.Contains(address, "/")) {
					cidr := strings.SplitN(address, "/", 2)
					if len(cidr) != 2 { return nil,fmt.Errorf("Unknown ipv6 format : %v", address) }
					address := net.ParseIP(cidr[0])
					bits,err := strconv.Atoi(cidr[1])
					if err != nil { return nil,err }
					mask := net.CIDRMask(bits, net.IPv6len*8)

					// figure out the network address
					network := address.Mask(mask)

					// make a broadcast address
					broadcast := network
					networkSize,totalSize := mask.Size()
					hostSize := totalSize-networkSize
					for i := networkSize / 8; i < totalSize / 8; i++ {
						broadcast[i] = byte(0xff)
					}
					octetIndex := network[networkSize / 8]
					bitsLeft := (uint32)(hostSize%8)
					broadcast[octetIndex] = network[octetIndex] | ((1<<bitsLeft)-1)

					// FIXME: check that the broadcast address was made correctly
					return pParameterRange6{min: network, max: broadcast},nil
				}
				res := net.ParseIP(address)
				return pParameterRange6{min: res, max:res},nil
			}
			if len(val.operand) == 2 {
				addr := net.ParseIP(val.operand[0])
				if strings.ToLower(val.operand[1]) == "temporary" {
					return pParameterRange6{min: addr, max: addr},nil
				}
				other := net.ParseIP(val.operand[1])
				return pParameterRange6{min: addr, max: other},nil
			}
			return nil,fmt.Errorf("Invalid number of parameters for pParameterRange6 : %v",val.operand)

		case "prefix6":
			if len(val.operand) != 3 {
				return nil,fmt.Errorf("Invalid number of parameters for pParameterRange6 : %v",val.operand)
			}
			bits,err := strconv.Atoi(val.operand[2])
			if err != nil {
				return nil,fmt.Errorf("Invalid bits for pParameterPrefix6 : %v",val.operand[2])
			}
			minaddr := net.ParseIP(val.operand[0])
			maxaddr := net.ParseIP(val.operand[1])
			return pParameterPrefix6{min: minaddr, max: maxaddr, bits:bits},nil

		case "hardware":
			if len(val.operand) != 2 {
				return nil,fmt.Errorf("Invalid number of parameters for pParameterHardware : %v",val.operand)
			}
			class := val.operand[0]
			octets := strings.Split(val.operand[1], ":")
			address := make([]byte, 0)
			for _,v := range octets {
				b,err := strconv.ParseInt(v, 16, 0)
				if err != nil { return nil,err }
				address = append(address, byte(b))
			}
			return pParameterHardware{class: class, address: address},nil

		case "fixed-address":
			ip4addrs := make(pParameterAddress4,len(val.operand))
			copy(ip4addrs, val.operand)
			return ip4addrs,nil

		case "fixed-address6":
			ip6addrs := make(pParameterAddress6,len(val.operand))
			copy(ip6addrs, val.operand)
			return ip6addrs,nil

		case "host-identifier":
			if len(val.operand) != 3 {
				return nil,fmt.Errorf("Invalid number of parameters for pParameterClientMatch : %v",val.operand)
			}
			if val.operand[0] != "option" {
				return nil,fmt.Errorf("Invalid match parameter : %v",val.operand[0])
			}
			optionName := val.operand[1]
			optionData := val.operand[2]
			return pParameterClientMatch{name: optionName, data: optionData},nil

		default:
			length := len(val.operand)
			if length < 1 {
				return pParameterBoolean{parameter: val.name, truancy: true},nil
			} else if length > 1 {
				if val.operand[0] == "=" {
					return pParameterExpression{parameter: val.name, expression: strings.Join(val.operand[1:],"")},nil
				}
			}
			if length != 1 {
				return nil,fmt.Errorf("Invalid number of parameters for pParameterOther : %v",val.operand)
			}
			if strings.ToLower(val.name) == "not" {
				return pParameterBoolean{parameter: val.operand[0], truancy: false},nil
			}
			return pParameterOther{parameter: val.name, value: val.operand[0]}, nil
	}
}

func parseTokenGroup(val tkGroup) (*pDeclaration,error) {
	params := val.id.operand
	switch val.id.name {
		case "group":
			return &pDeclaration{id:pDeclarationGroup{}},nil

		case "pool":
			return &pDeclaration{id:pDeclarationPool{}},nil

		case "host":
			if len(params) == 1 {
				return &pDeclaration{id:pDeclarationHost{name: params[0]}},nil
			}

		case "subnet":
			if len(params) == 3 && strings.ToLower(params[1]) == "netmask" {
				addr := make([]byte, 4)
				for i,v := range strings.SplitN(params[2], ".", 4) {
					res,err := strconv.ParseInt(v, 10, 0)
					if err != nil { return nil,err }
					addr[i] = byte(res)
				}
				oc1,oc2,oc3,oc4 := addr[0],addr[1],addr[2],addr[3]
				if subnet,mask := net.ParseIP(params[0]),net.IPv4Mask(oc1,oc2,oc3,oc4); subnet != nil && mask != nil {
					return &pDeclaration{id:pDeclarationSubnet4{net.IPNet{IP:subnet,Mask:mask}}},nil
				}
			}
		case "subnet6":
			if len(params) == 1 {
				ip6 := strings.SplitN(params[0], "/", 2)
				if len(ip6) == 2 && strings.Contains(ip6[0], ":") {
					address := net.ParseIP(ip6[0])
					prefix,err := strconv.Atoi(ip6[1])
					if err != nil { return nil, err }
					return &pDeclaration{id:pDeclarationSubnet6{net.IPNet{IP:address,Mask:net.CIDRMask(prefix, net.IPv6len*8)}}},nil
				}
			}
		case "shared-network":
			if len(params) == 1 {
				return &pDeclaration{id:pDeclarationShared{name: params[0]}},nil
			}
		case "":
			return &pDeclaration{id:pDeclarationGlobal{}},nil
	}
	return nil,fmt.Errorf("Invalid pDeclaration : %v : %v", val.id.name, params)
}

func flattenDhcpConfig(root tkGroup) (*pDeclaration,error) {
	var result *pDeclaration
	result,err := parseTokenGroup(root)
	if err != nil { return nil,err }

	for _,p := range root.params {
		param,err := parseParameter(p)
		if err != nil { return nil,err }
		result.parameters = append(result.parameters, param)
	}
	for _,p := range root.groups {
		group,err := flattenDhcpConfig(*p)
		if err != nil { return nil,err }
		group.parent = result
		result.declarations = append(result.declarations, *group)
	}
	return result,nil
}

/** reduce the tree into the things that we care about */
type grant uint
const (
	ALLOW grant = iota
	IGNORE grant = iota
	DENY grant = iota
)
type configDeclaration struct {
	id []pDeclarationIdentifier
	composites []pDeclaration

	address []pParameter

	options map[string]string
	grants map[string]grant
	attributes map[string]bool
	parameters map[string]string
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
	for i := len(hierarchy)-1; i >= 0; i-- {
		result.composites = append(result.composites, hierarchy[(len(hierarchy)-1) - i])
		result.id = append(result.id, hierarchy[(len(hierarchy)-1) - i].id)

		// update configDeclaration parameters
		for _,p := range hierarchy[i].parameters {
			switch p.(type) {
				case pParameterOption:
					result.options[p.(pParameterOption).name] = p.(pParameterOption).value
				case pParameterGrant:
					Grant := map[string]grant{"ignore":IGNORE, "allow":ALLOW, "deny":DENY}
					result.grants[p.(pParameterGrant).attribute] = Grant[p.(pParameterGrant).verb]
				case pParameterBoolean:
					result.attributes[p.(pParameterBoolean).parameter] = p.(pParameterBoolean).truancy
				case pParameterClientMatch:
					result.hostid = append(result.hostid, p.(pParameterClientMatch))
				case pParameterExpression:
					result.expressions[p.(pParameterExpression).parameter] = p.(pParameterExpression).expression
				case pParameterOther:
					result.parameters[p.(pParameterOther).parameter] = p.(pParameterOther).value
				default:
					result.address = append(result.address, p)
			}
		}
	}
	return result
}

func (e *configDeclaration) repr() string {
	var result []string

	var res []string

	res = make([]string, 0)
	for _,v := range e.id { res = append(res, v.repr()) }
	result = append(result, strings.Join(res, ","))

	if len(e.address) > 0 {
		res = make([]string, 0)
		for _,v := range e.address { res = append(res, v.repr()) }
		result = append(result, fmt.Sprintf("address : %v", strings.Join(res, ",")))
	}

	if len(e.options) > 0 		{ result = append(result, fmt.Sprintf("options : %v", e.options)) }
	if len(e.grants) > 0 		{ result = append(result, fmt.Sprintf("grants : %v", e.grants)) }
	if len(e.attributes) > 0 	{ result = append(result, fmt.Sprintf("attributes : %v", e.attributes)) }
	if len(e.parameters) > 0 	{ result = append(result, fmt.Sprintf("parameters : %v", e.parameters)) }
	if len(e.expressions) > 0 	{ result = append(result, fmt.Sprintf("parameter-expressions : %v", e.expressions)) }

	if len(e.hostid) > 0 {
		res = make([]string, 0)
		for _,v := range e.hostid { res = append(res, v.repr()) }
		result = append(result, fmt.Sprintf("hostid : %v", strings.Join(res, " ")))
	}
	return strings.Join(result, "\n") + "\n"
}

func (e *configDeclaration) IP4() (net.IP,error) {
	var result []string
	for _,entry := range e.address {
		switch entry.(type) {
			case pParameterAddress4:
				for _,s := range entry.(pParameterAddress4) {
					result = append(result, s)
				}
		}
	}
	if len(result) > 1 {
		return nil,fmt.Errorf("More than one address4 returned : %v", result)
	} else if len(result) == 0 {
		return nil,fmt.Errorf("No IP4 addresses found")
	}

	if res := net.ParseIP(result[0]); res != nil { return res,nil }
	res,err := net.ResolveIPAddr("ip4", result[0])
	if err != nil { return nil,err }
	return res.IP,nil
}
func (e *configDeclaration) IP6() (net.IP,error) {
	var result []string
	for _,entry := range e.address {
		switch entry.(type) {
			case pParameterAddress6:
				for _,s := range entry.(pParameterAddress6) {
					result = append(result, s)
				}
		}
	}
	if len(result) > 1 {
		return nil,fmt.Errorf("More than one address6 returned : %v", result)
	} else if len(result) == 0 {
		return nil,fmt.Errorf("No IP6 addresses found")
	}

	if res := net.ParseIP(result[0]); res != nil { return res,nil }
	res,err := net.ResolveIPAddr("ip6", result[0])
	if err != nil { return nil,err }
	return res.IP,nil
}
func (e *configDeclaration) Hardware() (net.HardwareAddr,error) {
	var result []pParameterHardware
	for _,addr := range e.address {
		switch addr.(type) {
			case pParameterHardware:
				result = append(result, addr.(pParameterHardware))
		}
	}
	if len(result) > 0 {
		return nil,fmt.Errorf("More than one hardware address returned : %v", result)
	}
	res := make(net.HardwareAddr, 0)
	for _,by := range result[0].address {
		res = append(res, by)
	}
	return res,nil
}

/*** Dhcp Configuration */
type DhcpConfiguration []configDeclaration
func ReadDhcpConfiguration(fd *os.File) (DhcpConfiguration,error) {
	fromfile,eof := consumeFile(fd)
	uncommented := uncomment(eof, fromfile)
	tokenized := tokenizeDhcpConfig(eof, uncommented)
	parsetree,err := parseDhcpConfig(eof, tokenized)
	if err != nil { return nil,err }

	global,err := flattenDhcpConfig(parsetree)
	if err != nil { return nil,err }

	var walkDeclarations func(root pDeclaration, out chan*configDeclaration);
	walkDeclarations = func(root pDeclaration, out chan*configDeclaration) {
		res := createDeclaration(root)
		out <- &res
		for _,p := range root.declarations {
			walkDeclarations(p, out)
		}
	}

	each := make(chan*configDeclaration)
	go func(out chan*configDeclaration) {
		walkDeclarations(*global, out)
		out <- nil
	}(each)

	var result DhcpConfiguration
	for decl := <-each; decl != nil; decl = <-each {
		result = append(result, *decl)
	}
	return result,nil
}

func (e *DhcpConfiguration) Global() configDeclaration {
	result := (*e)[0]
	if len(result.id) != 1 {
		panic(fmt.Errorf("Something that can't happen happened"))
	}
	return result
}

func (e *DhcpConfiguration) SubnetByAddress(address net.IP) (configDeclaration,error) {
	var result []configDeclaration
	for _,entry := range *e {
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
		return configDeclaration{},fmt.Errorf("No network declarations containing %s found", address.String())
	}
	if len(result) > 1 {
		return configDeclaration{},fmt.Errorf("More than 1 network declaration found : %v", result)
	}
	return result[0],nil
}

func (e *DhcpConfiguration) HostByName(host string) (configDeclaration,error) {
	var result []configDeclaration
	for _,entry := range *e {
		switch entry.id[0].(type) {
			case pDeclarationHost:
				id := entry.id[0].(pDeclarationHost)
				if strings.ToLower(id.name) == strings.ToLower(host) {
					result = append(result, entry)
				}
		}
	}
	if len(result) == 0 {
		return configDeclaration{},fmt.Errorf("No host declarations containing %s found", host)
	}
	if len(result) > 1 {
		return configDeclaration{},fmt.Errorf("More than 1 host declaration found : %v", result)
	}
	return result[0],nil
}

/*** Network Map */
type NetworkMap []map[string]string
func ReadNetworkMap(fd *os.File) (NetworkMap,error) {

	fromfile,eof := consumeFile(fd)
	uncommented := uncomment(eof,fromfile)
	tokenized := tokenizeNetworkMapConfig(eof, uncommented)

	result,err := parseNetworkMapConfig(eof, tokenized)
	if err != nil { return nil,err }
	return result,nil
}

func (e *NetworkMap) NameIntoDevice(name string) (string,error) {
	for _,val := range *e {
		if strings.ToLower(val["name"]) == strings.ToLower(name) {
			return val["device"],nil
		}
	}
	return "",fmt.Errorf("Network name not found : %v", name)
}
func (e *NetworkMap) DeviceIntoName(device string) (string,error) {
	for _,val := range *e {
		if strings.ToLower(val["device"]) == strings.ToLower(device) {
			return val["name"],nil
		}
	}
	return "",fmt.Errorf("Device name not found : %v", device)
}
func (e *NetworkMap) repr() string {
	var result []string
	for idx,val := range *e {
		result = append(result, fmt.Sprintf("network%d.name = \"%s\"", idx, val["name"]))
		result = append(result, fmt.Sprintf("network%d.device = \"%s\"", idx, val["device"]))
	}
	return strings.Join(result, "\n")
}

/** main */
func consumeFile(fd *os.File) (chan byte,sentinelSignaller) {
	fromfile := make(chan byte)
	eof := make(sentinelSignaller)
	go func() {
		b := make([]byte, 1)
		for {
			_,err := fd.Read(b)
			if err == io.EOF { break }
			fromfile <- b[0]
		}
		close(eof)
	}()
	return fromfile,eof
}
