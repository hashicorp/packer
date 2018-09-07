package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStepSourceAmiInfo_PVImage(t *testing.T) {
	err := new(StepSourceAMIInfo).canEnableEnhancedNetworking(&ec2.Image{
		VirtualizationType: aws.String("paravirtual"),
	})
	assert.Error(t, err)
}

func TestStepSourceAmiInfo_HVMImage(t *testing.T) {
	err := new(StepSourceAMIInfo).canEnableEnhancedNetworking(&ec2.Image{
		VirtualizationType: aws.String("hvm"),
	})
	assert.NoError(t, err)
}

func TestStepSourceAmiInfo_PVImageWithAMIVirtPV(t *testing.T) {
	stepSourceAMIInfo := StepSourceAMIInfo{
		AMIVirtType: "paravirtual",
	}
	err := stepSourceAMIInfo.canEnableEnhancedNetworking(&ec2.Image{
		VirtualizationType: aws.String("paravirtual"),
	})
	assert.Error(t, err)
}

func TestStepSourceAmiInfo_PVImageWithAMIVirtHVM(t *testing.T) {
	stepSourceAMIInfo := StepSourceAMIInfo{
		AMIVirtType: "hvm",
	}
	err := stepSourceAMIInfo.canEnableEnhancedNetworking(&ec2.Image{
		VirtualizationType: aws.String("paravirtual"),
	})
	assert.NoError(t, err)
}
