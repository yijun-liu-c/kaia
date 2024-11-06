package vm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kaiachain/kaia/common"
	"github.com/kaiachain/kaia/common/hexutil"
	"github.com/kaiachain/kaia/crypto"
	"math/big"
	"sync/atomic"
)

type state = map[common.Address]*account

type account struct {
	Balance *big.Int                    `json:"balance,omitempty"`
	Code    []byte                      `json:"code,omitempty"`
	Nonce   uint64                      `json:"nonce,omitempty"`
	Storage map[common.Hash]common.Hash `json:"storage,omitempty"`
}

func (a *account) exists() bool {
	return a.Nonce > 0 || len(a.Code) > 0 || len(a.Storage) > 0 || (a.Balance != nil && a.Balance.Sign() != 0)
}

type accountMarshaling struct {
	Balance *hexutil.Big
	Code    hexutil.Bytes
}

type PrestateTracer struct {
	env       *EVM
	pre       state
	post      state
	create    bool
	to        common.Address
	gasLimit  uint64 // Amount of gas bought for the whole tx
	config    prestateTracerConfig
	interrupt atomic.Bool // Atomic flag to signal execution interruption
	reason    error       // Textual reason for the interruption
	created   map[common.Address]bool
	deleted   map[common.Address]bool
}

type prestateTracerConfig struct {
	DiffMode bool `json:"diffMode"` // If true, this tracer will return state modifications
}

func NewPrestateTracer(cfg json.RawMessage) (Tracer, error) {
	var config prestateTracerConfig
	if cfg != nil {
		if err := json.Unmarshal(cfg, &config); err != nil {
			return nil, err
		}
	}
	pt := &PrestateTracer{
		pre:     make(state),
		post:    make(state),
		config:  config,
		created: make(map[common.Address]bool),
		deleted: make(map[common.Address]bool),
	}
	return pt, nil
}

// Transaction start
func (pt *PrestateTracer) CaptureTxStart(gasLimit uint64) {
	pt.gasLimit = gasLimit
}

func (pt *PrestateTracer) CaptureTxEnd(restGas uint64) {
	if !pt.config.DiffMode {
		return
	}

	for addr, state := range pt.pre {
		// The deleted account's state is pruned from `post` but kept in `pre`
		if _, ok := pt.deleted[addr]; ok {
			continue
		}
		modified := false
		postAccount := &account{Storage: make(map[common.Hash]common.Hash)}
		newBalance := pt.env.StateDB.GetBalance(addr)
		newNonce := pt.env.StateDB.GetNonce(addr)
		newCode := pt.env.StateDB.GetCode(addr)

		if newBalance.Cmp(pt.pre[addr].Balance) != 0 {
			modified = true
			postAccount.Balance = newBalance
		}
		if newNonce != pt.pre[addr].Nonce {
			modified = true
			postAccount.Nonce = newNonce
		}
		if !bytes.Equal(newCode, pt.pre[addr].Code) {
			modified = true
			postAccount.Code = newCode
		}

		// https://github.com/bnb-chain/bsc/blob/master/eth/tracers/native/prestate.go
		if len(postAccount.Code) > 0 {
			postAccount.Code = []byte("0")
		}

		for key, val := range state.Storage {
			// don't include the empty slot
			if val == (common.Hash{}) {
				delete(pt.pre[addr].Storage, key)
			}

			newVal := pt.env.StateDB.GetState(addr, key)
			if val == newVal {
				// Omit unchanged slots
				delete(pt.pre[addr].Storage, key)
			} else {
				modified = true
				if newVal != (common.Hash{}) {
					postAccount.Storage[key] = newVal
				}
			}
		}

		if modified {
			pt.post[addr] = postAccount
		} else {
			// if state is not modified, then no need to include into the pre state
			delete(pt.pre, addr)
		}
	}
	// the new created contracts' prestate were empty, so delete them
	for a := range pt.created {
		// the created contract maybe exists in statedb before the creating tx
		if s := pt.pre[a]; s != nil && !s.exists() {
			delete(pt.pre, a)
		}
	}
}

func (pt *PrestateTracer) CaptureStart(env *EVM, from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
	pt.env = env
	pt.create = create
	pt.to = to

	pt.lookupAccount(from)
	pt.lookupAccount(to)
	pt.lookupAccount(env.Context.Coinbase)

	// The recipient balance includes the value transferred.
	toBal := new(big.Int).Sub(pt.pre[to].Balance, value)
	pt.pre[to].Balance = toBal

	// The sender balance is after reducing: value and gasLimit.
	// We need to re-add them to get the pre-tx balance.
	fromBal := new(big.Int).Set(pt.pre[from].Balance)
	gasPrice := env.TxContext.GasPrice
	consumedGas := new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(pt.gasLimit))
	fromBal.Add(fromBal, new(big.Int).Add(value, consumedGas))

	pt.pre[from].Balance = fromBal
	pt.pre[from].Nonce--

	if create && pt.config.DiffMode {
		pt.created[to] = true
	}
}

