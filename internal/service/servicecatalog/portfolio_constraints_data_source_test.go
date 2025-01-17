package servicecatalog_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

func TestAccServiceCatalogPortfolioConstraintsDataSource_Constraint_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_constraint.test"
	dataSourceName := "data.aws_servicecatalog_portfolio_constraints.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccPortfolioConstraintDataSourceConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "accept_language", resourceName, "accept_language"),
					resource.TestCheckResourceAttr(dataSourceName, "details.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "details.0.constraint_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "details.0.description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "details.0.owner", resourceName, "owner"),
					resource.TestCheckResourceAttrPair(dataSourceName, "details.0.portfolio_id", resourceName, "portfolio_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "details.0.product_id", resourceName, "product_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "details.0.type", resourceName, "type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "portfolio_id", resourceName, "portfolio_id"),
				),
			},
		},
	})
}

func testAccPortfolioConstraintDataSourceConfig_basic(rName, description string) string {
	return acctest.ConfigCompose(testAccConstraintConfig_basic(rName, description), `
data "aws_servicecatalog_portfolio_constraints" "test" {
  portfolio_id = aws_servicecatalog_constraint.test.portfolio_id
}
`)
}
