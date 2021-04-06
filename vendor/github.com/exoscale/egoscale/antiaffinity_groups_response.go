// code generated; DO NOT EDIT.

package egoscale

import "fmt"

// Response returns the struct to unmarshal.
func (ListAntiAffinityGroups) Response() interface{} {
	return new(ListAntiAffinityGroupsResponse)
}

// ListRequest returns itself.
func (ls *ListAntiAffinityGroups) ListRequest() (ListCommand, error) {
	if ls == nil {
		return nil, fmt.Errorf("%T cannot be nil", ls)
	}
	return ls, nil
}

// SetPage sets the current page.
func (ls *ListAntiAffinityGroups) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size.
func (ls *ListAntiAffinityGroups) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// Each triggers the callback for each, valid answer or any non 404 issue.
func (ListAntiAffinityGroups) Each(resp interface{}, callback IterateItemFunc) {
	items, ok := resp.(*ListAntiAffinityGroupsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, ListAntiAffinityGroupsResponse was expected, got %T", resp))
		return
	}

	for i := range items.AntiAffinityGroup {
		if !callback(&items.AntiAffinityGroup[i], nil) {
			break
		}
	}
}
