// Copyright (C) 2024 Akave
// See LICENSE for copying information.

package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"akave.ai/akavesdk/private/memory"
)

type testCase struct {
	name           string
	args           []string
	expectedOutput []string
	expectError    bool
}

func TestMain(m *testing.M) {
	initFlags()

	// Run tests
	code := m.Run()

	os.Exit(code)
}

func TestCreateBucketCommand(t *testing.T) {
	bucketName, err := randString(10)
	require.NoError(t, err)
	nodeAddress := PickNodeRPCAddress(t)
	testCases := []testCase{
		{
			name:           "Create bucket successfully",
			args:           []string{"--node-address", nodeAddress, "bucket", "create", bucketName},
			expectedOutput: []string{"Bucket created"},
			expectError:    false,
		},
		{
			name:           "Create bucket already exists",
			args:           []string{"--node-address", nodeAddress, "bucket", "create", bucketName},
			expectedOutput: []string{"failed to create bucket: sdk: rpc error: code = AlreadyExists"},
			expectError:    true,
		},
		{
			name:           "Empty bucket name provided",
			args:           []string{"--node-address", nodeAddress, "bucket", "create", ""},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"--node-address", nodeAddress, "bucket", "create"},
			expectedOutput: []string{"Error: create bucket command expects exactly 1 argument [bucket name]; got 0"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestIPCCreateBucketCommand(t *testing.T) {
	bucketName, err := randString(10)
	require.NoError(t, err)
	nodeAddress := PickNodeRPCAddress(t)
	privateKey := PickPrivateKey(t)

	testCases := []testCase{
		{
			name:           "Create bucket successfully",
			args:           []string{"--node-address", nodeAddress, "ipc", "bucket", "create", "--private-key", privateKey, bucketName},
			expectedOutput: []string{"Bucket created"},
			expectError:    false,
		},
		{
			name:           "Create bucket already exists",
			args:           []string{"--node-address", nodeAddress, "ipc", "bucket", "create", "--private-key", privateKey, bucketName},
			expectedOutput: []string{"failed to create bucket: sdk: execution reverted"},
			expectError:    true,
		},
		{
			name:           "Empty bucket name provided",
			args:           []string{"--node-address", nodeAddress, "ipc", "bucket", "create", "--private-key", privateKey, ""},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"--node-address", nodeAddress, "ipc", "bucket", "create", "--private-key", privateKey},
			expectedOutput: []string{"create bucket command expects exactly 1 argument [bucket name]; got 0"},
			expectError:    true,
		},
		{
			name:           "Invalid private key provided",
			args:           []string{"--node-address", nodeAddress, "ipc", "bucket", "create", "--private-key", "51ccv2", bucketName},
			expectedOutput: []string{"Error: invalid hex character"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestDeleteBucketCommand(t *testing.T) {
	bucketName, err := randString(10)
	require.NoError(t, err)
	nodeAddress := PickNodeRPCAddress(t)
	_, _, err = captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName, "--node-address", nodeAddress})
	require.NoError(t, err)
	testCases := []testCase{
		{
			name:           "Delete bucket successfully",
			args:           []string{"bucket", "delete", bucketName, "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("Bucket deleted: Name=%s", bucketName)},
			expectError:    false,
		},
		{
			name:           "Delete not existing bucket",
			args:           []string{"bucket", "delete", bucketName, "--node-address", nodeAddress},
			expectedOutput: []string{"failed to delete bucket: sdk: rpc error: code = NotFound"},
			expectError:    true,
		},
		{
			name:           "Empty bucket name provided",
			args:           []string{"bucket", "delete", "", "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"bucket", "delete", "--node-address", nodeAddress},
			expectedOutput: []string{"Error: delete bucket command expects exactly 1 argument [bucket name]; got 0"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestIPCDeleteBucketCommand(t *testing.T) {
	bucketName, err := randString(10)
	require.NoError(t, err)
	nodeAddress := PickNodeRPCAddress(t)
	privateKey := PickPrivateKey(t)

	_, _, err = captureCobraOutput(rootCmd, []string{"--node-address", nodeAddress, "ipc", "bucket", "create", "--private-key", privateKey, bucketName})
	require.NoError(t, err)
	testCases := []testCase{
		{
			name:           "Delete bucket successfully",
			args:           []string{"ipc", "bucket", "delete", "--private-key", privateKey, bucketName, "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("Bucket deleted: Name=%s", bucketName)},
			expectError:    false,
		},
		{
			name:           "Delete not existing bucket",
			args:           []string{"ipc", "bucket", "delete", "--private-key", privateKey, bucketName, "--node-address", nodeAddress},
			expectedOutput: []string{"failed to get bucket: not found"},
			expectError:    true,
		},
		{
			name:           "Empty bucket name provided",
			args:           []string{"ipc", "bucket", "delete", "--private-key", privateKey, "", "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"ipc", "bucket", "delete", "--private-key", privateKey, "--node-address", nodeAddress},
			expectedOutput: []string{"Error: create bucket command expects exactly 1 argument [bucket name]; got 0"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestViewBucketCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, stderr, err := captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)
	bucketID, err := extractBucketID(stderr)
	assert.NoError(t, err)

	testCases := []testCase{
		{
			name:           "View bucket successfully",
			args:           []string{"bucket", "view", bucketName, "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("Bucket: ID=%s, Name=%s", bucketID, bucketName)},
			expectError:    false,
		},
		{
			name:           "View non-existent bucket",
			args:           []string{"bucket", "view", "error", "--node-address", nodeAddress},
			expectedOutput: []string{"failed to get bucket: sdk: rpc error: code = NotFound"},
			expectError:    true,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"bucket", "view", "", "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestIPCViewBucketCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	privateKey := PickPrivateKey(t)

	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, stderr, err := captureCobraOutput(rootCmd, []string{"--node-address", nodeAddress, "ipc", "bucket", "create", "--private-key", privateKey, bucketName})
	assert.NoError(t, err)
	bucketID, err := extractBucketID(stderr)
	assert.NoError(t, err)

	testCases := []testCase{
		{
			name:           "View bucket successfully",
			args:           []string{"ipc", "bucket", "view", "--private-key", privateKey, bucketName, "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("Bucket: ID=%s, Name=%s", bucketID, bucketName)},
			expectError:    false,
		},
		{
			name:           "View non-existent bucket",
			args:           []string{"ipc", "bucket", "view", "--private-key", privateKey, "error", "--node-address", nodeAddress},
			expectedOutput: []string{"failed to get bucket: not found"},
			expectError:    true,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"ipc", "bucket", "view", "--private-key", privateKey, "", "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "Private key not provided",
			args:           []string{"ipc", "bucket", "view", "--private-key", "", bucketName, "--node-address", nodeAddress},
			expectedOutput: []string{"invalid length, need 256 bits"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestListBucketsCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	bucketName1, err := randString(10)
	assert.NoError(t, err)
	_, stderr, err := captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName1, "--node-address", nodeAddress})
	assert.NoError(t, err)
	bucketID1, err := extractBucketID(stderr)
	assert.NoError(t, err)

	bucketName2, err := randString(10)
	assert.NoError(t, err)
	_, stderr, err = captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName2, "--node-address", nodeAddress})
	assert.NoError(t, err)
	bucketID2, err := extractBucketID(stderr)
	assert.NoError(t, err)

	testCase := testCase{
		name:           "List buckets successfully",
		args:           []string{"bucket", "list", "--node-address", nodeAddress},
		expectedOutput: []string{fmt.Sprintf("Bucket: ID=%s, Name=%s", bucketID1, bucketName1), fmt.Sprintf("Bucket: ID=%s, Name=%s", bucketID2, bucketName2)},
		expectError:    false,
	}

	t.Run(testCase.name, func(t *testing.T) {
		stdout, stderr, err := captureCobraOutput(rootCmd, testCase.args)

		if testCase.expectError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		for _, expected := range testCase.expectedOutput {
			assert.Contains(t, stdout+stderr, expected)
		}
	})
}

func TestIPCListBucketsCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	privateKey := PickPrivateKey(t)

	bucketName1, err := randString(10)
	assert.NoError(t, err)
	_, stderr, err := captureCobraOutput(rootCmd, []string{"--node-address", nodeAddress, "ipc", "bucket", "create", "--private-key", privateKey, bucketName1})
	assert.NoError(t, err)
	bucketID1, err := extractBucketID(stderr)
	assert.NoError(t, err)

	bucketName2, err := randString(10)
	assert.NoError(t, err)
	_, stderr, err = captureCobraOutput(rootCmd, []string{"--node-address", nodeAddress, "ipc", "bucket", "create", "--private-key", privateKey, bucketName2})
	assert.NoError(t, err)
	bucketID2, err := extractBucketID(stderr)
	assert.NoError(t, err)

	testCase := testCase{
		name:           "List buckets successfully",
		args:           []string{"ipc", "bucket", "list", "--node-address", nodeAddress, "--private-key", privateKey},
		expectedOutput: []string{fmt.Sprintf("Bucket: ID=%s, Name=%s", bucketID1, bucketName1), fmt.Sprintf("Bucket: ID=%s, Name=%s", bucketID2, bucketName2)},
		expectError:    false,
	}

	t.Run(testCase.name, func(t *testing.T) {
		stdout, stderr, err := captureCobraOutput(rootCmd, testCase.args)

		if testCase.expectError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		for _, expected := range testCase.expectedOutput {
			assert.Contains(t, stdout+stderr, expected)
		}
	})
}

func TestListFilesCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file1, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"file", "upload", bucketName, file1, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file2, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"file", "upload", bucketName, file2, "--node-address", nodeAddress})
	assert.NoError(t, err)

	testCases := []testCase{
		{
			name:           "List files successfully",
			args:           []string{"file", "list", bucketName, "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("File: Name=%s", filepath.Base(file1)), fmt.Sprintf("File: Name=%s", filepath.Base(file2))},
			expectError:    false,
		},
		{
			name:           "List files for non-existent bucket",
			args:           []string{"file", "list", "nonexistent-bucket", "--node-address", nodeAddress},
			expectedOutput: []string{"failed to list files: sdk: rpc error: code = NotFound"},
			expectError:    true,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"file", "list", "", "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestIPCListFilesCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	privateKey := PickPrivateKey(t)

	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"ipc", "bucket", "create", "--private-key", privateKey, bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file1, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"ipc", "file", "upload", "--private-key", privateKey, bucketName, file1, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file2, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"ipc", "file", "upload", "--private-key", privateKey, bucketName, file2, "--node-address", nodeAddress})
	assert.NoError(t, err)

	testCases := []testCase{
		{
			name:           "List files successfully",
			args:           []string{"ipc", "file", "list", "--private-key", privateKey, bucketName, "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("File: Name=%s", filepath.Base(file1)), fmt.Sprintf("File: Name=%s", filepath.Base(file2))},
			expectError:    false,
		},
		{
			name:           "List files for non-existent bucket",
			args:           []string{"ipc", "file", "list", "--private-key", privateKey, "nonexistent-bucket", "--node-address", nodeAddress},
			expectedOutput: []string{"No files"},
			expectError:    false,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"ipc", "file", "list", "--private-key", privateKey, "", "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "Private key not provided",
			args:           []string{"ipc", "file", "list", "--private-key", "", bucketName, "--node-address", nodeAddress},
			expectedOutput: []string{"Error: invalid length, need 256 bits"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestStreamingListFilesCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file1, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)
	file2, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)

	_, _, err = captureCobraOutput(rootCmd, []string{"files-streaming", "upload", bucketName, file1, "--node-address", nodeAddress})
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"files-streaming", "upload", bucketName, file2, "--node-address", nodeAddress})
	assert.NoError(t, err)

	testCases := []testCase{
		// Streaming API
		{
			name:           "List files successfully",
			args:           []string{"files-streaming", "list", bucketName, "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("File: Name=%s", filepath.Base(file1)), fmt.Sprintf("File: Name=%s", filepath.Base(file2))},
			expectError:    false,
		},
		{
			name:           "List files for non-existent bucket",
			args:           []string{"files-streaming", "list", "nonexistent-bucket", "--node-address", nodeAddress},
			expectedOutput: []string{"failed to list files: sdk: rpc error: code = NotFound"},
			expectError:    true,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"files-streaming", "list", "", "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestFileInfoCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)
	_, stderr, err := captureCobraOutput(rootCmd, []string{"file", "upload", bucketName, file, "--node-address", nodeAddress})
	assert.NoError(t, err)
	rootCID, err := extractRootCID(stderr)
	assert.NoError(t, err)

	testCases := []testCase{
		{
			name:           "File info successfully",
			args:           []string{"file", "info", bucketName, filepath.Base(file), "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("File: Name=%s, RootCID=%s", filepath.Base(file), rootCID)},
			expectError:    false,
		},
		{
			name:           "File info for non-existent file",
			args:           []string{"file", "info", bucketName, "nonexistent-file", "--node-address", nodeAddress},
			expectedOutput: []string{"failed to get file info: sdk: rpc error: code = NotFound"},
			expectError:    true,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"file", "info", "", filepath.Base(file), "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "File name not provided",
			args:           []string{"file", "info", bucketName, "", "--node-address", nodeAddress},
			expectedOutput: []string{"file name is required"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestIPCFileInfoCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	privateKey := PickPrivateKey(t)

	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"ipc", "bucket", "create", "--private-key", privateKey, bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)
	_, stderr, err := captureCobraOutput(rootCmd, []string{"ipc", "file", "upload", "--private-key", privateKey, bucketName, file, "--node-address", nodeAddress})
	assert.NoError(t, err)
	rootCID, err := extractRootCID(stderr)
	assert.NoError(t, err)

	testCases := []testCase{
		{
			name:           "File info successfully",
			args:           []string{"ipc", "file", "info", "--private-key", privateKey, bucketName, filepath.Base(file), "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("File: Name=%s, RootCID=%s", filepath.Base(file), rootCID)},
			expectError:    false,
		},
		{
			name:           "File info for non-existent file",
			args:           []string{"ipc", "file", "info", "--private-key", privateKey, bucketName, "nonexistent-file", "--node-address", nodeAddress},
			expectedOutput: []string{"invalid cid: varints malformed, could not reach the end"},
			expectError:    true,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"ipc", "file", "info", "--private-key", privateKey, "", filepath.Base(file), "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "File name not provided",
			args:           []string{"ipc", "file", "info", "--private-key", privateKey, bucketName, "", "--node-address", nodeAddress},
			expectedOutput: []string{"file name is required"},
			expectError:    true,
		},
		{
			name:           "Private key not provided",
			args:           []string{"ipc", "file", "info", "--private-key", "", bucketName, filepath.Base(file), "--node-address", nodeAddress},
			expectedOutput: []string{"Error: invalid length, need 256 bits"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestStreamingFileInfoCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)
	_, stderr, err := captureCobraOutput(rootCmd, []string{"files-streaming", "upload", bucketName, file, "--node-address", nodeAddress})
	assert.NoError(t, err)
	rootCID, err := extractRootCID(stderr)
	assert.NoError(t, err)

	testCases := []testCase{
		// Streaming API
		{
			name:           "File info successfully",
			args:           []string{"files-streaming", "info", bucketName, filepath.Base(file), "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("File: Name=%s, RootCID=%s", filepath.Base(file), rootCID)},
			expectError:    false,
		},
		{
			name:           "File info for non-existent file",
			args:           []string{"files-streaming", "info", bucketName, "nonexistent-file", "--node-address", nodeAddress},
			expectedOutput: []string{"failed to get file info: sdk: rpc error: code = NotFound"},
			expectError:    true,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"files-streaming", "info", "", filepath.Base(file), "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "File name not provided",
			args:           []string{"files-streaming", "info", bucketName, "", "--node-address", nodeAddress},
			expectedOutput: []string{"file name is required"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestFileUploadCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)

	testCases := []testCase{
		{
			name:           "File upload successfully",
			args:           []string{"file", "upload", bucketName, file, "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("File uploaded successfully: Name=%s", filepath.Base(file))},
			expectError:    false,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"file", "upload", "", file, "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "File path not provided",
			args:           []string{"file", "upload", bucketName, "", "--node-address", nodeAddress},
			expectedOutput: []string{"file path is required"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestIPCFileUploadCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	privateKey := PickPrivateKey(t)

	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"ipc", "bucket", "create", "--private-key", privateKey, bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)

	testCases := []testCase{
		{
			name:           "File upload successfully",
			args:           []string{"ipc", "file", "upload", "--private-key", privateKey, bucketName, file, "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("File uploaded successfully: Name=%s", filepath.Base(file))},
			expectError:    false,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"ipc", "file", "upload", "--private-key", privateKey, "", file, "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "File path not provided",
			args:           []string{"ipc", "file", "upload", "--private-key", privateKey, bucketName, "", "--node-address", nodeAddress},
			expectedOutput: []string{"file path is required"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestStreamingFileUploadCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)

	testCases := []testCase{
		{
			name:           "File upload successfully",
			args:           []string{"files-streaming", "upload", bucketName, file, "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("File uploaded successfully: Name=%s", filepath.Base(file))},
			expectError:    false,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"files-streaming", "upload", "", file, "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "File path not provided",
			args:           []string{"files-streaming", "upload", bucketName, "", "--node-address", nodeAddress},
			expectedOutput: []string{"file path is required"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestFileDownloadCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)

	_, _, err = captureCobraOutput(rootCmd, []string{"file", "upload", bucketName, file, "--node-address", nodeAddress})
	assert.NoError(t, err)

	tempDir := t.TempDir()

	testCases := []testCase{
		{
			name:           "File download successfully",
			args:           []string{"file", "download", bucketName, filepath.Base(file), tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("File downloaded successfully: Name=%s", filepath.Base(file))},
			expectError:    false,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"file", "download", "", filepath.Base(file), tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "File name not provided",
			args:           []string{"file", "download", bucketName, "", tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{"file name is required"},
			expectError:    true,
		},
		{
			name:           "Destination path not provided",
			args:           []string{"file", "download", bucketName, filepath.Base(file), "", "--node-address", nodeAddress},
			expectedOutput: []string{"destination path is required"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestRangeFileDownloadCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file, err := createTempFile(t, 10*memory.MB.ToInt64())
	assert.NoError(t, err)

	_, _, err = captureCobraOutput(rootCmd, []string{"file", "upload", bucketName, file, "--node-address", nodeAddress})
	assert.NoError(t, err)

	tempDir := t.TempDir()

	testCases := []testCase{
		{
			name:           "File download successfully",
			args:           []string{"file", "download-range", bucketName, filepath.Base(file), "0-3", tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("File part downloaded successfully: Name=%s, Blocks=0-3", filepath.Base(file))},
			expectError:    false,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"file", "download-range", "", filepath.Base(file), "0-3", tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "File name not provided",
			args:           []string{"file", "download-range", bucketName, "", "0-3", tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{"file name is required"},
			expectError:    true,
		},
		{
			name:           "Range not provided",
			args:           []string{"file", "download-range", bucketName, filepath.Base(file), "", tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{"range should be in the format start-end"},
			expectError:    true,
		},
		{
			name:           "Destination path not provided",
			args:           []string{"file", "download-range", bucketName, filepath.Base(file), "0-3", "", "--node-address", nodeAddress},
			expectedOutput: []string{"destination path is required"},
			expectError:    true,
		},
		{
			name:           "Not nmbers in ranges",
			args:           []string{"file", "download-range", bucketName, filepath.Base(file), "42-bar", tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{"failed to parse ranges in 42-bar"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestCmdFileDownloadV2(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	bucketName, err := randString(10)
	assert.NoError(t, err)

	_, _, err = captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)

	_, _, err = captureCobraOutput(rootCmd, []string{"file", "upload", bucketName, file, "--node-address", nodeAddress})
	assert.NoError(t, err)

	tempDir := t.TempDir()

	testCases := []struct {
		name           string
		args           []string
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "File download V2 successfully",
			args: []string{"file", "download2", bucketName, filepath.Base(file), tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{
				fmt.Sprintf("File downloaded successfully: Name=%s, Path=%s", filepath.Base(file), filepath.Join(tempDir, filepath.Base(file))),
			},
			expectError: false,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"file", "download2", "", filepath.Base(file), tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{"Error: bucket name is required"},
			expectError:    true,
		},
		{
			name:           "File name not provided",
			args:           []string{"file", "download2", bucketName, "", tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{"Error: file name is required"},
			expectError:    true,
		},
		{
			name:           "Destination path not provided",
			args:           []string{"file", "download2", bucketName, filepath.Base(file), "", "--node-address", nodeAddress},
			expectedOutput: []string{"Error: destination path is required"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			output := stdout + stderr

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, output, expected)
			}

			if !tc.expectError && tc.name == "File download V2 successfully" {
				downloadedFilePath := filepath.Join(tempDir, filepath.Base(file))
				_, err := os.Stat(downloadedFilePath)
				assert.NoError(t, err)

				originalContent, err := os.ReadFile(file)
				assert.NoError(t, err)

				downloadedContent, err := os.ReadFile(downloadedFilePath)
				assert.NoError(t, err)

				assert.Equal(t, originalContent, downloadedContent)
			}
		})
	}
}

func TestIPCFileDownloadCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	privateKey := PickPrivateKey(t)

	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"ipc", "bucket", "create", "--private-key", privateKey, bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)

	_, _, err = captureCobraOutput(rootCmd, []string{"ipc", "file", "upload", "--private-key", privateKey, bucketName, file, "--node-address", nodeAddress})
	assert.NoError(t, err)

	tempDir, err := os.MkdirTemp("", "test-download")
	assert.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		assert.NoError(t, err)
	}()

	testCases := []testCase{
		{
			name:           "File download successfully",
			args:           []string{"ipc", "file", "download", "--private-key", privateKey, bucketName, filepath.Base(file), tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("File downloaded successfully: Name=%s", filepath.Base(file))},
			expectError:    false,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"ipc", "file", "download", "--private-key", privateKey, "", filepath.Base(file), tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "File name not provided",
			args:           []string{"ipc", "file", "download", "--private-key", privateKey, bucketName, "", tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{"file name is required"},
			expectError:    true,
		},
		{
			name:           "Destination path not provided",
			args:           []string{"ipc", "file", "download", "--private-key", privateKey, bucketName, filepath.Base(file), "", "--node-address", nodeAddress},
			expectedOutput: []string{"destination path is required"},
			expectError:    true,
		},
		{
			name:           "Private key not provided",
			args:           []string{"ipc", "file", "download", "--private-key", "", bucketName, filepath.Base(file), tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{"Error: invalid length, need 256 bits"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestStreamingFileDownloadCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"files-streaming", "upload", bucketName, file, "--node-address", nodeAddress})
	assert.NoError(t, err)

	tempDir, err := os.MkdirTemp("", "test-download")
	assert.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		assert.NoError(t, err)
	}()

	testCases := []testCase{
		{
			name:           "File download successfully",
			args:           []string{"files-streaming", "download", bucketName, filepath.Base(file), tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("File downloaded successfully: Name=%s", filepath.Base(file))},
			expectError:    false,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"files-streaming", "download", "", filepath.Base(file), tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "File name not provided",
			args:           []string{"files-streaming", "download", bucketName, "", tempDir, "--node-address", nodeAddress},
			expectedOutput: []string{"file name is required"},
			expectError:    true,
		},
		{
			name:           "Destination path not provided",
			args:           []string{"files-streaming", "download", bucketName, filepath.Base(file), "", "--node-address", nodeAddress},
			expectedOutput: []string{"destination path is required"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestStreamingFileDeleteCommand(t *testing.T) {
	nodeAddress := PickNodeRPCAddress(t)
	bucketName, err := randString(10)
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"bucket", "create", bucketName, "--node-address", nodeAddress})
	assert.NoError(t, err)

	file, err := createTempFile(t, 2*memory.MB.ToInt64())
	assert.NoError(t, err)
	_, _, err = captureCobraOutput(rootCmd, []string{"files-streaming", "upload", bucketName, file, "--node-address", nodeAddress})
	assert.NoError(t, err)

	testCases := []testCase{
		// Streaming API
		{
			name:           "File delete successfully",
			args:           []string{"files-streaming", "delete", bucketName, filepath.Base(file), "--node-address", nodeAddress},
			expectedOutput: []string{fmt.Sprintf("File deleted successfully: Name=%s", filepath.Base(file))},
			expectError:    false,
		},
		{
			name:           "File delete for non-existent file",
			args:           []string{"files-streaming", "delete", bucketName, "nonexistent-file", "--node-address", nodeAddress},
			expectedOutput: []string{"failed to delete file: sdk: rpc error: code = NotFound"},
			expectError:    true,
		},
		{
			name:           "Bucket name not provided",
			args:           []string{"files-streaming", "delete", "", filepath.Base(file), "--node-address", nodeAddress},
			expectedOutput: []string{"bucket name is required"},
			expectError:    true,
		},
		{
			name:           "File name not provided",
			args:           []string{"files-streaming", "delete", bucketName, "", "--node-address", nodeAddress},
			expectedOutput: []string{"file name is required"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func TestCmdCarCreate(t *testing.T) {
	t.Skip("Fix me, this test does not work")

	nodeAddress := PickNodeRPCAddress(t)
	fileName := "test.car"
	destination := t.TempDir()
	path := filepath.Join(destination, fileName)
	_, _, err := captureCobraOutput(rootCmd, []string{"car", "create", path, "--node-address", nodeAddress})
	assert.NoError(t, err)

	testCases := []testCase{
		{
			name:           "CAR file created successfully",
			args:           []string{"car", "create", path, "--node-address", nodeAddress},
			expectedOutput: []string{"CAR file: Name=test.car"},
			expectError:    false,
		},
		{
			name:           "Destination path is required",
			args:           []string{"car", "create", "", "--node-address", nodeAddress},
			expectedOutput: []string{"destination file path is required"},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureCobraOutput(rootCmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout+stderr, expected)
			}
		})
	}
}

func randString(i int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, i)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b), nil
}

func createTempFile(t *testing.T, size int64) (string, error) {
	t.Helper()
	tempFile, err := os.CreateTemp(t.TempDir(), "test-file")
	if err != nil {
		return "", err
	}
	txt, err := randString(int(size))
	if err != nil {
		return "", err
	}
	_, err = tempFile.WriteString(txt)
	if err != nil {
		return "", err
	}
	err = tempFile.Close()
	if err != nil {
		return "", err
	}
	return tempFile.Name(), nil
}

func extractRootCID(output string) (string, error) {
	re := regexp.MustCompile(`RootCID=([a-zA-Z0-9]+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to extract root CID")
	}
	return matches[1], nil
}

func captureCobraOutput(cmd *cobra.Command, args []string) (string, string, error) {
	var stdoutBuf, stderrBuf bytes.Buffer

	cmd.SetOut(&stdoutBuf)
	cmd.SetErr(&stderrBuf)

	cmd.SetArgs(args)

	err := cmd.Execute()

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	return stdout, stderr, err
}

func extractBucketID(logMessage string) (string, error) {
	pattern := `Bucket created: ID=([a-zA-Z0-9\-]+), CreatedAt=`

	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to compile regex: %w", err)
	}

	matches := re.FindStringSubmatch(logMessage)
	if len(matches) < 2 {
		return "", fmt.Errorf("no ID found in log message")
	}

	id := matches[1]
	return id, nil
}

var (
	// nodeAddress is flag to set RPC address of akave node.
	nodeAddress = flag.String("node-rpc-address", os.Getenv("NODE_RPC_ADDRESS"), "flag to set node rpc address")
	// privateKeyTest is flag to set deployers hex private key.
	privateKeyTest = flag.String("cmd-private-key", os.Getenv("PRIVATE_KEY"), "flag to set deployers hex private key for client tests")
)

// PickNodeRPCAddress picks node PRC address from flag.
func PickNodeRPCAddress(t testing.TB) string {
	if *nodeAddress == "" || strings.EqualFold(*nodeAddress, "omit") {
		t.Skip("node rpc address flag missing, example: -NODE_RPC_ADDRESS=<node rpc address>")
	}
	return *nodeAddress
}

// PickPrivateKey picks hex private key of deployer.
func PickPrivateKey(t testing.TB) string {
	if *privateKeyTest == "" || strings.EqualFold(*privateKeyTest, "omit") {
		t.Skip("private key flag missing, example: -PRIVATE_KEY=<deployers hex private key>")
	}
	return *privateKeyTest
}
