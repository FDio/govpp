package performance

import (
	"testing"

	"go.fd.io/govpp/binapi/memclnt"

	"go.fd.io/govpp/test/vpptesting"
)

func BenchmarkBinapiControlPing(b *testing.B) {

	b.Run("1", func(b *testing.B) {
		benchBinapiControlPing(b, 1)
	})
	b.Run("10", func(b *testing.B) {
		benchBinapiControlPing(b, 10)
	})
	b.Run("100", func(b *testing.B) {
		benchBinapiControlPing(b, 100)
	})
}

func benchBinapiControlPing(b *testing.B, repeatN int) {
	test := vpptesting.SetupVPP(b)
	ctx := test.Context

	c := memclnt.NewServiceClient(test.Conn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for r := 0; r < repeatN; r++ {
			_, err := c.ControlPing(ctx, &memclnt.ControlPing{})
			if err != nil {
				b.Fatalf("getting version failed: %v", err)
			}
		}
	}
	b.StopTimer()
}
