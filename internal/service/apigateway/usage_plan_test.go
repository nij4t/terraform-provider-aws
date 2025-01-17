package apigateway_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/nij4t/terraform-provider-aws/internal/service/apigateway"
)

func TestAccAPIGatewayUsagePlan_basic(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedName := sdkacctest.RandomWithPrefix("tf-acc-test-2")
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/usageplans/+.`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "api_stages.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUsagePlanBasicConfig(updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayUsagePlan_tags(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanBasicTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
				Config: testAccUsagePlanBasicTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccUsagePlanBasicTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayUsagePlan_description(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUsagePlanDescriptionConfig(rName, "This is a description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a description"),
				),
			},
			{
				Config: testAccUsagePlanDescriptionConfig(rName, "This is a new description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "This is a new description"),
				),
			},
			{
				Config: testAccUsagePlanDescriptionConfig(rName, "This is a description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a description"),
				),
			},
			{
				Config: testAccUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayUsagePlan_productCode(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanProductCodeConfig(rName, "MYCODE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "product_code", "MYCODE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUsagePlanProductCodeConfig(rName, "MYCODE2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "product_code", "MYCODE2"),
				),
			},
			{
				Config: testAccUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "product_code", ""),
				),
			},
			{
				Config: testAccUsagePlanProductCodeConfig(rName, "MYCODE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "product_code", "MYCODE"),
				),
			},
		},
	})
}

func TestAccAPIGatewayUsagePlan_throttling(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "throttle_settings"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUsagePlanThrottlingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.burst_limit", "2"),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.rate_limit", "5"),
				),
			},
			{
				Config: testAccUsagePlanThrottlingModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.burst_limit", "3"),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.rate_limit", "6"),
				),
			},
			{
				Config: testAccUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "throttle_settings"),
				),
			},
		},
	})
}

// https://github.com/nij4t/terraform-provider-aws/issues/2057
func TestAccAPIGatewayUsagePlan_throttlingInitialRateLimit(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanThrottlingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.rate_limit", "5"),
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

func TestAccAPIGatewayUsagePlan_quota(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "quota_settings"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUsagePlanQuotaConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.limit", "100"),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.offset", "6"),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.period", "WEEK"),
				),
			},
			{
				Config: testAccUsagePlanQuotaModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.limit", "200"),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.offset", "20"),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.period", "MONTH"),
				),
			},
			{
				Config: testAccUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "quota_settings"),
				),
			},
		},
	})
}

func TestAccAPIGatewayUsagePlan_apiStages(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUsagePlanDestroy,
		Steps: []resource.TestStep{
			// Create UsagePlan WITH Stages as the API calls are different
			// when creating or updating.
			{
				Config: testAccUsagePlanAPIStagesConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage": "test",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Handle api stages removal
			{
				Config: testAccUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "api_stages"),
				),
			},
			// Handle api stages additions
			{
				Config: testAccUsagePlanAPIStagesConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage": "test",
					}),
				),
			},
			// Handle api stages updates
			{
				Config: testAccUsagePlanAPIStagesMultipleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage": "foo",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage": "test",
					}),
				),
			},
			{
				Config: testAccUsagePlanAPIStagesModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage": "foo",
					}),
				),
			},
			{
				Config: testAccUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "api_stages"),
				),
			},
		},
	})
}

func TestAccAPIGatewayUsagePlan_APIStages_multiple(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanAPIStagesMultipleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage": "foo",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage": "test",
					}),
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

func TestAccAPIGatewayUsagePlan_APIStages_throttle(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanAPIStagesConfigThrottle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "api_stages.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage":      "test",
						"throttle.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.0.throttle.*", map[string]string{
						"path":        "/test/GET",
						"burst_limit": "3",
						"rate_limit":  "6",
					}),
				),
			},
			{
				Config: testAccUsagePlanAPIStagesConfigThrottleMulti(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "api_stages.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage":      "foo",
						"throttle.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage":      "test",
						"throttle.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.0.throttle.*", map[string]string{
						"path":        "/test/GET",
						"burst_limit": "3",
						"rate_limit":  "6",
					}),
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

func TestAccAPIGatewayUsagePlan_disappears(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceUsagePlan(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceUsagePlan(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUsagePlanExists(n string, res *apigateway.UsagePlan) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Usage Plan ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

		req := &apigateway.GetUsagePlanInput{
			UsagePlanId: aws.String(rs.Primary.ID),
		}
		up, err := conn.GetUsagePlan(req)
		if err != nil {
			return err
		}

		if aws.StringValue(up.Id) != rs.Primary.ID {
			return fmt.Errorf("APIGateway Usage Plan not found")
		}

		*res = *up

		return nil
	}
}

func testAccCheckUsagePlanDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_usage_plan" {
			continue
		}

		req := &apigateway.GetUsagePlanInput{
			UsagePlanId: aws.String(rs.Primary.ID),
		}
		describe, err := conn.GetUsagePlan(req)

		if err == nil {
			if describe.Id != nil && aws.StringValue(describe.Id) == rs.Primary.ID {
				return fmt.Errorf("API Gateway Usage Plan still exists")
			}
		}

		if !tfawserr.ErrMessageContains(err, apigateway.ErrCodeNotFoundException, "") {
			return err
		}

		return nil
	}

	return nil
}

func testAccUsagePlanConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "test"
  description = "This is a test"

  variables = {
    "a" = "2"
  }
}

resource "aws_api_gateway_deployment" "foo" {
  depends_on = [
    aws_api_gateway_deployment.test,
    aws_api_gateway_integration.test,
  ]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "foo"
  description = "This is a prod stage"
}
`, rName)
}

