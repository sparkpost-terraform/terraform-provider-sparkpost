package provider

// createOrAdopt runs create and, if it fails, reports adopted as true when the
// resource exists anyway: SparkPost can register a domain while answering 400
// "Restricted domain". A 409 means it existed beforehand, so that needs an import.
func createOrAdopt(create func() error, exists func() (bool, error)) (adopted bool, err error) {
	err = create()
	if err == nil || isConflict(err) {
		return false, err
	}

	found, existsErr := exists()
	if existsErr != nil || !found {
		return false, err
	}

	return true, err
}
