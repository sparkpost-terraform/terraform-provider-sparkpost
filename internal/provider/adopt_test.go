package provider

import (
	"errors"
	"net/http"
	"testing"
)

func TestCreateOrAdopt_Success(t *testing.T) {
	checked := false
	adopted, err := createOrAdopt(
		func() error { return nil },
		func() (bool, error) { checked = true; return true, nil },
	)
	if err != nil || adopted || checked {
		t.Errorf("createOrAdopt() = (%v, %v), exists checked = %v; want (false, nil), not checked", adopted, err, checked)
	}
}

func TestCreateOrAdopt_AdoptsWhenRegisteredDespiteError(t *testing.T) {
	createErr := &apiError{StatusCode: http.StatusBadRequest, Status: "400 Bad Request", Message: "Restricted domain"}
	adopted, err := createOrAdopt(
		func() error { return createErr },
		func() (bool, error) { return true, nil },
	)
	if !adopted || !errors.Is(err, createErr) {
		t.Errorf("createOrAdopt() = (%v, %v); want (true, the create error)", adopted, err)
	}
}

func TestCreateOrAdopt_FailsWhenNotRegistered(t *testing.T) {
	createErr := &apiError{StatusCode: http.StatusBadRequest, Status: "400 Bad Request"}
	adopted, err := createOrAdopt(
		func() error { return createErr },
		func() (bool, error) { return false, nil },
	)
	if adopted || !errors.Is(err, createErr) {
		t.Errorf("createOrAdopt() = (%v, %v); want (false, the create error)", adopted, err)
	}
}

func TestCreateOrAdopt_FailsWhenExistenceCheckFails(t *testing.T) {
	createErr := errors.New("connection reset")
	adopted, err := createOrAdopt(
		func() error { return createErr },
		func() (bool, error) { return false, errors.New("timeout") },
	)
	if adopted || !errors.Is(err, createErr) {
		t.Errorf("createOrAdopt() = (%v, %v); want (false, the create error)", adopted, err)
	}
}

func TestCreateOrAdopt_NeverAdoptsOnConflict(t *testing.T) {
	createErr := &apiError{StatusCode: http.StatusConflict, Status: "409 Conflict"}
	checked := false
	adopted, err := createOrAdopt(
		func() error { return createErr },
		func() (bool, error) { checked = true; return true, nil },
	)
	if adopted || checked || !errors.Is(err, createErr) {
		t.Errorf("createOrAdopt() = (%v, %v), exists checked = %v; want (false, the create error), not checked", adopted, err, checked)
	}
}
