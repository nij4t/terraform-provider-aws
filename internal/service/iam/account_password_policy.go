package iam

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func ResourceAccountPasswordPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccountPasswordPolicyUpdate,
		Read:   resourceAccountPasswordPolicyRead,
		Update: resourceAccountPasswordPolicyUpdate,
		Delete: resourceAccountPasswordPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"allow_users_to_change_password": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"expire_passwords": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"hard_expiry": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"max_password_age": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"minimum_password_length": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  6,
			},
			"password_reuse_prevention": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"require_lowercase_characters": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"require_numbers": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"require_symbols": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"require_uppercase_characters": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceAccountPasswordPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	input := &iam.UpdateAccountPasswordPolicyInput{}

	if v, ok := d.GetOk("allow_users_to_change_password"); ok {
		input.AllowUsersToChangePassword = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("hard_expiry"); ok {
		input.HardExpiry = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("max_password_age"); ok {
		input.MaxPasswordAge = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("minimum_password_length"); ok {
		input.MinimumPasswordLength = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("password_reuse_prevention"); ok {
		input.PasswordReusePrevention = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("require_lowercase_characters"); ok {
		input.RequireLowercaseCharacters = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("require_numbers"); ok {
		input.RequireNumbers = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("require_symbols"); ok {
		input.RequireSymbols = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("require_uppercase_characters"); ok {
		input.RequireUppercaseCharacters = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Updating IAM account password policy: %s", input)
	_, err := conn.UpdateAccountPasswordPolicy(input)
	if err != nil {
		return fmt.Errorf("Error updating IAM Password Policy: %s", err)
	}
	log.Println("[DEBUG] IAM account password policy updated")

	d.SetId("iam-account-password-policy")

	return resourceAccountPasswordPolicyRead(d, meta)
}

func resourceAccountPasswordPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	input := &iam.GetAccountPasswordPolicyInput{}
	resp, err := conn.GetAccountPasswordPolicy(input)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "NoSuchEntity" {
			log.Printf("[WARN] IAM account password policy is gone (i.e. default)")
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading IAM account password policy: %s", err)
	}

	log.Printf("[DEBUG] Received IAM account password policy: %s", resp)

	policy := resp.PasswordPolicy

	d.Set("allow_users_to_change_password", policy.AllowUsersToChangePassword)
	d.Set("expire_passwords", policy.ExpirePasswords)
	d.Set("hard_expiry", policy.HardExpiry)
	d.Set("max_password_age", policy.MaxPasswordAge)
	d.Set("minimum_password_length", policy.MinimumPasswordLength)
	d.Set("password_reuse_prevention", policy.PasswordReusePrevention)
	d.Set("require_lowercase_characters", policy.RequireLowercaseCharacters)
	d.Set("require_numbers", policy.RequireNumbers)
	d.Set("require_symbols", policy.RequireSymbols)
	d.Set("require_uppercase_characters", policy.RequireUppercaseCharacters)

	return nil
}

func resourceAccountPasswordPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	log.Println("[DEBUG] Deleting IAM account password policy")
	input := &iam.DeleteAccountPasswordPolicyInput{}
	if _, err := conn.DeleteAccountPasswordPolicy(input); err != nil {
		return fmt.Errorf("Error deleting IAM Password Policy: %s", err)
	}
	log.Println("[DEBUG] Deleted IAM account password policy")

	return nil
}
