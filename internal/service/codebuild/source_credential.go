package codebuild

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func ResourceSourceCredential() *schema.Resource {
	return &schema.Resource{
		Create: resourceSourceCredentialCreate,
		Read:   resourceSourceCredentialRead,
		Delete: resourceSourceCredentialDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					codebuild.AuthTypeBasicAuth,
					codebuild.AuthTypePersonalAccessToken,
				}, false),
			},
			"server_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					codebuild.ServerTypeGithub,
					codebuild.ServerTypeBitbucket,
					codebuild.ServerTypeGithubEnterprise,
				}, false),
			},
			"token": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSourceCredentialCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn

	authType := d.Get("auth_type").(string)

	createOpts := &codebuild.ImportSourceCredentialsInput{
		AuthType:   aws.String(authType),
		ServerType: aws.String(d.Get("server_type").(string)),
		Token:      aws.String(d.Get("token").(string)),
	}

	if attr, ok := d.GetOk("user_name"); ok && authType == codebuild.AuthTypeBasicAuth {
		createOpts.Username = aws.String(attr.(string))
	}

	resp, err := conn.ImportSourceCredentials(createOpts)
	if err != nil {
		return fmt.Errorf("Error importing source credentials: %s", err)
	}

	d.SetId(aws.StringValue(resp.Arn))

	return resourceSourceCredentialRead(d, meta)
}

func resourceSourceCredentialRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn

	resp, err := conn.ListSourceCredentials(&codebuild.ListSourceCredentialsInput{})
	if err != nil {
		return fmt.Errorf("Error List CodeBuild Source Credential: %s", err)
	}

	var info *codebuild.SourceCredentialsInfo

	for _, sourceCredentialsInfo := range resp.SourceCredentialsInfos {
		if d.Id() == aws.StringValue(sourceCredentialsInfo.Arn) {
			info = sourceCredentialsInfo
			break
		}
	}

	if info == nil {
		log.Printf("[WARN] CodeBuild Source Credential (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", info.Arn)
	d.Set("auth_type", info.AuthType)
	d.Set("server_type", info.ServerType)

	return nil
}

func resourceSourceCredentialDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn

	deleteOpts := &codebuild.DeleteSourceCredentialsInput{
		Arn: aws.String(d.Id()),
	}

	if _, err := conn.DeleteSourceCredentials(deleteOpts); err != nil {
		if tfawserr.ErrMessageContains(err, codebuild.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting Source Credentials(%s): %s", d.Id(), err)
	}

	return nil
}
