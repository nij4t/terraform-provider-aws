package autoscaling

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func DataSourceLaunchConfiguration() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceLaunchConfigurationRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"image_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"iam_instance_profile": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"key_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"user_data": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"security_groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"vpc_classic_link_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"vpc_classic_link_security_groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"associate_public_ip_address": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"spot_price": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ebs_optimized": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"placement_tenancy": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"enable_monitoring": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"ebs_block_device": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"encrypted": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"iops": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"no_device": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"snapshot_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"throughput": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"volume_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"volume_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"ephemeral_block_device": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"virtual_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"metadata_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"http_tokens": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"http_put_response_hop_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},

			"root_block_device": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"encrypted": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"iops": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"throughput": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"volume_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"volume_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceLaunchConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	autoscalingconn := meta.(*conns.AWSClient).AutoScalingConn
	ec2conn := meta.(*conns.AWSClient).EC2Conn

	if v, ok := d.GetOk("name"); ok {
		d.SetId(v.(string))
	}

	describeOpts := autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []*string{aws.String(d.Id())},
	}

	log.Printf("[DEBUG] launch configuration describe configuration: %s", describeOpts)
	describConfs, err := autoscalingconn.DescribeLaunchConfigurations(&describeOpts)
	if err != nil {
		return fmt.Errorf("Error retrieving launch configuration: %w", err)
	}

	if describConfs == nil || len(describConfs.LaunchConfigurations) == 0 {
		return errors.New("No matching Launch Configuration found")
	}

	if len(describConfs.LaunchConfigurations) > 1 {
		return errors.New("Multiple matching Launch Configurations found")
	}

	lc := describConfs.LaunchConfigurations[0]

	d.Set("key_name", lc.KeyName)
	d.Set("image_id", lc.ImageId)
	d.Set("instance_type", lc.InstanceType)
	d.Set("arn", lc.LaunchConfigurationARN)
	d.Set("name", lc.LaunchConfigurationName)
	d.Set("user_data", lc.UserData)
	d.Set("iam_instance_profile", lc.IamInstanceProfile)
	d.Set("ebs_optimized", lc.EbsOptimized)
	d.Set("spot_price", lc.SpotPrice)
	d.Set("associate_public_ip_address", lc.AssociatePublicIpAddress)
	d.Set("vpc_classic_link_id", lc.ClassicLinkVPCId)
	d.Set("enable_monitoring", false)

	if lc.InstanceMonitoring != nil {
		d.Set("enable_monitoring", lc.InstanceMonitoring.Enabled)
	}

	vpcSGs := make([]string, 0, len(lc.SecurityGroups))
	for _, sg := range lc.SecurityGroups {
		vpcSGs = append(vpcSGs, *sg)
	}
	if err := d.Set("security_groups", vpcSGs); err != nil {
		return fmt.Errorf("error setting security_groups: %w", err)
	}

	if err := d.Set("metadata_options", flattenLaunchConfigInstanceMetadataOptions(lc.MetadataOptions)); err != nil {
		return fmt.Errorf("error setting metadata_options: %w", err)
	}

	classicSGs := make([]string, 0, len(lc.ClassicLinkVPCSecurityGroups))
	for _, sg := range lc.ClassicLinkVPCSecurityGroups {
		classicSGs = append(classicSGs, *sg)
	}
	if err := d.Set("vpc_classic_link_security_groups", classicSGs); err != nil {
		return fmt.Errorf("error setting vpc_classic_link_security_groups: %w", err)
	}

	if err := readLCBlockDevices(d, lc, ec2conn); err != nil {
		return err
	}

	return nil
}
