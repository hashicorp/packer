//go:generate mapstructure-to-hcl2 -type Config

package ebssnapshot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Tags                map[string]string `mapstructure:"tags"`
	Description         string            `mapstructure:"description"`
	WaiterMaxAttempts   int               `mapstruction:"waiterMaxAttempts"`
	WaiterDelayInSecond int               `mapstruction:"waiterDelayInSecond"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

func (p *PostProcessor) Configure(raw ...interface{}) error {
	return config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raw...)
}

func cleanUpVolume(ec2_client *ec2.EC2, deleteVolumeInput *ec2.DeleteVolumeInput, ui packer.Ui) {
	_, err := ec2_client.DeleteVolume(deleteVolumeInput)
	if err != nil {
		ui.Message("[Warning] error cleanining up volume : " + err.Error())
	}
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (a packer.Artifact, keep, mustKeep bool, err error) {
	ui.Message("Executing EBS snapshot post-processer")
	ui.Message(fmt.Sprintf("Creating EBS snapshot for %s", artifact.Id()))

	volume_id := artifact.Id()
	bits := strings.Split(volume_id, ":")
	if len(bits) != 2 {
		err = errors.New("Getting mal-formed volume ID, instead " + volume_id)
		return artifact, true, false, err
	}
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(bits[0])},
	)
	ec2_client := ec2.New(sess)

	// clean up the volume
	defer cleanUpVolume(ec2_client, &ec2.DeleteVolumeInput{VolumeId: aws.String(bits[1])}, ui)

	description := p.config.Description
	if description == "" {
		ui.Message("[Warning] provided empty description")
	}

	input := &ec2.CreateSnapshotInput{
		Description: aws.String(description),
		VolumeId:    aws.String(bits[1]),
	}

	if len(p.config.Tags) > 0 {
		input.TagSpecifications = map_to_ec2_tags(p.config.Tags)
	}

	snapshot, err := ec2_client.CreateSnapshot(input)
	if err != nil {
		return artifact, true, false, err
	}

	ui.Message("Created EBS snapshot with tags : " + *snapshot.SnapshotId)
	ui.Message("Waiting for EBS snapshot " + *snapshot.SnapshotId + " to be available")

	// have default value like 10 minutes
	waiterMaxAttempts := 20
	if p.config.WaiterMaxAttempts != 0 {
		waiterMaxAttempts = p.config.WaiterMaxAttempts
	}

	waiterDelayInSecond := 30
	if p.config.WaiterDelayInSecond != 0 {
		waiterDelayInSecond = p.config.WaiterDelayInSecond
	}

	err = ec2_client.WaitUntilSnapshotCompletedWithContext(aws.BackgroundContext(),
		&ec2.DescribeSnapshotsInput{SnapshotIds: []*string{snapshot.SnapshotId}},
		request.WithWaiterMaxAttempts(waiterMaxAttempts),
		request.WithWaiterDelay(request.ConstantWaiterDelay(time.Duration(waiterDelayInSecond)*time.Second)))

	if err != nil {
		return artifact, true, false, err
	}

	ui.Message(fmt.Sprintf("Removing EBS Volume %s", volume_id))
	return artifact, true, false, err
}

func map_to_ec2_tags(tags map[string]string) []*ec2.TagSpecification {
	var ec2_tags []*ec2.Tag
	for key, val := range tags {
		ec2_tags = append(ec2_tags, &ec2.Tag{
			Key:   aws.String(key),
			Value: aws.String(val),
		})
	}
	return []*ec2.TagSpecification{&ec2.TagSpecification{
		ResourceType: aws.String("snapshot"),
		Tags:         ec2_tags,
	}}
}
