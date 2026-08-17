export interface MaterialResult { materialID: string; weightedScore: number; vetoed: boolean; passed: boolean; reviewerCount: number; disagreements: string[] }
export interface ReviewerPosition { reviewerID: string; score: number; comment: string; evidence: string[] }
export interface DivergenceCase { criterionID: string; spread: number; positions: ReviewerPosition[] }
export interface FinalDecision { materialID: string; passed: boolean; conclusion: string; rationale: string; resolvedBy: string; resolvedAt: string }
export interface DeliberationCase { batchID: string; materialID: string; templateID: string; templateVersion: number; assignedReviewers: number; submittedReviewers: number; quorumMet: boolean; preliminary: MaterialResult; divergences: DivergenceCase[]; decision?: FinalDecision }
export interface ReviewStatistics { averageCycleHours: number; returnReasons: Array<{ reason: string; count: number }>; criteria: Array<{ criterionID: string; useCount: number; vetoCount: number; disagreementCount: number }> }
