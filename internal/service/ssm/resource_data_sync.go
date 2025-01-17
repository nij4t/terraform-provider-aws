package ssm

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
)

func ResourceResourceDataSync() *schema.Resource {
	return &schema.Resource{
		Create: resourceResourceDataSyncCreate,
		Read:   resourceResourceDataSyncRead,
		Delete: resourceResourceDataSyncDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"s3_destination": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_arn": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"bucket_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"sync_format": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  ssm.ResourceDataSyncS3FormatJsonSerDe,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourceResourceDataSyncCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn
	input := &ssm.CreateResourceDataSyncInput{
		S3Destination: expandSsmResourceDataSyncS3Destination(d),
		SyncName:      aws.String(d.Get("name").(string)),
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateResourceDataSync(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, ssm.ErrCodeResourceDataSyncInvalidConfigurationException, "S3 write failed for bucket") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.CreateResourceDataSync(input)
	}

	if err != nil {
		return err
	}

	d.SetId(d.Get("name").(string))
	return resourceResourceDataSyncRead(d, meta)
}

func resourceResourceDataSyncRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	syncItem, err := FindResourceDataSyncItem(conn, d.Id())
	if err != nil {
		return err
	}
	if syncItem == nil {
		d.SetId("")
		return nil
	}
	d.Set("name", syncItem.SyncName)
	d.Set("s3_destination", flattenSsmResourceDataSyncS3Destination(syncItem.S3Destination))
	return nil
}

func resourceResourceDataSyncDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	input := &ssm.DeleteResourceDataSyncInput{
		SyncName: aws.String(d.Get("name").(string)),
	}

	_, err := conn.DeleteResourceDataSync(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, ssm.ErrCodeResourceDataSyncNotFoundException, "") {
			return nil
		}
		return err
	}
	return nil
}

func FindResourceDataSyncItem(conn *ssm.SSM, name string) (*ssm.ResourceDataSyncItem, error) {
	nextToken := ""
	for {
		input := &ssm.ListResourceDataSyncInput{}
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}
		resp, err := conn.ListResourceDataSync(input)
		if err != nil {
			return nil, err
		}
		for _, v := range resp.ResourceDataSyncItems {
			if *v.SyncName == name {
				return v, nil
			}
		}
		if resp.NextToken == nil {
			break
		}
		nextToken = *resp.NextToken
	}
	return nil, nil
}

func flattenSsmResourceDataSyncS3Destination(dest *ssm.ResourceDataSyncS3Destination) []interface{} {
	result := make(map[string]interface{})
	result["bucket_name"] = *dest.BucketName
	result["region"] = *dest.Region
	result["sync_format"] = *dest.SyncFormat
	if dest.AWSKMSKeyARN != nil {
		result["kms_key_arn"] = *dest.AWSKMSKeyARN
	}
	if dest.Prefix != nil {
		result["prefix"] = *dest.Prefix
	}
	return []interface{}{result}
}

func expandSsmResourceDataSyncS3Destination(d *schema.ResourceData) *ssm.ResourceDataSyncS3Destination {
	raw := d.Get("s3_destination").([]interface{})[0].(map[string]interface{})
	s3dest := &ssm.ResourceDataSyncS3Destination{
		BucketName: aws.String(raw["bucket_name"].(string)),
		Region:     aws.String(raw["region"].(string)),
		SyncFormat: aws.String(raw["sync_format"].(string)),
	}
	if v, ok := raw["kms_key_arn"].(string); ok && v != "" {
		s3dest.AWSKMSKeyARN = aws.String(v)
	}
	if v, ok := raw["prefix"].(string); ok && v != "" {
		s3dest.Prefix = aws.String(v)
	}
	return s3dest
}
