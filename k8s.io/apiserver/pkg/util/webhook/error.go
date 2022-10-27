package webhook

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type ErrCallingWebhook struct {
	WebhookName string
	Reason      error
	Status      *apierrors.StatusError
}

func (e *ErrCallingWebhook) Error() string {
	if e.Reason != nil {
		return fmt.Sprintf("failed calling webhook %q: %v", e.WebhookName, e.Reason)
	}
	return fmt.Sprintf("failed calling webhook %q; no further details available", e.WebhookName)
}

type ErrWebhookRejection struct {
	Status *apierrors.StatusError
}

func (e *ErrWebhookRejection) Error() string {
	return e.Status.Error()
}
