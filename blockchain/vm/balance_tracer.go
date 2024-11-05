package vm

import (
	"encoding/json"
	"github.com/kaiachain/kaia/common"
	"github.com/kaiachain/kaia/common/hexutil"
	//"github.com/kaiachain/kaia/node/cn/tracers"
	"math/big"
	"sync/atomic"
)

var (
	TokenContractCaller  = common.HexToAddress("0x1992111111111111111111111111111111111110")
	TokenContractAddress = common.HexToAddress("0x1992111111111111111111111111111111111111")
)

type BalanceTracer struct {
	interrupt    atomic.Bool
	contracts    map[common.Address]map[string]struct{}
	topContracts map[common.Address]common.Address
	checkTop     bool
}

func NewBalanceTracer() Tracer {
	ct := &BalanceTracer{
		contracts:    make(map[common.Address]map[string]struct{}),
		topContracts: make(map[common.Address]common.Address),
		checkTop:     false,
	}
	return ct
}

// Transaction start
func (ct *BalanceTracer) CaptureTxStart(uint64) {}
func (ct *BalanceTracer) CaptureTxEnd(uint64)   {}
func (ct *BalanceTracer) CaptureStart(env *EVM, from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
}
func (ct *BalanceTracer) CaptureEnd(output []byte, gasUsed uint64, err error) {}
func (ct *BalanceTracer) CaptureEnter(typ OpCode, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
}
func (ct *BalanceTracer) CaptureExit(output []byte, gasUsed uint64, err error) {}
func (ct *BalanceTracer) CaptureState(env *EVM, pc uint64, op OpCode, gas, cost, ccLeft, ccOpcode uint64, scope *ScopeContext, depth int, err error) {
	if err != nil {
		return
	}
	stackData := scope.Stack.Data()
	stackLen := len(stackData)
	contractAddress := scope.Contract.Address()
	if _, ok := ct.contracts[contractAddress]; !ok {
		ct.contracts[contractAddress] = make(map[string]struct{})
	}

	if ct.checkTop {
		caller := scope.Contract.CallerAddress
		if caller != TokenContractCaller && caller != TokenContractAddress {
			if _, ok := ct.topContracts[contractAddress]; !ok {
				if topContract, hasCaller := ct.topContracts[caller]; !hasCaller {
					ct.topContracts[contractAddress] = caller
				} else {
					ct.topContracts[contractAddress] = topContract
				}
			}
		}
	}

	switch {
	case stackLen >= 2 && op == SHA3:
		offset := stackData[stackLen-1]
		size := stackData[stackLen-2]
		data, err := GetMemoryCopyPadded(scope.Memory.Data(), int64(offset.Uint64()), int64(size.Uint64()))
		if err != nil {
			return
		}
		if _, ok := ct.contracts[contractAddress]; !ok {
			ct.contracts[contractAddress] = make(map[string]struct{})
		}
		ct.contracts[contractAddress][hexutil.Encode(data)] = struct{}{}
	}
}

func (ct *BalanceTracer) CaptureFault(env *EVM, pc uint64, op OpCode, gas, cost, ccLeft, ccOpcode uint64, scope *ScopeContext, depth int, err error) {
}

type TokenBalanceResult struct {
	Contracts    map[common.Address][]string       `json:"contracts"`
	TopContracts map[common.Address]common.Address `json:"topContracts"`
}

func (ct *BalanceTracer) GetResult() (json.RawMessage, error) {
	// remove empty key
	for k, v := range ct.contracts {
		if len(v) == 0 {
			delete(ct.contracts, k)
		}
	}

	contracts := make(map[common.Address][]string)
	for k, vs := range ct.contracts {
		contracts[k] = make([]string, 0)
		for v := range vs {
			contracts[k] = append(contracts[k], v)
		}
	}

	tbr := TokenBalanceResult{
		Contracts:    contracts,
		TopContracts: ct.topContracts,
	}

	res, err := json.Marshal(tbr)
	if err != nil {
		return nil, err
	}
	// clear result
	ct.contracts = make(map[common.Address]map[string]struct{})
	ct.topContracts = make(map[common.Address]common.Address)

	return json.RawMessage(res), nil
}

func (t *BalanceTracer) Stop(err error) {
	t.interrupt.Store(true)
}
