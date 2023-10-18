// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"log"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	// Previously core-bundled components, split into their own plugins but
	// still vendored with Packer for now. Importing as library instead of
	// forcing use of packer init.
)

// VendoredDatasources are datasource components that were once bundled with the
// Packer core, but are now being imported from their counterpart plugin repos
var VendoredDatasources = map[string]packersdk.Datasource{}

// VendoredBuilders are builder components that were once bundled with the
// Packer core, but are now being imported from their counterpart plugin repos
var VendoredBuilders = map[string]packersdk.Builder{}

// VendoredProvisioners are provisioner components that were once bundled with the
// Packer core, but are now being imported from their counterpart plugin repos
var VendoredProvisioners = map[string]packersdk.Provisioner{}

// VendoredPostProcessors are post-processor components that were once bundled with the
// Packer core, but are now being imported from their counterpart plugin repos
var VendoredPostProcessors = map[string]packersdk.PostProcessor{}

// bundledStatus is used to know if one of the bundled components is loaded from
// an external plugin, or from the bundled plugins.
//
// We keep track of this to produce a warning if a user relies on one
// such plugin, as they will be removed in a later version of Packer.
var bundledStatus = map[string]bool{}

// TrackBundledPlugin marks a component as loaded from Packer's bundled plugins
// instead of from an externally loaded plugin.
//
// NOTE: `pluginName' must be in the format `packer-<type>-<component_name>'
func TrackBundledPlugin(pluginName string) {
	_, exists := bundledStatus[pluginName]
	if !exists {
		return
	}

	bundledStatus[pluginName] = true
}

var componentPluginMap = map[string]string{}

// compileBundledPluginList returns a list of plugins to import in a config
//
// This only works on bundled plugins and serves as a way to inform users that
// they should not rely on a bundled plugin anymore, but give them recommendations
// on how to manage those plugins instead.
func compileBundledPluginList(componentMap map[string]struct{}) []string {
	plugins := map[string]struct{}{}
	for component := range componentMap {
		plugin, ok := componentPluginMap[component]
		if !ok {
			log.Printf("Unknown bundled plugin component: %q", component)
			continue
		}

		plugins[plugin] = struct{}{}
	}

	pluginList := make([]string, 0, len(plugins))
	for plugin := range plugins {
		pluginList = append(pluginList, plugin)
	}

	return pluginList
}

// Upon init lets load up any plugins that were vendored manually into the default
// set of plugins.
func init() {
	for k, v := range VendoredDatasources {
		if _, ok := Datasources[k]; ok {
			continue
		}
		Datasources[k] = v
	}

	for k, v := range VendoredBuilders {
		if _, ok := Builders[k]; ok {
			continue
		}
		Builders[k] = v
	}

	for k, v := range VendoredProvisioners {
		if _, ok := Provisioners[k]; ok {
			continue
		}
		Provisioners[k] = v
	}

	for k, v := range VendoredPostProcessors {
		if _, ok := PostProcessors[k]; ok {
			continue
		}
		PostProcessors[k] = v
	}
}
