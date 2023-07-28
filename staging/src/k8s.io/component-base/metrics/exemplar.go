/*
Copyright 2023 The Kubernetes Authors.

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
	"strings"
	"time"
	"unicode/utf8"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ExemplarMaxRunes is the maximum number of runes allowed in exemplar label
// names and values.
const ExemplarMaxRunes = prometheus.ExemplarMaxRunes

// ReservedLabelPrefix is a prefix which is not legal in user-supplied
// label names.
const ReservedLabelPrefix = "__"

// newExemplar creates a new dto.Exemplar from the provided values. An error is
// returned if any of the label names or values are invalid or if the total
// number of runes in the label names and values exceeds ExemplarMaxRunes.
func newExemplar(value float64, ts time.Time, l Labels) (*dto.Exemplar, error) {
	// Set exemplar value, this is same as the value that was added to the counter.
	e := &dto.Exemplar{}
	e.Value = proto.Float64(value)

	// Set exemplar timestamp.
	tsProto := timestamppb.New(ts)
	if err := tsProto.CheckValid(); err != nil {
		return nil, err
	}
	e.Timestamp = tsProto

	// Set exemplar labels.
	labelPairs := make([]*dto.LabelPair, 0, len(l))
	var runes int
	for name, value := range l {
		if !model.LabelName(name).IsValid() && !strings.HasPrefix(name, ReservedLabelPrefix) {
			return nil, fmt.Errorf("exemplar label name %q is invalid", name)
		}
		runes += utf8.RuneCountInString(name)
		if !utf8.ValidString(value) {
			return nil, fmt.Errorf("exemplar label value %q is not valid UTF-8", value)
		}
		runes += utf8.RuneCountInString(value)
		labelPairs = append(
			labelPairs, &dto.LabelPair{
				Name:  proto.String(name),
				Value: proto.String(value),
			},
		)
	}
	if runes > ExemplarMaxRunes {
		return nil, fmt.Errorf("exemplar labels have %d runes, exceeding the limit of %d", runes, ExemplarMaxRunes)
	}
	e.Label = labelPairs

	return e, nil
}
