package diff

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/corneliusweig/rakkess/internal/client/result"
)

func TestDiffPreservesRequestErrors(t *testing.T) {
	for _, tc := range []struct {
		name        string
		left, right result.Access
		want        string
	}{
		{"error to allowed", result.RequestErr, result.Allowed, "ERR"},
		{"allowed to error", result.Allowed, result.RequestErr, "ERR"},
		{"error to denied", result.RequestErr, result.Denied, "ERR"},
		{"denied to error", result.Denied, result.RequestErr, "ERR"},
		{"error to not applicable", result.RequestErr, result.NotApplicable, "ERR"},
		{"not applicable to error", result.NotApplicable, result.RequestErr, "ERR"},
		{"both errors", result.RequestErr, result.RequestErr, "ERR"},
		{"gained access", result.Denied, result.Allowed, "yes"},
		{"lost access", result.Allowed, result.Denied, "no"},
		{"unchanged allowed", result.Allowed, result.Allowed, ""},
		{"unchanged denied", result.Denied, result.Denied, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			left := result.ResourceAccess{"pods": {"get": result.Allowed, "create": tc.left}}
			right := result.ResourceAccess{"pods": {"get": result.Allowed, "create": tc.right}}
			var out bytes.Buffer
			Diff(left, right, []string{"get", "create"}).Render(&out, "ascii-table")
			want := []string{"NAME", "GET", "CREATE"}
			if tc.want != "" {
				want = append(want, "pods", "n/a", tc.want)
			}
			if got := strings.Fields(out.String()); !reflect.DeepEqual(got, want) {
				t.Fatalf("diff output = %v, want %v", got, want)
			}
		})
	}
}
