// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controlapi

import (
	"context"
	"errors"
	"testing"

	"go.ate.dev/substrate/internal/installdefaults"
	"go.ate.dev/substrate/internal/resources"
	"go.ate.dev/substrate/pkg/proto/ateapipb"
)

func TestEnsureAteletTerminated_NodeGone(t *testing.T) {
	actorRef := resources.ActorRef{Atespace: "team-a", Name: "id1"}
	actor := &ateapipb.Actor{
		Metadata: &ateapipb.ResourceMetadata{Atespace: "team-a", Name: "id1", Uid: "uid-1"},
		Status:   &ateapipb.ActorStatus{State: ateapipb.ActorState_ACTOR_STATE_DELETING, AssignedNode: "node-1"},
	}
	for _, tc := range []struct {
		name    string
		gone    func(context.Context, string) (bool, error)
		wantErr bool
	}{
		{name: "deleted node needs no terminate", gone: func(context.Context, string) (bool, error) { return true, nil }},
		{name: "node still there waits for its atelet", gone: func(context.Context, string) (bool, error) { return false, nil }, wantErr: true},
		{name: "unknown node state waits", gone: func(context.Context, string) (bool, error) { return false, errors.New("api down") }, wantErr: true},
		{name: "no checker waits", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var opts []DialerOption
			if tc.gone != nil {
				opts = append(opts, WithNodeGone(tc.gone))
			}
			dialer := NewAteletDialer(newTestAteletIndexer(t), installdefaults.AteletSPIFFEID(installdefaults.SystemNamespace), "", "", opts...)
			w := &ActorWorkflow{dialer: dialer}
			err := w.ensureAteletTerminated(context.Background(), actorRef, actor, nil, "test")
			if tc.wantErr && !errors.Is(err, ErrNoAteletOnNode) {
				t.Fatalf("ensureAteletTerminated = %v, want ErrNoAteletOnNode", err)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("ensureAteletTerminated = %v, want nil", err)
			}
		})
	}
}
