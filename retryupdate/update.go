//go:build !solution

package retryupdate

import (
	"errors"

	"github.com/gofrs/uuid"
	"gitlab.com/slon/shad-go/retryupdate/kvapi"
)

var (
	authError     *kvapi.AuthError
	conflictError *kvapi.ConflictError
)

func UpdateValue(c kvapi.Client, key string, updateFn func(*string) (string, error)) error {
	var oldValue *string
	var oldVersion *uuid.UUID
	var onErrorVersion *uuid.UUID
	var returnError error
Loop:
	for {
		for {
			gReq := kvapi.GetRequest{Key: key}
			getRes, getErr := c.Get(&gReq)
			if getErr == nil {
				oldValue = &getRes.Value
				oldVersion = &getRes.Version
				break
			} else if errors.Is(getErr, kvapi.ErrKeyNotFound) {
				oldValue = nil
				oldVersion = new(uuid.UUID)
				break
			} else if errors.As(getErr, &authError) {
				returnError = &kvapi.APIError{Method: "get", Err: authError}
				break Loop
			} else {
				continue // retry
			}
		}

		for {
			newValue, updateErr := updateFn(oldValue)
			if updateErr != nil {
				return updateErr
			}
			newVersion := uuid.Must(uuid.NewV4())
			sReq := kvapi.SetRequest{Key: key,
				Value:      newValue,
				OldVersion: *oldVersion,
				NewVersion: newVersion,
			}
			_, setErr := c.Set(&sReq)

			if setErr == nil {
				returnError = nil
				break Loop
			} else if errors.As(setErr, &authError) {
				returnError = &kvapi.APIError{Method: "set", Err: authError}
				break Loop
			} else if errors.As(setErr, &conflictError) {
				// repeat read operation
				if conflictError.ExpectedVersion == *onErrorVersion {
					break Loop
				}
				break
			} else if errors.Is(setErr, kvapi.ErrKeyNotFound) {
				oldValue = nil
				oldVersion = new(uuid.UUID)
				continue
			} else {
				// retry
				onErrorVersion = &newVersion
				continue
			}
		}
	}

	return returnError
}
