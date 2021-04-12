package v2

import (
	"context"
)

// ListZones returns the list of Exoscale zones.
func (c *Client) ListZones(ctx context.Context) ([]string, error) {
	list := make([]string, 0)

	resp, err := c.ListZonesWithResponse(ctx)
	if err != nil {
		return nil, err
	}

	if resp.JSON200.Zones != nil {
		for i := range *resp.JSON200.Zones {
			zone := &(*resp.JSON200.Zones)[i]
			list = append(list, *zone.Name)
		}
	}

	return list, nil
}
