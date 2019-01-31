/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhook

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// admissionWebhook contains bits needed for generating a admissionWebhook Configuration
type admissionWebhook struct {
	// name is the name of the webhook
	name string
	// t is the webhook type, i.e. mutating, validating
	t webhookType
	// path is the path this webhook will serve.
	path string
	// rules maps to the rules field in admissionregistrationv1beta1.admissionWebhook
	rules []admissionregistrationv1beta1.RuleWithOperations
	// failurePolicy maps to the failurePolicy field in admissionregistrationv1beta1.admissionWebhook
	// This optional. If not set, will be defaulted to Ignore (fail-open) by the server.
	// More details: https://github.com/kubernetes/api/blob/f5c295feaba2cbc946f0bbb8b535fc5f6a0345ee/admissionregistration/v1beta1/types.go#L144-L147
	failurePolicy *admissionregistrationv1beta1.FailurePolicyType
	// namespaceSelector maps to the namespaceSelector field in admissionregistrationv1beta1.admissionWebhook
	// This optional.
	namespaceSelector *metav1.LabelSelector

	once sync.Once
}

func (w *admissionWebhook) setDefaults() {
	if len(w.path) == 0 {
		if len(w.rules) == 0 || len(w.rules[0].Resources) == 0 {
			// can't do defaulting, skip it.
			return
		}
		if w.t == webhookTypeMutating {
			w.path = "/mutate-" + w.rules[0].Resources[0]
		} else if w.t == webhookTypeValidating {
			w.path = "/validate-" + w.rules[0].Resources[0]
		}
	}
	if len(w.name) == 0 {
		reg := regexp.MustCompile("[^a-zA-Z0-9]+")
		processedPath := strings.ToLower(reg.ReplaceAllString(w.path, ""))
		w.name = processedPath + ".example.com"
	}
}

// GetName returns the name of the webhook.
func (w *admissionWebhook) GetName() string {
	w.once.Do(w.setDefaults)
	return w.name
}

// GetPath returns the path that the webhook registered.
func (w *admissionWebhook) GetPath() string {
	w.once.Do(w.setDefaults)
	return w.path
}

// GetType returns the type of the webhook.
func (w *admissionWebhook) GetType() webhookType {
	w.once.Do(w.setDefaults)
	return w.t
}

// Validate validates if the webhook is valid.
func (w *admissionWebhook) Validate() error {
	w.once.Do(w.setDefaults)
	if len(w.rules) == 0 {
		return errors.New("field rules should not be empty")
	}
	if len(w.name) == 0 {
		return errors.New("field name should not be empty")
	}
	if w.t != webhookTypeMutating && w.t != webhookTypeValidating {
		return fmt.Errorf("unsupported Type: %v, only webhookTypeMutating and webhookTypeValidating are supported", w.t)
	}
	if len(w.path) == 0 {
		return errors.New("field path should not be empty")
	}
	return nil
}