func testAccUsagePlanBasicConfig(rName string) string {
	return testAccUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"
}
`, rName)
}

func testAccUsagePlanBasicTags1(rName, tagKey1, tagValue1 string) string {
	return testAccUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  tags = {
    %q = %q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccUsagePlanBasicTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccUsagePlanDescriptionConfig(rName, desc string) string {
	return testAccUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, desc)
}

func testAccUsagePlanProductCodeConfig(rName, code string) string {
	return testAccUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name         = %[1]q
  product_code = %[2]q
}
`, rName, code)
}

func testAccUsagePlanThrottlingConfig(rName string) string {
	return testAccUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  throttle_settings {
    burst_limit = 2
    rate_limit  = 5
  }
}
`, rName)
}

func testAccUsagePlanThrottlingModifiedConfig(rName string) string {
	return testAccUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  throttle_settings {
    burst_limit = 3
    rate_limit  = 6
  }
}
`, rName)
}

func testAccUsagePlanQuotaConfig(rName string) string {
	return testAccUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  quota_settings {
    limit  = 100
    offset = 6
    period = "WEEK"
  }
}
`, rName)
}

func testAccUsagePlanQuotaModifiedConfig(rName string) string {
	return testAccUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  quota_settings {
    limit  = 200
    offset = 20
    period = "MONTH"
  }
}
`, rName)
}

func testAccUsagePlanAPIStagesConfig(rName string) string {
	return testAccUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.test.stage_name
  }
}
`, rName)
}

func testAccUsagePlanAPIStagesConfigThrottle(rName string) string {
	return testAccUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  throttle_settings {
    burst_limit = 3
    rate_limit  = 6
  }

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.test.stage_name

    throttle {
      path        = "${aws_api_gateway_resource.test.path}/${aws_api_gateway_method.test.http_method}"
      burst_limit = 3
      rate_limit  = 6
    }
  }
}
`, rName)
}

func testAccUsagePlanAPIStagesConfigThrottleMulti(rName string) string {
	return testAccUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  throttle_settings {
    burst_limit = 3
    rate_limit  = 6
  }

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.test.stage_name

    throttle {
      path        = "${aws_api_gateway_resource.test.path}/${aws_api_gateway_method.test.http_method}"
      burst_limit = 3
      rate_limit  = 6
    }
  }

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.foo.stage_name

    throttle {
      path        = "${aws_api_gateway_resource.test.path}/${aws_api_gateway_method.test.http_method}"
      burst_limit = 3
      rate_limit  = 6
    }
  }
}
`, rName)
}

func testAccUsagePlanAPIStagesModifiedConfig(rName string) string {
	return testAccUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.foo.stage_name
  }
}
`, rName)
}

func testAccUsagePlanAPIStagesMultipleConfig(rName string) string {
	return testAccUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.foo.stage_name
  }

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.test.stage_name
  }
}
`, rName)
}
