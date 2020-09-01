package probe

// creates copy of a slice
func copySlice(slice []byte) []byte {
	sliceCopy := make([]byte, len(slice))
	copy(sliceCopy, slice)
	return sliceCopy
}
