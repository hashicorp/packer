// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkresolvers

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/yandex-cloud/go-sdk/pkg/sdkerrors"
)

const DefaultResolverPageSize = 100

func CreateResolverFilter(nameField string, value string) string {
	// TODO(novikoff): should we add escaping or value validation?
	return fmt.Sprintf(`%s = "%s"`, nameField, value)
}

type resolveOptions struct {
	out          *string
	folderID     string
	cloudID      string
	clusterID    string
	federationID string
}

type ResolveOption func(*resolveOptions)

func Out(out *string) ResolveOption {
	return func(o *resolveOptions) {
		o.out = out
	}
}

// FolderID specifies folder id for resolvers that need it (most of the resolvers).
func FolderID(folderID string) ResolveOption {
	return func(o *resolveOptions) {
		o.folderID = folderID
	}
}

// CloudID specifies cloud id for resolvers that need it, e.g. FolderResolver
func CloudID(cloudID string) ResolveOption {
	return func(o *resolveOptions) {
		o.cloudID = cloudID
	}
}

// ClusterID specifies cluster id for resolvers that need it, e.g. DataprocSubclusterResolver
func ClusterID(clusterID string) ResolveOption {
	return func(o *resolveOptions) {
		o.clusterID = clusterID
	}
}

// FederationID specifies federation id for resolvers that need it, e.g. CertificateResolver
func FederationID(federationID string) ResolveOption {
	return func(o *resolveOptions) {
		o.federationID = federationID
	}
}

func combineOpts(opts ...ResolveOption) *resolveOptions {
	o := &resolveOptions{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type BaseResolver struct {
	Name string
	id   string
	err  error
	opts *resolveOptions
}

func NewBaseResolver(name string, opts ...ResolveOption) BaseResolver {
	return BaseResolver{
		Name: name,
		opts: combineOpts(opts...),
	}
}

type BaseNameResolver struct {
	BaseResolver

	resolvingObjectType string
}

func NewBaseNameResolver(name string, resolvingObjectType string, opts ...ResolveOption) BaseNameResolver {
	return BaseNameResolver{
		BaseResolver:        NewBaseResolver(name, opts...),
		resolvingObjectType: resolvingObjectType,
	}
}

func (r *BaseResolver) ID() string {
	return r.id
}
func (r *BaseResolver) Err() error {
	return r.err
}

func (r *BaseResolver) SetErr(err error) error {
	if r.err != nil {
		panic(fmt.Sprintf("Trying to change error. Old: %v; New: %v", r.err, err))
	}
	r.err = err
	return r.err
}

func (r *BaseResolver) SetID(id string) {
	r.id = id
	r.writeOut()
}

func (r *BaseResolver) Set(entity Entity, err error) error {
	if err != nil {
		return r.SetErr(err)
	}
	r.SetID(entity.GetId())
	return nil
}

type Entity interface {
	GetId() string
}

func (r *BaseResolver) FolderID() string {
	return r.opts.folderID
}

func (r *BaseResolver) CloudID() string {
	return r.opts.cloudID
}

func (r *BaseResolver) ClusterID() string {
	return r.opts.clusterID
}

func (r *BaseResolver) FederationID() string {
	return r.opts.federationID
}

func (r *BaseResolver) coordinates() string {
	buf := bytes.Buffer{}
	if r.FederationID() != "" {
		buf.WriteString(fmt.Sprintf("in the federation \"%s\" ", r.FederationID()))
	}
	if r.ClusterID() != "" {
		buf.WriteString(fmt.Sprintf("in the cluster \"%s\" ", r.ClusterID()))
	}
	if r.CloudID() != "" {
		buf.WriteString(fmt.Sprintf("in the cloud \"%s\" ", r.CloudID()))
	}
	if r.FolderID() != "" {
		buf.WriteString(fmt.Sprintf("in the folder \"%s\" ", r.FolderID()))
	}
	return strings.TrimSpace(buf.String())
}

func (r *BaseResolver) writeOut() {
	if r.opts.out != nil {
		*r.opts.out = r.id
	}
}

func (r *BaseNameResolver) findName(slice interface{}, err error) error {
	return r.SetErr(r.findNameImpl(slice, err))
}

func (r *BaseNameResolver) ensureFolderID() error {
	if r.FolderID() == "" {
		err := &ErrNotFound{error: fmt.Sprintf("can't resolve %v without folder id specified", r.resolvingObjectType)}
		return r.SetErr(err)
	}

	return nil
}

func (r *BaseNameResolver) ensureCloudID() error {
	if r.CloudID() == "" {
		err := &ErrNotFound{error: fmt.Sprintf("can't resolve %v without cloud id specified", r.resolvingObjectType)}
		return r.SetErr(err)
	}

	return nil
}

func NewErrNotFound(err string) error {
	return &ErrNotFound{error: err}
}

type ErrNotFound struct {
	error string
}

func (e *ErrNotFound) Error() string {
	return e.error
}

func errNotFound(caption, name string) error {
	return &ErrNotFound{error: fmt.Sprintf("%v with name \"%v\" not found", caption, name)}
}

func (r *BaseNameResolver) findNameImpl(slice interface{}, err error) error {
	if err != nil {
		return sdkerrors.WithMessagef(err, "failed to find %v with name \"%v\" %v", r.resolvingObjectType, r.Name, r.coordinates())
	}
	rv := reflect.ValueOf(slice)
	var found nameAndID
	for i := 0; i < rv.Len(); i++ {
		v := rv.Index(i).Interface().(nameAndID)
		if v.GetName() == r.Name {
			if found != nil {
				return fmt.Errorf("multiple %v items with name \"%v\" found %v", r.resolvingObjectType, r.Name, r.coordinates())
			}
			found = v
		}
	}
	if found == nil {
		return errNotFound(r.resolvingObjectType, r.Name)
	}
	r.SetID(found.GetId())
	return nil
}

type nameAndID interface {
	GetId() string
	GetName() string
}
