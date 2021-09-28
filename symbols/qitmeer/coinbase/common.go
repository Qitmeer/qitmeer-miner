package coinbase

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain/opreturn"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
)

// standardCoinbaseOpReturn creates a standard OP_RETURN output to insert into
// coinbase to use as extranonces. The OP_RETURN pushes 32 bytes.
func standardCoinbaseOpReturn(enData []byte) ([]byte, error) {
	if len(enData) == 0 {
		return nil, nil
	}
	extraNonceScript, err := txscript.GenerateProvablyPruneableOut(enData)
	if err != nil {
		return nil, err
	}
	return extraNonceScript, nil
}

func standardCoinbaseScript(randStr string, nextBlockHeight uint64, extraNonce uint64) ([]byte, error) {
	return txscript.NewScriptBuilder().AddInt64(int64(nextBlockHeight)).
		AddInt64(int64(extraNonce)).AddData([]byte(randStr)).
		Script()
}

// CalcBlockTaxSubsidy calculates the subsidy for the organization address in the
// coinbase.
func CalcBlockTaxSubsidy(coinbaseVal uint64, params *params.Params) uint64 {
	_, _, tax := calcBlockProportion(coinbaseVal, params)
	return tax
}

func calcSubsidyByCoinBase(coinbaseVal uint64, params *params.Params) uint64 {
	workPro := float64(params.WorkRewardProportion)
	proportions := float64(params.TotalSubsidyProportions())
	subsidy := float64(coinbaseVal) * proportions / workPro
	return uint64(subsidy)
}

func calcBlockProportion(coinbaseVal uint64, params *params.Params) (uint64, uint64, uint64) {
	subsidy := calcSubsidyByCoinBase(coinbaseVal, params)
	workPro := float64(params.WorkRewardProportion)
	stakePro := float64(params.StakeRewardProportion)
	proportions := float64(params.TotalSubsidyProportions())
	work := uint64(workPro / proportions * float64(subsidy))
	stake := uint64(stakePro / proportions * float64(subsidy))
	tax := subsidy - work - stake
	return work, stake, tax
}

// createCoinbaseTx returns a coinbase transaction paying an appropriate subsidy
// based on the passed block height to the provided address.  When the address
// is nil, the coinbase transaction will instead be redeemable by anyone.
//
// See the comment for NewBlockTemplate for more information about why the nil
// address handling is useful.
func createCoinbaseTx(subsidy uint64, coinbaseScript []byte, opReturnPkScript []byte,
	addr types.Address, params *params.Params) (*types.Tx, *types.TxOutput, error) {
	tx := types.NewTransaction()
	tx.AddTxIn(&types.TxInput{
		// Coinbase085 transactions have no inputs, so previous outpoint is
		// zero hash and max index.
		PreviousOut: *types.NewOutPoint(&hash.Hash{},
			types.MaxPrevOutIndex),
		Sequence:   types.MaxTxInSequenceNum,
		SignScript: coinbaseScript,
	})

	hasTax := false
	if params.BlockTaxProportion > 0 &&
		len(params.OrganizationPkScript) > 0 {
		hasTax = true
	}
	// Create a coinbase with correct block subsidy and extranonce.
	tax := CalcBlockTaxSubsidy(subsidy, params)
	// output
	// Create the script to pay to the provided payment address if one was
	// specified.  Otherwise create a script that allows the coinbase to be
	// redeemable by anyone.
	var pksSubsidy []byte
	var err error
	if addr != nil {
		pksSubsidy, err = txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, nil, err
		}
	} else {
		scriptBuilder := txscript.NewScriptBuilder()
		pksSubsidy, err = scriptBuilder.AddOp(txscript.OP_TRUE).Script()
		if err != nil {
			return nil, nil, err
		}
	}
	if !hasTax {
		subsidy += uint64(tax)
		tax = 0
	}
	// Subsidy paid to miner.
	tx.AddTxOut(&types.TxOutput{
		Amount: types.Amount{
			Id:    types.MEERID,
			Value: int64(subsidy),
		},
		PkScript: pksSubsidy,
	})

	// Tax output.
	if hasTax {
		tx.AddTxOut(&types.TxOutput{
			Amount: types.Amount{
				Id:    types.MEERID,
				Value: int64(tax),
			},
			PkScript: params.OrganizationPkScript,
		})
	}
	// nulldata.
	if opReturnPkScript != nil {
		tx.AddTxOut(&types.TxOutput{
			Amount: types.Amount{
				Id:    types.MEERID,
				Value: int64(tax),
			},
			PkScript: opReturnPkScript,
		})
	}
	// opReturnPkScript
	var opReturnOutput *types.TxOutput
	if len(opReturnPkScript) > 0 {
		opReturnOutput = &types.TxOutput{
			PkScript: opReturnPkScript,
		}
	} else {
		opReturnOutput = opreturn.GetOPReturnTxOutput(opreturn.NewShowAmount(int64(subsidy)))
	}
	// AmountIn.
	//tx.TxIn[0].AmountIn = subsidy + uint64(tax)  //TODO, remove type conversion
	return types.NewTx(tx), opReturnOutput, nil
}
