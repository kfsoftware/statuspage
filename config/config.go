package config

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TLSHealthCheck struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TLSHealthCheckSpec `json:"spec,omitempty"`
}
type TLSHealthCheckSpec struct {
	Host      string  `json:"host"`
	Port      int     `json:"port"`
	Frequency string  `json:"frequency,omitempty"`
	RootCAs   *string `json:"rootCAs,omitempty"`
}

type StatusPage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec StatusPageSpec `json:"spec,omitempty"`
}
type StatusPageSpec struct {
	Title    string   `json:"title"`
	Services []string `json:"services"`
}

type HttpHealthCheck struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HttpHealthCheckSpec `json:"spec,omitempty"`
}
type HttpHealthCheckSpec struct {
	URL        string `json:"url"`
	Frequency  string `json:"frequency,omitempty"`
	StatusCode int    `json:"statusCode"`
}
