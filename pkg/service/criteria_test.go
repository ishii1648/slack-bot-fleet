package service

import (
	"testing"
)

func TestParseReactionEventCriteriaYml(t *testing.T) {
	trigger, err := ParseReactionEventCriteriaYml("../test/criteria_test.yml")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", trigger)
}
