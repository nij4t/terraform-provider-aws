package elasticsearch

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
)

func ResourceDomainSAMLOptions() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainSAMLOptionsPut,
		Read:   resourceDomainSAMLOptionsRead,
		Update: resourceDomainSAMLOptionsPut,
		Delete: resourceDomainSAMLOptionsDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("domain_name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"saml_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"idp": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"entity_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"metadata_content": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
						"master_backend_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"master_user_name": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"roles_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"session_timeout_minutes": {
							Type:             schema.TypeInt,
							Optional:         true,
							Default:          60,
							ValidateFunc:     validation.IntBetween(1, 1440),
							DiffSuppressFunc: elasticsearchDomainSamlOptionsDiffSupress,
						},
						"subject_key": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "NameID",
							DiffSuppressFunc: elasticsearchDomainSamlOptionsDiffSupress,
						},
					},
				},
			},
		},
	}
}
func elasticsearchDomainSamlOptionsDiffSupress(k, old, new string, d *schema.ResourceData) bool {
	if v, ok := d.Get("saml_options").([]interface{}); ok && len(v) > 0 {
		if enabled, ok := v[0].(map[string]interface{})["enabled"].(bool); ok && !enabled {
			return true
		}
	}
	return false
}

func resourceDomainSAMLOptionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticsearchConn

	input := &elasticsearch.DescribeElasticsearchDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	domain, err := conn.DescribeElasticsearchDomain(input)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] Elasticsearch Domain %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] Received Elasticsearch domain: %s", domain)

	ds := domain.DomainStatus
	options := ds.AdvancedSecurityOptions.SAMLOptions

	if err := d.Set("saml_options", flattenESSAMLOptions(d, options)); err != nil {
		return fmt.Errorf("error setting saml_options for Elasticsearch Configuration: %w", err)
	}

	return nil
}

func resourceDomainSAMLOptionsPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticsearchConn

	domainName := d.Get("domain_name").(string)
	config := elasticsearch.AdvancedSecurityOptionsInput{}
	config.SetSAMLOptions(expandESSAMLOptions(d.Get("saml_options").([]interface{})))

	log.Printf("[DEBUG] Updating Elasticsearch domain SAML Options %s", config)

	_, err := conn.UpdateElasticsearchDomainConfig(&elasticsearch.UpdateElasticsearchDomainConfigInput{
		DomainName:              aws.String(domainName),
		AdvancedSecurityOptions: &config,
	})

	if err != nil {
		return err
	}

	d.SetId(domainName)

	input := &elasticsearch.DescribeElasticsearchDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	}
	var out *elasticsearch.DescribeElasticsearchDomainOutput
	err = resource.Retry(50*time.Minute, func() *resource.RetryError {
		var err error
		out, err = conn.DescribeElasticsearchDomain(input)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if !*out.DomainStatus.Processing {
			return nil
		}

		return resource.RetryableError(
			fmt.Errorf("%q: Timeout while waiting for changes to be processed", d.Id()))
	})
	if tfresource.TimedOut(err) {
		out, err = conn.DescribeElasticsearchDomain(input)
		if err == nil && !*out.DomainStatus.Processing {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error updating Elasticsearch domain SAML Options: %s", err)
	}

	return resourceDomainSAMLOptionsRead(d, meta)
}

func resourceDomainSAMLOptionsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticsearchConn

	domainName := d.Get("domain_name").(string)
	config := elasticsearch.AdvancedSecurityOptionsInput{}
	config.SetSAMLOptions(nil)

	_, err := conn.UpdateElasticsearchDomainConfig(&elasticsearch.UpdateElasticsearchDomainConfigInput{
		DomainName:              aws.String(domainName),
		AdvancedSecurityOptions: &config,
	})
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Waiting for Elasticsearch domain SAML Options %q to be deleted", d.Get("domain_name").(string))

	input := &elasticsearch.DescribeElasticsearchDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	}
	var out *elasticsearch.DescribeElasticsearchDomainOutput
	err = resource.Retry(60*time.Minute, func() *resource.RetryError {
		var err error
		out, err = conn.DescribeElasticsearchDomain(input)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if !*out.DomainStatus.Processing {
			return nil
		}

		return resource.RetryableError(
			fmt.Errorf("%q: Timeout while waiting for SAML Options to be deleted", d.Id()))
	})
	if tfresource.TimedOut(err) {
		out, err := conn.DescribeElasticsearchDomain(input)
		if err == nil && !*out.DomainStatus.Processing {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error deleting Elasticsearch domain SAML Options: %s", err)
	}
	return nil
}
