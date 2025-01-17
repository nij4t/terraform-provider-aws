package configservice

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tftags "github.com/nij4t/terraform-provider-aws/internal/tags"
	"github.com/nij4t/terraform-provider-aws/internal/verify"
)

func ResourceAggregateAuthorization() *schema.Resource {
	return &schema.Resource{
		Create: resourceAggregateAuthorizationPut,
		Read:   resourceAggregateAuthorizationRead,
		Update: resourceAggregateAuthorizationUpdate,
		Delete: resourceAggregateAuthorizationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAggregateAuthorizationPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	accountId := d.Get("account_id").(string)
	region := d.Get("region").(string)

	req := &configservice.PutAggregationAuthorizationInput{
		AuthorizedAccountId: aws.String(accountId),
		AuthorizedAwsRegion: aws.String(region),
		Tags:                Tags(tags.IgnoreAWS()),
	}

	_, err := conn.PutAggregationAuthorization(req)
	if err != nil {
		return fmt.Errorf("Error creating aggregate authorization: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", accountId, region))

	return resourceAggregateAuthorizationRead(d, meta)
}

func resourceAggregateAuthorizationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	accountId, region, err := AggregateAuthorizationParseID(d.Id())
	if err != nil {
		return err
	}

	d.Set("account_id", accountId)
	d.Set("region", region)

	aggregateAuthorizations, err := DescribeAggregateAuthorizations(conn)
	if err != nil {
		return fmt.Errorf("Error retrieving list of aggregate authorizations: %s", err)
	}

	var aggregationAuthorization *configservice.AggregationAuthorization
	// Check for existing authorization
	for _, auth := range aggregateAuthorizations {
		if accountId == aws.StringValue(auth.AuthorizedAccountId) && region == aws.StringValue(auth.AuthorizedAwsRegion) {
			aggregationAuthorization = auth
		}
	}

	if aggregationAuthorization == nil {
		log.Printf("[WARN] Aggregate Authorization not found, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", aggregationAuthorization.AggregationAuthorizationArn)

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for Config Aggregate Authorization (%s): %s", d.Get("arn").(string), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAggregateAuthorizationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Config Aggregate Authorization (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAggregateAuthorizationRead(d, meta)
}

func resourceAggregateAuthorizationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

	accountId, region, err := AggregateAuthorizationParseID(d.Id())
	if err != nil {
		return err
	}

	req := &configservice.DeleteAggregationAuthorizationInput{
		AuthorizedAccountId: aws.String(accountId),
		AuthorizedAwsRegion: aws.String(region),
	}

	_, err = conn.DeleteAggregationAuthorization(req)
	if err != nil {
		return fmt.Errorf("Error deleting aggregate authorization: %s", err)
	}

	return nil
}

func DescribeAggregateAuthorizations(conn *configservice.ConfigService) ([]*configservice.AggregationAuthorization, error) {
	aggregationAuthorizations := []*configservice.AggregationAuthorization{}
	input := &configservice.DescribeAggregationAuthorizationsInput{}

	for {
		output, err := conn.DescribeAggregationAuthorizations(input)
		if err != nil {
			return aggregationAuthorizations, err
		}
		aggregationAuthorizations = append(aggregationAuthorizations, output.AggregationAuthorizations...)
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return aggregationAuthorizations, nil
}

func AggregateAuthorizationParseID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("Please make sure the ID is in the form account_id:region (i.e. 123456789012:us-east-1") // lintignore:AWSAT003
	}
	accountId := idParts[0]
	region := idParts[1]
	return accountId, region, nil
}
