/*
Copyright 2021 The Kubernetes Authors.

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

package v1alpha5

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestSpecIsValid(t *testing.T) {
	cases := []struct {
		Spec          Metal3MachineSpec
		ErrorExpected bool
		Name          string
	}{
		{
			Spec:          Metal3MachineSpec{},
			ErrorExpected: true,
			Name:          "empty spec",
		},
		{
			Spec: Metal3MachineSpec{
				Image: Image{
					URL:      "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2",
					Checksum: "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2.sha256sum",
				},
				UserData: &corev1.SecretReference{
					Name: "worker-user-data",
				},
			},
			ErrorExpected: false,
			Name:          "Valid spec without UserData.Namespace",
		},
		{
			Spec: Metal3MachineSpec{
				Image: Image{
					URL:      "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2",
					Checksum: "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2.sha256sum",
				},
				UserData: &corev1.SecretReference{
					Name:      "worker-user-data",
					Namespace: "otherns",
				},
			},
			ErrorExpected: false,
			Name:          "Valid spec with UserData.Namespace",
		},
		{
			Spec: Metal3MachineSpec{
				Image: Image{
					Checksum: "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2.sha256sum",
				},
				UserData: &corev1.SecretReference{
					Name: "worker-user-data",
				},
			},
			ErrorExpected: true,
			Name:          "missing Image.URL",
		},
		{
			Spec: Metal3MachineSpec{
				Image: Image{
					URL: "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2",
				},
				UserData: &corev1.SecretReference{
					Name: "worker-user-data",
				},
			},
			ErrorExpected: true,
			Name:          "missing Image.Checksum",
		},
		{
			Spec: Metal3MachineSpec{
				Image: Image{
					URL:      "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2",
					Checksum: "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2.sha256sum",
				},
			},
			ErrorExpected: false,
			Name:          "missing optional UserData",
		},
		{
			Spec: Metal3MachineSpec{
				Image: Image{
					URL:      "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2",
					Checksum: "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2.sha256sum",
				},
				UserData: &corev1.SecretReference{
					Namespace: "otherns",
				},
			},
			ErrorExpected: false,
			Name:          "missing optional UserData.Name",
		},
		{
			Spec: Metal3MachineSpec{
				Image: Image{
					URL:      "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2",
					Checksum: "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2.sha256sum",
				},
				HostSelector: HostSelector{},
			},
			ErrorExpected: false,
			Name:          "Empty HostSelector provided",
		},
		{
			Spec: Metal3MachineSpec{
				Image: Image{
					URL:      "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2",
					Checksum: "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2.sha256sum",
				},
				HostSelector: HostSelector{
					MatchLabels: map[string]string{"key": "value"},
				},
			},
			ErrorExpected: false,
			Name:          "HostSelector Single MatchLabel provided",
		},
		{
			Spec: Metal3MachineSpec{
				Image: Image{
					URL:      "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2",
					Checksum: "http://172.22.0.1/images/rhcos-ootpa-latest.qcow2.sha256sum",
				},
				HostSelector: HostSelector{
					MatchLabels: map[string]string{"key": "value", "key2": "value2"},
				},
			},
			ErrorExpected: false,
			Name:          "HostSelector Multiple MatchLabels provided",
		},
	}

	for _, tc := range cases {
		err := tc.Spec.IsValid()
		if tc.ErrorExpected && err == nil {
			t.Errorf("Did not get error from case \"%v\"", tc.Name)
		}
		if !tc.ErrorExpected && err != nil {
			t.Errorf("Got unexpected error from case \"%v\": %v", tc.Name, err)
		}
	}
}
