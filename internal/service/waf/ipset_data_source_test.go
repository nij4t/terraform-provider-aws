package waf_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/waf"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

func TestAccWAFIPSetDataSource_basic(t *testing.T) {
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.ipset"
	datasourceName := "data.aws_waf_ipset.ipset"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(waf.EndpointsID, t) },
		ErrorCheck: acctest.ErrorCheck(t, waf.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccIPSetDataSource_NonExistent,
				ExpectError: regexp.MustCompile(`WAF IP Set not found`),
			},
			{
				Config: testAccIPSetDataSource_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccIPSetDataSource_Name(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = %[1]q
}

data "aws_waf_ipset" "ipset" {
  name = aws_waf_ipset.ipset.name
}
`, name)
}

const testAccIPSetDataSource_NonExistent = `
data "aws_waf_ipset" "ipset" {
  name = "tf-acc-test-does-not-exist"
}
`
