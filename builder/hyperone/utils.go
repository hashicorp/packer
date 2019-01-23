package hyperone

import (
	"fmt"

	"github.com/hyperonecom/h1-client-go"
)

func formatOpenAPIError(err error) string {
	openAPIError, ok := err.(openapi.GenericOpenAPIError)
	if !ok {
		return err.Error()
	}

	return fmt.Sprintf("%s (body: %s)", openAPIError.Error(), openAPIError.Body())
}
