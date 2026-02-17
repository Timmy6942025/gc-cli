package preview

import "strings"

const (
	FeatureRubrics        = "rubrics"
	FeatureGradingPeriods = "grading_periods"
	FeatureStudentGroups  = "student_groups"
)

type Capability struct {
	Feature   string `json:"feature"`
	Available bool   `json:"available"`
	Preview   bool   `json:"preview"`
	Reason    string `json:"reason,omitempty"`
}

type CapabilityDetector interface {
	Capability(feature string) Capability
	List() []Capability
}

type Detector struct {
	PreviewEnabled bool
}

func (d Detector) Capability(feature string) Capability {
	f := strings.ToLower(strings.TrimSpace(feature))
	switch f {
	case FeatureRubrics, FeatureGradingPeriods, FeatureStudentGroups:
		if d.PreviewEnabled {
			return Capability{Feature: f, Available: true, Preview: true, Reason: "enabled by preview flag"}
		}
		return Capability{Feature: f, Available: false, Preview: true, Reason: "requires preview_enabled=true"}
	default:
		return Capability{Feature: f, Available: true, Preview: false}
	}
}

func (d Detector) List() []Capability {
	features := []string{FeatureRubrics, FeatureGradingPeriods, FeatureStudentGroups}
	out := make([]Capability, 0, len(features))
	for _, f := range features {
		out = append(out, d.Capability(f))
	}
	return out
}
