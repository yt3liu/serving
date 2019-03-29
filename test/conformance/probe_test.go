// +build e2e

/*
Copyright 2018 The Knative Authors

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
	"strings"
	"testing"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/test"
	corev1 "k8s.io/api/core/v1"
)

func TestContainerExitingMsg3(t *testing.T) {
	t.Parallel()

	const (
		// The given image will always exit with an exit code of 5
		exitCodeReason = "ExitCode5"
		// ... and will print "Crashed..." before it exits
		errorLog = "Crashed..."
	)

	tests := []struct {
		Name           string
		ReadinessProbe *corev1.Probe
	}{
		{
			// container-exiting-msg3-http1-dbyjwewn
			Name: "http1",
			ReadinessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{},
				},
			},
		},
		{
			// container-exiting-msg3-http2-vlpyuscw
			Name: "http2",
			ReadinessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{},
				},
				InitialDelaySeconds: 0,
				TimeoutSeconds:      1,
				PeriodSeconds:       5,
				SuccessThreshold:    1,
				FailureThreshold:    3,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			clients := setup(t)

			names := test.ResourceNames{
				Config: test.ObjectNameForTest(t),
				Image:  "failing",
			}

			t.Logf("Creating a new Configuration %s", names.Image)

			if _, err := test.CreateConfiguration(t, clients, names, &test.Options{ReadinessProbe: tt.ReadinessProbe}); err != nil {
				t.Fatalf("Failed to create configuration %s: %v", names.Config, err)
			}
			//defer test.TearDown(clients, names)
			//test.CleanupOnInterrupt(func() { test.TearDown(clients, names) })

			t.Log("When the containers keep crashing, the Configuration should have error status.")

			err := test.WaitForConfigurationState(clients.ServingClient, names.Config, func(r *v1alpha1.Configuration) (bool, error) {
				cond := r.Status.GetCondition(v1alpha1.ConfigurationConditionReady)
				if cond != nil && !cond.IsUnknown() {
					if strings.Contains(cond.Message, errorLog) && cond.IsFalse() {
						return true, nil
					}
					t.Logf("Reason: %s ; Message: %s ; Status: %s", cond.Reason, cond.Message, cond.Status)
					return true, fmt.Errorf("The configuration %s was not marked with expected error condition (Reason=\"%s\", Message=\"%s\", Status=\"%s\"), but with (Reason=\"%s\", Message=\"%s\", Status=\"%s\")",
						names.Config, containerMissing, errorLog, "False", cond.Reason, cond.Message, cond.Status)
				}
				return false, nil
			}, "ConfigContainersCrashing")

			if err != nil {
				t.Fatalf("Failed to validate configuration state: %s", err)
			}

			revisionName, err := getRevisionFromConfiguration(clients, names.Config)
			if err != nil {
				t.Fatalf("Failed to get revision from configuration %s: %v", names.Config, err)
			}

			t.Log("When the containers keep crashing, the revision should have error status.")
			err = test.WaitForRevisionState(clients.ServingClient, revisionName, func(r *v1alpha1.Revision) (bool, error) {
				cond := r.Status.GetCondition(v1alpha1.RevisionConditionReady)
				if cond != nil {
					if cond.Reason == exitCodeReason && strings.Contains(cond.Message, errorLog) {
						return true, nil
					}
					return true, fmt.Errorf("The revision %s was not marked with expected error condition (Reason=%q, Message=%q), but with (Reason=%q, Message=%q)",
						revisionName, exitCodeReason, errorLog, cond.Reason, cond.Message)
				}
				return false, nil
			}, "RevisionContainersCrashing")

			if err != nil {
				t.Fatalf("Failed to validate revision state: %s", err)
			}

			t.Log("When the revision has error condition, logUrl should be populated.")
			if _, err = getLogURLFromRevision(clients, revisionName); err != nil {
				t.Fatalf("Failed to get logUrl from revision %s: %v", revisionName, err)
			}
		})
	}
}

func TestContainerExitingMsg4(t *testing.T) {
	t.Parallel()

	const (
		// The given image will always exit with an exit code of 5
		exitCodeReason = "ExitCode5"
		// ... and will print "Crashed..." before it exits
		errorLog = "Crashed..."
	)

	tests := []struct {
		Name          string
		LivenessProbe *corev1.Probe
	}{
		{
			// Deployment: container-exiting-msg4-http1-pskmvwph
			Name: "http1",
			LivenessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{},
				},
			},
		},
		{
			// Deployment: container-exiting-msg4-http2-fnqpjhrw
			Name: "http2",
			LivenessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{},
				},
				InitialDelaySeconds: 0,
				TimeoutSeconds:      1,
				PeriodSeconds:       5,
				SuccessThreshold:    1,
				FailureThreshold:    3,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			clients := setup(t)

			names := test.ResourceNames{
				Config: test.ObjectNameForTest(t),
				Image:  "failing",
			}

			t.Logf("Creating a new Configuration %s", names.Image)

			if _, err := test.CreateConfiguration(t, clients, names, &test.Options{LivenessProbe: tt.LivenessProbe}); err != nil {
				t.Fatalf("Failed to create configuration %s: %v", names.Config, err)
			}
			//defer test.TearDown(clients, names)
			//test.CleanupOnInterrupt(func() { test.TearDown(clients, names) })

			t.Log("When the containers keep crashing, the Configuration should have error status.")

			err := test.WaitForConfigurationState(clients.ServingClient, names.Config, func(r *v1alpha1.Configuration) (bool, error) {
				cond := r.Status.GetCondition(v1alpha1.ConfigurationConditionReady)
				if cond != nil && !cond.IsUnknown() {
					if strings.Contains(cond.Message, errorLog) && cond.IsFalse() {
						return true, nil
					}
					t.Logf("Reason: %s ; Message: %s ; Status: %s", cond.Reason, cond.Message, cond.Status)
					return true, fmt.Errorf("The configuration %s was not marked with expected error condition (Reason=\"%s\", Message=\"%s\", Status=\"%s\"), but with (Reason=\"%s\", Message=\"%s\", Status=\"%s\")",
						names.Config, containerMissing, errorLog, "False", cond.Reason, cond.Message, cond.Status)
				}
				return false, nil
			}, "ConfigContainersCrashing")

			if err != nil {
				t.Fatalf("Failed to validate configuration state: %s", err)
			}

			revisionName, err := getRevisionFromConfiguration(clients, names.Config)
			if err != nil {
				t.Fatalf("Failed to get revision from configuration %s: %v", names.Config, err)
			}

			t.Log("When the containers keep crashing, the revision should have error status.")
			err = test.WaitForRevisionState(clients.ServingClient, revisionName, func(r *v1alpha1.Revision) (bool, error) {
				cond := r.Status.GetCondition(v1alpha1.RevisionConditionReady)
				if cond != nil {
					if cond.Reason == exitCodeReason && strings.Contains(cond.Message, errorLog) {
						return true, nil
					}
					return true, fmt.Errorf("The revision %s was not marked with expected error condition (Reason=%q, Message=%q), but with (Reason=%q, Message=%q)",
						revisionName, exitCodeReason, errorLog, cond.Reason, cond.Message)
				}
				return false, nil
			}, "RevisionContainersCrashing")

			if err != nil {
				t.Fatalf("Failed to validate revision state: %s", err)
			}

			t.Log("When the revision has error condition, logUrl should be populated.")
			if _, err = getLogURLFromRevision(clients, revisionName); err != nil {
				t.Fatalf("Failed to get logUrl from revision %s: %v", revisionName, err)
			}
		})
	}
}

func TestContainerExitingMsg5(t *testing.T) {
	t.Parallel()

	const (
		// The given image will always exit with an exit code of 5
		exitCodeReason = "ExitCode5"
		// ... and will print "Crashed..." before it exits
		errorLog = "Crashed..."
	)

	tests := []struct {
		Name           string
		ReadinessProbe *corev1.Probe
		LivenessProbe  *corev1.Probe
	}{
		{
			// container-exiting-msg5-both-default-ztlhnrzw
			Name: "both-default",
			ReadinessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{},
				},
			},
			LivenessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{},
				},
			},
		},
		{
			// container-exiting-msg5-r-less-than-g-mgohlbhz
			Name: "r-less-than-g",
			ReadinessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{},
				},
				InitialDelaySeconds: 0,
				TimeoutSeconds:      1,
				PeriodSeconds:       5,
				SuccessThreshold:    1,
				FailureThreshold:    3,
			},
			LivenessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{},
				},
				InitialDelaySeconds: 1000,
				TimeoutSeconds:      1000,
				PeriodSeconds:       1000,
				SuccessThreshold:    1,
				FailureThreshold:    10,
			},
		},
		{
			// container-exiting-msg5-r-greater-than-g-aqcfdxzv
			Name: "r-greater-than-g",
			ReadinessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{},
				},
				InitialDelaySeconds: 1000,
				TimeoutSeconds:      1000,
				PeriodSeconds:       1000,
				SuccessThreshold:    1,
				FailureThreshold:    10,
			},
			LivenessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{},
				},
				InitialDelaySeconds: 0,
				TimeoutSeconds:      1,
				PeriodSeconds:       5,
				SuccessThreshold:    1,
				FailureThreshold:    3,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			clients := setup(t)

			names := test.ResourceNames{
				Config: test.ObjectNameForTest(t),
				Image:  "failing",
			}

			t.Logf("Creating a new Configuration %s", names.Image)

			if _, err := test.CreateConfiguration(t, clients, names,
				&test.Options{ReadinessProbe: tt.ReadinessProbe, LivenessProbe: tt.LivenessProbe}); err != nil {
				t.Fatalf("Failed to create configuration %s: %v", names.Config, err)
			}
			//defer test.TearDown(clients, names)
			//test.CleanupOnInterrupt(func() { test.TearDown(clients, names) })

			t.Log("When the containers keep crashing, the Configuration should have error status.")

			err := test.WaitForConfigurationState(clients.ServingClient, names.Config, func(r *v1alpha1.Configuration) (bool, error) {
				cond := r.Status.GetCondition(v1alpha1.ConfigurationConditionReady)
				if cond != nil && !cond.IsUnknown() {
					if strings.Contains(cond.Message, errorLog) && cond.IsFalse() {
						return true, nil
					}
					t.Logf("Reason: %s ; Message: %s ; Status: %s", cond.Reason, cond.Message, cond.Status)
					return true, fmt.Errorf("The configuration %s was not marked with expected error condition (Reason=\"%s\", Message=\"%s\", Status=\"%s\"), but with (Reason=\"%s\", Message=\"%s\", Status=\"%s\")",
						names.Config, containerMissing, errorLog, "False", cond.Reason, cond.Message, cond.Status)
				}
				return false, nil
			}, "ConfigContainersCrashing")

			if err != nil {
				t.Fatalf("Failed to validate configuration state: %s", err)
			}

			revisionName, err := getRevisionFromConfiguration(clients, names.Config)
			if err != nil {
				t.Fatalf("Failed to get revision from configuration %s: %v", names.Config, err)
			}

			t.Log("When the containers keep crashing, the revision should have error status.")
			err = test.WaitForRevisionState(clients.ServingClient, revisionName, func(r *v1alpha1.Revision) (bool, error) {
				cond := r.Status.GetCondition(v1alpha1.RevisionConditionReady)
				if cond != nil {
					if cond.Reason == exitCodeReason && strings.Contains(cond.Message, errorLog) {
						return true, nil
					}
					return true, fmt.Errorf("The revision %s was not marked with expected error condition (Reason=%q, Message=%q), but with (Reason=%q, Message=%q)",
						revisionName, exitCodeReason, errorLog, cond.Reason, cond.Message)
				}
				return false, nil
			}, "RevisionContainersCrashing")

			if err != nil {
				t.Fatalf("Failed to validate revision state: %s", err)
			}

			t.Log("When the revision has error condition, logUrl should be populated.")
			if _, err = getLogURLFromRevision(clients, revisionName); err != nil {
				t.Fatalf("Failed to get logUrl from revision %s: %v", revisionName, err)
			}
		})
	}
}
