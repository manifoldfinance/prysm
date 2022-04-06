package endtoend

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/prysmaticlabs/prysm/config/params"
	ev "github.com/prysmaticlabs/prysm/testing/endtoend/evaluators"
	e2eParams "github.com/prysmaticlabs/prysm/testing/endtoend/params"
	"github.com/prysmaticlabs/prysm/testing/endtoend/types"
	"github.com/prysmaticlabs/prysm/testing/require"
)

func TestCheckpointSync_MainnetConfig(t *testing.T) {
	params.UseE2EMainnetConfig()
	require.NoError(t, e2eParams.InitMultiClient(e2eParams.StandardBeaconCount, e2eParams.StandardLighthouseNodeCount))

	// Run for 10 epochs if not in long-running to confirm long-running has no issues.
	var err error
	epochsToRun := 10
	epochStr, longRunning := os.LookupEnv("E2E_EPOCHS")
	if longRunning {
		epochsToRun, err = strconv.Atoi(epochStr)
		require.NoError(t, err)
	}
	_, crossClient := os.LookupEnv("RUN_CROSS_CLIENT")
	tracingPort := e2eParams.TestParams.Ports.JaegerTracingPort
	tracingEndpoint := fmt.Sprintf("127.0.0.1:%d", tracingPort)
	evals := []types.Evaluator{
		ev.PeersConnect,
		ev.HealthzCheck,
		ev.MetricsCheck,
		ev.ValidatorsAreActive,
		ev.ValidatorsParticipatingAtEpoch(2),
		ev.FinalizationOccurs(3),
		ev.ProposeVoluntaryExit,
		ev.ValidatorHasExited,
		ev.ColdStateCheckpoint,
		ev.ForkTransition,
		ev.APIMiddlewareVerifyIntegrity,
		ev.APIGatewayV1Alpha1VerifyIntegrity,
		ev.FinishedSyncing,
		ev.AllNodesHaveSameHead,
	}
	testConfig := &types.E2EConfig{
		BeaconFlags: []string{
			fmt.Sprintf("--slots-per-archive-point=%d", params.BeaconConfig().SlotsPerEpoch*16),
			fmt.Sprintf("--tracing-endpoint=http://%s", tracingEndpoint),
			"--enable-tracing",
			"--trace-sample-fraction=1.0",
		},
		ValidatorFlags:          []string{},
		EpochsToRun:             uint64(epochsToRun),
		TestSync:                true,
		TestFeature:             true,
		TestDeposits:            true,
		UseFixedPeerIDs:         true,
		UseValidatorCrossClient: crossClient,
		UsePrysmShValidator:     false,
		UsePprof:                !longRunning,
		TracingSinkEndpoint:     tracingEndpoint,
		Evaluators:              evals,
	}

	newTestRunner(t, testConfig).run()
}