package sagemaker_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(sagemaker.EndpointsID, testAccErrorCheckSkipSagemaker)
}

func testAccErrorCheckSkipSagemaker(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"is not supported in region",
		"is not supported for the chosen region",
	)
}

// Tests are serialized as SagmMaker Domain resources are limited to 1 per account by default.
// SageMaker UserProfile and App depend on the Domain resources and as such are also part of the serialized test suite.
// Sagemaker Workteam tests must also be serialized
func TestAccSageMaker_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"App": {
			"basic":        testAccApp_basic,
			"disappears":   testAccApp_tags,
			"tags":         testAccApp_disappears,
			"resourceSpec": testAccApp_resourceSpec,
		},
		"Domain": {
			"basic":                                    testAccDomain_basic,
			"disappears":                               testAccDomain_tags,
			"tags":                                     testAccDomain_disappears,
			"tensorboardAppSettings":                   testAccDomain_tensorboardAppSettings,
			"tensorboardAppSettingsWithImage":          testAccDomain_tensorboardAppSettingsWithImage,
			"kernelGatewayAppSettings":                 testAccDomain_kernelGatewayAppSettings,
			"kernelGatewayAppSettings_customImage":     testAccDomain_kernelGatewayAppSettings_customImage,
			"kernelGatewayAppSettings_lifecycleConfig": testAccDomain_kernelGatewayAppSettings_lifecycleConfig,
			"jupyterServerAppSettings":                 testAccDomain_jupyterServerAppSettings,
			"kms":                                      testAccDomain_kms,
			"securityGroup":                            testAccDomain_securityGroup,
			"sharingSettings":                          testAccDomain_sharingSettings,
		},
		"FlowDefinition": {
			"basic":                          testAccFlowDefinition_basic,
			"disappears":                     testAccFlowDefinition_disappears,
			"HumanLoopConfigPublicWorkforce": testAccFlowDefinition_humanLoopConfig_publicWorkforce,
			"HumanLoopRequestSource":         testAccFlowDefinition_humanLoopRequestSource,
			"Tags":                           testAccFlowDefinition_tags,
		},
		"UserProfile": {
			"basic":                           testAccUserProfile_basic,
			"disappears":                      testAccUserProfile_tags,
			"tags":                            testAccUserProfile_disappears,
			"tensorboardAppSettings":          testAccUserProfile_tensorboardAppSettings,
			"tensorboardAppSettingsWithImage": testAccUserProfile_tensorboardAppSettingsWithImage,
			"kernelGatewayAppSettings":        testAccUserProfile_kernelGatewayAppSettings,
			"kernelGatewayAppSettings_lifecycleConfig": testAccUserProfile_kernelGatewayAppSettings_lifecycleconfig,
			"jupyterServerAppSettings":                 testAccUserProfile_jupyterServerAppSettings,
		},
		"Workforce": {
			"disappears":     testAccWorkforce_disappears,
			"CognitoConfig":  testAccWorkforce_cognitoConfig,
			"OidcConfig":     testAccWorkforce_oidcConfig,
			"SourceIpConfig": testAccWorkforce_sourceIPConfig,
		},
		"Workteam": {
			"disappears":         testAccWorkteam_disappears,
			"CognitoConfig":      testAccWorkteam_cognitoConfig,
			"NotificationConfig": testAccWorkteam_notificationConfig,
			"OidcConfig":         testAccWorkteam_oidcConfig,
			"Tags":               testAccWorkteam_tags,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
