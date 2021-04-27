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

package metrics

import (
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	metricNamespace                      = "security_profiles_operator"
	metricNameSeccompProfile             = "seccomp_profile"
	metricLabelValueSeccompProfileUpdate = "update"
	metricLabelValueSeccompProfileDelete = "delete"
	metricLabelOperation                 = "operation"
)

// Metrics is the main structure of this package.
type Metrics struct {
	impl                 impl
	log                  logr.Logger
	metricSeccompProfile *prometheus.CounterVec
}

// New returns a new Metrics instance.
func New() *Metrics {
	return &Metrics{
		impl: &defaultImpl{},
		log:  ctrl.Log.WithName("metrics"),
		metricSeccompProfile: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      metricNameSeccompProfile,
				Namespace: metricNamespace,
				Help:      "Counter about seccomp profile operations.",
			},
			[]string{metricLabelOperation},
		),
	}
}

// Register iterates over all available metrics and registers them.
func (m *Metrics) Register() error {
	for name, collector := range map[string]prometheus.Collector{
		metricNameSeccompProfile: m.metricSeccompProfile,
	} {
		m.log.Info(fmt.Sprintf("Registering metric: %s", name))
		if err := m.impl.Register(collector); err != nil {
			return errors.Wrapf(err, "register collector for %s metric", name)
		}
	}
	return nil
}

// Serve creates an HTTP endpoint "/metrics" and starts serving them.
func (m *Metrics) Serve() error {
	const (
		addr = ":8081" // controller-runtime is already serving metrics on :8080
		path = "/metrics"
	)
	m.log.Info(fmt.Sprintf("Serving metrics on %s%s", addr, path))

	handler := &http.ServeMux{}
	handler.Handle(path, promhttp.Handler())
	err := m.impl.ListenAndServe(addr, handler)
	return errors.Wrap(err, "serve metrics")
}

// IncSeccompProfileUpdate increments the seccomp profile update counter.
func (m *Metrics) IncSeccompProfileUpdate() {
	m.metricSeccompProfile.
		WithLabelValues(metricLabelValueSeccompProfileUpdate).Inc()
}

// IncSeccompProfileDelete increments the seccomp profile deletion counter.
func (m *Metrics) IncSeccompProfileDelete() {
	m.metricSeccompProfile.
		WithLabelValues(metricLabelValueSeccompProfileDelete).Inc()
}