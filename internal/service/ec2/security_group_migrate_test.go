package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfec2 "github.com/nij4t/terraform-provider-aws/internal/service/ec2"
)

func TestSecurityGroupMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0": {
			StateVersion: 0,
			Attributes: map[string]string{
				"name": "test",
			},
			Expected: map[string]string{
				"name":                   "test",
				"revoke_rules_on_delete": "false",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "i-abc123",
			Attributes: tc.Attributes,
		}
		is, err := tfec2.SecurityGroupMigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		for k, v := range tc.Expected {
			if is.Attributes[k] != v {
				t.Fatalf(
					"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
					tn, k, v, k, is.Attributes[k], is.Attributes)
			}
		}
	}
}

func TestSecurityGroupMigrateState_empty(t *testing.T) {
	var is *terraform.InstanceState
	var meta interface{}

	// should handle nil
	is, err := tfec2.SecurityGroupMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
	if is != nil {
		t.Fatalf("expected nil instancestate, got: %#v", is)
	}

	// should handle non-nil but empty
	is = &terraform.InstanceState{}
	_, err = tfec2.SecurityGroupMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
}
