package domain

import (
	"testing"
	"time"
)

func TestSnapshotDoesNotAliasTemplateInputs(t *testing.T) {
	now := time.Now()
	template := StandardTemplate{ID: "quality", Version: 1, Variables: map[string]string{"region": "cn"}, ApplicableWhen: map[string]string{"tier": "gold"}, Criteria: []Criterion{{ID: "quality", Conditions: map[string]string{"mode": "strict"}, Variables: []string{"evidence"}}}}
	snapshot := Snapshot(template, map[string]string{"region": "us"}, now)
	template.Variables["region"] = "eu"
	template.Criteria[0].Conditions["mode"] = "loose"
	if snapshot.Variables["region"] != "us" || snapshot.Criteria[0].Conditions["mode"] != "strict" {
		t.Fatalf("snapshot changed with template input: %#v", snapshot)
	}
}
