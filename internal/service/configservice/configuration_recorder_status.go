package configservice

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func ResourceConfigurationRecorderStatus() *schema.Resource {
	return &schema.Resource{
		Create: resourceConfigurationRecorderStatusPut,
		Read:   resourceConfigurationRecorderStatusRead,
		Update: resourceConfigurationRecorderStatusPut,
		Delete: resourceConfigurationRecorderStatusDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"is_enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceConfigurationRecorderStatusPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

	name := d.Get("name").(string)
	d.SetId(name)

	if d.HasChange("is_enabled") {
		isEnabled := d.Get("is_enabled").(bool)
		if isEnabled {
			log.Printf("[DEBUG] Starting AWSConfig Configuration recorder %q", name)
			startInput := configservice.StartConfigurationRecorderInput{
				ConfigurationRecorderName: aws.String(name),
			}
			_, err := conn.StartConfigurationRecorder(&startInput)
			if err != nil {
				return fmt.Errorf("Failed to start Configuration Recorder: %s", err)
			}
		} else {
			log.Printf("[DEBUG] Stopping AWSConfig Configuration recorder %q", name)
			stopInput := configservice.StopConfigurationRecorderInput{
				ConfigurationRecorderName: aws.String(name),
			}
			_, err := conn.StopConfigurationRecorder(&stopInput)
			if err != nil {
				return fmt.Errorf("Failed to stop Configuration Recorder: %s", err)
			}
		}
	}

	return resourceConfigurationRecorderStatusRead(d, meta)
}

func resourceConfigurationRecorderStatusRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

	name := d.Id()
	statusInput := configservice.DescribeConfigurationRecorderStatusInput{
		ConfigurationRecorderNames: []*string{aws.String(name)},
	}
	statusOut, err := conn.DescribeConfigurationRecorderStatus(&statusInput)
	if err != nil {
		if tfawserr.ErrMessageContains(err, configservice.ErrCodeNoSuchConfigurationRecorderException, "") {
			log.Printf("[WARN] Configuration Recorder (status) %q is gone (NoSuchConfigurationRecorderException)", name)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Failed describing Configuration Recorder %q status: %s",
			name, err)
	}

	numberOfStatuses := len(statusOut.ConfigurationRecordersStatus)
	if numberOfStatuses < 1 {
		log.Printf("[WARN] Configuration Recorder (status) %q is gone (no recorders found)", name)
		d.SetId("")
		return nil
	}

	if numberOfStatuses > 1 {
		return fmt.Errorf("Expected exactly 1 Configuration Recorder (status), received %d: %#v",
			numberOfStatuses, statusOut.ConfigurationRecordersStatus)
	}

	d.Set("is_enabled", statusOut.ConfigurationRecordersStatus[0].Recording)

	return nil
}

func resourceConfigurationRecorderStatusDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn
	input := configservice.StopConfigurationRecorderInput{
		ConfigurationRecorderName: aws.String(d.Get("name").(string)),
	}
	_, err := conn.StopConfigurationRecorder(&input)
	if err != nil {
		return fmt.Errorf("Stopping Configuration Recorder failed: %s", err)
	}

	return nil
}
