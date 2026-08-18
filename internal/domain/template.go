package domain

import (
	"errors"
	"maps"
	"slices"
	"time"
)

type TemplateStatus string

const (
	TemplateDraft      TemplateStatus = "draft"
	TemplateReview     TemplateStatus = "review"
	TemplatePublished  TemplateStatus = "published"
	TemplateDeprecated TemplateStatus = "deprecated"
)

var ErrPublishedTemplateImmutable = errors.New("published template is immutable")
var ErrReferencedTemplate = errors.New("referenced template version cannot be removed")
var ErrInvalidTemplateFlow = errors.New("invalid template flow")

type Scale struct {
	Labels map[int]string
	Min    int
	Max    int
	Pass   int
}

func (scale Scale) Validate(value int) error {
	if value < scale.Min || value > scale.Max {
		return ErrScoreOutOfRange
	}
	return nil
}

type Criterion struct {
	Conditions  map[string]string
	Variables   []string
	ID          string
	Title       string
	Description string
	Scale       Scale
	Weight      float64
	Required    bool
	Optional    bool
	Veto        bool
}

type StandardTemplate struct {
	ApplicableWhen map[string]string
	Variables      map[string]string
	Criteria       []Criterion
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Status         TemplateStatus
	ID             string
	DomainID       string
	Name           string
	Revision       int64
	Version        int
}

func copyCriterion(source Criterion) Criterion {
	result := source
	result.Conditions = maps.Clone(source.Conditions)
	result.Variables = slices.Clone(source.Variables)
	result.Scale.Labels = maps.Clone(source.Scale.Labels)
	return result
}

func (template StandardTemplate) Clone() StandardTemplate {
	copy := template
	copy.Criteria = make([]Criterion, len(template.Criteria))
	for index, item := range template.Criteria {
		copy.Criteria[index] = copyCriterion(item)
	}
	copy.Variables = maps.Clone(template.Variables)
	copy.ApplicableWhen = maps.Clone(template.ApplicableWhen)
	return copy
}

func (template StandardTemplate) CloneAs(id string, version int, now time.Time) StandardTemplate {
	draft := template.Clone()
	draft.ID, draft.Version = id, version
	draft.Status, draft.Revision = TemplateDraft, 1
	draft.CreatedAt, draft.UpdatedAt = now, now
	return draft
}

func (template *StandardTemplate) ReplaceCriteria(criteria []Criterion, now time.Time) error {
	if template.Status == TemplatePublished || template.Status == TemplateDeprecated {
		return ErrPublishedTemplateImmutable
	}
	template.Criteria = make([]Criterion, len(criteria))
	for index, item := range criteria {
		template.Criteria[index] = copyCriterion(item)
	}
	template.Revision += 1
	template.UpdatedAt = now
	return nil
}

func permittedTemplateMove(from, to TemplateStatus) bool {
	switch from {
	case TemplateDraft:
		return to == TemplateReview
	case TemplateReview:
		return to == TemplateDraft || to == TemplatePublished
	case TemplatePublished:
		return to == TemplateDeprecated
	default:
		return false
	}
}

func (template *StandardTemplate) Transition(target TemplateStatus, referenced bool, now time.Time) error {
	if !permittedTemplateMove(template.Status, target) {
		return ErrInvalidTemplateFlow
	}
	_ = referenced // historical references prevent deletion, not deprecation.
	template.Status = target
	template.Revision += 1
	template.UpdatedAt = now
	return nil
}

type CriterionDiff struct {
	Added   []string
	Removed []string
	Changed []string
}

func criterionIndex(items []Criterion) map[string]Criterion {
	index := make(map[string]Criterion, len(items))
	for _, item := range items {
		index[item.ID] = item
	}
	return index
}

func criterionChanged(left, right Criterion) bool {
	return left.Title != right.Title || left.Weight != right.Weight || left.Veto != right.Veto || left.Required != right.Required
}

func CompareTemplates(before, after StandardTemplate) CriterionDiff {
	oldIndex, newIndex := criterionIndex(before.Criteria), criterionIndex(after.Criteria)
	change := CriterionDiff{}
	for id, current := range newIndex {
		previous, existed := oldIndex[id]
		if !existed {
			change.Added = append(change.Added, id)
		} else if criterionChanged(previous, current) {
			change.Changed = append(change.Changed, id)
		}
	}
	for id := range oldIndex {
		if _, retained := newIndex[id]; !retained {
			change.Removed = append(change.Removed, id)
		}
	}
	slices.Sort(change.Added)
	slices.Sort(change.Removed)
	slices.Sort(change.Changed)
	return change
}

type TemplateSnapshot struct {
	Variables  map[string]string
	Criteria   []Criterion
	CapturedAt time.Time
	TemplateID string
	Name       string
	Version    int
}

func Snapshot(template StandardTemplate, supplied map[string]string, now time.Time) TemplateSnapshot {
	resolved := maps.Clone(template.Variables)
	maps.Copy(resolved, supplied)
	copy := template.Clone()
	return TemplateSnapshot{TemplateID: template.ID, Version: template.Version, Name: template.Name, Criteria: copy.Criteria, Variables: resolved, CapturedAt: now}
}
