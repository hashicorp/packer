package oapi

type OAPIClient interface {
	POST_AcceptNetPeering(
		acceptnetpeeringrequest AcceptNetPeeringRequest,
	) (
		response *POST_AcceptNetPeeringResponses,
		err error,
	)

	POST_AuthenticateAccount(
		authenticateaccountrequest AuthenticateAccountRequest,
	) (
		response *POST_AuthenticateAccountResponses,
		err error,
	)

	POST_CheckSignature(
		checksignaturerequest CheckSignatureRequest,
	) (
		response *POST_CheckSignatureResponses,
		err error,
	)

	POST_CopyAccount(
		copyaccountrequest CopyAccountRequest,
	) (
		response *POST_CopyAccountResponses,
		err error,
	)

	POST_CreateAccount(
		createaccountrequest CreateAccountRequest,
	) (
		response *POST_CreateAccountResponses,
		err error,
	)

	POST_CreateApiKey(
		createapikeyrequest CreateApiKeyRequest,
	) (
		response *POST_CreateApiKeyResponses,
		err error,
	)

	POST_CreateClientGateway(
		createclientgatewayrequest CreateClientGatewayRequest,
	) (
		response *POST_CreateClientGatewayResponses,
		err error,
	)

	POST_CreateDhcpOptions(
		createdhcpoptionsrequest CreateDhcpOptionsRequest,
	) (
		response *POST_CreateDhcpOptionsResponses,
		err error,
	)

	POST_CreateDirectLink(
		createdirectlinkrequest CreateDirectLinkRequest,
	) (
		response *POST_CreateDirectLinkResponses,
		err error,
	)

	POST_CreateDirectLinkInterface(
		createdirectlinkinterfacerequest CreateDirectLinkInterfaceRequest,
	) (
		response *POST_CreateDirectLinkInterfaceResponses,
		err error,
	)

	POST_CreateImage(
		createimagerequest CreateImageRequest,
	) (
		response *POST_CreateImageResponses,
		err error,
	)

	POST_CreateImageExportTask(
		createimageexporttaskrequest CreateImageExportTaskRequest,
	) (
		response *POST_CreateImageExportTaskResponses,
		err error,
	)

	POST_CreateInternetService(
		createinternetservicerequest CreateInternetServiceRequest,
	) (
		response *POST_CreateInternetServiceResponses,
		err error,
	)

	POST_CreateKeypair(
		createkeypairrequest CreateKeypairRequest,
	) (
		response *POST_CreateKeypairResponses,
		err error,
	)

	POST_CreateListenerRule(
		createlistenerrulerequest CreateListenerRuleRequest,
	) (
		response *POST_CreateListenerRuleResponses,
		err error,
	)

	POST_CreateLoadBalancer(
		createloadbalancerrequest CreateLoadBalancerRequest,
	) (
		response *POST_CreateLoadBalancerResponses,
		err error,
	)

	POST_CreateLoadBalancerListeners(
		createloadbalancerlistenersrequest CreateLoadBalancerListenersRequest,
	) (
		response *POST_CreateLoadBalancerListenersResponses,
		err error,
	)

	POST_CreateLoadBalancerPolicy(
		createloadbalancerpolicyrequest CreateLoadBalancerPolicyRequest,
	) (
		response *POST_CreateLoadBalancerPolicyResponses,
		err error,
	)

	POST_CreateNatService(
		createnatservicerequest CreateNatServiceRequest,
	) (
		response *POST_CreateNatServiceResponses,
		err error,
	)

	POST_CreateNet(
		createnetrequest CreateNetRequest,
	) (
		response *POST_CreateNetResponses,
		err error,
	)

	POST_CreateNetAccessPoint(
		createnetaccesspointrequest CreateNetAccessPointRequest,
	) (
		response *POST_CreateNetAccessPointResponses,
		err error,
	)

	POST_CreateNetPeering(
		createnetpeeringrequest CreateNetPeeringRequest,
	) (
		response *POST_CreateNetPeeringResponses,
		err error,
	)

	POST_CreateNic(
		createnicrequest CreateNicRequest,
	) (
		response *POST_CreateNicResponses,
		err error,
	)

	POST_CreatePolicy(
		createpolicyrequest CreatePolicyRequest,
	) (
		response *POST_CreatePolicyResponses,
		err error,
	)

	POST_CreatePublicIp(
		createpubliciprequest CreatePublicIpRequest,
	) (
		response *POST_CreatePublicIpResponses,
		err error,
	)

	POST_CreateRoute(
		createrouterequest CreateRouteRequest,
	) (
		response *POST_CreateRouteResponses,
		err error,
	)

	POST_CreateRouteTable(
		createroutetablerequest CreateRouteTableRequest,
	) (
		response *POST_CreateRouteTableResponses,
		err error,
	)

	POST_CreateSecurityGroup(
		createsecuritygrouprequest CreateSecurityGroupRequest,
	) (
		response *POST_CreateSecurityGroupResponses,
		err error,
	)

	POST_CreateSecurityGroupRule(
		createsecuritygrouprulerequest CreateSecurityGroupRuleRequest,
	) (
		response *POST_CreateSecurityGroupRuleResponses,
		err error,
	)

	POST_CreateServerCertificate(
		createservercertificaterequest CreateServerCertificateRequest,
	) (
		response *POST_CreateServerCertificateResponses,
		err error,
	)

	POST_CreateSnapshot(
		createsnapshotrequest CreateSnapshotRequest,
	) (
		response *POST_CreateSnapshotResponses,
		err error,
	)

	POST_CreateSnapshotExportTask(
		createsnapshotexporttaskrequest CreateSnapshotExportTaskRequest,
	) (
		response *POST_CreateSnapshotExportTaskResponses,
		err error,
	)

	POST_CreateSubnet(
		createsubnetrequest CreateSubnetRequest,
	) (
		response *POST_CreateSubnetResponses,
		err error,
	)

	POST_CreateTags(
		createtagsrequest CreateTagsRequest,
	) (
		response *POST_CreateTagsResponses,
		err error,
	)

	POST_CreateUser(
		createuserrequest CreateUserRequest,
	) (
		response *POST_CreateUserResponses,
		err error,
	)

	POST_CreateUserGroup(
		createusergrouprequest CreateUserGroupRequest,
	) (
		response *POST_CreateUserGroupResponses,
		err error,
	)

	POST_CreateVirtualGateway(
		createvirtualgatewayrequest CreateVirtualGatewayRequest,
	) (
		response *POST_CreateVirtualGatewayResponses,
		err error,
	)

	POST_CreateVms(
		createvmsrequest CreateVmsRequest,
	) (
		response *POST_CreateVmsResponses,
		err error,
	)

	POST_CreateVolume(
		createvolumerequest CreateVolumeRequest,
	) (
		response *POST_CreateVolumeResponses,
		err error,
	)

	POST_CreateVpnConnection(
		createvpnconnectionrequest CreateVpnConnectionRequest,
	) (
		response *POST_CreateVpnConnectionResponses,
		err error,
	)

	POST_CreateVpnConnectionRoute(
		createvpnconnectionrouterequest CreateVpnConnectionRouteRequest,
	) (
		response *POST_CreateVpnConnectionRouteResponses,
		err error,
	)

	POST_DeleteApiKey(
		deleteapikeyrequest DeleteApiKeyRequest,
	) (
		response *POST_DeleteApiKeyResponses,
		err error,
	)

	POST_DeleteClientGateway(
		deleteclientgatewayrequest DeleteClientGatewayRequest,
	) (
		response *POST_DeleteClientGatewayResponses,
		err error,
	)

	POST_DeleteDhcpOptions(
		deletedhcpoptionsrequest DeleteDhcpOptionsRequest,
	) (
		response *POST_DeleteDhcpOptionsResponses,
		err error,
	)

	POST_DeleteDirectLink(
		deletedirectlinkrequest DeleteDirectLinkRequest,
	) (
		response *POST_DeleteDirectLinkResponses,
		err error,
	)

	POST_DeleteDirectLinkInterface(
		deletedirectlinkinterfacerequest DeleteDirectLinkInterfaceRequest,
	) (
		response *POST_DeleteDirectLinkInterfaceResponses,
		err error,
	)

	POST_DeleteExportTask(
		deleteexporttaskrequest DeleteExportTaskRequest,
	) (
		response *POST_DeleteExportTaskResponses,
		err error,
	)

	POST_DeleteImage(
		deleteimagerequest DeleteImageRequest,
	) (
		response *POST_DeleteImageResponses,
		err error,
	)

	POST_DeleteInternetService(
		deleteinternetservicerequest DeleteInternetServiceRequest,
	) (
		response *POST_DeleteInternetServiceResponses,
		err error,
	)

	POST_DeleteKeypair(
		deletekeypairrequest DeleteKeypairRequest,
	) (
		response *POST_DeleteKeypairResponses,
		err error,
	)

	POST_DeleteListenerRule(
		deletelistenerrulerequest DeleteListenerRuleRequest,
	) (
		response *POST_DeleteListenerRuleResponses,
		err error,
	)

	POST_DeleteLoadBalancer(
		deleteloadbalancerrequest DeleteLoadBalancerRequest,
	) (
		response *POST_DeleteLoadBalancerResponses,
		err error,
	)

	POST_DeleteLoadBalancerListeners(
		deleteloadbalancerlistenersrequest DeleteLoadBalancerListenersRequest,
	) (
		response *POST_DeleteLoadBalancerListenersResponses,
		err error,
	)

	POST_DeleteLoadBalancerPolicy(
		deleteloadbalancerpolicyrequest DeleteLoadBalancerPolicyRequest,
	) (
		response *POST_DeleteLoadBalancerPolicyResponses,
		err error,
	)

	POST_DeleteNatService(
		deletenatservicerequest DeleteNatServiceRequest,
	) (
		response *POST_DeleteNatServiceResponses,
		err error,
	)

	POST_DeleteNet(
		deletenetrequest DeleteNetRequest,
	) (
		response *POST_DeleteNetResponses,
		err error,
	)

	POST_DeleteNetAccessPoints(
		deletenetaccesspointsrequest DeleteNetAccessPointsRequest,
	) (
		response *POST_DeleteNetAccessPointsResponses,
		err error,
	)

	POST_DeleteNetPeering(
		deletenetpeeringrequest DeleteNetPeeringRequest,
	) (
		response *POST_DeleteNetPeeringResponses,
		err error,
	)

	POST_DeleteNic(
		deletenicrequest DeleteNicRequest,
	) (
		response *POST_DeleteNicResponses,
		err error,
	)

	POST_DeletePolicy(
		deletepolicyrequest DeletePolicyRequest,
	) (
		response *POST_DeletePolicyResponses,
		err error,
	)

	POST_DeletePublicIp(
		deletepubliciprequest DeletePublicIpRequest,
	) (
		response *POST_DeletePublicIpResponses,
		err error,
	)

	POST_DeleteRoute(
		deleterouterequest DeleteRouteRequest,
	) (
		response *POST_DeleteRouteResponses,
		err error,
	)

	POST_DeleteRouteTable(
		deleteroutetablerequest DeleteRouteTableRequest,
	) (
		response *POST_DeleteRouteTableResponses,
		err error,
	)

	POST_DeleteSecurityGroup(
		deletesecuritygrouprequest DeleteSecurityGroupRequest,
	) (
		response *POST_DeleteSecurityGroupResponses,
		err error,
	)

	POST_DeleteSecurityGroupRule(
		deletesecuritygrouprulerequest DeleteSecurityGroupRuleRequest,
	) (
		response *POST_DeleteSecurityGroupRuleResponses,
		err error,
	)

	POST_DeleteServerCertificate(
		deleteservercertificaterequest DeleteServerCertificateRequest,
	) (
		response *POST_DeleteServerCertificateResponses,
		err error,
	)

	POST_DeleteSnapshot(
		deletesnapshotrequest DeleteSnapshotRequest,
	) (
		response *POST_DeleteSnapshotResponses,
		err error,
	)

	POST_DeleteSubnet(
		deletesubnetrequest DeleteSubnetRequest,
	) (
		response *POST_DeleteSubnetResponses,
		err error,
	)

	POST_DeleteTags(
		deletetagsrequest DeleteTagsRequest,
	) (
		response *POST_DeleteTagsResponses,
		err error,
	)

	POST_DeleteUser(
		deleteuserrequest DeleteUserRequest,
	) (
		response *POST_DeleteUserResponses,
		err error,
	)

	POST_DeleteUserGroup(
		deleteusergrouprequest DeleteUserGroupRequest,
	) (
		response *POST_DeleteUserGroupResponses,
		err error,
	)

	POST_DeleteVirtualGateway(
		deletevirtualgatewayrequest DeleteVirtualGatewayRequest,
	) (
		response *POST_DeleteVirtualGatewayResponses,
		err error,
	)

	POST_DeleteVms(
		deletevmsrequest DeleteVmsRequest,
	) (
		response *POST_DeleteVmsResponses,
		err error,
	)

	POST_DeleteVolume(
		deletevolumerequest DeleteVolumeRequest,
	) (
		response *POST_DeleteVolumeResponses,
		err error,
	)

	POST_DeleteVpnConnection(
		deletevpnconnectionrequest DeleteVpnConnectionRequest,
	) (
		response *POST_DeleteVpnConnectionResponses,
		err error,
	)

	POST_DeleteVpnConnectionRoute(
		deletevpnconnectionrouterequest DeleteVpnConnectionRouteRequest,
	) (
		response *POST_DeleteVpnConnectionRouteResponses,
		err error,
	)

	POST_DeregisterUserInUserGroup(
		deregisteruserinusergrouprequest DeregisterUserInUserGroupRequest,
	) (
		response *POST_DeregisterUserInUserGroupResponses,
		err error,
	)

	POST_DeregisterVmsInLoadBalancer(
		deregistervmsinloadbalancerrequest DeregisterVmsInLoadBalancerRequest,
	) (
		response *POST_DeregisterVmsInLoadBalancerResponses,
		err error,
	)

	POST_LinkInternetService(
		linkinternetservicerequest LinkInternetServiceRequest,
	) (
		response *POST_LinkInternetServiceResponses,
		err error,
	)

	POST_LinkNic(
		linknicrequest LinkNicRequest,
	) (
		response *POST_LinkNicResponses,
		err error,
	)

	POST_LinkPolicy(
		linkpolicyrequest LinkPolicyRequest,
	) (
		response *POST_LinkPolicyResponses,
		err error,
	)

	POST_LinkPrivateIps(
		linkprivateipsrequest LinkPrivateIpsRequest,
	) (
		response *POST_LinkPrivateIpsResponses,
		err error,
	)

	POST_LinkPublicIp(
		linkpubliciprequest LinkPublicIpRequest,
	) (
		response *POST_LinkPublicIpResponses,
		err error,
	)

	POST_LinkRouteTable(
		linkroutetablerequest LinkRouteTableRequest,
	) (
		response *POST_LinkRouteTableResponses,
		err error,
	)

	POST_LinkVirtualGateway(
		linkvirtualgatewayrequest LinkVirtualGatewayRequest,
	) (
		response *POST_LinkVirtualGatewayResponses,
		err error,
	)

	POST_LinkVolume(
		linkvolumerequest LinkVolumeRequest,
	) (
		response *POST_LinkVolumeResponses,
		err error,
	)

	POST_PurchaseReservedVmsOffer(
		purchasereservedvmsofferrequest PurchaseReservedVmsOfferRequest,
	) (
		response *POST_PurchaseReservedVmsOfferResponses,
		err error,
	)

	POST_ReadAccount(
		readaccountrequest ReadAccountRequest,
	) (
		response *POST_ReadAccountResponses,
		err error,
	)

	POST_ReadAccountConsumption(
		readaccountconsumptionrequest ReadAccountConsumptionRequest,
	) (
		response *POST_ReadAccountConsumptionResponses,
		err error,
	)

	POST_ReadAdminPassword(
		readadminpasswordrequest ReadAdminPasswordRequest,
	) (
		response *POST_ReadAdminPasswordResponses,
		err error,
	)

	POST_ReadApiKeys(
		readapikeysrequest ReadApiKeysRequest,
	) (
		response *POST_ReadApiKeysResponses,
		err error,
	)

	POST_ReadApiLogs(
		readapilogsrequest ReadApiLogsRequest,
	) (
		response *POST_ReadApiLogsResponses,
		err error,
	)

	POST_ReadBillableDigest(
		readbillabledigestrequest ReadBillableDigestRequest,
	) (
		response *POST_ReadBillableDigestResponses,
		err error,
	)

	POST_ReadCatalog(
		readcatalogrequest ReadCatalogRequest,
	) (
		response *POST_ReadCatalogResponses,
		err error,
	)

	POST_ReadClientGateways(
		readclientgatewaysrequest ReadClientGatewaysRequest,
	) (
		response *POST_ReadClientGatewaysResponses,
		err error,
	)

	POST_ReadConsoleOutput(
		readconsoleoutputrequest ReadConsoleOutputRequest,
	) (
		response *POST_ReadConsoleOutputResponses,
		err error,
	)

	POST_ReadDhcpOptions(
		readdhcpoptionsrequest ReadDhcpOptionsRequest,
	) (
		response *POST_ReadDhcpOptionsResponses,
		err error,
	)

	POST_ReadDirectLinkInterfaces(
		readdirectlinkinterfacesrequest ReadDirectLinkInterfacesRequest,
	) (
		response *POST_ReadDirectLinkInterfacesResponses,
		err error,
	)

	POST_ReadDirectLinks(
		readdirectlinksrequest ReadDirectLinksRequest,
	) (
		response *POST_ReadDirectLinksResponses,
		err error,
	)

	POST_ReadImageExportTasks(
		readimageexporttasksrequest ReadImageExportTasksRequest,
	) (
		response *POST_ReadImageExportTasksResponses,
		err error,
	)

	POST_ReadImages(
		readimagesrequest ReadImagesRequest,
	) (
		response *POST_ReadImagesResponses,
		err error,
	)

	POST_ReadInternetServices(
		readinternetservicesrequest ReadInternetServicesRequest,
	) (
		response *POST_ReadInternetServicesResponses,
		err error,
	)

	POST_ReadKeypairs(
		readkeypairsrequest ReadKeypairsRequest,
	) (
		response *POST_ReadKeypairsResponses,
		err error,
	)

	POST_ReadListenerRules(
		readlistenerrulesrequest ReadListenerRulesRequest,
	) (
		response *POST_ReadListenerRulesResponses,
		err error,
	)

	POST_ReadLoadBalancers(
		readloadbalancersrequest ReadLoadBalancersRequest,
	) (
		response *POST_ReadLoadBalancersResponses,
		err error,
	)

	POST_ReadLocations(
		readlocationsrequest ReadLocationsRequest,
	) (
		response *POST_ReadLocationsResponses,
		err error,
	)

	POST_ReadNatServices(
		readnatservicesrequest ReadNatServicesRequest,
	) (
		response *POST_ReadNatServicesResponses,
		err error,
	)

	POST_ReadNetAccessPointServices(
		readnetaccesspointservicesrequest ReadNetAccessPointServicesRequest,
	) (
		response *POST_ReadNetAccessPointServicesResponses,
		err error,
	)

	POST_ReadNetAccessPoints(
		readnetaccesspointsrequest ReadNetAccessPointsRequest,
	) (
		response *POST_ReadNetAccessPointsResponses,
		err error,
	)

	POST_ReadNetPeerings(
		readnetpeeringsrequest ReadNetPeeringsRequest,
	) (
		response *POST_ReadNetPeeringsResponses,
		err error,
	)

	POST_ReadNets(
		readnetsrequest ReadNetsRequest,
	) (
		response *POST_ReadNetsResponses,
		err error,
	)

	POST_ReadNics(
		readnicsrequest ReadNicsRequest,
	) (
		response *POST_ReadNicsResponses,
		err error,
	)

	POST_ReadPolicies(
		readpoliciesrequest ReadPoliciesRequest,
	) (
		response *POST_ReadPoliciesResponses,
		err error,
	)

	POST_ReadPrefixLists(
		readprefixlistsrequest ReadPrefixListsRequest,
	) (
		response *POST_ReadPrefixListsResponses,
		err error,
	)

	POST_ReadProductTypes(
		readproducttypesrequest ReadProductTypesRequest,
	) (
		response *POST_ReadProductTypesResponses,
		err error,
	)

	POST_ReadPublicCatalog(
		readpubliccatalogrequest ReadPublicCatalogRequest,
	) (
		response *POST_ReadPublicCatalogResponses,
		err error,
	)

	POST_ReadPublicIpRanges(
		readpubliciprangesrequest ReadPublicIpRangesRequest,
	) (
		response *POST_ReadPublicIpRangesResponses,
		err error,
	)

	POST_ReadPublicIps(
		readpublicipsrequest ReadPublicIpsRequest,
	) (
		response *POST_ReadPublicIpsResponses,
		err error,
	)

	POST_ReadQuotas(
		readquotasrequest ReadQuotasRequest,
	) (
		response *POST_ReadQuotasResponses,
		err error,
	)

	POST_ReadRegionConfig(
		readregionconfigrequest ReadRegionConfigRequest,
	) (
		response *POST_ReadRegionConfigResponses,
		err error,
	)

	POST_ReadRegions(
		readregionsrequest ReadRegionsRequest,
	) (
		response *POST_ReadRegionsResponses,
		err error,
	)

	POST_ReadReservedVmOffers(
		readreservedvmoffersrequest ReadReservedVmOffersRequest,
	) (
		response *POST_ReadReservedVmOffersResponses,
		err error,
	)

	POST_ReadReservedVms(
		readreservedvmsrequest ReadReservedVmsRequest,
	) (
		response *POST_ReadReservedVmsResponses,
		err error,
	)

	POST_ReadRouteTables(
		readroutetablesrequest ReadRouteTablesRequest,
	) (
		response *POST_ReadRouteTablesResponses,
		err error,
	)

	POST_ReadSecurityGroups(
		readsecuritygroupsrequest ReadSecurityGroupsRequest,
	) (
		response *POST_ReadSecurityGroupsResponses,
		err error,
	)

	POST_ReadServerCertificates(
		readservercertificatesrequest ReadServerCertificatesRequest,
	) (
		response *POST_ReadServerCertificatesResponses,
		err error,
	)

	POST_ReadSnapshotExportTasks(
		readsnapshotexporttasksrequest ReadSnapshotExportTasksRequest,
	) (
		response *POST_ReadSnapshotExportTasksResponses,
		err error,
	)

	POST_ReadSnapshots(
		readsnapshotsrequest ReadSnapshotsRequest,
	) (
		response *POST_ReadSnapshotsResponses,
		err error,
	)

	POST_ReadSubnets(
		readsubnetsrequest ReadSubnetsRequest,
	) (
		response *POST_ReadSubnetsResponses,
		err error,
	)

	POST_ReadSubregions(
		readsubregionsrequest ReadSubregionsRequest,
	) (
		response *POST_ReadSubregionsResponses,
		err error,
	)

	POST_ReadTags(
		readtagsrequest ReadTagsRequest,
	) (
		response *POST_ReadTagsResponses,
		err error,
	)

	POST_ReadUserGroups(
		readusergroupsrequest ReadUserGroupsRequest,
	) (
		response *POST_ReadUserGroupsResponses,
		err error,
	)

	POST_ReadUsers(
		readusersrequest ReadUsersRequest,
	) (
		response *POST_ReadUsersResponses,
		err error,
	)

	POST_ReadVirtualGateways(
		readvirtualgatewaysrequest ReadVirtualGatewaysRequest,
	) (
		response *POST_ReadVirtualGatewaysResponses,
		err error,
	)
	POST_ReadVmTypes(
		readvmtypesrequest ReadVmTypesRequest,
	) (
		response *POST_ReadVmTypesResponses,
		err error,
	)

	POST_ReadVms(
		readvmsrequest ReadVmsRequest,
	) (
		response *POST_ReadVmsResponses,
		err error,
	)

	POST_ReadVmsHealth(
		readvmshealthrequest ReadVmsHealthRequest,
	) (
		response *POST_ReadVmsHealthResponses,
		err error,
	)

	POST_ReadVmsState(
		readvmsstaterequest ReadVmsStateRequest,
	) (
		response *POST_ReadVmsStateResponses,
		err error,
	)

	POST_ReadVolumes(
		readvolumesrequest ReadVolumesRequest,
	) (
		response *POST_ReadVolumesResponses,
		err error,
	)

	POST_ReadVpnConnections(
		readvpnconnectionsrequest ReadVpnConnectionsRequest,
	) (
		response *POST_ReadVpnConnectionsResponses,
		err error,
	)

	POST_RebootVms(
		rebootvmsrequest RebootVmsRequest,
	) (
		response *POST_RebootVmsResponses,
		err error,
	)

	POST_RegisterUserInUserGroup(
		registeruserinusergrouprequest RegisterUserInUserGroupRequest,
	) (
		response *POST_RegisterUserInUserGroupResponses,
		err error,
	)

	POST_RegisterVmsInLoadBalancer(
		registervmsinloadbalancerrequest RegisterVmsInLoadBalancerRequest,
	) (
		response *POST_RegisterVmsInLoadBalancerResponses,
		err error,
	)

	POST_RejectNetPeering(
		rejectnetpeeringrequest RejectNetPeeringRequest,
	) (
		response *POST_RejectNetPeeringResponses,
		err error,
	)

	POST_ResetAccountPassword(
		resetaccountpasswordrequest ResetAccountPasswordRequest,
	) (
		response *POST_ResetAccountPasswordResponses,
		err error,
	)

	POST_SendResetPasswordEmail(
		sendresetpasswordemailrequest SendResetPasswordEmailRequest,
	) (
		response *POST_SendResetPasswordEmailResponses,
		err error,
	)

	POST_StartVms(
		startvmsrequest StartVmsRequest,
	) (
		response *POST_StartVmsResponses,
		err error,
	)

	POST_StopVms(
		stopvmsrequest StopVmsRequest,
	) (
		response *POST_StopVmsResponses,
		err error,
	)

	POST_UnlinkInternetService(
		unlinkinternetservicerequest UnlinkInternetServiceRequest,
	) (
		response *POST_UnlinkInternetServiceResponses,
		err error,
	)

	POST_UnlinkNic(
		unlinknicrequest UnlinkNicRequest,
	) (
		response *POST_UnlinkNicResponses,
		err error,
	)

	POST_UnlinkPolicy(
		unlinkpolicyrequest UnlinkPolicyRequest,
	) (
		response *POST_UnlinkPolicyResponses,
		err error,
	)

	POST_UnlinkPrivateIps(
		unlinkprivateipsrequest UnlinkPrivateIpsRequest,
	) (
		response *POST_UnlinkPrivateIpsResponses,
		err error,
	)

	POST_UnlinkPublicIp(
		unlinkpubliciprequest UnlinkPublicIpRequest,
	) (
		response *POST_UnlinkPublicIpResponses,
		err error,
	)

	POST_UnlinkRouteTable(
		unlinkroutetablerequest UnlinkRouteTableRequest,
	) (
		response *POST_UnlinkRouteTableResponses,
		err error,
	)

	POST_UnlinkVirtualGateway(
		unlinkvirtualgatewayrequest UnlinkVirtualGatewayRequest,
	) (
		response *POST_UnlinkVirtualGatewayResponses,
		err error,
	)

	POST_UnlinkVolume(
		unlinkvolumerequest UnlinkVolumeRequest,
	) (
		response *POST_UnlinkVolumeResponses,
		err error,
	)

	POST_UpdateAccount(
		updateaccountrequest UpdateAccountRequest,
	) (
		response *POST_UpdateAccountResponses,
		err error,
	)

	POST_UpdateApiKey(
		updateapikeyrequest UpdateApiKeyRequest,
	) (
		response *POST_UpdateApiKeyResponses,
		err error,
	)

	POST_UpdateHealthCheck(
		updatehealthcheckrequest UpdateHealthCheckRequest,
	) (
		response *POST_UpdateHealthCheckResponses,
		err error,
	)

	POST_UpdateImage(
		updateimagerequest UpdateImageRequest,
	) (
		response *POST_UpdateImageResponses,
		err error,
	)

	POST_UpdateKeypair(
		updatekeypairrequest UpdateKeypairRequest,
	) (
		response *POST_UpdateKeypairResponses,
		err error,
	)

	POST_UpdateListenerRule(
		updatelistenerrulerequest UpdateListenerRuleRequest,
	) (
		response *POST_UpdateListenerRuleResponses,
		err error,
	)

	POST_UpdateLoadBalancer(
		updateloadbalancerrequest UpdateLoadBalancerRequest,
	) (
		response *POST_UpdateLoadBalancerResponses,
		err error,
	)

	POST_UpdateNet(
		updatenetrequest UpdateNetRequest,
	) (
		response *POST_UpdateNetResponses,
		err error,
	)

	POST_UpdateNetAccessPoint(
		updatenetaccesspointrequest UpdateNetAccessPointRequest,
	) (
		response *POST_UpdateNetAccessPointResponses,
		err error,
	)

	POST_UpdateNic(
		updatenicrequest UpdateNicRequest,
	) (
		response *POST_UpdateNicResponses,
		err error,
	)

	POST_UpdateRoute(
		updaterouterequest UpdateRouteRequest,
	) (
		response *POST_UpdateRouteResponses,
		err error,
	)

	POST_UpdateRoutePropagation(
		updateroutepropagationrequest UpdateRoutePropagationRequest,
	) (
		response *POST_UpdateRoutePropagationResponses,
		err error,
	)

	POST_UpdateServerCertificate(
		updateservercertificaterequest UpdateServerCertificateRequest,
	) (
		response *POST_UpdateServerCertificateResponses,
		err error,
	)

	POST_UpdateSnapshot(
		updatesnapshotrequest UpdateSnapshotRequest,
	) (
		response *POST_UpdateSnapshotResponses,
		err error,
	)

	POST_UpdateUser(
		updateuserrequest UpdateUserRequest,
	) (
		response *POST_UpdateUserResponses,
		err error,
	)

	POST_UpdateUserGroup(
		updateusergrouprequest UpdateUserGroupRequest,
	) (
		response *POST_UpdateUserGroupResponses,
		err error,
	)

	POST_UpdateVm(
		updatevmrequest UpdateVmRequest,
	) (
		response *POST_UpdateVmResponses,
		err error,
	)
}
