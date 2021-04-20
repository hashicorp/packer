package ncloud

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vpc"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/olekukonko/tablewriter"
)

//StepValidateTemplate : struct for Validation a template
type StepValidateTemplate struct {
	Conn                      *NcloudAPIClient
	Validate                  func() error
	getZone                   func() error
	getMemberServerImageList  func() ([]*server.MemberServerImage, error)
	getServerImageProductList func() ([]*server.Product, error)
	getServerProductList      func(serverImageProductCode string) ([]*server.Product, error)
	Say                       func(message string)
	Error                     func(e error)
	Config                    *Config
	zoneNo                    string
	zoneCode                  string
	regionNo                  string
	regionCode                string
	FeeSystemTypeCode         string
}

// NewStepValidateTemplate : function for Validation a template
func NewStepValidateTemplate(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepValidateTemplate {
	var step = &StepValidateTemplate{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	if config.SupportVPC {
		step.getZone = step.getZoneCode
		step.getMemberServerImageList = step.getVpcMemberServerImageList
		step.getServerImageProductList = step.getVpcServerImageProductList
		step.getServerProductList = step.getVpcServerProductList
	} else {
		step.getZone = step.getZoneNo
		step.getMemberServerImageList = step.getClassicMemberServerImageList
		step.getServerImageProductList = step.getClassicServerImageProductList
		step.getServerProductList = step.getClassicServerProductList
	}

	step.Validate = step.validateTemplate

	return step
}

// getZoneNo : get zoneNo
func (s *StepValidateTemplate) getZoneNo() error {
	if s.Config.Region == "" {
		return nil
	}

	regionList, err := s.Conn.server.V2Api.GetRegionList(&server.GetRegionListRequest{})
	if err != nil {
		return err
	}

	for _, region := range regionList.RegionList {
		if strings.EqualFold(*region.RegionName, s.Config.Region) {
			s.regionNo = *region.RegionNo
			s.regionCode = *region.RegionCode
		}
	}

	if s.regionNo == "" {
		return fmt.Errorf("region %s is invalid", s.Config.Region)
	}

	// Get ZoneNo
	resp, err := s.Conn.server.V2Api.GetZoneList(&server.GetZoneListRequest{RegionNo: &s.regionNo})
	if err != nil {
		return err
	}

	if len(resp.ZoneList) > 0 {
		s.zoneNo = *resp.ZoneList[0].ZoneNo
	}

	return nil
}

// getZoneCode : get zoneCode
func (s *StepValidateTemplate) getZoneCode() error {
	if s.Config.Region == "" {
		return nil
	}

	regionList, err := s.Conn.vserver.V2Api.GetRegionList(&vserver.GetRegionListRequest{})
	if err != nil {
		return err
	}

	for _, region := range regionList.RegionList {
		if strings.EqualFold(*region.RegionName, s.Config.Region) {
			s.regionCode = *region.RegionCode
			s.Config.RegionCode = *region.RegionCode
		}
	}

	if s.regionCode == "" {
		return fmt.Errorf("region %s is invalid", s.Config.Region)
	}

	// Get ZoneNo
	resp, err := s.Conn.vserver.V2Api.GetZoneList(&vserver.GetZoneListRequest{RegionCode: &s.regionCode})
	if err != nil {
		return err
	}

	if len(resp.ZoneList) > 0 {
		s.zoneCode = *resp.ZoneList[0].ZoneCode
	}

	return nil
}

func (s *StepValidateTemplate) validateMemberServerImage(fnGetServerImageList func() ([]*server.MemberServerImage, error)) error {
	var serverImageName = s.Config.ServerImageName

	memberServerImageList, err := fnGetServerImageList()
	if err != nil {
		return err
	}

	var isExistMemberServerImageNo = false
	for _, image := range memberServerImageList {
		// Check duplicate server_image_name
		if *image.MemberServerImageName == serverImageName {
			return fmt.Errorf("server_image_name %s is exists", serverImageName)
		}

		if *image.MemberServerImageNo == s.Config.MemberServerImageNo {
			isExistMemberServerImageNo = true
			if s.Config.ServerProductCode == "" {
				s.Config.ServerProductCode = *image.OriginalServerProductCode
				s.Say("server_product_code for member server image '" + *image.OriginalServerProductCode + "' is configured automatically")
			}
			s.Config.ServerImageProductCode = *image.OriginalServerImageProductCode
		}
	}

	if s.Config.MemberServerImageNo != "" && !isExistMemberServerImageNo {
		return fmt.Errorf("member_server_image_no %s does not exist", s.Config.MemberServerImageNo)
	}

	return nil
}

func (s *StepValidateTemplate) getClassicMemberServerImageList() ([]*server.MemberServerImage, error) {
	reqParams := &server.GetMemberServerImageListRequest{
		RegionNo: &s.regionNo,
	}

	resp, err := s.Conn.server.V2Api.GetMemberServerImageList(reqParams)
	if err != nil {
		return nil, err
	}

	return resp.MemberServerImageList, nil
}

func (s *StepValidateTemplate) getVpcMemberServerImageList() ([]*server.MemberServerImage, error) {
	reqParams := &vserver.GetMemberServerImageInstanceListRequest{
		RegionCode: &s.regionCode,
	}

	memberServerImageList, err := s.Conn.vserver.V2Api.GetMemberServerImageInstanceList(reqParams)
	if err != nil {
		return nil, err
	}

	var results []*server.MemberServerImage
	for _, r := range memberServerImageList.MemberServerImageInstanceList {
		results = append(results, &server.MemberServerImage{
			MemberServerImageNo:          r.MemberServerImageInstanceNo,
			MemberServerImageName:        r.MemberServerImageName,
			MemberServerImageDescription: r.MemberServerImageDescription,
			OriginalServerInstanceNo:     r.OriginalServerInstanceNo,
			OriginalServerProductCode:    r.OriginalServerImageProductCode,
		})
	}

	return results, nil
}

func (s *StepValidateTemplate) getClassicServerImageProductList() ([]*server.Product, error) {
	reqParams := &server.GetServerImageProductListRequest{
		RegionNo: &s.regionNo,
	}

	resp, err := s.Conn.server.V2Api.GetServerImageProductList(reqParams)
	if err != nil {
		return nil, err
	}

	return resp.ProductList, nil
}

func (s *StepValidateTemplate) getVpcServerImageProductList() ([]*server.Product, error) {
	reqParams := &vserver.GetServerImageProductListRequest{
		RegionCode: &s.regionCode,
	}

	resp, err := s.Conn.vserver.V2Api.GetServerImageProductList(reqParams)
	if err != nil {
		return nil, err
	}

	var results []*server.Product
	for _, r := range resp.ProductList {
		results = append(results, &server.Product{
			ProductCode: r.ProductCode,
			ProductName: r.ProductName,
		})
	}

	return results, nil
}

func (s *StepValidateTemplate) validateServerImageProduct() error {
	var serverImageProductCode = s.Config.ServerImageProductCode
	if serverImageProductCode == "" {
		return nil
	}

	serverImageProductList, err := s.getServerImageProductList()
	if err != nil {
		return err
	}

	var isExistServerImage = false
	var buf bytes.Buffer
	var productName string
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Name", "Code"})

	for _, product := range serverImageProductList {
		// Check exist server image product code
		if *product.ProductCode == serverImageProductCode {
			isExistServerImage = true
			productName = *product.ProductName
			break
		}

		table.Append([]string{*product.ProductName, *product.ProductCode})
	}

	if !isExistServerImage {
		table.Render()
		s.Say(buf.String())

		return fmt.Errorf("server_image_product_code %s does not exist", serverImageProductCode)
	}

	if strings.Contains(productName, "mssql") {
		s.FeeSystemTypeCode = "FXSUM"
	}

	return nil
}

func (s *StepValidateTemplate) getClassicServerProductList(serverImageProductCode string) ([]*server.Product, error) {
	reqParams := &server.GetServerProductListRequest{
		ServerImageProductCode: &serverImageProductCode,
		RegionNo:               &s.regionNo,
	}

	resp, err := s.Conn.server.V2Api.GetServerProductList(reqParams)

	if err != nil {
		return nil, err
	}

	return resp.ProductList, nil
}

func (s *StepValidateTemplate) getVpcServerProductList(serverImageProductCode string) ([]*server.Product, error) {
	reqParams := &vserver.GetServerProductListRequest{
		ServerImageProductCode: &serverImageProductCode,
		RegionCode:             &s.regionCode,
	}

	resp, err := s.Conn.vserver.V2Api.GetServerProductList(reqParams)

	if err != nil {
		return nil, err
	}

	var results []*server.Product
	for _, r := range resp.ProductList {
		results = append(results, &server.Product{
			ProductCode: r.ProductCode,
			ProductName: r.ProductName,
			ProductType: &server.CommonCode{
				Code:     r.ProductType.Code,
				CodeName: r.ProductType.CodeName,
			},
		})
	}

	return results, nil
}

func (s *StepValidateTemplate) validateServerProductCode() error {
	var serverImageProductCode = s.Config.ServerImageProductCode
	var productCode = s.Config.ServerProductCode

	productList, err := s.getServerProductList(serverImageProductCode)
	if err != nil {
		return err
	}

	if productCode == "" {
		return nil
	}

	var isExistProductCode = false
	for _, product := range productList {
		// Check exist server image product code
		if *product.ProductCode == productCode {
			isExistProductCode = true
			if strings.Contains(*product.ProductName, "mssql") {
				s.FeeSystemTypeCode = "FXSUM"
			}

			if *product.ProductType.Code == "VDS" {
				return errors.New("You cannot create my server image for VDS servers")
			}

			break
		}
	}

	if !isExistProductCode {
		var buf bytes.Buffer
		table := tablewriter.NewWriter(&buf)
		table.SetHeader([]string{"Name", "Code"})
		for _, product := range productList {
			table.Append([]string{*product.ProductName, *product.ProductCode})
		}
		table.Render()

		s.Say(buf.String())

		return fmt.Errorf("server_product_code %s does not exist", productCode)
	}

	return nil
}

func (s *StepValidateTemplate) validateVpc() error {
	if !s.Config.SupportVPC {
		return nil
	}

	if s.Config.VpcNo != "" {
		reqParam := &vpc.GetVpcDetailRequest{
			RegionCode: &s.Config.RegionCode,
			VpcNo:      &s.Config.VpcNo,
		}

		resp, err := s.Conn.vpc.V2Api.GetVpcDetail(reqParam)
		if err != nil {
			return err
		}

		if resp == nil || *resp.TotalRows == 0 {
			return fmt.Errorf("cloud not found VPC `vpc_no` [%s]", s.Config.VpcNo)
		}
	}

	if s.Config.SubnetNo != "" {
		reqParam := &vpc.GetSubnetDetailRequest{
			RegionCode: &s.Config.RegionCode,
			SubnetNo:   &s.Config.SubnetNo,
		}

		resp, err := s.Conn.vpc.V2Api.GetSubnetDetail(reqParam)
		if err != nil {
			return err
		}

		if resp != nil && *resp.TotalRows > 0 && *resp.SubnetList[0].SubnetType.Code == "PUBLIC" {
			s.Config.VpcNo = *resp.SubnetList[0].VpcNo
			s.Say("Set `vpc_no` is " + s.Config.VpcNo)
		} else {
			return fmt.Errorf("cloud not found public subnet in `subnet_no` [%s]", s.Config.SubnetNo)
		}
	}

	if s.Config.VpcNo != "" && s.Config.SubnetNo == "" {
		reqParam := &vpc.GetSubnetListRequest{
			RegionCode:     &s.Config.RegionCode,
			VpcNo:          &s.Config.VpcNo,
			SubnetTypeCode: ncloud.String("PUBLIC"),
		}

		resp, err := s.Conn.vpc.V2Api.GetSubnetList(reqParam)
		if err != nil {
			return err
		}

		if resp != nil && *resp.TotalRows > 0 {
			s.Config.SubnetNo = *resp.SubnetList[0].SubnetNo
			s.Say("Set `subnet_no` is " + s.Config.SubnetNo)
		} else {
			return fmt.Errorf("cloud not found public subnet in `vpc_no` [%s]", s.Config.VpcNo)
		}
	}

	return nil
}

// Check ImageName / Product Code / Server Image Product Code / Server Product Code...
func (s *StepValidateTemplate) validateTemplate() error {
	// Get RegionNo, ZoneNo
	if err := s.getZone(); err != nil {
		return err
	}

	// Validate member_server_image_no and member_server_image_no
	if err := s.validateMemberServerImage(s.getMemberServerImageList); err != nil {
		return err
	}

	// Validate server_image_product_code
	if err := s.validateServerImageProduct(); err != nil {
		return err
	}

	// Validate VPC
	if err := s.validateVpc(); err != nil {
		return err
	}

	// Validate server_product_code
	return s.validateServerProductCode()
}

// Run : main function for validation a template
func (s *StepValidateTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Validating deployment template ...")

	err := s.Validate()

	state.Put("zone_no", s.zoneNo)

	if s.FeeSystemTypeCode != "" {
		state.Put("fee_system_type_code", s.FeeSystemTypeCode)
	}

	return processStepResult(err, s.Error, state)
}

// Cleanup : cleanup on error
func (s *StepValidateTemplate) Cleanup(multistep.StateBag) {
}
