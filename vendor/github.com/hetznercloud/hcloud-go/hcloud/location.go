package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

// Location represents a location in the Hetzner Cloud.
type Location struct {
	ID          int
	Name        string
	Description string
	Country     string
	City        string
	Latitude    float64
	Longitude   float64
}

// LocationClient is a client for the location API.
type LocationClient struct {
	client *Client
}

// GetByID retrieves a location by its ID.
func (c *LocationClient) GetByID(ctx context.Context, id int) (*Location, *Response, error) {
	req, err := c.client.NewRequest(ctx, "GET", fmt.Sprintf("/locations/%d", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.LocationGetResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}
	return LocationFromSchema(body.Location), resp, nil
}

// GetByName retrieves an location by its name.
func (c *LocationClient) GetByName(ctx context.Context, name string) (*Location, *Response, error) {
	path := "/locations?name=" + url.QueryEscape(name)
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.LocationListResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		return nil, nil, err
	}

	if len(body.Locations) == 0 {
		return nil, resp, nil
	}
	return LocationFromSchema(body.Locations[0]), resp, nil
}

// Get retrieves a location by its ID if the input can be parsed as an integer, otherwise it retrieves a location by its name.
func (c *LocationClient) Get(ctx context.Context, idOrName string) (*Location, *Response, error) {
	if id, err := strconv.Atoi(idOrName); err == nil {
		return c.GetByID(ctx, int(id))
	}
	return c.GetByName(ctx, idOrName)
}

// LocationListOpts specifies options for listing location.
type LocationListOpts struct {
	ListOpts
}

// List returns a list of locations for a specific page.
func (c *LocationClient) List(ctx context.Context, opts LocationListOpts) ([]*Location, *Response, error) {
	path := "/locations?" + valuesForListOpts(opts.ListOpts).Encode()
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.LocationListResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		return nil, nil, err
	}
	locations := make([]*Location, 0, len(body.Locations))
	for _, i := range body.Locations {
		locations = append(locations, LocationFromSchema(i))
	}
	return locations, resp, nil
}

// All returns all locations.
func (c *LocationClient) All(ctx context.Context) ([]*Location, error) {
	allLocations := []*Location{}

	opts := LocationListOpts{}
	opts.PerPage = 50

	_, err := c.client.all(func(page int) (*Response, error) {
		opts.Page = page
		locations, resp, err := c.List(ctx, opts)
		if err != nil {
			return resp, err
		}
		allLocations = append(allLocations, locations...)
		return resp, nil
	})
	if err != nil {
		return nil, err
	}

	return allLocations, nil
}
