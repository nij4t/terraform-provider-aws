package ses_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfses "github.com/nij4t/terraform-provider-aws/internal/service/ses"
)

func TestAccSESTemplate_basic(t *testing.T) {
	resourceName := "aws_ses_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var template ses.Template

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSesTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckTemplateResourceBasic1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSesTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "text", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSESTemplate_update(t *testing.T) {
	acctest.Skip(t, "Skip due to SES.UpdateTemplate eventual consistency issues")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_template.test"
	var template ses.Template

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSesTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckTemplateResourceBasic1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSesTemplateExists(resourceName, &template),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ses", fmt.Sprintf("template/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "text", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckTemplateResourceBasic2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSesTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "text", "text"),
				),
			},
			{
				Config: testAccCheckTemplateResourceBasic3Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSesTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html update"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "text", ""),
				),
			},
		},
	})
}

func TestAccSESTemplate_disappears(t *testing.T) {
	resourceName := "aws_ses_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var template ses.Template

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSesTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckTemplateResourceBasic1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSesTemplateExists(resourceName, &template),
					acctest.CheckResourceDisappears(acctest.Provider, tfses.ResourceTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSesTemplateExists(pr string, template *ses.Template) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		input := ses.GetTemplateInput{
			TemplateName: aws.String(rs.Primary.ID),
		}

		templateOutput, err := conn.GetTemplate(&input)
		if err != nil {
			return err
		}

		if templateOutput == nil || templateOutput.Template == nil {
			return fmt.Errorf("SES Template (%s) not found", rs.Primary.ID)
		}

		*template = *templateOutput.Template

		return nil
	}
}

func testAccCheckSesTemplateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_template" {
			continue
		}
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			input := ses.GetTemplateInput{
				TemplateName: aws.String(rs.Primary.ID),
			}

			gto, err := conn.GetTemplate(&input)
			if err != nil {
				if tfawserr.ErrMessageContains(err, ses.ErrCodeTemplateDoesNotExistException, "") {
					return nil
				}
				return resource.NonRetryableError(err)
			}
			if gto.Template != nil {
				return resource.RetryableError(fmt.Errorf("Template exists: %v", gto.Template))
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckTemplateResourceBasic1Config(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_template" "test" {
  name    = "%s"
  subject = "subject"
  html    = "html"
}
`, name)
}

func testAccCheckTemplateResourceBasic2Config(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_template" "test" {
  name    = "%s"
  subject = "subject"
  html    = "html"
  text    = "text"
}
`, name)
}

func testAccCheckTemplateResourceBasic3Config(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_template" "test" {
  name    = "%s"
  subject = "subject"
  html    = "html update"
}
`, name)
}
