package types

import (
	"encoding/json"
	"math/rand"
	"time"

	log "github.com/inconshreveable/log15"
	"gitlab.33.cn/chain33/chain33/common"
	"gitlab.33.cn/chain33/chain33/common/address"
	"gitlab.33.cn/chain33/chain33/types"
)

var nameX string

var (
	hlog = log.New("module", "exectype.hashlock")

	actionName = map[string]int32{
		"Hlock":   HashlockActionLock,
		"Hsend":   HashlockActionSend,
		"Hunlock": HashlockActionUnlock,
	}
)

func init() {
	nameX = types.ExecName("hashlock")
	// init executor type
	types.RegistorExecutor("hashlock", NewType())

	// init log
	//types.RegistorLog(types.TyLogDeposit, &CoinsDepositLog{})

	// init query rpc
	//types.RegisterRPCQueryHandle("q2", &CoinsGetTxsByAddr{})
}

type HashlockType struct {
	types.ExecTypeBase
}

func NewType() *HashlockType {
	c := &HashlockType{}
	c.SetChild(c)
	return c
}

func (hashlock *HashlockType) GetPayload() types.Message {
	return &HashlockAction{}
}

func (hashlock *HashlockType) ActionName(tx *types.Transaction) string {
	var action HashlockAction
	err := types.Decode(tx.Payload, &action)
	if err != nil {
		return "unknown-err"
	}
	if action.Ty == HashlockActionLock && action.GetHlock() != nil {
		return "lock"
	} else if action.Ty == HashlockActionUnlock && action.GetHunlock() != nil {
		return "unlock"
	} else if action.Ty == HashlockActionSend && action.GetHsend() != nil {
		return "send"
	}
	return "unknown"
}

// TODO 暂时不修改实现， 先完成结构的重构
func (hashlock *HashlockType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	hlog.Debug("hashlock.CreateTx", "action", action)
	var tx *types.Transaction
	if action == "HashlockLock" {
		var param HashlockLockTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			hlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInputPara
		}
		return CreateRawHashlockLockTx(&param)
	} else if action == "HashlockUnlock" {
		var param HashlockUnlockTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			hlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInputPara
		}
		return CreateRawHashlockUnlockTx(&param)
	} else if action == "HashlockSend" {
		var param HashlockSendTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			hlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInputPara
		}
		return CreateRawHashlockSendTx(&param)
	} else {
		return nil, types.ErrNotSupport
	}
	return tx, nil
}

func (hashlock *HashlockType) GetTypeMap() map[string]int32 {
	return actionName
}

type CoinsDepositLog struct {
}

func (l CoinsDepositLog) Name() string {
	return "LogDeposit"
}

func (l CoinsDepositLog) Decode(msg []byte) (interface{}, error) {
	var logTmp types.ReceiptAccountTransfer
	err := types.Decode(msg, &logTmp)
	if err != nil {
		return nil, err
	}
	return logTmp, err
}

type CoinsGetTxsByAddr struct {
}

func (t *CoinsGetTxsByAddr) JsonToProto(message json.RawMessage) ([]byte, error) {
	var req types.ReqAddr
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}
	return types.Encode(&req), nil
}

func (t *CoinsGetTxsByAddr) ProtoToJson(reply *types.Message) (interface{}, error) {
	return reply, nil
}

func CreateRawHashlockLockTx(parm *HashlockLockTx) (*types.Transaction, error) {
	if parm == nil {
		hlog.Error("CreateRawHashlockLockTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &HashlockLock{
		Amount:        parm.Amount,
		Time:          parm.Time,
		Hash:          common.Sha256([]byte(parm.Secret)),
		ToAddress:     parm.ToAddr,
		ReturnAddress: parm.ReturnAddr,
	}
	lock := &HashlockAction{
		Ty:    HashlockActionLock,
		Value: &HashlockAction_Hlock{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(nameX)),
		Payload: types.Encode(lock),
		Fee:     parm.Fee,
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		To:      address.ExecAddress(types.ExecName(nameX)),
	}

	err := tx.SetRealFee(types.MinFee)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func CreateRawHashlockUnlockTx(parm *HashlockUnlockTx) (*types.Transaction, error) {
	if parm == nil {
		hlog.Error("CreateRawHashlockUnlockTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &HashlockUnlock{
		Secret: []byte(parm.Secret),
	}
	unlock := &HashlockAction{
		Ty:    HashlockActionUnlock,
		Value: &HashlockAction_Hunlock{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(nameX),
		Payload: types.Encode(unlock),
		Fee:     parm.Fee,
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		To:      address.ExecAddress(nameX),
	}

	err := tx.SetRealFee(types.MinFee)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func CreateRawHashlockSendTx(parm *HashlockSendTx) (*types.Transaction, error) {
	if parm == nil {
		hlog.Error("CreateRawHashlockSendTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &HashlockSend{
		Secret: []byte(parm.Secret),
	}
	send := &HashlockAction{
		Ty:    HashlockActionSend,
		Value: &HashlockAction_Hsend{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(nameX),
		Payload: types.Encode(send),
		Fee:     parm.Fee,
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		To:      address.ExecAddress(nameX),
	}

	err := tx.SetRealFee(types.MinFee)
	if err != nil {
		return nil, err
	}

	return tx, nil
}