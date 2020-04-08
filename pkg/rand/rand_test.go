package rand_test

import (
	"testing"

	"github.com/zitryss/aye-and-nay/pkg/rand"
)

func TestRandId(t *testing.T) {
	tests := []struct {
		give int
		want int
	}{
		{
			give: 1,
			want: 1,
		},
		{
			give: 2,
			want: 2,
		},
		{
			give: 3,
			want: 3,
		},
		{
			give: 4,
			want: 4,
		},
		{
			give: 5,
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			id, err := rand.Id(tt.give)
			if err != nil {
				t.Error("err != nil")
			}
			got := len(id)
			if got != tt.want {
				t.Errorf("len(uid.New(%v)) = %v, want %v", tt.give, got, tt.want)
			}
		})
	}
}
