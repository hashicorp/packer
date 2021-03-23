package publicapi

import (
	"context"
	"net/http"
)

// MultiRequestsEditor is an oapi-codegen compatible RequestEditorFn function that executes multiple
// RequestEditorFn functions sequentially.
func MultiRequestsEditor(fns ...RequestEditorFn) RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		for _, fn := range fns {
			if err := fn(ctx, req); err != nil {
				return err
			}
		}

		return nil
	}
}
