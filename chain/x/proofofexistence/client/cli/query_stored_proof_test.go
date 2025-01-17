package cli_test

import (
	"fmt"
	"strconv"
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/empowerchain/empowerchain/testutil/network"
	"github.com/empowerchain/empowerchain/testutil/nullify"
	"github.com/empowerchain/empowerchain/x/proofofexistence/client/cli"
	"github.com/empowerchain/empowerchain/x/proofofexistence/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func networkWithProofObjects(t *testing.T, n int) (*network.Network, []types.Proof) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	for i := 0; i < n; i++ {
		proof := types.Proof{
			Hash: strconv.Itoa(i),
		}
		nullify.Fill(&proof)
		state.ProofList = append(state.ProofList, proof)
	}
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), state.ProofList
}

func TestShowProof(t *testing.T) {
	net, objs := networkWithProofObjects(t, 2)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc   string
		idHash string

		args []string
		err  error
		obj  types.Proof
	}{
		{
			desc:   "found",
			idHash: objs[0].Hash,

			args: common,
			obj:  objs[0],
		},
		{
			desc:   "not found",
			idHash: strconv.Itoa(100000),

			args: common,
			err:  status.Error(codes.NotFound, "not found"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{
				tc.idHash,
			}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowProof(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetProofResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.Proof)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.Proof),
				)
			}
		})
	}
}
