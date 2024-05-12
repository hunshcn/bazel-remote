package server

import (
	"crypto"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/buchgr/bazel-remote/v2/cache"
	asset "github.com/buchgr/bazel-remote/v2/genproto/build/bazel/remote/asset/v1"
	//pb "github.com/buchgr/bazel-remote/v2/genproto/build/bazel/remote/execution/v2"

	"google.golang.org/grpc/codes"

	testutils "github.com/buchgr/bazel-remote/v2/utils"
)

func testAssetFetchBlob(hashType crypto.Hash) func(t *testing.T) {
	return func(t *testing.T) {
		fixture := grpcTestSetup(t)
		defer os.Remove(fixture.tempdir)

		ts := newTestGetServer(hashType)

		hash := strings.TrimSuffix(ts.path, ".tar.gz")
		hashBytes, err := hex.DecodeString(hash)
		if err != nil {
			t.Fatal(err)
		}

		req := asset.FetchBlobRequest{
			Uris: []string{
				ts.srv.URL + "/404.unrecognisedextension",
				ts.srv.URL + "/404.tar.gz",
				ts.srv.URL + "/" + ts.path, // This URL should work.
			},
			Qualifiers: []*asset.Qualifier{
				{
					Name: "checksum.sri",
					Value: cache.GetHashTypePrefix(hashType) + "-" +
						base64.StdEncoding.EncodeToString(hashBytes),
				},
			},
		}
		fmt.Println(req)

		resp, err := fixture.assetClient.FetchBlob(ctx, &req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.Status.GetCode() != int32(codes.OK) {
			t.Fatal("expected successful fetch", resp)
		}
		if resp.BlobDigest == nil {
			t.Fatal("expected non-bil BlobDigest")
		}
		if resp.BlobDigest.Hash != hash {
			t.Fatal("mismatching BlobDigest hash returned")
		}
	}
}

func TestAssetFetchBlob(t *testing.T) {
	t.Parallel()
	//cache.EmptyHashes
	for hashType := range map[crypto.Hash]bool{crypto.SHA256: false} {
		t.Run("hash-type-"+cache.GetHashTypePrefix(hashType), testAssetFetchBlob(hashType))
	}
}

type testGetServer struct {
	srv *httptest.Server

	blob []byte
	path string
}

func (s *testGetServer) handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Unsupported method for this test",
			http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/"+s.path {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)

	if r.Method == http.MethodHead {
		w.Header().Set("ContentLength", fmt.Sprintf("%d", len(s.blob)))
	}

	if r.Method == http.MethodGet {
		_, _ = w.Write(s.blob)
	}
}

func newTestGetServer(hashType crypto.Hash) *testGetServer {
	blob, hash := testutils.RandomDataAndHash(hashType, 256)

	ts := testGetServer{
		blob: blob,
		path: hash + ".tar.gz",
	}
	ts.srv = httptest.NewServer(http.HandlerFunc(ts.handler))

	return &ts
}
