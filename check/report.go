package check

// Report is a single emission of a check and it's one time evaluation.
type Report struct {
	Check  Check
	Result Result
}
