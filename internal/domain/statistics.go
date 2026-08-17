package domain

import "time"

type ReturnReasonCount struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}
type CriterionEffectiveness struct {
	CriterionID       string `json:"criterionID"`
	UseCount          int    `json:"useCount"`
	VetoCount         int    `json:"vetoCount"`
	DisagreementCount int    `json:"disagreementCount"`
}
type ReviewStatistics struct {
	AverageCycleHours float64                  `json:"averageCycleHours"`
	ReturnReasons     []ReturnReasonCount      `json:"returnReasons"`
	Criteria          []CriterionEffectiveness `json:"criteria"`
	From              time.Time                `json:"from"`
	To                time.Time                `json:"to"`
}
