// Copyright © 2020, 2021 Attestant Limited.
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

package phase0_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/goccy/go-yaml"
	"github.com/golang/snappy"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestSignedBeaconBlockHeaderJSON(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		err   string
	}{
		{
			name: "Empty",
			err:  "unexpected end of JSON input",
		},
		{
			name:  "JSONBad",
			input: []byte("[]"),
			err:   "invalid JSON: json: cannot unmarshal array into Go value of type phase0.signedBeaconBlockHeaderJSON",
		},
		{
			name:  "MessageMissing",
			input: []byte(`{"signature":"0x606162636465666768696a6b6c6d6e6f707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f909192939495969798999a9b9c9d9e9fa0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebf"}`),
			err:   "message missing",
		},
		{
			name:  "MessageWrongType",
			input: []byte(`{"message":true,"signature":"0x606162636465666768696a6b6c6d6e6f707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f909192939495969798999a9b9c9d9e9fa0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebf"}`),
			err:   "invalid JSON: invalid JSON: json: cannot unmarshal bool into Go value of type phase0.beaconBlockHeaderJSON",
		},
		{
			name:  "MessageInvalid",
			input: []byte(`{"message":{"invalid":true},"signature":"0x606162636465666768696a6b6c6d6e6f707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f909192939495969798999a9b9c9d9e9fa0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebf"}`),
			err:   "invalid JSON: slot missing",
		},
		{
			name:  "SignatureMissing",
			input: []byte(`{"message":{"slot":"1","proposer_index":"2","parent_root":"0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f","state_root":"0x202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f","body_root":"0x404142434445464748494a4b4c4d4e4f505152535455565758595a5b5c5d5e5f"}}`),
			err:   "incorrect length for signature",
		},
		{
			name:  "SignatureWrongType",
			input: []byte(`{"message":{"slot":"1","proposer_index":"2","parent_root":"0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f","state_root":"0x202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f","body_root":"0x404142434445464748494a4b4c4d4e4f505152535455565758595a5b5c5d5e5f"},"signature":true}`),
			err:   "invalid JSON: json: cannot unmarshal bool into Go struct field signedBeaconBlockHeaderJSON.signature of type string",
		},
		{
			name:  "SignatureInvalid",
			input: []byte(`{"message":{"slot":"1","proposer_index":"2","parent_root":"0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f","state_root":"0x202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f","body_root":"0x404142434445464748494a4b4c4d4e4f505152535455565758595a5b5c5d5e5f"},"signature":"invalid"}`),
			err:   "invalid value for signature: encoding/hex: invalid byte: U+0069 'i'",
		},
		{
			name:  "SignatureShort",
			input: []byte(`{"message":{"slot":"1","proposer_index":"2","parent_root":"0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f","state_root":"0x202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f","body_root":"0x404142434445464748494a4b4c4d4e4f505152535455565758595a5b5c5d5e5f"},"signature":"0x6162636465666768696a6b6c6d6e6f707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f909192939495969798999a9b9c9d9e9fa0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebf"}`),
			err:   "incorrect length for signature",
		},
		{
			name:  "SignatureLong",
			input: []byte(`{"message":{"slot":"1","proposer_index":"2","parent_root":"0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f","state_root":"0x202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f","body_root":"0x404142434445464748494a4b4c4d4e4f505152535455565758595a5b5c5d5e5f"},"signature":"0x60606162636465666768696a6b6c6d6e6f707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f909192939495969798999a9b9c9d9e9fa0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebf"}`),
			err:   "incorrect length for signature",
		},
		{
			name:  "Good",
			input: []byte(`{"message":{"slot":"1","proposer_index":"2","parent_root":"0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f","state_root":"0x202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f","body_root":"0x404142434445464748494a4b4c4d4e4f505152535455565758595a5b5c5d5e5f"},"signature":"0x606162636465666768696a6b6c6d6e6f707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f909192939495969798999a9b9c9d9e9fa0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebf"}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var res phase0.SignedBeaconBlockHeader
			err := json.Unmarshal(test.input, &res)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				rt, err := json.Marshal(&res)
				require.NoError(t, err)
				assert.Equal(t, string(test.input), string(rt))
			}
		})
	}
}

func TestSignedBeaconBlockHeaderYAML(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		root  []byte
		err   string
	}{
		{
			name:  "Good",
			input: []byte(`{message: {slot: 1, proposer_index: 2, parent_root: '0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f', state_root: '0x202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f', body_root: '0x404142434445464748494a4b4c4d4e4f505152535455565758595a5b5c5d5e5f'}, signature: '0x606162636465666768696a6b6c6d6e6f707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f909192939495969798999a9b9c9d9e9fa0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebf'}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var res phase0.SignedBeaconBlockHeader
			err := yaml.Unmarshal(test.input, &res)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				rt, err := yaml.Marshal(&res)
				require.NoError(t, err)
				assert.Equal(t, string(rt), res.String())
				rt = bytes.TrimSuffix(rt, []byte("\n"))
				assert.Equal(t, string(test.input), string(rt))
			}
		})
	}
}

func TestSignedBeaconBlockHeaderSpec(t *testing.T) {
	if os.Getenv("ETH2_SPEC_TESTS_DIR") == "" {
		t.Skip("ETH2_SPEC_TESTS_DIR not suppplied, not running spec tests")
	}
	baseDir := filepath.Join(os.Getenv("ETH2_SPEC_TESTS_DIR"), "tests", "mainnet", "phase0", "ssz_static", "SignedBeaconBlockHeader", "ssz_random")
	require.NoError(t, filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if path == baseDir {
			// Only interested in subdirectories.
			return nil
		}
		require.NoError(t, err)
		if info.IsDir() {
			t.Run(info.Name(), func(t *testing.T) {
				specYAML, err := os.ReadFile(filepath.Join(path, "value.yaml"))
				require.NoError(t, err)
				var res phase0.SignedBeaconBlockHeader
				require.NoError(t, yaml.Unmarshal(specYAML, &res))

				compressedSpecSSZ, err := os.ReadFile(filepath.Join(path, "serialized.ssz_snappy"))
				require.NoError(t, err)
				var specSSZ []byte
				specSSZ, err = snappy.Decode(specSSZ, compressedSpecSSZ)
				require.NoError(t, err)

				// Ensure this matches the expected hash tree root.
				ssz, err := res.MarshalSSZ()
				require.NoError(t, err)
				require.Equal(t, specSSZ, ssz)
			})
		}
		return nil
	}))
}
