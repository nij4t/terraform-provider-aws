package iot

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func ResourceRoleAlias() *schema.Resource {
	return &schema.Resource{
		Create: resourceRoleAliasCreate,
		Read:   resourceRoleAliasRead,
		Update: resourceRoleAliasUpdate,
		Delete: resourceRoleAliasDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"credential_duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3600,
				ValidateFunc: validation.IntBetween(900, 3600),
			},
		},
	}
}

func resourceRoleAliasCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	roleAlias := d.Get("alias").(string)
	roleArn := d.Get("role_arn").(string)
	credentialDuration := d.Get("credential_duration").(int)

	_, err := conn.CreateRoleAlias(&iot.CreateRoleAliasInput{
		RoleAlias:                 aws.String(roleAlias),
		RoleArn:                   aws.String(roleArn),
		CredentialDurationSeconds: aws.Int64(int64(credentialDuration)),
	})

	if err != nil {
		return fmt.Errorf("error creating role alias %s for role %s: %s", roleAlias, roleArn, err)
	}

	d.SetId(roleAlias)
	return resourceRoleAliasRead(d, meta)
}

func GetRoleAliasDescription(conn *iot.IoT, alias string) (*iot.RoleAliasDescription, error) {

	roleAliasDescriptionOutput, err := conn.DescribeRoleAlias(&iot.DescribeRoleAliasInput{
		RoleAlias: aws.String(alias),
	})

	if err != nil {
		return nil, err
	}

	if roleAliasDescriptionOutput == nil {
		return nil, nil
	}

	return roleAliasDescriptionOutput.RoleAliasDescription, nil
}

func resourceRoleAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	var roleAliasDescription *iot.RoleAliasDescription

	roleAliasDescription, err := GetRoleAliasDescription(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error describing role alias %s: %s", d.Id(), err)
	}

	if roleAliasDescription == nil {
		log.Printf("[WARN] Role alias (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", roleAliasDescription.RoleAliasArn)
	d.Set("alias", roleAliasDescription.RoleAlias)
	d.Set("role_arn", roleAliasDescription.RoleArn)
	d.Set("credential_duration", roleAliasDescription.CredentialDurationSeconds)

	return nil
}

func resourceRoleAliasDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	alias := d.Get("alias").(string)

	_, err := conn.DeleteRoleAlias(&iot.DeleteRoleAliasInput{
		RoleAlias: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error deleting role alias %s: %s", alias, err)
	}

	return nil
}

func resourceRoleAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	if d.HasChange("credential_duration") {
		roleAliasInput := &iot.UpdateRoleAliasInput{
			RoleAlias:                 aws.String(d.Id()),
			CredentialDurationSeconds: aws.Int64(int64(d.Get("credential_duration").(int))),
		}
		_, err := conn.UpdateRoleAlias(roleAliasInput)
		if err != nil {
			return fmt.Errorf("Error updating role alias %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("role_arn") {
		roleAliasInput := &iot.UpdateRoleAliasInput{
			RoleAlias: aws.String(d.Id()),
			RoleArn:   aws.String(d.Get("role_arn").(string)),
		}
		_, err := conn.UpdateRoleAlias(roleAliasInput)
		if err != nil {
			return fmt.Errorf("Error updating role alias %s: %s", d.Id(), err)
		}
	}

	return resourceRoleAliasRead(d, meta)
}
