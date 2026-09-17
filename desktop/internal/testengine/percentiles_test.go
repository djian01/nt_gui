package testengine

import (
	"math/rand"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func TestLatencyPercentilesExactRanks(t *testing.T) {
	for _, order := range []string{"ascending", "descending", "random", "duplicates"} {
		t.Run(order, func(t *testing.T) {
			var tracker latencyPercentiles
			var session Session
			var values []float64
			rng := rand.New(rand.NewSource(42))
			tracker.add(Sample{RTT: 9999}, &session)
			if session.P95RTT != nil || session.P99RTT != nil {
				t.Fatal("failure produced percentiles")
			}
			for i := 0; i < 1200; i++ {
				value := float64(i)
				switch order {
				case "descending":
					value = 1200 - value
				case "random":
					value = rng.Float64() * 10000
				case "duplicates":
					value = float64(i % 7)
				}
				previous := session
				tracker.add(Sample{Success: true, RTT: value}, &session)
				values = append(values, value)
				sorted := append([]float64(nil), values...)
				sort.Float64s(sorted)
				if *session.P95RTT != sorted[(len(sorted)*95+99)/100-1] || *session.P99RTT != sorted[(len(sorted)*99+99)/100-1] {
					t.Fatalf("incorrect rank at %d", i)
				}
				if previous.P95RTT != nil && previous.P95RTT == session.P95RTT {
					t.Fatal("mutated emitted snapshot")
				}
				tracker.add(Sample{RTT: 1e9}, &session)
				if *session.P99RTT != sorted[(len(sorted)*99+99)/100-1] {
					t.Fatal("failure altered percentile")
				}
			}
		})
	}
}

func TestPercentilesSavedAndLegacyAllProtocols(t *testing.T) {
	for _, protocol := range []string{"http", "icmp", "tcp", "dns"} {
		t.Run(protocol, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "results.db")
			store, err := OpenStore(path)
			if err != nil {
				t.Fatal(err)
			}
			session := Session{ID: protocol, Config: Config{Type: protocol, Recording: true}, Revision: 1}
			var tracker latencyPercentiles
			for i := 1; i <= 120; i++ {
				// Large early values must remain represented after leaving Recent's six samples.
				sample := Sample{Sequence: i, Time: time.Now(), Success: i <= 100, RTT: float64(101 - i)}
				updateImportedSummary(&session, sample)
				tracker.add(sample, &session)
				if err := store.Save(session, &sample); err != nil {
					t.Fatal(err)
				}
			}
			if err := store.Close(); err != nil {
				t.Fatal(err)
			}
			store, err = OpenStore(path)
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			check := func() {
				t.Helper()
				got, err := store.Get(session.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got.P95RTT == nil || got.P99RTT == nil || *got.P95RTT != 95 || *got.P99RTT != 99 {
					t.Fatalf("incorrect percentiles: %+v", got)
				}
			}
			check()
			session.P95RTT, session.P99RTT = nil, nil
			if err := store.Save(session, nil); err != nil {
				t.Fatal(err)
			}
			check()
			session.Succeeded++ // Older partial recordings cannot recover missing samples.
			if err := store.Save(session, nil); err != nil {
				t.Fatal(err)
			}
			got, err := store.Get(session.ID)
			if err != nil {
				t.Fatal(err)
			}
			if got.P95RTT != nil || got.P99RTT != nil {
				t.Fatal("partial history reported full-session percentiles")
			}
		})
	}
}
