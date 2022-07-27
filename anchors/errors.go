package anchors

import "github.com/centrifuge/go-centrifuge/errors"

const (
	ErrAnchorRetrieval   = errors.Error("couldn't retrieve anchor")
	ErrEmptyDocumentRoot = errors.Error("document root is empty")
)
