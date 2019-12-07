package transaction

import (
	"context"
	"github.com/ququzone/ckb-sdk-go/rpc"
	"github.com/ququzone/ckb-sdk-go/types"
)

type SystemScriptCell struct {
	CellHash types.Hash
	OutPoint *types.OutPoint
}

type SystemScripts struct {
	SecpCell     *SystemScriptCell
	MultiSigCell *SystemScriptCell
	DaoCell      *SystemScriptCell
}

func NewSystemScript(client rpc.Client) (*SystemScripts, error) {
	genesis, err := client.GetBlockByNumber(context.Background(), 0)
	if err != nil {
		return nil, err
	}

	secpHash, err := genesis.Transactions[0].Outputs[1].Type.Hash()
	if err != nil {
		return nil, err
	}
	multiSigHash, err := genesis.Transactions[0].Outputs[4].Type.Hash()
	if err != nil {
		return nil, err
	}
	daoHash, err := genesis.Transactions[0].Outputs[2].Type.Hash()
	if err != nil {
		return nil, err
	}

	return &SystemScripts{
		SecpCell: &SystemScriptCell{
			CellHash: secpHash,
			OutPoint: &types.OutPoint{
				TxHash: genesis.Transactions[1].Hash,
				Index:  0,
			},
		},
		MultiSigCell: &SystemScriptCell{
			CellHash: multiSigHash,
			OutPoint: &types.OutPoint{
				TxHash: genesis.Transactions[1].Hash,
				Index:  1,
			},
		},
		DaoCell: &SystemScriptCell{
			CellHash: daoHash,
			OutPoint: &types.OutPoint{
				TxHash: genesis.Transactions[0].Hash,
				Index:  2,
			},
		},
	}, nil
}