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
		expectedExitCode int
	}{{
		name:             "happy_path",
		pathToDb:         "./tests/transaction_validation/happy_path",
		txHash:           "tx",
		expectedExitCode: 0,
	},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, "transaction", tc.pathToDb, tc.txHash)
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
		expectedExitCode int
	}{} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, "proposed-block", tc.pathToDb)
			cmd.Stderr = os.Stderr
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
