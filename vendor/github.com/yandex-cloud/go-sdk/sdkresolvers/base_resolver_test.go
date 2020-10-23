// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkresolvers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	compute "github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func TestBaseNameResolverFindName(t *testing.T) {
	base := func() *BaseNameResolver {
		x := &BaseNameResolver{
			BaseResolver:        BaseResolver{Name: "name1"},
			resolvingObjectType: "test",
		}
		x.opts = &resolveOptions{
			out:      &x.id,
			folderID: "folder_id_value",
		}
		return x
	}
	t.Run("only one correct name", func(t *testing.T) {
		x := base()
		err := x.findName([]*compute.Instance{
			{Id: "id1", Name: "name1"},
		}, nil)
		require.NoError(t, err)
		assert.Equal(t, x.id, "id1")
	})
	t.Run("two records with same name", func(t *testing.T) {
		x := base()
		err := x.findName([]*compute.Instance{
			{Id: "id1", Name: "name1"},
			{Id: "id2", Name: "name1"},
		}, nil)
		require.Error(t, err)
		assert.Equal(t, "multiple test items with name \"name1\" found in the folder \"folder_id_value\"", err.Error())
	})
	t.Run("two records with same name but not found", func(t *testing.T) {
		x := base()
		err := x.findName([]*compute.Instance{
			{Id: "id1", Name: "name2"},
			{Id: "id2", Name: "name2"},
		}, nil)
		require.Error(t, err)
		assert.Equal(t, errNotFound("test", "name1"), err)
	})
	t.Run("resolve error", func(t *testing.T) {
		x := base()
		err := x.findName(nil, errors.New("forward this"))
		require.Error(t, err)
		assert.Equal(t, "failed to find test with name \"name1\" in the folder \"folder_id_value\": forward this", err.Error())
	})
	t.Run("multiple items returned 1", func(t *testing.T) {
		x := base()
		err := x.findName([]*compute.Instance{
			{Id: "id1", Name: "name1"},
			{Id: "id2", Name: "name2"},
		}, nil)
		require.NoError(t, err)
		assert.Equal(t, x.id, "id1")
	})
	t.Run("multiple items returned 2", func(t *testing.T) {
		x := base()
		err := x.findName([]*compute.Instance{
			{Id: "id2", Name: "name2"},
			{Id: "id1", Name: "name1"},
		}, nil)
		require.NoError(t, err)
		assert.Equal(t, x.id, "id1")
	})
}
