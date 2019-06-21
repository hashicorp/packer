package ecs

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

type ClientWrapper struct {
	*ecs.Client
}

const (
	InstanceStatusRunning  = "Running"
	InstanceStatusStarting = "Starting"
	InstanceStatusStopped  = "Stopped"
	InstanceStatusStopping = "Stopping"
)

const (
	ImageStatusWaiting      = "Waiting"
	ImageStatusCreating     = "Creating"
	ImageStatusCreateFailed = "CreateFailed"
	ImageStatusAvailable    = "Available"
)

var ImageStatusQueried = fmt.Sprintf("%s,%s,%s,%s", ImageStatusWaiting, ImageStatusCreating, ImageStatusCreateFailed, ImageStatusAvailable)

const (
	SnapshotStatusAll          = "all"
	SnapshotStatusProgressing  = "progressing"
	SnapshotStatusAccomplished = "accomplished"
	SnapshotStatusFailed       = "failed"
)

const (
	DiskStatusInUse     = "In_use"
	DiskStatusAvailable = "Available"
	DiskStatusAttaching = "Attaching"
	DiskStatusDetaching = "Detaching"
	DiskStatusCreating  = "Creating"
	DiskStatusReIniting = "ReIniting"
)

const (
	VpcStatusPending   = "Pending"
	VpcStatusAvailable = "Available"
)

const (
	VSwitchStatusPending   = "Pending"
	VSwitchStatusAvailable = "Available"
)

const (
	EipStatusAssociating   = "Associating"
	EipStatusUnassociating = "Unassociating"
	EipStatusInUse         = "InUse"
	EipStatusAvailable     = "Available"
)

const (
	ImageOwnerSystem      = "system"
	ImageOwnerSelf        = "self"
	ImageOwnerOthers      = "others"
	ImageOwnerMarketplace = "marketplace"
)

const (
	IOOptimizedNone      = "none"
	IOOptimizedOptimized = "optimized"
)

const (
	InstanceNetworkClassic = "classic"
	InstanceNetworkVpc     = "vpc"
)

const (
	DiskTypeSystem = "system"
	DiskTypeData   = "data"
)

const (
	TagResourceImage    = "image"
	TagResourceInstance = "instance"
	TagResourceSnapshot = "snapshot"
	TagResourceDisk     = "disk"
)

const (
	IpProtocolAll  = "all"
	IpProtocolTCP  = "tcp"
	IpProtocolUDP  = "udp"
	IpProtocolICMP = "icmp"
	IpProtocolGRE  = "gre"
)

const (
	NicTypeInternet = "internet"
	NicTypeIntranet = "intranet"
)

const (
	DefaultPortRange = "-1/-1"
	DefaultCidrIp    = "0.0.0.0/0"
	DefaultCidrBlock = "172.16.0.0/24"
)

const (
	defaultRetryInterval = 5 * time.Second
	defaultRetryTimes    = 12
	shortRetryTimes      = 36
	mediumRetryTimes     = 360
	longRetryTimes       = 720
)

type WaitForExpectEvalResult struct {
	evalPass  bool
	stopRetry bool
}

var (
	WaitForExpectSuccess = WaitForExpectEvalResult{
		evalPass:  true,
		stopRetry: true,
	}

	WaitForExpectToRetry = WaitForExpectEvalResult{
		evalPass:  false,
		stopRetry: false,
	}

	WaitForExpectFailToStop = WaitForExpectEvalResult{
		evalPass:  false,
		stopRetry: true,
	}
)

type WaitForExpectArgs struct {
	RequestFunc   func() (responses.AcsResponse, error)
	EvalFunc      func(response responses.AcsResponse, err error) WaitForExpectEvalResult
	RetryInterval time.Duration
	RetryTimes    int
	RetryTimeout  time.Duration
}

