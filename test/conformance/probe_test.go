// +build e2e

/*
Copyright 2019 The Knative Authors

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

package conformance

import (
	"fmt"
	"testing"

	pkgTest "github.com/knative/pkg/test"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeploymentProbe(t *testing.T) {
	t.Parallel()
	clients := setup(t)

	names := test.ResourceNames{
		Config: test.ObjectNameForTest(t),
		Image:  "probe",
	}

	t.Logf("Creating a new Configuration %s", names.Image)

	probe := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{},
		},
	}
	if _, err := test.CreateConfiguration(t, clients, names, &test.Options{LivenessProbe: probe}); err != nil {
		t.Fatalf("Failed to create configuration %s: %v", names.Config, err)
	}
	defer test.TearDown(clients, names)
	test.CleanupOnInterrupt(func() { test.TearDown(clients, names) })

	err := test.WaitForConfigurationState(clients.ServingClient, names.Config, func(r *v1alpha1.Configuration) (bool, error) {
		cond := r.Status.GetCondition(v1alpha1.ConfigurationConditionReady)
		t.Logf("cond: %v", cond)
		if cond != nil && !cond.IsUnknown() {
			t.Logf("Reason: %s ; Message: %s ; Status: %s", cond.Reason, cond.Message, cond.Status)
			if cond.IsTrue() {
				return true, nil
			}
			return false, fmt.Errorf("The configuration %s was not marked as Ready (Reason=\"%s\", Message=\"%s\", Status=\"%s\"), but with (Reason=\"%s\", Message=\"%s\", Status=\"%s\")",
				names.Config, "", "", "True", cond.Reason, cond.Message, cond.Status)
		}

		return false, nil
	}, "ConfigContainersReady")

	if err != nil {
		t.Fatalf("Failed to validate configuration state: %s", err)
	}

	revisionName, err := getRevisionFromConfiguration(clients, names.Config)
	if err != nil {
		t.Fatalf("Failed to get revision from configuration %s: %v", names.Config, err)
	}
	t.Logf("revisionName: %v", revisionName)

	// TODO send a request to crash the container
	route, err := clients.ServingClient.Routes.Get(names.Route, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Error fetching Route %s: %v", names.Route, err)
	}
	domain := route.Status.Domain
	client, err := pkgTest.NewSpoofingClient(clients.KubeClient, t.Logf, domain, test.ServingFlags.ResolvableDomain)
	if err != nil {
		t.Fatal("Failed to create the spoofing client. Error: %s", err)
	}
	actualResponses, err := sendRequests(client, domain, 1)
	if err != nil {
		t.Fatalf("Failed to send the request. Error: %s", err)
	}
	t.Logf("actual Response: %v", actualResponses)

	t.Log("When the containers crash, the revision should have error status.")
	err = test.WaitForRevisionState(clients.ServingClient, revisionName, func(r *v1alpha1.Revision) (bool, error) {
		cond := r.Status.GetCondition(v1alpha1.RevisionConditionReady)
	})

}
