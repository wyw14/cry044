package domain

import "testing"

func TestTemplateComparisonDetectsDecisionSemanticChanges(t *testing.T) {
	baseline := StandardTemplate{ID: "quality", Version: 1, Criteria: []Criterion{{ID: "safety", Title: "Safety", Required: true, Veto: true, Weight: 1, Scale: Scale{Min: 0, Max: 5, Pass: 3, Labels: map[int]string{3: "pass"}}, Conditions: map[string]string{"area": "assembly"}, Variables: []string{"shift"}}}}
	changed := baseline.Clone()
	changed.Criteria[0].Scale.Pass = 4
	changed.Criteria[0].Scale.Labels[3] = "review"
	changed.Criteria[0].Veto = false
	changed.Criteria[0].Conditions["area"] = "welding"
	changed.Criteria[0].Variables = []string{"line"}
	diff := CompareTemplates(baseline, changed)
	if len(diff.Changed) != 1 || diff.Changed[0] != "safety" {
		t.Fatalf("decision-semantic change disappeared from diff: %#v", diff)
	}
}
