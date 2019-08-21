// GENERATED FILE: DO NOT EDIT!

package oapi

// To create a server, first write a class that implements this interface.
// Then pass an instance of it to Initialize().
type Provider interface {

	//
	POST_AcceptNetPeering(parameters *POST_AcceptNetPeeringParameters, responses *POST_AcceptNetPeeringResponses) (err error)

	//
	POST_AuthenticateAccount(parameters *POST_AuthenticateAccountParameters, responses *POST_AuthenticateAccountResponses) (err error)

	//
	POST_CheckSignature(parameters *POST_CheckSignatureParameters, responses *POST_CheckSignatureResponses) (err error)

	//
	POST_CopyAccount(parameters *POST_CopyAccountParameters, responses *POST_CopyAccountResponses) (err error)

	//
	POST_CreateAccount(parameters *POST_CreateAccountParameters, responses *POST_CreateAccountResponses) (err error)

	//
	POST_CreateApiKey(parameters *POST_CreateApiKeyParameters, responses *POST_CreateApiKeyResponses) (err error)

	//
	POST_CreateClientGateway(parameters *POST_CreateClientGatewayParameters, responses *POST_CreateClientGatewayResponses) (err error)

	//
	POST_CreateDhcpOptions(parameters *POST_CreateDhcpOptionsParameters, responses *POST_CreateDhcpOptionsResponses) (err error)

	//
	POST_CreateDirectLink(parameters *POST_CreateDirectLinkParameters, responses *POST_CreateDirectLinkResponses) (err error)

	//
	POST_CreateDirectLinkInterface(parameters *POST_CreateDirectLinkInterfaceParameters, responses *POST_CreateDirectLinkInterfaceResponses) (err error)

	//
	POST_CreateImage(parameters *POST_CreateImageParameters, responses *POST_CreateImageResponses) (err error)

	//
	POST_CreateImageExportTask(parameters *POST_CreateImageExportTaskParameters, responses *POST_CreateImageExportTaskResponses) (err error)

	//
	POST_CreateInternetService(parameters *POST_CreateInternetServiceParameters, responses *POST_CreateInternetServiceResponses) (err error)

	//
	POST_CreateKeypair(parameters *POST_CreateKeypairParameters, responses *POST_CreateKeypairResponses) (err error)

	//
	POST_CreateListenerRule(parameters *POST_CreateListenerRuleParameters, responses *POST_CreateListenerRuleResponses) (err error)

	//
	POST_CreateLoadBalancer(parameters *POST_CreateLoadBalancerParameters, responses *POST_CreateLoadBalancerResponses) (err error)

	//
	POST_CreateLoadBalancerListeners(parameters *POST_CreateLoadBalancerListenersParameters, responses *POST_CreateLoadBalancerListenersResponses) (err error)

	//
	POST_CreateLoadBalancerPolicy(parameters *POST_CreateLoadBalancerPolicyParameters, responses *POST_CreateLoadBalancerPolicyResponses) (err error)

	//
	POST_CreateNatService(parameters *POST_CreateNatServiceParameters, responses *POST_CreateNatServiceResponses) (err error)

	//
	POST_CreateNet(parameters *POST_CreateNetParameters, responses *POST_CreateNetResponses) (err error)

	//
	POST_CreateNetAccessPoint(parameters *POST_CreateNetAccessPointParameters, responses *POST_CreateNetAccessPointResponses) (err error)

	//
	POST_CreateNetPeering(parameters *POST_CreateNetPeeringParameters, responses *POST_CreateNetPeeringResponses) (err error)

	//
	POST_CreateNic(parameters *POST_CreateNicParameters, responses *POST_CreateNicResponses) (err error)

	//
	POST_CreatePolicy(parameters *POST_CreatePolicyParameters, responses *POST_CreatePolicyResponses) (err error)

	//
	POST_CreatePublicIp(parameters *POST_CreatePublicIpParameters, responses *POST_CreatePublicIpResponses) (err error)

	//
	POST_CreateRoute(parameters *POST_CreateRouteParameters, responses *POST_CreateRouteResponses) (err error)

	//
	POST_CreateRouteTable(parameters *POST_CreateRouteTableParameters, responses *POST_CreateRouteTableResponses) (err error)

	//
	POST_CreateSecurityGroup(parameters *POST_CreateSecurityGroupParameters, responses *POST_CreateSecurityGroupResponses) (err error)

	//
	POST_CreateSecurityGroupRule(parameters *POST_CreateSecurityGroupRuleParameters, responses *POST_CreateSecurityGroupRuleResponses) (err error)

	//
	POST_CreateServerCertificate(parameters *POST_CreateServerCertificateParameters, responses *POST_CreateServerCertificateResponses) (err error)

	//
	POST_CreateSnapshot(parameters *POST_CreateSnapshotParameters, responses *POST_CreateSnapshotResponses) (err error)

	//
	POST_CreateSnapshotExportTask(parameters *POST_CreateSnapshotExportTaskParameters, responses *POST_CreateSnapshotExportTaskResponses) (err error)

	//
	POST_CreateSubnet(parameters *POST_CreateSubnetParameters, responses *POST_CreateSubnetResponses) (err error)

	//
	POST_CreateTags(parameters *POST_CreateTagsParameters, responses *POST_CreateTagsResponses) (err error)

	//
	POST_CreateUser(parameters *POST_CreateUserParameters, responses *POST_CreateUserResponses) (err error)

	//
	POST_CreateUserGroup(parameters *POST_CreateUserGroupParameters, responses *POST_CreateUserGroupResponses) (err error)

	//
	POST_CreateVirtualGateway(parameters *POST_CreateVirtualGatewayParameters, responses *POST_CreateVirtualGatewayResponses) (err error)

	//
	POST_CreateVms(parameters *POST_CreateVmsParameters, responses *POST_CreateVmsResponses) (err error)

	//
	POST_CreateVolume(parameters *POST_CreateVolumeParameters, responses *POST_CreateVolumeResponses) (err error)

	//
	POST_CreateVpnConnection(parameters *POST_CreateVpnConnectionParameters, responses *POST_CreateVpnConnectionResponses) (err error)

	//
	POST_CreateVpnConnectionRoute(parameters *POST_CreateVpnConnectionRouteParameters, responses *POST_CreateVpnConnectionRouteResponses) (err error)

	//
	POST_DeleteApiKey(parameters *POST_DeleteApiKeyParameters, responses *POST_DeleteApiKeyResponses) (err error)

	//
	POST_DeleteClientGateway(parameters *POST_DeleteClientGatewayParameters, responses *POST_DeleteClientGatewayResponses) (err error)

	//
	POST_DeleteDhcpOptions(parameters *POST_DeleteDhcpOptionsParameters, responses *POST_DeleteDhcpOptionsResponses) (err error)

	//
	POST_DeleteDirectLink(parameters *POST_DeleteDirectLinkParameters, responses *POST_DeleteDirectLinkResponses) (err error)

	//
	POST_DeleteDirectLinkInterface(parameters *POST_DeleteDirectLinkInterfaceParameters, responses *POST_DeleteDirectLinkInterfaceResponses) (err error)

	//
	POST_DeleteExportTask(parameters *POST_DeleteExportTaskParameters, responses *POST_DeleteExportTaskResponses) (err error)

	//
	POST_DeleteImage(parameters *POST_DeleteImageParameters, responses *POST_DeleteImageResponses) (err error)

	//
	POST_DeleteInternetService(parameters *POST_DeleteInternetServiceParameters, responses *POST_DeleteInternetServiceResponses) (err error)

	//
	POST_DeleteKeypair(parameters *POST_DeleteKeypairParameters, responses *POST_DeleteKeypairResponses) (err error)

	//
	POST_DeleteListenerRule(parameters *POST_DeleteListenerRuleParameters, responses *POST_DeleteListenerRuleResponses) (err error)

	//
	POST_DeleteLoadBalancer(parameters *POST_DeleteLoadBalancerParameters, responses *POST_DeleteLoadBalancerResponses) (err error)

	//
	POST_DeleteLoadBalancerListeners(parameters *POST_DeleteLoadBalancerListenersParameters, responses *POST_DeleteLoadBalancerListenersResponses) (err error)

	//
	POST_DeleteLoadBalancerPolicy(parameters *POST_DeleteLoadBalancerPolicyParameters, responses *POST_DeleteLoadBalancerPolicyResponses) (err error)

	//
	POST_DeleteNatService(parameters *POST_DeleteNatServiceParameters, responses *POST_DeleteNatServiceResponses) (err error)

	//
	POST_DeleteNet(parameters *POST_DeleteNetParameters, responses *POST_DeleteNetResponses) (err error)

	//
	POST_DeleteNetAccessPoints(parameters *POST_DeleteNetAccessPointsParameters, responses *POST_DeleteNetAccessPointsResponses) (err error)

	//
	POST_DeleteNetPeering(parameters *POST_DeleteNetPeeringParameters, responses *POST_DeleteNetPeeringResponses) (err error)

	//
	POST_DeleteNic(parameters *POST_DeleteNicParameters, responses *POST_DeleteNicResponses) (err error)

	//
	POST_DeletePolicy(parameters *POST_DeletePolicyParameters, responses *POST_DeletePolicyResponses) (err error)

	//
	POST_DeletePublicIp(parameters *POST_DeletePublicIpParameters, responses *POST_DeletePublicIpResponses) (err error)

	//
	POST_DeleteRoute(parameters *POST_DeleteRouteParameters, responses *POST_DeleteRouteResponses) (err error)

	//
	POST_DeleteRouteTable(parameters *POST_DeleteRouteTableParameters, responses *POST_DeleteRouteTableResponses) (err error)

	//
	POST_DeleteSecurityGroup(parameters *POST_DeleteSecurityGroupParameters, responses *POST_DeleteSecurityGroupResponses) (err error)

	//
	POST_DeleteSecurityGroupRule(parameters *POST_DeleteSecurityGroupRuleParameters, responses *POST_DeleteSecurityGroupRuleResponses) (err error)

	//
	POST_DeleteServerCertificate(parameters *POST_DeleteServerCertificateParameters, responses *POST_DeleteServerCertificateResponses) (err error)

	//
	POST_DeleteSnapshot(parameters *POST_DeleteSnapshotParameters, responses *POST_DeleteSnapshotResponses) (err error)

	//
	POST_DeleteSubnet(parameters *POST_DeleteSubnetParameters, responses *POST_DeleteSubnetResponses) (err error)

	//
	POST_DeleteTags(parameters *POST_DeleteTagsParameters, responses *POST_DeleteTagsResponses) (err error)

	//
	POST_DeleteUser(parameters *POST_DeleteUserParameters, responses *POST_DeleteUserResponses) (err error)

	//
	POST_DeleteUserGroup(parameters *POST_DeleteUserGroupParameters, responses *POST_DeleteUserGroupResponses) (err error)

	//
	POST_DeleteVirtualGateway(parameters *POST_DeleteVirtualGatewayParameters, responses *POST_DeleteVirtualGatewayResponses) (err error)

	//
	POST_DeleteVms(parameters *POST_DeleteVmsParameters, responses *POST_DeleteVmsResponses) (err error)

	//
	POST_DeleteVolume(parameters *POST_DeleteVolumeParameters, responses *POST_DeleteVolumeResponses) (err error)

	//
	POST_DeleteVpnConnection(parameters *POST_DeleteVpnConnectionParameters, responses *POST_DeleteVpnConnectionResponses) (err error)

	//
	POST_DeleteVpnConnectionRoute(parameters *POST_DeleteVpnConnectionRouteParameters, responses *POST_DeleteVpnConnectionRouteResponses) (err error)

	//
	POST_DeregisterUserInUserGroup(parameters *POST_DeregisterUserInUserGroupParameters, responses *POST_DeregisterUserInUserGroupResponses) (err error)

	//
	POST_DeregisterVmsInLoadBalancer(parameters *POST_DeregisterVmsInLoadBalancerParameters, responses *POST_DeregisterVmsInLoadBalancerResponses) (err error)

	//
	POST_LinkInternetService(parameters *POST_LinkInternetServiceParameters, responses *POST_LinkInternetServiceResponses) (err error)

	//
	POST_LinkNic(parameters *POST_LinkNicParameters, responses *POST_LinkNicResponses) (err error)

	//
	POST_LinkPolicy(parameters *POST_LinkPolicyParameters, responses *POST_LinkPolicyResponses) (err error)

	//
	POST_LinkPrivateIps(parameters *POST_LinkPrivateIpsParameters, responses *POST_LinkPrivateIpsResponses) (err error)

	//
	POST_LinkPublicIp(parameters *POST_LinkPublicIpParameters, responses *POST_LinkPublicIpResponses) (err error)

	//
	POST_LinkRouteTable(parameters *POST_LinkRouteTableParameters, responses *POST_LinkRouteTableResponses) (err error)

	//
	POST_LinkVirtualGateway(parameters *POST_LinkVirtualGatewayParameters, responses *POST_LinkVirtualGatewayResponses) (err error)

	//
	POST_LinkVolume(parameters *POST_LinkVolumeParameters, responses *POST_LinkVolumeResponses) (err error)

	//
	POST_PurchaseReservedVmsOffer(parameters *POST_PurchaseReservedVmsOfferParameters, responses *POST_PurchaseReservedVmsOfferResponses) (err error)

	//
	POST_ReadAccount(parameters *POST_ReadAccountParameters, responses *POST_ReadAccountResponses) (err error)

	//
	POST_ReadAccountConsumption(parameters *POST_ReadAccountConsumptionParameters, responses *POST_ReadAccountConsumptionResponses) (err error)

	//
	POST_ReadAdminPassword(parameters *POST_ReadAdminPasswordParameters, responses *POST_ReadAdminPasswordResponses) (err error)

	//
	POST_ReadApiKeys(parameters *POST_ReadApiKeysParameters, responses *POST_ReadApiKeysResponses) (err error)

	//
	POST_ReadApiLogs(parameters *POST_ReadApiLogsParameters, responses *POST_ReadApiLogsResponses) (err error)

	//
	POST_ReadBillableDigest(parameters *POST_ReadBillableDigestParameters, responses *POST_ReadBillableDigestResponses) (err error)

	//
	POST_ReadCatalog(parameters *POST_ReadCatalogParameters, responses *POST_ReadCatalogResponses) (err error)

	//
	POST_ReadClientGateways(parameters *POST_ReadClientGatewaysParameters, responses *POST_ReadClientGatewaysResponses) (err error)

	//
	POST_ReadConsoleOutput(parameters *POST_ReadConsoleOutputParameters, responses *POST_ReadConsoleOutputResponses) (err error)

	//
	POST_ReadDhcpOptions(parameters *POST_ReadDhcpOptionsParameters, responses *POST_ReadDhcpOptionsResponses) (err error)

	//
	POST_ReadDirectLinkInterfaces(parameters *POST_ReadDirectLinkInterfacesParameters, responses *POST_ReadDirectLinkInterfacesResponses) (err error)

	//
	POST_ReadDirectLinks(parameters *POST_ReadDirectLinksParameters, responses *POST_ReadDirectLinksResponses) (err error)

	//
	POST_ReadImageExportTasks(parameters *POST_ReadImageExportTasksParameters, responses *POST_ReadImageExportTasksResponses) (err error)

	//
	POST_ReadImages(parameters *POST_ReadImagesParameters, responses *POST_ReadImagesResponses) (err error)

	//
	POST_ReadInternetServices(parameters *POST_ReadInternetServicesParameters, responses *POST_ReadInternetServicesResponses) (err error)

	//
	POST_ReadKeypairs(parameters *POST_ReadKeypairsParameters, responses *POST_ReadKeypairsResponses) (err error)

	//
	POST_ReadListenerRules(parameters *POST_ReadListenerRulesParameters, responses *POST_ReadListenerRulesResponses) (err error)

	//
	POST_ReadLoadBalancers(parameters *POST_ReadLoadBalancersParameters, responses *POST_ReadLoadBalancersResponses) (err error)

	//
	POST_ReadLocations(parameters *POST_ReadLocationsParameters, responses *POST_ReadLocationsResponses) (err error)

	//
	POST_ReadNatServices(parameters *POST_ReadNatServicesParameters, responses *POST_ReadNatServicesResponses) (err error)

	//
	POST_ReadNetAccessPointServices(parameters *POST_ReadNetAccessPointServicesParameters, responses *POST_ReadNetAccessPointServicesResponses) (err error)

	//
	POST_ReadNetAccessPoints(parameters *POST_ReadNetAccessPointsParameters, responses *POST_ReadNetAccessPointsResponses) (err error)

	//
	POST_ReadNetPeerings(parameters *POST_ReadNetPeeringsParameters, responses *POST_ReadNetPeeringsResponses) (err error)

	//
	POST_ReadNets(parameters *POST_ReadNetsParameters, responses *POST_ReadNetsResponses) (err error)

	//
	POST_ReadNics(parameters *POST_ReadNicsParameters, responses *POST_ReadNicsResponses) (err error)

	//
	POST_ReadPolicies(parameters *POST_ReadPoliciesParameters, responses *POST_ReadPoliciesResponses) (err error)

	//
	POST_ReadPrefixLists(parameters *POST_ReadPrefixListsParameters, responses *POST_ReadPrefixListsResponses) (err error)

	//
	POST_ReadProductTypes(parameters *POST_ReadProductTypesParameters, responses *POST_ReadProductTypesResponses) (err error)

	//
	POST_ReadPublicCatalog(parameters *POST_ReadPublicCatalogParameters, responses *POST_ReadPublicCatalogResponses) (err error)

	//
	POST_ReadPublicIpRanges(parameters *POST_ReadPublicIpRangesParameters, responses *POST_ReadPublicIpRangesResponses) (err error)

	//
	POST_ReadPublicIps(parameters *POST_ReadPublicIpsParameters, responses *POST_ReadPublicIpsResponses) (err error)

	//
	POST_ReadQuotas(parameters *POST_ReadQuotasParameters, responses *POST_ReadQuotasResponses) (err error)

	//
	POST_ReadRegionConfig(parameters *POST_ReadRegionConfigParameters, responses *POST_ReadRegionConfigResponses) (err error)

	//
	POST_ReadRegions(parameters *POST_ReadRegionsParameters, responses *POST_ReadRegionsResponses) (err error)

	//
	POST_ReadReservedVmOffers(parameters *POST_ReadReservedVmOffersParameters, responses *POST_ReadReservedVmOffersResponses) (err error)

	//
	POST_ReadReservedVms(parameters *POST_ReadReservedVmsParameters, responses *POST_ReadReservedVmsResponses) (err error)

	//
	POST_ReadRouteTables(parameters *POST_ReadRouteTablesParameters, responses *POST_ReadRouteTablesResponses) (err error)

	//
	POST_ReadSecurityGroups(parameters *POST_ReadSecurityGroupsParameters, responses *POST_ReadSecurityGroupsResponses) (err error)

	//
	POST_ReadServerCertificates(parameters *POST_ReadServerCertificatesParameters, responses *POST_ReadServerCertificatesResponses) (err error)

	//
	POST_ReadSnapshotExportTasks(parameters *POST_ReadSnapshotExportTasksParameters, responses *POST_ReadSnapshotExportTasksResponses) (err error)

	//
	POST_ReadSnapshots(parameters *POST_ReadSnapshotsParameters, responses *POST_ReadSnapshotsResponses) (err error)

	//
	POST_ReadSubnets(parameters *POST_ReadSubnetsParameters, responses *POST_ReadSubnetsResponses) (err error)

	//
	POST_ReadSubregions(parameters *POST_ReadSubregionsParameters, responses *POST_ReadSubregionsResponses) (err error)

	//
	POST_ReadTags(parameters *POST_ReadTagsParameters, responses *POST_ReadTagsResponses) (err error)

	//
	POST_ReadUserGroups(parameters *POST_ReadUserGroupsParameters, responses *POST_ReadUserGroupsResponses) (err error)

	//
	POST_ReadUsers(parameters *POST_ReadUsersParameters, responses *POST_ReadUsersResponses) (err error)

	//
	POST_ReadVirtualGateways(parameters *POST_ReadVirtualGatewaysParameters, responses *POST_ReadVirtualGatewaysResponses) (err error)

	//
	POST_ReadVmTypes(parameters *POST_ReadVmTypesParameters, responses *POST_ReadVmTypesResponses) (err error)

	//
	POST_ReadVms(parameters *POST_ReadVmsParameters, responses *POST_ReadVmsResponses) (err error)

	//
	POST_ReadVmsHealth(parameters *POST_ReadVmsHealthParameters, responses *POST_ReadVmsHealthResponses) (err error)

	//
	POST_ReadVmsState(parameters *POST_ReadVmsStateParameters, responses *POST_ReadVmsStateResponses) (err error)

	//
	POST_ReadVolumes(parameters *POST_ReadVolumesParameters, responses *POST_ReadVolumesResponses) (err error)

	//
	POST_ReadVpnConnections(parameters *POST_ReadVpnConnectionsParameters, responses *POST_ReadVpnConnectionsResponses) (err error)

	//
	POST_RebootVms(parameters *POST_RebootVmsParameters, responses *POST_RebootVmsResponses) (err error)

	//
	POST_RegisterUserInUserGroup(parameters *POST_RegisterUserInUserGroupParameters, responses *POST_RegisterUserInUserGroupResponses) (err error)

	//
	POST_RegisterVmsInLoadBalancer(parameters *POST_RegisterVmsInLoadBalancerParameters, responses *POST_RegisterVmsInLoadBalancerResponses) (err error)

	//
	POST_RejectNetPeering(parameters *POST_RejectNetPeeringParameters, responses *POST_RejectNetPeeringResponses) (err error)

	//
	POST_ResetAccountPassword(parameters *POST_ResetAccountPasswordParameters, responses *POST_ResetAccountPasswordResponses) (err error)

	//
	POST_SendResetPasswordEmail(parameters *POST_SendResetPasswordEmailParameters, responses *POST_SendResetPasswordEmailResponses) (err error)

	//
	POST_StartVms(parameters *POST_StartVmsParameters, responses *POST_StartVmsResponses) (err error)

	//
	POST_StopVms(parameters *POST_StopVmsParameters, responses *POST_StopVmsResponses) (err error)

	//
	POST_UnlinkInternetService(parameters *POST_UnlinkInternetServiceParameters, responses *POST_UnlinkInternetServiceResponses) (err error)

	//
	POST_UnlinkNic(parameters *POST_UnlinkNicParameters, responses *POST_UnlinkNicResponses) (err error)

	//
	POST_UnlinkPolicy(parameters *POST_UnlinkPolicyParameters, responses *POST_UnlinkPolicyResponses) (err error)

	//
	POST_UnlinkPrivateIps(parameters *POST_UnlinkPrivateIpsParameters, responses *POST_UnlinkPrivateIpsResponses) (err error)

	//
	POST_UnlinkPublicIp(parameters *POST_UnlinkPublicIpParameters, responses *POST_UnlinkPublicIpResponses) (err error)

	//
	POST_UnlinkRouteTable(parameters *POST_UnlinkRouteTableParameters, responses *POST_UnlinkRouteTableResponses) (err error)

	//
	POST_UnlinkVirtualGateway(parameters *POST_UnlinkVirtualGatewayParameters, responses *POST_UnlinkVirtualGatewayResponses) (err error)

	//
	POST_UnlinkVolume(parameters *POST_UnlinkVolumeParameters, responses *POST_UnlinkVolumeResponses) (err error)

	//
	POST_UpdateAccount(parameters *POST_UpdateAccountParameters, responses *POST_UpdateAccountResponses) (err error)

	//
	POST_UpdateApiKey(parameters *POST_UpdateApiKeyParameters, responses *POST_UpdateApiKeyResponses) (err error)

	//
	POST_UpdateHealthCheck(parameters *POST_UpdateHealthCheckParameters, responses *POST_UpdateHealthCheckResponses) (err error)

	//
	POST_UpdateImage(parameters *POST_UpdateImageParameters, responses *POST_UpdateImageResponses) (err error)

	//
	POST_UpdateKeypair(parameters *POST_UpdateKeypairParameters, responses *POST_UpdateKeypairResponses) (err error)

	//
	POST_UpdateListenerRule(parameters *POST_UpdateListenerRuleParameters, responses *POST_UpdateListenerRuleResponses) (err error)

	//
	POST_UpdateLoadBalancer(parameters *POST_UpdateLoadBalancerParameters, responses *POST_UpdateLoadBalancerResponses) (err error)

	//
	POST_UpdateNet(parameters *POST_UpdateNetParameters, responses *POST_UpdateNetResponses) (err error)

	//
	POST_UpdateNetAccessPoint(parameters *POST_UpdateNetAccessPointParameters, responses *POST_UpdateNetAccessPointResponses) (err error)

	//
	POST_UpdateNic(parameters *POST_UpdateNicParameters, responses *POST_UpdateNicResponses) (err error)

	//
	POST_UpdateRoute(parameters *POST_UpdateRouteParameters, responses *POST_UpdateRouteResponses) (err error)

	//
	POST_UpdateRoutePropagation(parameters *POST_UpdateRoutePropagationParameters, responses *POST_UpdateRoutePropagationResponses) (err error)

	//
	POST_UpdateServerCertificate(parameters *POST_UpdateServerCertificateParameters, responses *POST_UpdateServerCertificateResponses) (err error)

	//
	POST_UpdateSnapshot(parameters *POST_UpdateSnapshotParameters, responses *POST_UpdateSnapshotResponses) (err error)

	//
	POST_UpdateUser(parameters *POST_UpdateUserParameters, responses *POST_UpdateUserResponses) (err error)

	//
	POST_UpdateUserGroup(parameters *POST_UpdateUserGroupParameters, responses *POST_UpdateUserGroupResponses) (err error)

	//
	POST_UpdateVm(parameters *POST_UpdateVmParameters, responses *POST_UpdateVmResponses) (err error)
}
