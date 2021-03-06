package transaction

import (
	"encoding/binary"
	"errors"

	"github.com/ququzone/ckb-sdk-go/crypto"
	"github.com/ququzone/ckb-sdk-go/crypto/blake2b"
	"github.com/ququzone/ckb-sdk-go/types"
	"github.com/ququzone/ckb-sdk-go/utils"
)

var (
	EmptyWitnessArg = &types.WitnessArgs{
		Lock:       make([]byte, 65),
		InputType:  nil,
		OutputType: nil,
	}
	EmptyWitnessArgPlaceholder = make([]byte, 89)
	SignaturePlaceholder       = make([]byte, 65)
)

func NewSecp256k1SingleSigTx(scripts *utils.SystemScripts) *types.Transaction {
	return &types.Transaction{
		Version:    0,
		HeaderDeps: []types.Hash{},
		CellDeps: []*types.CellDep{
			{
				OutPoint: scripts.SecpSingleSigCell.OutPoint,
				DepType:  types.DepTypeDepGroup,
			},
		},
	}
}

func NewSecp256k1MultiSigTx(scripts *utils.SystemScripts) *types.Transaction {
	return &types.Transaction{
		Version:    0,
		HeaderDeps: []types.Hash{},
		CellDeps: []*types.CellDep{
			{
				OutPoint: scripts.SecpMultiSigCell.OutPoint,
				DepType:  types.DepTypeDepGroup,
			},
		},
	}
}

func NewSecp256k1HybirdSigTx(scripts *utils.SystemScripts) *types.Transaction {
	return &types.Transaction{
		Version:    0,
		HeaderDeps: []types.Hash{},
		CellDeps: []*types.CellDep{

			{
				OutPoint: scripts.SecpSingleSigCell.OutPoint,
				DepType:  types.DepTypeDepGroup,
			},
			{
				OutPoint: scripts.SecpMultiSigCell.OutPoint,
				DepType:  types.DepTypeDepGroup,
			},
		},
	}
}

func AddInputsForTransaction(transaction *types.Transaction, cells []*types.Cell) ([]int, *types.WitnessArgs, error) {
	if len(cells) == 0 {
		return nil, nil, errors.New("input cells empty")
	}
	group := make([]int, len(cells))
	start := len(transaction.Witnesses)
	for i := 0; i < len(cells); i++ {
		cell := cells[i]
		input := &types.CellInput{
			Since: 0,
			PreviousOutput: &types.OutPoint{
				TxHash: cell.OutPoint.TxHash,
				Index:  cell.OutPoint.Index,
			},
		}
		transaction.Inputs = append(transaction.Inputs, input)
		transaction.Witnesses = append(transaction.Witnesses, []byte{})
		group[i] = start + i
	}
	transaction.Witnesses[start] = EmptyWitnessArgPlaceholder
	return group, EmptyWitnessArg, nil
}

func SingleSignTransaction(transaction *types.Transaction, group []int, witnessArgs *types.WitnessArgs, key crypto.Key) error {
	data, err := witnessArgs.Serialize()
	if err != nil {
		return err
	}
	length := make([]byte, 8)
	binary.LittleEndian.PutUint64(length, uint64(len(data)))

	hash, err := transaction.ComputeHash()
	if err != nil {
		return err
	}

	message := append(hash.Bytes(), length...)
	message = append(message, data...)

	if len(group) > 1 {
		for i := 1; i < len(group); i++ {
			var data []byte
			length := make([]byte, 8)
			binary.LittleEndian.PutUint64(length, uint64(len(data)))
			message = append(message, length...)
			message = append(message, data...)
		}
	}

	message, err = blake2b.Blake256(message)
	if err != nil {
		return err
	}

	signed, err := key.Sign(message)
	if err != nil {
		return err
	}

	wa := &types.WitnessArgs{
		Lock:       signed,
		InputType:  witnessArgs.InputType,
		OutputType: witnessArgs.OutputType,
	}
	wab, err := wa.Serialize()
	if err != nil {
		return err
	}

	transaction.Witnesses[group[0]] = wab

	return nil
}

func MultiSignTransaction(transaction *types.Transaction, group []int, witnessArgs *types.WitnessArgs, serialize []byte, keys ...crypto.Key) error {
	var emptySignature []byte
	for range keys {
		emptySignature = append(emptySignature, SignaturePlaceholder...)
	}
	witnessArgs.Lock = append(serialize, emptySignature...)

	data, err := witnessArgs.Serialize()
	if err != nil {
		return err
	}
	length := make([]byte, 8)
	binary.LittleEndian.PutUint64(length, uint64(len(data)))

	hash, err := transaction.ComputeHash()
	if err != nil {
		return err
	}

	message := append(hash.Bytes(), length...)
	message = append(message, data...)

	if len(group) > 1 {
		for i := 1; i < len(group); i++ {
			var data []byte
			length := make([]byte, 8)
			binary.LittleEndian.PutUint64(length, uint64(len(data)))
			message = append(message, length...)
			message = append(message, data...)
		}
	}

	message, err = blake2b.Blake256(message)
	if err != nil {
		return err
	}

	var signed []byte
	for _, key := range keys {
		s, err := key.Sign(message)
		if err != nil {
			return err
		}
		signed = append(signed, s...)
	}

	wa := &types.WitnessArgs{
		Lock:       append(serialize, signed...),
		InputType:  witnessArgs.InputType,
		OutputType: witnessArgs.OutputType,
	}
	wab, err := wa.Serialize()
	if err != nil {
		return err
	}

	transaction.Witnesses[group[0]] = wab

	return nil
}

func SingleSegmentSignMessage(transaction *types.Transaction, start int, end int, witnessArgs *types.WitnessArgs) ([]byte, error) {
	data, err := witnessArgs.Serialize()
	if err != nil {
		return nil, err
	}
	length := make([]byte, 8)
	binary.LittleEndian.PutUint64(length, uint64(len(data)))

	hash, err := transaction.ComputeHash()
	if err != nil {
		return nil, err
	}

	message := append(hash.Bytes(), length...)
	message = append(message, data...)

	for i := start + 1; i < end; i++ {
		var data []byte
		length := make([]byte, 8)
		binary.LittleEndian.PutUint64(length, uint64(len(data)))
		message = append(message, length...)
		message = append(message, data...)
	}

	return blake2b.Blake256(message)
}

func SingleSegmentSignTransaction(transaction *types.Transaction, start int, end int, witnessArgs *types.WitnessArgs, key crypto.Key) error {
	message, err := SingleSegmentSignMessage(transaction, start, end, witnessArgs)
	if err != nil {
		return err
	}

	signed, err := key.Sign(message)
	if err != nil {
		return err
	}

	wa := &types.WitnessArgs{
		Lock:       signed,
		InputType:  witnessArgs.InputType,
		OutputType: witnessArgs.OutputType,
	}
	wab, err := wa.Serialize()
	if err != nil {
		return err
	}

	transaction.Witnesses[start] = wab

	return nil
}

func CalculateTransactionFee(tx *types.Transaction, feeRate uint64) (uint64, error) {
	bytes, err := tx.Serialize()
	if err != nil {
		return 0, err
	}

	// raw tx serialize
	txSize := uint64(len(bytes))
	// witness serialize
	txSize += uint64(len(types.SerializeDynVec(tx.Witnesses)))
	// raw tx + witness offset
	txSize += 12
	// tx offset
	txSize += 4
	fee := txSize * feeRate / 1000
	if fee*1000 < txSize*feeRate {
		fee += 1
	}
	return fee, nil
}