func (pt *PrestateTracer) CaptureEnd(output []byte, gasUsed uint64, err error) {
	if pt.config.DiffMode {
		return
	}

	if pt.create {
		// Keep existing account prior to contract creation at that address
		if s := pt.pre[pt.to]; s != nil && !s.exists() {
			// Exclude newly created contract.
			delete(pt.pre, pt.to)
		}
	}
}

func (pt *PrestateTracer) CaptureEnter(typ OpCode, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
}
func (pt *PrestateTracer) CaptureExit(output []byte, gasUsed uint64, err error) {}

// CaptureState implements the EVMLogger interface to trace a single step of VM execution.
func (pt *PrestateTracer) CaptureState(env *EVM, pc uint64, op OpCode, gas, cost, ccLeft, ccOpcode uint64, scope *ScopeContext, depth int, err error) {
	if err != nil {
		return
	}
	// Skip if tracing was interrupted
	if pt.interrupt.Load() {
		return
	}
	stack := scope.Stack
	stackData := stack.Data()
	stackLen := len(stackData)
	caller := scope.Contract.Address()
	switch {
	case stackLen >= 1 && (op == SLOAD || op == SSTORE):
		slot := common.Hash(stackData[stackLen-1].Bytes32())
		pt.lookupStorage(caller, slot)
	case stackLen >= 1 && (op == EXTCODECOPY || op == EXTCODEHASH || op == EXTCODESIZE || op == BALANCE || op == SELFDESTRUCT):
		addr := common.Address(stackData[stackLen-1].Bytes20())
		pt.lookupAccount(addr)
		if op == SELFDESTRUCT {
			pt.deleted[caller] = true
		}
	case stackLen >= 5 && (op == DELEGATECALL || op == CALL || op == STATICCALL || op == CALLCODE):
		addr := common.Address(stackData[stackLen-2].Bytes20())
		pt.lookupAccount(addr)
	case op == CREATE:
		nonce := pt.env.StateDB.GetNonce(caller)
		addr := crypto.CreateAddress(caller, nonce)
		pt.lookupAccount(addr)
		pt.created[addr] = true
	case stackLen >= 4 && op == CREATE2:
		offset := stackData[stackLen-2]
		size := stackData[stackLen-3]
		init, err := GetMemoryCopyPadded(scope.Memory.Data(), int64(offset.Uint64()), int64(size.Uint64()))
		if err != nil {
			fmt.Println("failed to copy CREATE2 input", "err", err, "tracer", "prestateTracer", "offset", offset, "size", size)
			return
		}
		inithash := crypto.Keccak256(init)
		salt := stackData[stackLen-4]
		addr := crypto.CreateAddress2(caller, salt.Bytes32(), inithash)
		pt.lookupAccount(addr)
		pt.created[addr] = true
	}
}

func (pt *PrestateTracer) CaptureFault(env *EVM, pc uint64, op OpCode, gas, cost, ccLeft, ccOpcode uint64, scope *ScopeContext, depth int, err error) {
}

func (pt *PrestateTracer) GetResult() (json.RawMessage, error) {
	var res []byte
	var err error
	if pt.config.DiffMode {
		res, err = json.Marshal(struct {
			Post state `json:"post"`
			Pre  state `json:"pre"`
		}{pt.post, pt.pre})
	} else {
		res, err = json.Marshal(pt.pre)
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(res), pt.reason
}

func (pt *PrestateTracer) Stop(err error) {
	pt.interrupt.Store(true)
}

// lookupAccount fetches details of an account and adds it to the prestate
// if it doesn't exist there.
func (pt *PrestateTracer) lookupAccount(addr common.Address) {
	if _, ok := pt.pre[addr]; ok {
		return
	}
	// https://github.com/bnb-chain/bsc/blob/master/eth/tracers/native/prestate.go
	code := pt.env.StateDB.GetCode(addr)
	if len(code) > 0 {
		code = []byte("0")
	}

	pt.pre[addr] = &account{
		Balance: pt.env.StateDB.GetBalance(addr),
		Nonce:   pt.env.StateDB.GetNonce(addr),
		Code:    code,
		Storage: make(map[common.Hash]common.Hash),
	}
}

// lookupStorage fetches the requested storage slot and adds
// it to the prestate of the given contract. It assumes `lookupAccount`
// has been performed on the contract before.
func (pt *PrestateTracer) lookupStorage(addr common.Address, key common.Hash) {
	if _, ok := pt.pre[addr].Storage[key]; ok {
		return
	}
	pt.pre[addr].Storage[key] = pt.env.StateDB.GetState(addr, key)
}
