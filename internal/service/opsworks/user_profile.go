package opsworks

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func ResourceUserProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserProfileCreate,
		Read:   resourceUserProfileRead,
		Update: resourceUserProfileUpdate,
		Delete: resourceUserProfileDelete,

		Schema: map[string]*schema.Schema{
			"user_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"allow_self_management": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"ssh_username": {
				Type:     schema.TypeString,
				Required: true,
			},

			"ssh_public_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceUserProfileRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*conns.AWSClient).OpsWorksConn

	req := &opsworks.DescribeUserProfilesInput{
		IamUserArns: []*string{
			aws.String(d.Id()),
		},
	}

	log.Printf("[DEBUG] Reading OpsWorks user profile: %s", d.Id())

	resp, err := client.DescribeUserProfiles(req)
	if err != nil {
		if awserr, ok := err.(awserr.Error); ok {
			if awserr.Code() == "ResourceNotFoundException" {
				log.Printf("[DEBUG] OpsWorks user profile (%s) not found", d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	for _, profile := range resp.UserProfiles {
		d.Set("allow_self_management", profile.AllowSelfManagement)
		d.Set("user_arn", profile.IamUserArn)
		d.Set("ssh_public_key", profile.SshPublicKey)
		d.Set("ssh_username", profile.SshUsername)
		break
	}

	return nil
}

func resourceUserProfileCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*conns.AWSClient).OpsWorksConn

	req := &opsworks.CreateUserProfileInput{
		AllowSelfManagement: aws.Bool(d.Get("allow_self_management").(bool)),
		IamUserArn:          aws.String(d.Get("user_arn").(string)),
		SshPublicKey:        aws.String(d.Get("ssh_public_key").(string)),
		SshUsername:         aws.String(d.Get("ssh_username").(string)),
	}

	resp, err := client.CreateUserProfile(req)
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(resp.IamUserArn))

	return resourceUserProfileUpdate(d, meta)
}

func resourceUserProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*conns.AWSClient).OpsWorksConn

	req := &opsworks.UpdateUserProfileInput{
		AllowSelfManagement: aws.Bool(d.Get("allow_self_management").(bool)),
		IamUserArn:          aws.String(d.Get("user_arn").(string)),
		SshPublicKey:        aws.String(d.Get("ssh_public_key").(string)),
		SshUsername:         aws.String(d.Get("ssh_username").(string)),
	}

	log.Printf("[DEBUG] Updating OpsWorks user profile: %s", req)

	_, err := client.UpdateUserProfile(req)
	if err != nil {
		return err
	}

	return resourceUserProfileRead(d, meta)
}

func resourceUserProfileDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*conns.AWSClient).OpsWorksConn

	req := &opsworks.DeleteUserProfileInput{
		IamUserArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting OpsWorks user profile: %s", d.Id())

	_, err := client.DeleteUserProfile(req)

	return err
}
