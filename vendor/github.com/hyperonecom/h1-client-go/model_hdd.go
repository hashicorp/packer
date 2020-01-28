/*
 * HyperOne API
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 0.0.2
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// Hdd struct for Hdd
type Hdd struct {
	MaximumIOPS        float32 `json:"maximumIOPS,omitempty"`
	ControllerType     string  `json:"controllerType,omitempty"`
	ControllerNumber   string  `json:"controllerNumber,omitempty"`
	ControllerLocation float32 `json:"controllerLocation,omitempty"`
	Disk               HddDisk `json:"disk,omitempty"`
	Id                 string  `json:"id,omitempty"`
}
