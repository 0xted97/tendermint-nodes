package auth

import (
	"encoding/json"
	"fmt"
	"testing"
)

// Test data for verifiers.
var (
// Add your test data for verifiers here, for example:
// googleVerifierTestData = GoogleVerifierParams{IDToken: "test_id_token", VerifierID: "test_verifier_id"}
)

func TestNewAuthService(t *testing.T) {
	// Add your test cases here
}

func TestListVerifiers(t *testing.T) {
	// Add your test cases here
}

func TestLookup(t *testing.T) {
	googleVerifier := NewGoogleVerifier()
	testCases := []struct {
		name               string
		verifierIdentifier string
		expectedError      string
	}{
		{
			name:               "ValidVerifierIdentifier",
			verifierIdentifier: "google",
			expectedError:      "",
		},
		{
			name:               "InvalidVerifierIdentifier",
			verifierIdentifier: "something",
			expectedError:      "Verifier with verifierIdentifier something could not be found",
		},
	}

	authService := NewAuthService([]Verifier{
		// Add your test verifiers here, for example:
		googleVerifier,
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := authService.Lookup(tc.verifierIdentifier)
			if err != nil && err.Error() != tc.expectedError {
				t.Errorf("Expected error: %s, got: %s", tc.expectedError, err.Error())
			}
			if err == nil && tc.expectedError != "" {
				t.Errorf("Expected error: %s, got: nil", tc.expectedError)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	// Initialize your test verifiers here, for example:
	googleVerifier := NewGoogleVerifier()

	testCases := []struct {
		name          string
		rawMessage    string
		expected      bool
		expectedID    string
		expectedError string
	}{
		{
			name:          "ValidTokenAndVerifierIdentifier",
			rawMessage:    `{"verifier_id":"khiemnguyen@lecle.co.kr","id_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6ImM5YWZkYTM2ODJlYmYwOWViMzA1NWMxYzRiZDM5Yjc1MWZiZjgxOTUiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiI0MDc0MDg3MTgxOTIuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJhdWQiOiI0MDc0MDg3MTgxOTIuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJzdWIiOiIxMDk5NzA0NTE1MDQwNDYzNzcxNzgiLCJoZCI6ImxlY2xlLmNvLmtyIiwiZW1haWwiOiJraGllbW5ndXllbkBsZWNsZS5jby5rciIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJhdF9oYXNoIjoiR0pxSXo4YVQzbDFMQTBrMHA4YnJ3ZyIsIm5hbWUiOiJLaGllbSBOZ3V5ZW4iLCJwaWN0dXJlIjoiaHR0cHM6Ly9saDMuZ29vZ2xldXNlcmNvbnRlbnQuY29tL2EvQUdObXl4YnI4bUtYcHZhVEVYVnlpZzVPcERqV3gwc2xDYS1ZVl9KQ21sMzI9czk2LWMiLCJnaXZlbl9uYW1lIjoiS2hpZW0iLCJmYW1pbHlfbmFtZSI6Ik5ndXllbiIsImxvY2FsZSI6ImVuIiwiaWF0IjoxNjgzMjcwOTU3LCJleHAiOjE2ODMyNzQ1NTd9.BKo5et9Ke58WkddNKC3rtxn9CSq4Wx5_nQTqnJNWD7_Z5ihIuISio1e6cVJ-5lH6HjIAiuXt_r41u868ZoR8rvEojYDsGtxdZDY8b9segQvAUuYOap6O0ddlbxLDIb0a39uVEbjmivdgtdm3CheuBVrX7w5rEev6smUB8EpvlkeL804sBDKFsFljKuK2so8A7Ca37-o3HbpCLj-njbQHS9gn89b-m9dR6ILsYMdQ7qCQqBMWzq9vXw_0cUx68lNJ3ZrysLvGAbppWHkI2CE6CTDW8HBsSkq2zhNVEqc94bRsJ31xjHrzTzAkMuUeYhCYnsISUdAvkO7z-uEbIJlXSA", "verifieridentifier": "google"}`,
			expected:      true,
			expectedID:    "khiemnguyen@lecle.co.kr",
			expectedError: "",
		},
		// Test other case
	}

	authService := NewAuthService([]Verifier{
		// Add your test verifiers here, for example:
		googleVerifier,
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rawMessage := json.RawMessage(tc.rawMessage)
			verified, verifierID, err := authService.Verify(&rawMessage)
			fmt.Printf("verified: %v\n", verified)
			fmt.Printf("verifierID: %v\n", verifierID)
			fmt.Printf("err: %v\n", err)
		})
	}
}
