// Copyright (c) The Thanos Authors.
// Licensed under the Apache License 2.0.

package objtesting

import (
	"os"
	"strings"
	"testing"

	"github.com/swgupta98/objstore"
	"github.com/swgupta98/objstore/client"
	"github.com/swgupta98/objstore/providers/azure"
	"github.com/swgupta98/objstore/providers/bos"
	"github.com/swgupta98/objstore/providers/cos"
	"github.com/swgupta98/objstore/providers/filesystem"
	"github.com/swgupta98/objstore/providers/gcs"
	"github.com/swgupta98/objstore/providers/obs"
	"github.com/swgupta98/objstore/providers/oci"
	"github.com/swgupta98/objstore/providers/oss"
	"github.com/swgupta98/objstore/providers/s3"
	"github.com/swgupta98/objstore/providers/swift"

	"github.com/efficientgo/core/testutil"
)

// IsObjStoreSkipped returns true if given provider ID is found in THANOS_TEST_OBJSTORE_SKIP array delimited by comma e.g:
// THANOS_TEST_OBJSTORE_SKIP=GCS,S3,AZURE,SWIFT,COS,ALIYUNOSS,BOS,OCI.
func IsObjStoreSkipped(t *testing.T, provider client.ObjProvider) bool {
	if e, ok := os.LookupEnv("THANOS_TEST_OBJSTORE_SKIP"); ok {
		obstores := strings.Split(e, ",")
		for _, objstore := range obstores {
			if objstore == string(provider) {
				t.Logf("%s found in THANOS_TEST_OBJSTORE_SKIP array. Skipping.", provider)
				return true
			}
		}
	}

	return false
}

// ForeachStore runs given test using all available objstore implementations.
// For each it creates a new bucket with a random name and a cleanup function
// that deletes it after test was run.
// Use THANOS_TEST_OBJSTORE_SKIP to skip explicitly certain object storages.
func ForeachStore(t *testing.T, testFn func(t *testing.T, bkt objstore.Bucket)) {
	t.Parallel()

	// Mandatory Inmem. Not parallel, to detect problem early.
	if ok := t.Run("inmem", func(t *testing.T) {
		testFn(t, objstore.NewInMemBucket())
	}); !ok {
		return
	}

	// Mandatory Filesystem.
	t.Run("filesystem", func(t *testing.T) {
		t.Parallel()

		dir, err := os.MkdirTemp("", "filesystem-foreach-store-test")
		testutil.Ok(t, err)
		defer testutil.Ok(t, os.RemoveAll(dir))

		b, err := filesystem.NewBucket(dir)
		testutil.Ok(t, err)
		testFn(t, b)
		testFn(t, objstore.NewPrefixedBucket(b, "some_prefix"))
	})

	// Optional GCS.
	if !IsObjStoreSkipped(t, client.GCS) {
		t.Run("gcs", func(t *testing.T) {
			bkt, closeFn, err := gcs.NewTestBucket(t, os.Getenv("GCP_PROJECT"))
			testutil.Ok(t, err)

			t.Parallel()
			defer closeFn()

			// TODO(bwplotka): Add goleak when https://github.com/GoogleCloudPlatform/google-cloud-go/issues/1025 is resolved.
			testFn(t, bkt)
			testFn(t, objstore.NewPrefixedBucket(bkt, "some_prefix"))
		})
	}

	// Optional S3.
	if !IsObjStoreSkipped(t, client.S3) {
		t.Run("aws s3", func(t *testing.T) {
			// TODO(bwplotka): Allow taking location from envvar.
			bkt, closeFn, err := s3.NewTestBucket(t, "us-west-2")
			testutil.Ok(t, err)

			t.Parallel()
			defer closeFn()

			// TODO(bwplotka): Add goleak when we fix potential leak in minio library.
			// We cannot use goleak for detecting our own potential leaks, when goleak detects leaks in minio itself.
			// This needs to be investigated more.

			testFn(t, bkt)
			testFn(t, objstore.NewPrefixedBucket(bkt, "some_prefix"))
		})
	}

	// Optional Azure.
	if !IsObjStoreSkipped(t, client.AZURE) {
		t.Run("azure", func(t *testing.T) {
			bkt, closeFn, err := azure.NewTestBucket(t, "e2e-tests")
			testutil.Ok(t, err)

			t.Parallel()
			defer closeFn()

			testFn(t, bkt)
			testFn(t, objstore.NewPrefixedBucket(bkt, "some_prefix"))
		})
	}

	// Optional SWIFT.
	if !IsObjStoreSkipped(t, client.SWIFT) {
		t.Run("swift", func(t *testing.T) {
			container, closeFn, err := swift.NewTestContainer(t)
			testutil.Ok(t, err)

			t.Parallel()
			defer closeFn()

			testFn(t, container)
			testFn(t, objstore.NewPrefixedBucket(container, "some_prefix"))
		})
	}

	// Optional COS.
	if !IsObjStoreSkipped(t, client.COS) {
		t.Run("Tencent cos", func(t *testing.T) {
			bkt, closeFn, err := cos.NewTestBucket(t)
			testutil.Ok(t, err)

			t.Parallel()
			defer closeFn()

			testFn(t, bkt)
			testFn(t, objstore.NewPrefixedBucket(bkt, "some_prefix"))
		})
	}

	// Optional OSS.
	if !IsObjStoreSkipped(t, client.ALIYUNOSS) {
		t.Run("AliYun oss", func(t *testing.T) {
			bkt, closeFn, err := oss.NewTestBucket(t)
			testutil.Ok(t, err)

			t.Parallel()
			defer closeFn()

			testFn(t, bkt)
			testFn(t, objstore.NewPrefixedBucket(bkt, "some_prefix"))
		})
	}

	// Optional BOS.
	if !IsObjStoreSkipped(t, client.BOS) {
		t.Run("Baidu BOS", func(t *testing.T) {
			bkt, closeFn, err := bos.NewTestBucket(t)
			testutil.Ok(t, err)

			t.Parallel()
			defer closeFn()

			testFn(t, bkt)
			testFn(t, objstore.NewPrefixedBucket(bkt, "some_prefix"))
		})
	}

	// Optional OCI.
	if !IsObjStoreSkipped(t, client.OCI) {
		t.Run("oci", func(t *testing.T) {
			bkt, closeFn, err := oci.NewTestBucket(t)
			testutil.Ok(t, err)

			t.Parallel()
			defer closeFn()

			testFn(t, bkt)
		})
	}

	// Optional OBS.
	if !IsObjStoreSkipped(t, client.OBS) {
		t.Run("obs", func(t *testing.T) {
			bkt, closeFn, err := obs.NewTestBucket(t, "cn-south-1")
			testutil.Ok(t, err)

			t.Parallel()
			defer closeFn()

			testFn(t, bkt)
			testFn(t, objstore.NewPrefixedBucket(bkt, "some_prefix"))
		})
	}
}
