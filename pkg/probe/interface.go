package probe

import "github.com/glebarez/padre/pkg/client"

// PaddingErrorMatcher - tests if HTTP response matches with padding error
type PaddingErrorMatcher interface {
	IsPaddingError(*client.Response) (bool, error)
}
