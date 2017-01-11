package ebssnapshot

import (
	"errors"
	"fmt"
	"log"
	_ "strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/errwrap"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/builder/amazon/ebsvolume"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

// Configuration of this post processor
type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	awscommon.AccessConfig `mapstructure:",squash"`

	Tags map[string]string `mapstructure:"tags"`

	SnapDescription string `mapstructure:"description"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	p.config.ctx.Funcs = awscommon.TemplateFuncs
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	errs := new(packer.MultiError)

	// Check we have AWS access variables defined somewhere
	errs = packer.MultiErrorAppend(errs, p.config.AccessConfig.Prepare(&p.config.ctx)...)

	// Anything which flagged return back up the stack
	if len(errs.Errors) > 0 {
		return errs
	}

	log.Println(common.ScrubConfig(p.config, p.config.AccessKey, p.config.SecretKey))
	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	var err error

	switch artifact.BuilderId() {
	case ebsvolume.BuilderId:
		break
	default:
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from ebsvolume builder artifacts.",
			artifact.BuilderId())
		return nil, false, err
	}

	if len(artifact.Id()) == 0 {
		ui.Message("There is no ebsvolume artifact.")
		return nil, false, nil
	}

	volumeArtifacts := strings.Split(artifact.Id(), ",")
	if len(volumeArtifacts) == 0 {
		ui.Message("There is no ebsvolume artifact.")
		return nil, false, nil
	}

	config, err := p.config.Config()
	if err != nil {
		return nil, false, err
	}

	session, err := session.NewSession(config)
	if err != nil {
		return nil, false, errwrap.Wrapf("Error creating AWS Session: {{err}}", err)
	}

	ec2conn := ec2.New(session)

	var snapshotIdList []string
	for _, v := range volumeArtifacts {
		volumeMetadata := strings.Split(v, ":")

		if len(volumeMetadata) != 2 {
			return nil, false, fmt.Errorf("Unexpected artifact information:", v)
		}

		volumeRegion := volumeMetadata[0]
		volumeId := volumeMetadata[1]

		if strings.Compare(*config.Region, volumeRegion) == 0 {
			ui.Say(fmt.Sprintf("Creating snapshot of EBS Volume %s in region %s...", volumeId, volumeRegion))
			createSnapResp, err := ec2conn.CreateSnapshot(&ec2.CreateSnapshotInput{
				VolumeId:    &volumeId,
				Description: &p.config.SnapDescription,
			})

			if err != nil {
				return nil, false, errwrap.Wrapf("Error creating EBS Volume Snapshot: {{err}}", err)
			}

			snapshotId := *createSnapResp.SnapshotId
			snapshotIdList = append(snapshotIdList, snapshotId)
			ui.Say(fmt.Sprintf("Creating snapshot with ID %s", snapshotId))

			// Wait for snapshot to be completed

			stateChange := awscommon.StateChangeConf{
				Pending: []string{"pending"},
				Target:  "completed",
				Refresh: func() (interface{}, string, error) {
					resp, err := ec2conn.DescribeSnapshots(&ec2.DescribeSnapshotsInput{SnapshotIds: []*string{&snapshotId}})
					if err != nil {
						return nil, "", err
					}

					if len(resp.Snapshots) == 0 {
						return nil, "", errors.New("No snapshots found.")
					}

					s := resp.Snapshots[0]
					return s, *s.State, nil
					//return nil, "", nil
				},
			}

			_, err = awscommon.WaitForState(&stateChange)
			if err != nil {
				return nil, false, fmt.Errorf("Error waiting for snapshot: %s", err)
			}

			if len(p.config.Tags) > 0 {
				var ec2Tags []*ec2.Tag

				log.Printf("Repacking tags into AWS format")

				for key, value := range p.config.Tags {
					ui.Message(fmt.Sprintf("Adding tag \"%s\": \"%s\"", key, value))
					ec2Tags = append(ec2Tags, &ec2.Tag{
						Key:   aws.String(key),
						Value: aws.String(value),
					})
				}

				resourceIds := []*string{&snapshotId}

				ui.Message(fmt.Sprintf("Tagging snapshotIds %s", resourceIds))

				_, err = ec2conn.CreateTags(&ec2.CreateTagsInput{
					Resources: resourceIds,
					Tags:      ec2Tags,
				})

				if err != nil {
					return nil, false, fmt.Errorf("Failed to add tags to resources %#v: %s", resourceIds, err)
				}

			}
		} else {
			ui.Message("Can't make snapshot of volume " + volumeMetadata[1] + " as we are in region " + *config.Region + " and volume is in region " + volumeMetadata[0])
			return nil, true, nil
		}
	}

	// Build the artifact and return it
	newartifact := &Artifact{
		Snapshots: snapshotIdList,
	}

	return newartifact, true, nil
}
