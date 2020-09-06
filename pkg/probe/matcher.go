package probe

import (
	"regexp"

	"github.com/glebarez/padre/pkg/client"
)

type matcherByFingerprint struct {
	fingerprints []*ResponseFingerprint
}

func (m *matcherByFingerprint) IsPaddingError(resp *client.Response) (bool, error) {

	respFP, err := GetResponseFingerprint(resp)
	if err != nil {
		return false, err
	}

	for _, fp := range m.fingerprints {
		if &fp == &respFP {
			return true, nil
		}
	}

	return false, nil
}

type matcherByRegexp struct {
	re *regexp.Regexp
}

func (m *matcherByRegexp) IsPaddingError(resp *client.Response) (bool, error) {
	return m.re.Match(resp.Body), nil
}

func NewMatcherByRegexp(r string) (PaddingErrorMatcher, error) {
	re, err := regexp.Compile(r)
	if err != nil {
		return nil, err
	}

	return &matcherByRegexp{re}, nil
}
