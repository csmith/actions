package githubrelease

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-github/v93/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	commitSHA = "1111111111111111111111111111111111111111"
	otherSHA  = "2222222222222222222222222222222222222222"
	tagSHA    = "3333333333333333333333333333333333333333"
)

func testClient(t *testing.T, handler http.Handler) *github.Client {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := github.NewClient(
		github.WithHTTPClient(server.Client()),
		github.WithEnterpriseURLs(server.URL, server.URL),
	)
	require.NoError(t, err)
	return client
}

func refHandler(objType, sha string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, `{"ref":"refs/tags/v1.0.0","object":{"type":%q,"sha":%q}}`, objType, sha)
	}
}

func TestWaitForTag_lightweightTagMatches(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("GET /api/v3/repos/o/r/git/ref/tags/v1.0.0", refHandler("commit", commitSHA))

	require.NoError(t, waitForTag(testClient(t, mux), "o", "r", "v1.0.0", commitSHA, time.Millisecond, time.Second))
}

func TestWaitForTag_annotatedTagIsPeeled(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("GET /api/v3/repos/o/r/git/ref/tags/v1.0.0", refHandler("tag", tagSHA))
	mux.HandleFunc(
		"GET /api/v3/repos/o/r/git/tags/"+tagSHA,
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintf(w, `{"sha":%q,"object":{"type":"commit","sha":%q}}`, tagSHA, commitSHA)
		},
	)

	require.NoError(t, waitForTag(testClient(t, mux), "o", "r", "v1.0.0", commitSHA, time.Millisecond, time.Second))
}

func TestWaitForTag_wrongCommitFailsImmediately(t *testing.T) {
	var calls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc(
		"GET /api/v3/repos/o/r/git/ref/tags/v1.0.0",
		func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			refHandler("commit", otherSHA)(w, r)
		},
	)

	err := waitForTag(testClient(t, mux), "o", "r", "v1.0.0", commitSHA, time.Millisecond, time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), otherSHA)
	assert.Contains(t, err.Error(), commitSHA)
	assert.Equal(t, int32(1), calls.Load())
}

func TestWaitForTag_retriesUntilTagAppears(t *testing.T) {
	var calls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc(
		"GET /api/v3/repos/o/r/git/ref/tags/v1.0.0",
		func(w http.ResponseWriter, r *http.Request) {
			if calls.Add(1) < 3 {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"message":"Not Found"}`))
				return
			}
			refHandler("commit", commitSHA)(w, r)
		},
	)

	require.NoError(t, waitForTag(testClient(t, mux), "o", "r", "v1.0.0", commitSHA, time.Millisecond, time.Second))
	assert.Equal(t, int32(3), calls.Load())
}

func TestWaitForTag_timesOutWhenTagNeverAppears(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"GET /api/v3/repos/o/r/git/ref/tags/v1.0.0",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	)

	err := waitForTag(testClient(t, mux), "o", "r", "v1.0.0", commitSHA, time.Millisecond, 50*time.Millisecond)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "did not appear")
}

func TestWaitForTag_otherErrorsFail(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"GET /api/v3/repos/o/r/git/ref/tags/v1.0.0",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		},
	)

	err := waitForTag(testClient(t, mux), "o", "r", "v1.0.0", commitSHA, time.Millisecond, time.Second)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "did not appear")
}

func TestWaitForTag_retriesTransientErrors(t *testing.T) {
	tests := []struct {
		name string
		fail func(w http.ResponseWriter)
	}{
		{
			name: "server error",
			fail: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusBadGateway)
			},
		},
		{
			name: "too many requests",
			fail: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusTooManyRequests)
			},
		},
		{
			name: "primary rate limit",
			fail: func(w http.ResponseWriter) {
				w.Header().Set("X-Ratelimit-Remaining", "0")
				w.Header().Set("X-Ratelimit-Reset", strconv.FormatInt(time.Now().Unix(), 10))
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
			},
		},
		{
			name: "secondary rate limit",
			fail: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"documentation_url":"https://docs.github.com/#about-secondary-rate-limits"}`))
			},
		},
		{
			name: "network error",
			fail: func(w http.ResponseWriter) {
				conn, _, err := w.(http.Hijacker).Hijack()
				if err == nil {
					_ = conn.Close()
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				var calls atomic.Int32
				mux := http.NewServeMux()
				mux.HandleFunc(
					"GET /api/v3/repos/o/r/git/ref/tags/v1.0.0",
					func(w http.ResponseWriter, r *http.Request) {
						if calls.Add(1) == 1 {
							tt.fail(w)
							return
						}
						refHandler("commit", commitSHA)(w, r)
					},
				)

				require.NoError(
					t,
					waitForTag(testClient(t, mux), "o", "r", "v1.0.0", commitSHA, time.Millisecond, 5*time.Second),
				)
				assert.Equal(t, int32(2), calls.Load())
			},
		)
	}
}

func TestWaitForTag_timeoutIncludesLastTransientError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"GET /api/v3/repos/o/r/git/ref/tags/v1.0.0",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		},
	)

	err := waitForTag(testClient(t, mux), "o", "r", "v1.0.0", commitSHA, time.Millisecond, 50*time.Millisecond)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "did not appear")
	assert.Contains(t, err.Error(), "503")
}

func TestWaitForTag_honoursRetryAfter(t *testing.T) {
	var calls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc(
		"GET /api/v3/repos/o/r/git/ref/tags/v1.0.0",
		func(w http.ResponseWriter, r *http.Request) {
			if calls.Add(1) == 1 {
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			refHandler("commit", commitSHA)(w, r)
		},
	)

	start := time.Now()
	require.NoError(t, waitForTag(testClient(t, mux), "o", "r", "v1.0.0", commitSHA, time.Millisecond, 5*time.Second))
	assert.GreaterOrEqual(t, time.Since(start), time.Second)
}
