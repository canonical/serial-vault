package service

import "testing"

func TestClearSignFile(t *testing.T) {
	const assertions = "ABCD123456||聖誕快樂||A1234/L"

	// Get the test private key
	key, err := getPrivateKey(TestPrivateKeyPath)
	if err != nil {
		t.Errorf("Error reading the private key file: %v", err)
	}

	response, err := ClearSign(assertions, string(key), "")
	if err != nil {
		t.Errorf("Error signing the assertions text: %v", err)
	}
	if len(response) == 0 {
		t.Errorf("Empty signed data returned.")
	}
}
