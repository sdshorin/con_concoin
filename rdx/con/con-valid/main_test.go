package main

import (
	"con-valid/model"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/slon/shad-go/tools/testtool"
)

var binCache testtool.BinCache

func TestMain(m *testing.M) {
	os.Exit(func() int {
		var teardown testtool.CloseFunc
		binCache, teardown = testtool.NewBinCache()
		defer teardown()

		return m.Run()
	}())
}

func TestConValidValidateTransaction(t *testing.T) {
	binary, err := binCache.GetBinary("con-valid")
	require.NoError(t, err)

	for _, tc := range []struct {
		name             string
		pathToDb         string
		txHash           model.Hash
		maliciousMode    bool
		expectedExitCode int
	}{{
		name:             "happy_path",
		pathToDb:         "./tests/transaction_validation/happy_path",
		txHash:           "tx",
		maliciousMode:    false,
		expectedExitCode: 0,
	},
		{
			name:             "amount_is_more_than_balance",
			pathToDb:         "./tests/transaction_validation/amount_is_more_than_balance",
			txHash:           "tx",
			maliciousMode:    false,
			expectedExitCode: 1,
		},
		{
			name:             "negative_amount",
			pathToDb:         "./tests/transaction_validation/negative_amount",
			txHash:           "tx",
			maliciousMode:    false,
			expectedExitCode: 1,
		},
		{
			name:             "user_not_found",
			pathToDb:         "./tests/transaction_validation/user_not_found",
			txHash:           "tx",
			maliciousMode:    false,
			expectedExitCode: 1,
		},
		{
			name:             "signature_is_bad",
			pathToDb:         "./tests/transaction_validation/signature_is_bad",
			txHash:           "tx",
			maliciousMode:    false,
			expectedExitCode: 1,
		},
		{
			name:             "malicious_mode",
			pathToDb:         "./tests/transaction_validation/negative_amount",
			txHash:           "tx",
			maliciousMode:    true,
			expectedExitCode: 0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var cmd *exec.Cmd
			if tc.maliciousMode {
				cmd = exec.Command(binary, "--malicious", "transaction", tc.pathToDb, tc.txHash)
			} else {
				cmd = exec.Command(binary, "transaction", tc.pathToDb, tc.txHash)
			}
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout

			exit_code := 0
			if err := cmd.Run(); err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exit_code = exitError.ExitCode()
				}
			}

			require.Equal(t, exit_code, tc.expectedExitCode)
		})
	}
}

func TestConValidValidateProposedBlock(t *testing.T) {
	binary, err := binCache.GetBinary("con-valid")
	require.NoError(t, err)

	for _, tc := range []struct {
		name             string
		pathToDb         string
		maliciousMode    bool
		expectedExitCode int
	}{
		{
			name:             "happy_path",
			pathToDb:         "./tests/block_validation/happy_path",
			maliciousMode:    false,
			expectedExitCode: 0,
		},
		{
			name:             "tx_signature_is_bad",
			pathToDb:         "./tests/block_validation/tx_signature_is_bad",
			maliciousMode:    false,
			expectedExitCode: 1,
		},
		{
			name:             "malicious_mode",
			pathToDb:         "./tests/block_validation/tx_signature_is_bad",
			maliciousMode:    true,
			expectedExitCode: 0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var cmd *exec.Cmd
			if tc.maliciousMode {
				cmd = exec.Command(binary, "--malicious", "proposed-block", tc.pathToDb)
			} else {
				cmd = exec.Command(binary, "proposed-block", tc.pathToDb)
			}
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			exit_code := 0
			if err := cmd.Run(); err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exit_code = exitError.ExitCode()
				}
			}

			require.Equal(t, exit_code, tc.expectedExitCode)
		})
	}
}
