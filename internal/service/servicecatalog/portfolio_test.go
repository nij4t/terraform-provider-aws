package servicecatalog_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func TestAccServiceCatalogPortfolio_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_portfolio.test"
	name := sdkacctest.RandString(5)
	var dpo servicecatalog.DescribePortfolioOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckServiceCatlaogPortfolioDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPortfolioResourceBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolio(resourceName, &dpo),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "catalog", regexp.MustCompile(`portfolio/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "test-2"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", "test-3"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccServiceCatalogPortfolio_disappears(t *testing.T) {
	name := sdkacctest.RandString(5)
	resourceName := "aws_servicecatalog_portfolio.test"
	var dpo servicecatalog.DescribePortfolioOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckServiceCatlaogPortfolioDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPortfolioResourceBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolio(resourceName, &dpo),
					testAccCheckServiceCatlaogPortfolioDisappears(&dpo),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceCatalogPortfolio_tags(t *testing.T) {
	resourceName := "aws_servicecatalog_portfolio.test"
	name := sdkacctest.RandString(5)
	var dpo servicecatalog.DescribePortfolioOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckServiceCatlaogPortfolioDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPortfolioResourceTags1Config(name, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolio(resourceName, &dpo),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckPortfolioResourceTags2Config(name, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolio(resourceName, &dpo),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCheckPortfolioResourceTags1Config(name, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPortfolio(resourceName, &dpo),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckPortfolio(pr string, dpo *servicecatalog.DescribePortfolioOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		input := servicecatalog.DescribePortfolioInput{}
		input.Id = aws.String(rs.Primary.ID)

		resp, err := conn.DescribePortfolio(&input)
		if err != nil {
			return err
		}

		*dpo = *resp
		return nil
	}
}

func testAccCheckServiceCatlaogPortfolioDisappears(dpo *servicecatalog.DescribePortfolioOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

		input := servicecatalog.DeletePortfolioInput{}
		input.Id = dpo.PortfolioDetail.Id

		_, err := conn.DeletePortfolio(&input)
		return err
	}
}

func testAccCheckServiceCatlaogPortfolioDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_portfolio" {
			continue
		}
		input := servicecatalog.DescribePortfolioInput{}
		input.Id = aws.String(rs.Primary.ID)

		_, err := conn.DescribePortfolio(&input)
		if err == nil {
			return fmt.Errorf("Portfolio still exists")
		}
	}

	return nil
}

func testAccCheckPortfolioResourceBasicConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = "%s"
  description   = "test-2"
  provider_name = "test-3"
}
`, name)
}

func testAccCheckPortfolioResourceTags1Config(name, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = "test-b"
  provider_name = "test-c"

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccCheckPortfolioResourceTags2Config(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = "test-only-change-me"
  provider_name = "test-c"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}
