package hcl2template

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

// A refstring is any reference string that can point to a component of a config
//
// This includes anything in the following format:
//
// * data.<type>.<name>
// * var.<name>
// * local.<name>
type refString struct {
	// RefMType is the top-level label, that is used to know which type of
	// component we are to look for (i.e. data for Datasource, local for
	// Locals, etc.)
	MType string
	// RefType is the type of component, as in the name of the plugin that
	// will handle evaluation for the component.
	//
	// Only for Datasources now as var/local do not have a component type
	// (they're always evaluated by Packer itself, not a plugin).
	Type string
	// RefName is the name of the component to get.
	// For locals/vars this is the name of the variable to look for, while
	// for datasources this is the name of the block, which coupled with the
	// type is the identity of the datasource's execution.
	Name string
}

func NewRefStringFromDep(t hcl.Traversal) (refString, error) {
	root := t.RootName()

	switch root {
	case "local", "var":
		return NewRefString(fmt.Sprintf("%s.%s", root, t[1].(hcl.TraverseAttr).Name))
	case "data":
		return NewRefString(fmt.Sprintf("%s.%s.%s", root,
			t[1].(hcl.TraverseAttr).Name,
			t[2].(hcl.TraverseAttr).Name))
	}

	return refString{}, fmt.Errorf("unsupported refstring %q, must be of 'data', 'local' or 'var' type", t)
}

func NewRefString(rs string) (refString, error) {
	parts := strings.Split(rs, ".")

	switch parts[0] {
	case "local", "var":
		return refString{
			MType: parts[0],
			Name:  parts[1],
		}, nil
	case "data":
		return newDataSourceRefString(parts)
	}

	return refString{}, fmt.Errorf("unsupported reftype %q, must be either 'data', 'local' or 'var'", parts[0])
}

func (rs refString) String() string {
	if rs.Type == "" {
		return fmt.Sprintf("%s.%s", rs.MType, rs.Name)
	}

	return fmt.Sprintf("%s.%s.%s", rs.MType, rs.Type, rs.Name)
}

func newDataSourceRefString(parts []string) (refString, error) {
	if len(parts) != 3 {
		return refString{}, fmt.Errorf("malformed datasource reference %q, data sources must be composed of 3 parts",
			strings.Join(parts, "."))
	}

	return refString{
		MType: "data",
		Type:  parts[1],
		Name:  parts[2],
	}, nil
}

// getComponentByRef gets a registered component from the configuration from a refString
func (cfg *PackerConfig) getComponentByRef(rs refString) (interface{}, error) {
	switch rs.MType {
	case "data":
		for _, ds := range cfg.Datasources {
			if ds.Type != rs.Type {
				continue
			}
			if ds.DSName != rs.Name {
				continue
			}
			return ds, nil
		}
		return nil, fmt.Errorf("failed to get datasource '%s.%s': component unknown", rs.Type, rs.Name)
	case "local":
		for _, loc := range cfg.LocalBlocks {
			if loc.LocalName != rs.Name {
				continue
			}
			return loc, nil
		}
	case "var":
		for _, val := range cfg.InputVariables {
			if val.Name != rs.Name {
				continue
			}
			return val, nil
		}
	}

	return nil, fmt.Errorf("Unsupported component: %q, only vars, locals and datasources can be fetched by their refString", rs)
}

func (ds *DatasourceBlock) RegisterDependency(rs refString) error {
	switch rs.MType {
	case "data", "local":
		ds.Dependencies = append(ds.Dependencies, rs)
	// NOOP: vars are always evaluated beforehand for datasources
	case "var":
	default:
		return fmt.Errorf("unsupported dependency type %q; datasources can only depend on local, var or data.", rs.MType)
	}

	return nil
}

func (loc *LocalBlock) RegisterDependency(rs refString) error {
	switch rs.MType {
	case "data", "local":
		loc.dependencies = append(loc.dependencies, rs)
	// NOOP: vars are always evaluated beforehand for locals
	case "var":
	default:
		return fmt.Errorf("unsupported dependency type %q; locals can only depend on local, var or data.", rs.MType)
	}

	return nil
}
