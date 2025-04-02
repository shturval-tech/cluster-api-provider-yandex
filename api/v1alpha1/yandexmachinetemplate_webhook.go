/*
Copyright 2024.

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

package v1alpha1

import (
	"reflect"
	"regexp"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var ymtlog = logf.Log.WithName("yandexmachinetemplate-resource")

// regex for validating machine template webhook
//   - It must start with a lowercase letter (a-z).
//   - It can contain 0 to 55 (excluding the first and last character) additional characters, which may include lowercase letters (a-z), digits (0-9), and hyphens (-).
//   - If additional characters are present, the string must end with a letter or a digit (it cannot end with a hyphen).
//   - The 57-character limit exists because a YandexMachine name is generated from the YandexMachineTemplate name
//     with a 6-character postfix (e.g., "-12345"), and the total allowed length is 63 characters.
var nameRegex = regexp.MustCompile("^[a-z]([-a-z0-9]{0,55}[a-z0-9])?$")

// SetupWebhookWithManager creates an YandexMachineTemplate validation webhook.
func (t *YandexMachineTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(t).
		Complete()
}

//+kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1alpha1-yandexmachinetemplate,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=yandexmachinetemplates,verbs=create;update,versions=v1alpha1,name=validation.yandexmachinetemplates.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1beta1

var (
	_ webhook.Defaulter = &YandexMachineTemplate{}
	_ webhook.Validator = &YandexMachineTemplate{}
)

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (t *YandexMachineTemplate) Default() {
	ymtlog.Info("default", "name", t.Name)
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (t *YandexMachineTemplate) ValidateCreate() (admission.Warnings, error) {
	ymtlog.Info("validate create", "name", t.Name)
	var allErrs field.ErrorList

	if t.Spec.Template.Spec.ProviderID != nil {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec", "template", "spec", "providerID"), "cannot be set in templates"))
	}

	if !nameRegex.MatchString(t.Name) {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("metadata", "name"),
			t.Name,
			"may contain lowercase Latin letters, digits, and hyphens. The first character must be a letter, and the hyphen cannot be the last character, max 57 symbols"),
		)
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(GroupVersion.WithKind("YandexMachineTemplate").GroupKind(), t.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (t *YandexMachineTemplate) ValidateUpdate(oldRaw runtime.Object) (admission.Warnings, error) {
	ymtlog.Info("validate update", "name", t.Name)

	newYandexMachineTemplate, err := runtime.DefaultUnstructuredConverter.ToUnstructured(t)
	if err != nil {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind("YandexMachineTemplate").GroupKind(), t.Name, field.ErrorList{
			field.InternalError(nil, errors.Wrap(err, "failed to convert new YandexMachineTemplate to unstructured object")),
		})
	}
	oldYandexMachineTemplate, err := runtime.DefaultUnstructuredConverter.ToUnstructured(oldRaw)
	if err != nil {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind("YandexMachineTemplate").GroupKind(), t.Name, field.ErrorList{
			field.InternalError(nil, errors.Wrap(err, "failed to convert old YandexMachineTemplate to unstructured object")),
		})
	}

	newYandexMachineTemplateSpec := newYandexMachineTemplate["spec"].(map[string]interface{})
	oldYandexMachineTemplateSpec := oldYandexMachineTemplate["spec"].(map[string]interface{})

	if !reflect.DeepEqual(oldYandexMachineTemplateSpec, newYandexMachineTemplateSpec) {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind("YandexMachineTemplate").GroupKind(), t.Name, field.ErrorList{
			field.Forbidden(field.NewPath("spec"), "cannot be modified"),
		})
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (t *YandexMachineTemplate) ValidateDelete() (admission.Warnings, error) {
	ymtlog.Info("validate delete", "name", t.Name)

	return nil, nil
}
