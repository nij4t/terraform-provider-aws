package ecr

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfiam "github.com/nij4t/terraform-provider-aws/internal/service/iam"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
	"github.com/nij4t/terraform-provider-aws/internal/verify"
)

func ResourceRepositoryPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceRepositoryPolicyPut,
		Read:   resourceRepositoryPolicyRead,
		Update: resourceRepositoryPolicyPut,
		Delete: resourceRepositoryPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRepositoryPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn

	input := ecr.SetRepositoryPolicyInput{
		RepositoryName: aws.String(d.Get("repository").(string)),
		PolicyText:     aws.String(d.Get("policy").(string)),
	}

	log.Printf("[DEBUG] Creating ECR repository policy: %#v", input)

	// Retry due to IAM eventual consistency
	var err error
	var out *ecr.SetRepositoryPolicyOutput
	err = resource.Retry(tfiam.PropagationTimeout, func() *resource.RetryError {
		out, err = conn.SetRepositoryPolicy(&input)

		if tfawserr.ErrMessageContains(err, ecr.ErrCodeInvalidParameterException, "Invalid repository policy provided") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.SetRepositoryPolicy(&input)
	}
	if err != nil {
		return fmt.Errorf("error creating ECR Repository Policy: %w", err)
	}

	log.Printf("[DEBUG] ECR repository policy created: %s", aws.StringValue(out.RepositoryName))

	d.SetId(aws.StringValue(out.RepositoryName))

	return resourceRepositoryPolicyRead(d, meta)
}

func resourceRepositoryPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn

	input := &ecr.GetRepositoryPolicyInput{
		RepositoryName: aws.String(d.Id()),
	}

	var out *ecr.GetRepositoryPolicyOutput

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		out, err = conn.GetRepositoryPolicy(input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
			return resource.RetryableError(err)
		}

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryPolicyNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		out, err = conn.GetRepositoryPolicy(input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
		log.Printf("[WARN] ECR Repository Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryPolicyNotFoundException) {
		log.Printf("[WARN] ECR Repository Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading ECR Repository Policy (%s): %w", d.Id(), err)
	}

	if out == nil {
		return fmt.Errorf("error reading ECR Repository Policy (%s): empty response", d.Id())
	}

	log.Printf("[DEBUG] Received repository policy %s", out)

	d.Set("repository", out.RepositoryName)
	d.Set("registry_id", out.RegistryId)
	d.Set("policy", out.PolicyText)

	return nil
}

func resourceRepositoryPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn

	_, err := conn.DeleteRepositoryPolicy(&ecr.DeleteRepositoryPolicyInput{
		RepositoryName: aws.String(d.Id()),
		RegistryId:     aws.String(d.Get("registry_id").(string)),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, ecr.ErrCodeRepositoryNotFoundException, "") ||
			tfawserr.ErrMessageContains(err, ecr.ErrCodeRepositoryPolicyNotFoundException, "") {
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] repository policy %s deleted.", d.Id())

	return nil
}
