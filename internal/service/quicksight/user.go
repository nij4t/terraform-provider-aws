package quicksight

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserCreate,
		Read:   resourceUserRead,
		Update: resourceUserUpdate,
		Delete: resourceUserDelete,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"aws_account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"iam_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"identity_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					quicksight.IdentityTypeIam,
					quicksight.IdentityTypeQuicksight,
				}, false),
			},

			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "default",
				ValidateFunc: validation.StringInSlice([]string{
					"default",
				}, false),
			},

			"session_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"user_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"user_role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					quicksight.UserRoleReader,
					quicksight.UserRoleAuthor,
					quicksight.UserRoleAdmin,
				}, false),
			},
		},
	}
}

func resourceUserCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).QuickSightConn

	awsAccountID := meta.(*conns.AWSClient).AccountID

	namespace := d.Get("namespace").(string)

	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountID = v.(string)
	}

	createOpts := &quicksight.RegisterUserInput{
		AwsAccountId: aws.String(awsAccountID),
		Email:        aws.String(d.Get("email").(string)),
		IdentityType: aws.String(d.Get("identity_type").(string)),
		Namespace:    aws.String(namespace),
		UserRole:     aws.String(d.Get("user_role").(string)),
	}

	if v, ok := d.GetOk("iam_arn"); ok {
		createOpts.IamArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("session_name"); ok {
		createOpts.SessionName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("user_name"); ok {
		createOpts.UserName = aws.String(v.(string))
	}

	resp, err := conn.RegisterUser(createOpts)
	if err != nil {
		return fmt.Errorf("Error registering QuickSight user: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", awsAccountID, namespace, aws.StringValue(resp.User.UserName)))

	return resourceUserRead(d, meta)
}

func resourceUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).QuickSightConn

	awsAccountID, namespace, userName, err := UserParseID(d.Id())
	if err != nil {
		return err
	}

	descOpts := &quicksight.DescribeUserInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		UserName:     aws.String(userName),
	}

	resp, err := conn.DescribeUser(descOpts)
	if tfawserr.ErrMessageContains(err, quicksight.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] QuickSight User %s is not found", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error describing QuickSight User (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.User.Arn)
	d.Set("aws_account_id", awsAccountID)
	d.Set("email", resp.User.Email)
	d.Set("namespace", namespace)
	d.Set("user_role", resp.User.Role)
	d.Set("user_name", resp.User.UserName)

	return nil
}

func resourceUserUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).QuickSightConn

	awsAccountID, namespace, userName, err := UserParseID(d.Id())
	if err != nil {
		return err
	}

	updateOpts := &quicksight.UpdateUserInput{
		AwsAccountId: aws.String(awsAccountID),
		Email:        aws.String(d.Get("email").(string)),
		Namespace:    aws.String(namespace),
		Role:         aws.String(d.Get("user_role").(string)),
		UserName:     aws.String(userName),
	}

	_, err = conn.UpdateUser(updateOpts)
	if tfawserr.ErrMessageContains(err, quicksight.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] QuickSight User %s is not found", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error updating QuickSight User %s: %s", d.Id(), err)
	}

	return resourceUserRead(d, meta)
}

func resourceUserDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).QuickSightConn

	awsAccountID, namespace, userName, err := UserParseID(d.Id())
	if err != nil {
		return err
	}

	deleteOpts := &quicksight.DeleteUserInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		UserName:     aws.String(userName),
	}

	if _, err := conn.DeleteUser(deleteOpts); err != nil {
		if tfawserr.ErrMessageContains(err, quicksight.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting QuickSight User %s: %s", d.Id(), err)
	}

	return nil
}

func UserParseID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID/NAMESPACE/USER_NAME", id)
	}
	return parts[0], parts[1], parts[2], nil
}
