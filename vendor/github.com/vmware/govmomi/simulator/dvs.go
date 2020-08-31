/*
Copyright (c) 2017 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package simulator

import (
	"strconv"
	"strings"

	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

type DistributedVirtualSwitch struct {
	mo.DistributedVirtualSwitch
}

func (s *DistributedVirtualSwitch) AddDVPortgroupTask(ctx *Context, c *types.AddDVPortgroup_Task) soap.HasFault {
	task := CreateTask(s, "addDVPortgroup", func(t *Task) (types.AnyType, types.BaseMethodFault) {
		f := Map.getEntityParent(s, "Folder").(*Folder)

		portgroups := s.Portgroup
		portgroupNames := s.Summary.PortgroupName

		for _, spec := range c.Spec {
			pg := &DistributedVirtualPortgroup{}
			pg.Name = spec.Name
			pg.Entity().Name = pg.Name

			// Standard AddDVPortgroupTask() doesn't allow duplicate names, but NSX 3.0 does create some DVPGs with the same name.
			// Allow duplicate names using this prefix so we can reproduce and test this condition.
			if !strings.HasPrefix(pg.Name, "NSX-") {
				if obj := Map.FindByName(pg.Name, f.ChildEntity); obj != nil {
					return nil, &types.DuplicateName{
						Name:   pg.Name,
						Object: obj.Reference(),
					}
				}
			}

			folderPutChild(ctx, &f.Folder, pg)

			pg.Key = pg.Self.Value
			pg.Config = types.DVPortgroupConfigInfo{
				Key:                          pg.Key,
				Name:                         pg.Name,
				NumPorts:                     spec.NumPorts,
				DistributedVirtualSwitch:     &s.Self,
				DefaultPortConfig:            spec.DefaultPortConfig,
				Description:                  spec.Description,
				Type:                         spec.Type,
				Policy:                       spec.Policy,
				PortNameFormat:               spec.PortNameFormat,
				Scope:                        spec.Scope,
				VendorSpecificConfig:         spec.VendorSpecificConfig,
				ConfigVersion:                spec.ConfigVersion,
				AutoExpand:                   spec.AutoExpand,
				VmVnicNetworkResourcePoolKey: spec.VmVnicNetworkResourcePoolKey,
				LogicalSwitchUuid:            spec.LogicalSwitchUuid,
				BackingType:                  spec.BackingType,
			}

			if pg.Config.LogicalSwitchUuid != "" {
				if pg.Config.BackingType == "" {
					pg.Config.BackingType = "nsx"
				}
			}

			if pg.Config.DefaultPortConfig == nil {
				pg.Config.DefaultPortConfig = &types.VMwareDVSPortSetting{
					Vlan: new(types.VmwareDistributedVirtualSwitchVlanIdSpec),
					UplinkTeamingPolicy: &types.VmwareUplinkPortTeamingPolicy{
						Policy: &types.StringPolicy{
							Value: "loadbalance_srcid",
						},
						ReversePolicy: &types.BoolPolicy{
							Value: types.NewBool(true),
						},
						NotifySwitches: &types.BoolPolicy{
							Value: types.NewBool(true),
						},
						RollingOrder: &types.BoolPolicy{
							Value: types.NewBool(true),
						},
					},
				}
			}

			if pg.Config.Policy == nil {
				pg.Config.Policy = &types.VMwareDVSPortgroupPolicy{
					DVPortgroupPolicy: types.DVPortgroupPolicy{
						BlockOverrideAllowed:               true,
						ShapingOverrideAllowed:             false,
						VendorConfigOverrideAllowed:        false,
						LivePortMovingAllowed:              false,
						PortConfigResetAtDisconnect:        true,
						NetworkResourcePoolOverrideAllowed: types.NewBool(false),
						TrafficFilterOverrideAllowed:       types.NewBool(false),
					},
					VlanOverrideAllowed:           false,
					UplinkTeamingOverrideAllowed:  false,
					SecurityPolicyOverrideAllowed: false,
					IpfixOverrideAllowed:          types.NewBool(false),
				}
			}

			for i := 0; i < int(spec.NumPorts); i++ {
				pg.PortKeys = append(pg.PortKeys, strconv.Itoa(i))
			}

			portgroups = append(portgroups, pg.Self)
			portgroupNames = append(portgroupNames, pg.Name)

			for _, h := range s.Summary.HostMember {
				pg.Host = append(pg.Host, h)

				host := Map.Get(h).(*HostSystem)
				Map.AppendReference(host, &host.Network, pg.Reference())

				parent := Map.Get(*host.HostSystem.Parent)
				computeNetworks := append(hostParent(&host.HostSystem).Network, pg.Reference())
				Map.Update(parent, []types.PropertyChange{
					{Name: "network", Val: computeNetworks},
				})
			}
		}

		Map.Update(s, []types.PropertyChange{
			{Name: "portgroup", Val: portgroups},
			{Name: "summary.portgroupName", Val: portgroupNames},
		})

		return nil, nil
	})

	return &methods.AddDVPortgroup_TaskBody{
		Res: &types.AddDVPortgroup_TaskResponse{
			Returnval: task.Run(),
		},
	}
}

func (s *DistributedVirtualSwitch) ReconfigureDvsTask(req *types.ReconfigureDvs_Task) soap.HasFault {
	task := CreateTask(s, "reconfigureDvs", func(t *Task) (types.AnyType, types.BaseMethodFault) {
		spec := req.Spec.GetDVSConfigSpec()

		members := s.Summary.HostMember

		for _, member := range spec.Host {
			h := Map.Get(member.Host)
			if h == nil {
				return nil, &types.ManagedObjectNotFound{Obj: member.Host}
			}

			host := h.(*HostSystem)

			switch types.ConfigSpecOperation(member.Operation) {
			case types.ConfigSpecOperationAdd:
				if FindReference(s.Summary.HostMember, member.Host) != nil {
					return nil, &types.AlreadyExists{Name: host.Name}
				}

				hostNetworks := append(host.Network, s.Portgroup...)
				Map.Update(host, []types.PropertyChange{
					{Name: "network", Val: hostNetworks},
				})
				members = append(members, member.Host)
				parent := Map.Get(*host.HostSystem.Parent)

				var pgs []types.ManagedObjectReference
				for _, ref := range s.Portgroup {
					pg := Map.Get(ref).(*DistributedVirtualPortgroup)
					pgs = append(pgs, ref)

					pgHosts := append(pg.Host, member.Host)
					Map.Update(pg, []types.PropertyChange{
						{Name: "host", Val: pgHosts},
					})

					cr := hostParent(&host.HostSystem)
					if FindReference(cr.Network, ref) == nil {
						computeNetworks := append(cr.Network, ref)
						Map.Update(parent, []types.PropertyChange{
							{Name: "network", Val: computeNetworks},
						})
					}
				}

			case types.ConfigSpecOperationRemove:
				for _, ref := range host.Vm {
					vm := Map.Get(ref).(*VirtualMachine)
					if pg := FindReference(vm.Network, s.Portgroup...); pg != nil {
						return nil, &types.ResourceInUse{
							Type: pg.Type,
							Name: pg.Value,
						}
					}
				}

				RemoveReference(&members, member.Host)
			case types.ConfigSpecOperationEdit:
				return nil, &types.NotSupported{}
			}
		}

		Map.Update(s, []types.PropertyChange{
			{Name: "summary.hostMember", Val: members},
		})

		return nil, nil
	})

	return &methods.ReconfigureDvs_TaskBody{
		Res: &types.ReconfigureDvs_TaskResponse{
			Returnval: task.Run(),
		},
	}
}

func (s *DistributedVirtualSwitch) FetchDVPorts(req *types.FetchDVPorts) soap.HasFault {
	body := &methods.FetchDVPortsBody{}
	body.Res = &types.FetchDVPortsResponse{
		Returnval: s.dvPortgroups(req.Criteria),
	}
	return body
}

func (s *DistributedVirtualSwitch) DestroyTask(ctx *Context, req *types.Destroy_Task) soap.HasFault {
	task := CreateTask(s, "destroy", func(t *Task) (types.AnyType, types.BaseMethodFault) {
		f := Map.getEntityParent(s, "Folder").(*Folder)
		folderRemoveChild(ctx, &f.Folder, s.Reference())
		return nil, nil
	})

	return &methods.Destroy_TaskBody{
		Res: &types.Destroy_TaskResponse{
			Returnval: task.Run(),
		},
	}
}

func (s *DistributedVirtualSwitch) dvPortgroups(_ *types.DistributedVirtualSwitchPortCriteria) []types.DistributedVirtualPort {
	// TODO(agui): Filter is not implemented yet
	var res []types.DistributedVirtualPort
	for _, ref := range s.Portgroup {
		pg := Map.Get(ref).(*DistributedVirtualPortgroup)
		res = append(res, types.DistributedVirtualPort{
			DvsUuid: s.Uuid,
			Key:     pg.Key,
			Config: types.DVPortConfigInfo{
				Setting: pg.Config.DefaultPortConfig,
			},
		})

		for _, key := range pg.PortKeys {
			res = append(res, types.DistributedVirtualPort{
				DvsUuid: s.Uuid,
				Key:     key,
				Config: types.DVPortConfigInfo{
					Setting: pg.Config.DefaultPortConfig,
				},
			})
		}
	}
	return res
}