func (c *ClientWrapper) WaitForExpected(args *WaitForExpectArgs) (responses.AcsResponse, error) {
	if args.RetryInterval <= 0 {
		args.RetryInterval = defaultRetryInterval
	}
	if args.RetryTimes <= 0 {
		args.RetryTimes = defaultRetryTimes
	}

	var timeoutPoint time.Time
	if args.RetryTimeout > 0 {
		timeoutPoint = time.Now().Add(args.RetryTimeout)
	}

	var lastResponse responses.AcsResponse
	var lastError error

	for i := 0; ; i++ {
		if args.RetryTimeout > 0 && time.Now().After(timeoutPoint) {
			break
		}

		if args.RetryTimeout <= 0 && i >= args.RetryTimes {
			break
		}

		response, err := args.RequestFunc()
		lastResponse = response
		lastError = err

		evalResult := args.EvalFunc(response, err)
		if evalResult.evalPass {
			return response, nil
		}
		if evalResult.stopRetry {
			return response, err
		}

		time.Sleep(args.RetryInterval)
	}

	if lastError == nil {
		lastError = fmt.Errorf("<no error>")
	}

	if args.RetryTimeout > 0 {
		return lastResponse, fmt.Errorf("evaluate failed after %d seconds timeout with %d seconds retry interval: %s", int(args.RetryTimeout.Seconds()), int(args.RetryInterval.Seconds()), lastError)
	}

	return lastResponse, fmt.Errorf("evaluate failed after %d times retry with %d seconds retry interval: %s", args.RetryTimes, int(args.RetryInterval.Seconds()), lastError)
}

func (c *ClientWrapper) WaitForInstanceStatus(regionId string, instanceId string, expectedStatus string) (responses.AcsResponse, error) {
	return c.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			request := ecs.CreateDescribeInstancesRequest()
			request.RegionId = regionId
			request.InstanceIds = fmt.Sprintf("[\"%s\"]", instanceId)
			return c.DescribeInstances(request)
		},
		EvalFunc: func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
			if err != nil {
				return WaitForExpectToRetry
			}

			instancesResponse := response.(*ecs.DescribeInstancesResponse)
			instances := instancesResponse.Instances.Instance
			for _, instance := range instances {
				if instance.Status == expectedStatus {
					return WaitForExpectSuccess
				}
			}
			return WaitForExpectToRetry
		},
		RetryTimes: mediumRetryTimes,
	})
}

func (c *ClientWrapper) WaitForImageStatus(regionId string, imageId string, expectedStatus string, timeout time.Duration) (responses.AcsResponse, error) {
	return c.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			request := ecs.CreateDescribeImagesRequest()
			request.RegionId = regionId
			request.ImageId = imageId
			request.Status = ImageStatusQueried
			return c.DescribeImages(request)
		},
		EvalFunc: func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
			if err != nil {
				return WaitForExpectToRetry
			}

			imagesResponse := response.(*ecs.DescribeImagesResponse)
			images := imagesResponse.Images.Image
			for _, image := range images {
				if image.Status == expectedStatus {
					return WaitForExpectSuccess
				}
			}

			return WaitForExpectToRetry
		},
		RetryTimeout: timeout,
	})
}

func (c *ClientWrapper) WaitForSnapshotStatus(regionId string, snapshotId string, expectedStatus string, timeout time.Duration) (responses.AcsResponse, error) {
	return c.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			request := ecs.CreateDescribeSnapshotsRequest()
			request.RegionId = regionId
			request.SnapshotIds = fmt.Sprintf("[\"%s\"]", snapshotId)
			return c.DescribeSnapshots(request)
		},
		EvalFunc: func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
			if err != nil {
				return WaitForExpectToRetry
			}

			snapshotsResponse := response.(*ecs.DescribeSnapshotsResponse)
			snapshots := snapshotsResponse.Snapshots.Snapshot
			for _, snapshot := range snapshots {
				if snapshot.Status == expectedStatus {
					return WaitForExpectSuccess
				}
			}
			return WaitForExpectToRetry
		},
		RetryTimeout: timeout,
	})
}

type EvalErrorType bool

const (
	EvalRetryErrorType    = EvalErrorType(true)
	EvalNotRetryErrorType = EvalErrorType(false)
)

func (c *ClientWrapper) EvalCouldRetryResponse(evalErrors []string, evalErrorType EvalErrorType) func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
	return func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
		if err == nil {
			return WaitForExpectSuccess
		}

		e, ok := err.(errors.Error)
		if !ok {
			return WaitForExpectToRetry
		}

		if evalErrorType == EvalRetryErrorType && !ContainsInArray(evalErrors, e.ErrorCode()) {
			return WaitForExpectFailToStop
		}

		if evalErrorType == EvalNotRetryErrorType && ContainsInArray(evalErrors, e.ErrorCode()) {
			return WaitForExpectFailToStop
		}

		return WaitForExpectToRetry
	}
}
