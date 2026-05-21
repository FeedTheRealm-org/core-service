package errors

// ExportNotFound is returned when an export zip entry cannot be found.
type ExportNotFound struct {
	details string
}

func (e *ExportNotFound) Error() string {
	return e.details
}

func NewExportNotFound(details string) *ExportNotFound {
	return &ExportNotFound{details: details}
}

// ExportVersionConflict is returned when an export version already exists.
type ExportVersionConflict struct {
	details string
}

func (e *ExportVersionConflict) Error() string {
	return e.details
}

func NewExportVersionConflict(details string) *ExportVersionConflict {
	return &ExportVersionConflict{details: details}
}
