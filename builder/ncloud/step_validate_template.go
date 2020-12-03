package ncloud

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/olekukonko/tablewriter"
)

//StepValidateTemplate : struct for Validation a template
type StepValidateTemplate struct {
	Conn              *NcloudAPIClient
	Validate          func() error
	Say               func(message string)
	Error             func(e error)
	Config            *Config
	zoneNo            string
	regionNo          string
	FeeSystemTypeCode string
}

// NewStepValidateTemplate : function for Validation a template
func NewStepValidateTemplate(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepValidateTemplate {
	var step = &StepValidateTemplate{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
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

	var regionNo string
	for _, region := range regionList.RegionList {
		if strings.EqualFold(*region.RegionName, s.Config.Region) {
			regionNo = *region.RegionNo
		}
	}

	if regionNo == "" {
		return fmt.Errorf("region %s is invalid", s.Config.Region)
	}

	s.regionNo = regionNo

	// Get ZoneNo
	resp, err := s.Conn.server.V2Api.GetZoneList(&server.GetZoneListRequest{RegionNo: &regionNo})
	if err != nil {
		return err
	}

	if len(resp.ZoneList) > 0 {
		s.zoneNo = *resp.ZoneList[0].ZoneNo
	}

	return nil
}

func (s *StepValidateTemplate) validateMemberServerImage() error {
	var serverImageName = s.Config.ServerImageName

	reqParams := new(server.GetMemberServerImageListRequest)
	reqParams.RegionNo = &s.regionNo

	memberServerImageList, err := s.Conn.server.V2Api.GetMemberServerImageList(reqParams)
	if err != nil {
		return err
	}

	var isExistMemberServerImageNo = false
	for _, image := range memberServerImageList.MemberServerImageList {
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

func (s *StepValidateTemplate) validateServerImageProduct() error {
	var serverImageProductCode = s.Config.ServerImageProductCode
	if serverImageProductCode == "" {
		return nil
	}

	reqParams := new(server.GetServerImageProductListRequest)
	reqParams.RegionNo = &s.regionNo

	serverImageProductList, err := s.Conn.server.V2Api.GetServerImageProductList(reqParams)
	if err != nil {
		return err
	}

	var isExistServerImage = false
	var buf bytes.Buffer
	var productName string
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Name", "Code"})

	for _, product := range serverImageProductList.ProductList {
		// Check exist server image product code
		if *product.ProductCode == serverImageProductCode {
			isExistServerImage = true
			productName = *product.ProductName
			break
		}

		table.Append([]string{*product.ProductName, *product.ProductCode})
	}

	if !isExistServerImage {
		reqParams.BlockStorageSize = ncloud.Int32(100)

		serverImageProductList, err := s.Conn.server.V2Api.GetServerImageProductList(reqParams)
		if err != nil {
			return err
		}

		for _, product := range serverImageProductList.ProductList {
			// Check exist server image product code
			if *product.ProductCode == serverImageProductCode {
				isExistServerImage = true
				productName = *product.ProductName
				break
			}

			table.Append([]string{*product.ProductName, *product.ProductCode})
		}
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

func (s *StepValidateTemplate) validateServerProductCode() error {
	var serverImageProductCode = s.Config.ServerImageProductCode
	var productCode = s.Config.ServerProductCode

	reqParams := new(server.GetServerProductListRequest)
	reqParams.ServerImageProductCode = &serverImageProductCode
	reqParams.RegionNo = &s.regionNo

	resp, err := s.Conn.server.V2Api.GetServerProductList(reqParams)
	if err != nil {
		return err
	}

	var isExistProductCode = false
	for _, product := range resp.ProductList {
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
		} else if productCode == "" && *product.ProductType.Code == "STAND" {
			isExistProductCode = true
			s.Config.ServerProductCode = *product.ProductCode
			s.Say("server_product_code '" + *product.ProductCode + "' is configured automatically")
			break
		}
	}

	if !isExistProductCode {
		var buf bytes.Buffer
		table := tablewriter.NewWriter(&buf)
		table.SetHeader([]string{"Name", "Code"})
		for _, product := range resp.ProductList {
			table.Append([]string{*product.ProductName, *product.ProductCode})
		}
		table.Render()

		s.Say(buf.String())

		return fmt.Errorf("server_product_code %s does not exist", productCode)
	}

	return nil
}

// Check ImageName / Product Code / Server Image Product Code / Server Product Code...
func (s *StepValidateTemplate) validateTemplate() error {
	// Get RegionNo, ZoneNo
	if err := s.getZoneNo(); err != nil {
		return err
	}

	// Validate member_server_image_no and member_server_image_no
	if err := s.validateMemberServerImage(); err != nil {
		return err
	}

	// Validate server_image_product_code
	if err := s.validateServerImageProduct(); err != nil {
		return err
	}

	// Validate server_product_code
	return s.validateServerProductCode()
}

// Run : main function for validation a template
func (s *StepValidateTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Validating deployment template ...")

	err := s.Validate()

	state.Put("ZoneNo", s.zoneNo)

	if s.FeeSystemTypeCode != "" {
		state.Put("FeeSystemTypeCode", s.FeeSystemTypeCode)
	}

	return processStepResult(err, s.Error, state)
}

// Cleanup : cleanup on error
func (s *StepValidateTemplate) Cleanup(multistep.StateBag) {
}
