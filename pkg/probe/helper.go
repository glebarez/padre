// check if response contains padding error
func isPaddingError(resp *http.Response, body []byte) (bool, error) {
	// try regex matcher if pattern is set
	if *config.paddingErrorPattern != "" {
		matched, err := regexp.Match(*config.paddingErrorPattern, body)
		if err != nil {
			return false, err
		}
		return matched, nil
	}

	// otherwise fallback to fingerprint
	if config.paddingErrorFingerprint != nil {
		fp, err := getResponseFingerprint(resp, body)
		if err != nil {
			return false, err
		}
		return *fp == *config.paddingErrorFingerprint, nil
	}

	return false, fmt.Errorf("Neither fingerprint nor string pattern for padding error is set")
}