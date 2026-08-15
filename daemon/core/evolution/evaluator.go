package evolution

type FeedbackType string

const (
	FeedbackAccepted FeedbackType = "ACCEPTED"
	FeedbackRejected FeedbackType = "REJECTED"
	FeedbackModified FeedbackType = "MODIFIED"
)

type ScopeLevel string

const (
	ScopePersonal     ScopeLevel = "PERSONAL"
	ScopeProject      ScopeLevel = "PROJECT"
	ScopeOrganization ScopeLevel = "ORGANIZATION"
	ScopeGeneric      ScopeLevel = "GENERIC"
)

func CalculateScopeWeight(scope ScopeLevel) float64 {
	switch scope {
	case ScopePersonal:
		return 1.5
	case ScopeProject:
		return 1.3
	case ScopeOrganization:
		return 1.1
	default:
		return 1.0
	}
}
