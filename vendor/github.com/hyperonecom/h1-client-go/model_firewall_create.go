/*
 * HyperOne API
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 0.0.2
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type FirewallCreate struct {
	Name string `json:"name,omitempty"`
	Tag map[string]interface{} `json:"tag,omitempty"`
	Service string `json:"service,omitempty"`
}
