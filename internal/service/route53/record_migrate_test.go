package route53_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfroute53 "github.com/nij4t/terraform-provider-aws/internal/service/route53"
)

func TestRecordMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		ID           string
		Attributes   map[string]string
		Expected     string
		Meta         interface{}
	}{
		"v0_0": {
			StateVersion: 0,
			ID:           "some_id",
			Attributes: map[string]string{
				"name": "www",
			},
			Expected: "www",
		},
		"v0_1": {
			StateVersion: 0,
			ID:           "some_id",
			Attributes: map[string]string{
				"name": "www.example.com.",
			},
			Expected: "www.example.com",
		},
		"v0_2": {
			StateVersion: 0,
			ID:           "some_id",
			Attributes: map[string]string{
				"name": "www.example.com",
			},
			Expected: "www.example.com",
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         tc.ID,
			Attributes: tc.Attributes,
		}
		is, err := tfroute53.RecordMigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		if is.Attributes["name"] != tc.Expected {
			t.Fatalf("bad Route 53 Migrate: %s\n\n expected: %s", is.Attributes["name"], tc.Expected)
		}
	}
}

func TestRecordMigrateStateV1toV2(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0_1": {
			StateVersion: 1,
			Attributes: map[string]string{
				"weight":   "0",
				"failover": "PRIMARY",
			},
			Expected: map[string]string{
				"weighted_routing_policy.#":        "1",
				"weighted_routing_policy.0.weight": "0",
				"failover_routing_policy.#":        "1",
				"failover_routing_policy.0.type":   "PRIMARY",
			},
		},
		"v0_2": {
			StateVersion: 0,
			Attributes: map[string]string{
				"weight": "-1",
			},
			Expected: map[string]string{},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "route53_record",
			Attributes: tc.Attributes,
		}
		is, err := tfroute53.ResourceRecord().MigrateState(
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
