package client

import (
	"reflect"
	"testing"
)

func TestParseResourceID(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		want       Resource
		wantErr    bool
	}{
		{
			name:       "happy path",
			resourceID: "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/resourceGroups/rg/providers/Microsoft.Resources/resources/resource",
			want: Resource{
				Subscription:  "17c60680-0e49-465b-aa54-ece043ce5571",
				ResourceGroup: "rg",
				Provider:      "Microsoft.Resources",
				ResourceType:  CompoundName{"resources"},
				ResourceName:  CompoundName{"resource"},
			},
		},
		{
			name:       "sub resource",
			resourceID: "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/resourcegroups/rg/providers/Microsoft.Resources/resources/resource/subResources/child",
			want: Resource{
				Subscription:  "17c60680-0e49-465b-aa54-ece043ce5571",
				ResourceGroup: "rg",
				Provider:      "Microsoft.Resources",
				ResourceType:  CompoundName{"resources", "subResources"},
				ResourceName:  CompoundName{"resource", "child"},
			},
		},
		{
			name:       "incomplete",
			resourceID: "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/resourceGroups/rg/providers/Microsoft.Resources/resources/resource/subResources",
			wantErr:    true,
		},
		{
			name:       "incomplete 2",
			resourceID: "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/resourceGroups/rg/providers/Microsoft.Resources/resources/",
			wantErr:    true,
		},
		{
			name:       "extra slash",
			resourceID: "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/resourceGroups/rg/providers/Microsoft.Resources/resources//resources",
			wantErr:    true,
		},
		{
			name:       "empty resource name",
			resourceID: "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/resourceGroups/rg/providers/Microsoft.Resources/resources//subresources/child",
			wantErr:    true,
		},
		{
			name:       "empty sub resource type",
			resourceID: "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/resourceGroups/rg/providers/Microsoft.Resources/resources/resource//child",
			wantErr:    true,
		},
		{
			name:       "ungrouped resource path",
			resourceID: "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/providers/Microsoft.Resources/resources/resource",
			wantErr:    true,
		},
		{
			name:       "misspelled subscriptions",
			resourceID: "/subscription/17c60680-0e49-465b-aa54-ece043ce5571/resourceGroups/rg/providers/Microsoft.Resources/resources/resource/subResources/child",
			wantErr:    true,
		},
		{
			name:       "misspelled resourceGroups",
			resourceID: "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/resourceGroup/rg/providers/Microsoft.Resources/resources/resource/subResources/child",
			wantErr:    true,
		},
		{
			name:       "misspelled providers",
			resourceID: "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/resourceGroups/rg/provider/Microsoft.Resources/resources/resource/subResources/child",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseResourceID(tt.resourceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseResourceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseResourceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResource_String(t *testing.T) {
	type fields struct {
		Subscription  string
		ResourceGroup string
		Provider      string
		ResourceType  CompoundName
		ResourceName  CompoundName
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "happy path",
			fields: fields{
				Subscription:  "sub",
				ResourceGroup: "rg",
				Provider:      "provider",
				ResourceType:  CompoundName{"type"},
				ResourceName:  CompoundName{"name"},
			},
			want: "/subscriptions/sub/resourceGroups/rg/providers/provider/type/name",
		},
		{
			name: "happy path - child resource",
			fields: fields{
				Subscription:  "sub",
				ResourceGroup: "rg",
				Provider:      "provider",
				ResourceType:  CompoundName{"type", "sub"},
				ResourceName:  CompoundName{"name", "child"},
			},
			want: "/subscriptions/sub/resourceGroups/rg/providers/provider/type/name/sub/child",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Resource{
				Subscription:  tt.fields.Subscription,
				ResourceGroup: tt.fields.ResourceGroup,
				Provider:      tt.fields.Provider,
				ResourceType:  tt.fields.ResourceType,
				ResourceName:  tt.fields.ResourceName,
			}
			if got := r.String(); got != tt.want {
				t.Errorf("Resource.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResource_Parent(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		want       string
		wantErr    bool
	}{
		{
			name:       "happy path",
			resourceID: "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/resourceGroups/rg/providers/Microsoft.Resources/resources/resource/sub/child",
			want:       "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/resourceGroups/rg/providers/Microsoft.Resources/resources/resource",
		},
		{
			name:       "sub sub",
			resourceID: "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/resourceGroups/rg/providers/Microsoft.Resources/resources/resource/sub/child/subsub/grandchild",
			want:       "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/resourceGroups/rg/providers/Microsoft.Resources/resources/resource/sub/child",
		},
		{
			name:       "top level resource",
			resourceID: "/subscriptions/17c60680-0e49-465b-aa54-ece043ce5571/resourceGroups/rg/providers/Microsoft.Resources/resources/resource",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ParseResourceID(tt.resourceID)
			if err != nil {
				t.Fatalf("Error parsing test resource: %v", err)
			}
			got, err := r.Parent()
			if (err != nil) != tt.wantErr {
				t.Errorf("Resource.Parent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got.String() != tt.want {
				t.Errorf("Resource.Parent() = %v, want %v", got, tt.want)
			}
		})
	}
}
