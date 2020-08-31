/*
Copyright (c) 2020 VMware, Inc. All Rights Reserved.

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
	"net/url"
	"strings"

	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

type GuestOperationsManager struct {
	mo.GuestOperationsManager
}

func (m *GuestOperationsManager) init(r *Registry) {
	fm := new(GuestFileManager)
	if m.FileManager == nil {
		m.FileManager = &types.ManagedObjectReference{
			Type:  "GuestFileManager",
			Value: "guestOperationsFileManager",
		}
	}
	fm.Self = *m.FileManager
	r.Put(fm)

	pm := new(GuestProcessManager)
	if m.ProcessManager == nil {
		m.ProcessManager = &types.ManagedObjectReference{
			Type:  "GuestProcessManager",
			Value: "guestOperationsProcessManager",
		}
	}
	pm.Self = *m.ProcessManager
	r.Put(pm)
}

type GuestFileManager struct {
	mo.GuestFileManager
}

func guestURL(ctx *Context, vm *VirtualMachine, path string) string {
	return (&url.URL{
		Scheme: ctx.svc.Listen.Scheme,
		Host:   "*", // See guest.FileManager.TransferURL
		Path:   guestPrefix + strings.TrimPrefix(path, "/"),
		RawQuery: url.Values{
			"id":    []string{vm.run.id},
			"token": []string{ctx.Session.Key},
		}.Encode(),
	}).String()
}

func (m *GuestFileManager) InitiateFileTransferToGuest(ctx *Context, req *types.InitiateFileTransferToGuest) soap.HasFault {
	body := new(methods.InitiateFileTransferToGuestBody)

	vm := ctx.Map.Get(req.Vm).(*VirtualMachine)
	err := vm.run.prepareGuestOperation(vm, req.Auth)
	if err != nil {
		body.Fault_ = Fault("", err)
		return body
	}

	body.Res = &types.InitiateFileTransferToGuestResponse{
		Returnval: guestURL(ctx, vm, req.GuestFilePath),
	}

	return body
}

func (m *GuestFileManager) InitiateFileTransferFromGuest(ctx *Context, req *types.InitiateFileTransferFromGuest) soap.HasFault {
	body := new(methods.InitiateFileTransferFromGuestBody)

	vm := ctx.Map.Get(req.Vm).(*VirtualMachine)
	err := vm.run.prepareGuestOperation(vm, req.Auth)
	if err != nil {
		body.Fault_ = Fault("", err)
		return body
	}

	body.Res = &types.InitiateFileTransferFromGuestResponse{
		Returnval: types.FileTransferInformation{
			Attributes: nil, // TODO
			Size:       0,   // TODO
			Url:        guestURL(ctx, vm, req.GuestFilePath),
		},
	}

	return body
}

type GuestProcessManager struct {
	mo.GuestProcessManager
}
